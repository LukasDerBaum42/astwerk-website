//go:build js && wasm

// Command tutorial-game is the tutorial's game shell: a reactive + wasmwrap
// application that hydrates every tutorial page — the HUD, the quest hub, the
// level map on the overview, the quest-complete modal, the step lab's scoring
// and the badge toasts.
//
// The pages teach; the shell scores (tutorial-game/01-shell.md). The teaching
// components stay vanilla and announce themselves: the quiz script dispatches
// astwerk:quiz, the lab script dispatches astwerk:lab, the objective
// checkboxes fire change events, the try-it details fire toggle. The shell
// converts them into game events against the pure-Go ledger in scripts/game —
// the same code `go test ./scripts/game` runs on the host.
//
// The shell hydrates, it never renders: every container below is templ markup
// the page already placed. With JavaScript off — or wasm blocked — the pages
// read completely and the vanilla progress script keeps working.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
	"github.com/LukasDerBaum42/astwerk-website/scripts/game"
	"github.com/a-h/templ"
)

// saveKey is where the shell keeps the save game, a plain JSON value.
const saveKey = "astwerk-tutorial-save"

func main() {
	ls := wasmwrap.LocalStorage()

	// sessionStart is when this page load began — what separates "completed in
	// the visit that started it" from "completed on a later visit".
	sessionStart := time.Now().Unix()

	save := loadSave(ls, sessionStart)
	if hud := wasmwrap.Query(".game-hud"); hud.Exists() {
		gameHUD(ls, save, sessionStart, hud)
	}
	if m := wasmwrap.Query(".game-map"); m.Exists() {
		levelMap(ls, save, m)
	}

	// Keep the module alive: main returning would tear down every handler.
	select {}
}

// loadSave reads the save game, migrating v1 saves and the vanilla progress
// script's key on first boot. The legacy key is left alone: the no-wasm
// fallback still writes it.
func loadSave(ls wasmwrap.Storage, now int64) *game.Save {
	if raw := ls.Get(saveKey); raw != "" {
		var s game.Save
		if err := json.Unmarshal([]byte(raw), &s); err == nil {
			return game.MigrateV1(&s)
		}
	}

	var legacy []string
	if raw := ls.Get(game.LegacyKey); raw != "" {
		json.Unmarshal([]byte(raw), &legacy)
	}
	if s, ok := game.MigrateLegacy(legacy, 11, now); ok {
		return s
	}
	return game.NewSave()
}

// saveGame writes the save back after every award.
func saveGame(ls wasmwrap.Storage, s *game.Save) {
	if b, err := json.Marshal(s); err == nil {
		ls.Set(saveKey, string(b))
	}
}

// totalsFrom reads the build-time totals JSON the pages emit — the per-line
// counts, the dependency graph and the quest names the badge rules, the map
// and the "now available" lists compare against.
func totalsFrom(el wasmwrap.Element) game.Totals {
	var t game.Totals
	raw := el.Dataset().Get("game-totals")
	if raw == "" {
		return t
	}
	json.Unmarshal([]byte(raw), &t)
	return t
}

// pathsFrom reads the quest-id → href map the HUD emits — what the modal's
// "now available" list and the pager's suggested-next link point at.
func pathsFrom(el wasmwrap.Element) map[string]string {
	paths := map[string]string{}
	raw := el.Dataset().Get("game-paths")
	if raw == "" {
		return paths
	}
	json.Unmarshal([]byte(raw), &paths)
	return paths
}

// gameHUD hydrates one step page: the quest bar, the quest hub, the
// quest-complete modal, the step lab's scoring, the restore pass and the
// toast stack, and listens for the page's teaching mechanics.
func gameHUD(ls wasmwrap.Storage, save *game.Save, sessionStart int64, hud wasmwrap.Element) {
	quest := hud.Dataset().Get("quest")
	line := hud.Dataset().Get("line")
	objectiveTotal, _ := strconv.Atoi(hud.Dataset().Get("objectives"))
	quizTotal, _ := strconv.Atoi(hud.Dataset().Get("quizzes"))
	// The step lab is part of a quest's completion: its frame's slot count is
	// the lab half of the totals (empty frame = 0).
	labTotal, _ := strconv.Atoi(wasmwrap.Query("[data-lab]").Dataset().Get("lab-slots"))
	totals := totalsFrom(hud)
	paths := pathsFrom(hud)

	// The visit award: +5 once, silently — the HUD just fills in.
	if _, first := save.Visit(quest, sessionStart); first {
		saveGame(ls, save)
	}

	// refresh re-reads the save: every award mutates the plain struct, then
	// bumps this signal so the bindings below re-run.
	refresh := reactive.Num(0)

	reactive.BindText(hud.Find("[data-game-rank]"), reactive.Computed(func() string {
		refresh.Get()
		return save.Rank()
	}))
	reactive.BindText(hud.Find("[data-game-xp]"), reactive.Computed(func() string {
		refresh.Get()
		return fmt.Sprintf("%d XP", save.XP)
	}))
	reactive.BindStyle(hud.Find("[data-game-xpbar]"), "width", reactive.Computed(func() string {
		refresh.Get()
		from, to := save.RankProgress()
		pct := 100.0
		if to > from {
			pct = float64(save.XP-from) / float64(to-from) * 100
		}
		return fmt.Sprintf("%.0f%%", pct)
	}))
	reactive.BindText(hud.Find("[data-game-badges]"), reactive.Computed(func() string {
		refresh.Get()
		return fmt.Sprintf("%d badges", len(save.Badges))
	}))
	reactive.BindText(hud.Find("[data-game-sound]"), reactive.Computed(func() string {
		refresh.Get()
		if save.Sound {
			return "🔔"
		}
		return "🔕"
	}))
	reactive.BindAttr(hud.Find("[data-game-sound]"), "aria-pressed", reactive.Computed(func() string {
		refresh.Get()
		if save.Sound {
			return "true"
		}
		return "false"
	}))

	reactive.On(hud.Find("[data-game-sound]"), "click", func(wasmwrap.Event) {
		save.Sound = !save.Sound
		saveGame(ls, save)
		refresh.Inc(1)
	})

	toasts := wasmwrap.Query("[data-game-toasts]")

	// Every award lands the same way: mutate, save, refresh.
	settle := func() {
		saveGame(ls, save)
		refresh.Inc(1)
	}

	// The quest-complete modal: appears once per completion (the Complete
	// award gates it — never on re-visits), dismisses on Esc or click-outside.
	// The breakdown, the badge list and the "now available" list are BindTempl
	// output, re-rendered on modalTick from the save and the just-earned
	// badges.
	modal := wasmwrap.Query("[data-game-modal]")
	modalOpen := reactive.Bool(false)
	modal.RemoveAttr("hidden")
	reactive.BindClass(modal, "open", modalOpen)

	lastEarned := []string{}
	lastNext := []nextLink{}
	modalTick := reactive.Num(0)
	reactive.BindTempl(modal.Find("[data-game-modal-breakdown]"), func() templ.Component {
		modalTick.Get()
		return breakdownList(breakdown(save.Quests[quest]))
	})
	reactive.BindTempl(modal.Find("[data-game-modal-badges]"), func() templ.Component {
		modalTick.Get()
		return badgeList(lastEarned)
	})
	reactive.BindTempl(modal.Find("[data-game-modal-next]"), func() templ.Component {
		modalTick.Get()
		return nextList(lastNext)
	})

	closeModal := func() {
		modalOpen.Set(false)
		hud.Find("[data-hub-open]").Focus()
	}
	for _, closer := range wasmwrap.QueryAll("[data-game-modal-close]") {
		reactive.On(closer, "click", func(wasmwrap.Event) { closeModal() })
	}
	wasmwrap.Wrap(js.Global()).On("keydown", func(e wasmwrap.Event) {
		if e.Value().Get("key").String() == "Escape" && modalOpen.Get() {
			closeModal()
		}
	})

	// Completion is derived: a quest is complete when every objective is
	// ticked, every quiz answered correctly and the whole lab solved.
	// observe() re-evaluates that state after every objective, quiz and lab
	// event — the +50 fires the first time the page's own mechanics add up,
	// and never again. The lab's reveal-on-check keeps it never blocking.
	observe := func() {
		if save.Observe(quest, objectiveTotal, quizTotal, labTotal, sessionStart, time.Now().Unix()) {
			earned := save.UnlockBadges(totals)
			settle()
			if save.Sound {
				fanfare()
			}
			lastEarned = earned
			lastNext = computeNext(save, totals, quest, paths)
			modalTick.Inc(1)
			modalOpen.Set(true)
			// The completion moment: the quest track pulses once before the
			// modal opens (08-reactive-pages.md).
			if track := wasmwrap.Query("[data-quest-track]"); track.Exists() {
				pulse(track)
			}
			// Focus moves into the modal: the first onward link, else Close.
			if first := modal.Find("[data-game-modal-next] a"); first.Exists() {
				first.Focus()
			} else {
				modal.Find(".game-modal-actions [data-game-modal-close]").Focus()
			}
			for _, b := range earned {
				toast(toasts, "Badge earned: "+game.BadgeNames[b])
				if save.Sound {
					ding()
				}
			}
			// Keep the no-wasm fallback's checklist in step with the save.
			markLegacyDone(ls, quest)
		}
	}
	observe()

	// The objective checkboxes: +10 per objective, once each, and a derived
	// completion re-check. Unchecking never deducts. The quest hub's mirrors
	// are wired below, matched by data-mirror-of — the step body's boxes are
	// the single write point.
	for _, box := range wasmwrap.QueryAll("[data-objective]") {
		reactive.On(box, "change", func(e wasmwrap.Event) {
			if !e.Value().Get("target").Get("checked").Bool() {
				return
			}
			if _, awarded := save.Objective(quest, objectiveTotal, time.Now().Unix()); awarded {
				settle()
			}
			observe()
		})
	}

	// The quizzes announce themselves: first try +20, retry +5, wrong nothing.
	wasmwrap.Wrap(js.Global()).On("astwerk:quiz", func(e wasmwrap.Event) {
		detail := e.Value().Get("detail")
		correct := detail.Get("correct").Bool()
		firstTry := detail.Get("firstTry").Bool()

		if _, awarded := save.Quiz(quest, quizTotal, correct, firstTry, time.Now().Unix()); !awarded {
			return
		}
		earned := save.UnlockBadges(totals)
		settle()
		observe()

		if correct && firstTry {
			toast(toasts, "+20 XP — first try!")
			// The quiz's "Correct." line gains its XP tag.
			if quiz := detail.Get("quiz"); quiz.Truthy() {
				if fb := wasmwrap.Wrap(quiz).Find(".quiz-feedback"); fb.Exists() {
					tag := wasmwrap.Create("span")
					tag.Class().Add("quiz-xp")
					tag.SetText("+20 XP")
					fb.Append(tag)
				}
			}
		}
		if save.Sound {
			ding()
		}
		for _, b := range earned {
			toast(toasts, "Badge earned: "+game.BadgeNames[b])
			if save.Sound {
				ding()
			}
		}
	})

	// The step lab announces itself: +5 per slot, once. Peeking pays the same
	// XP — only the toast's label differs. Wrong answers cost nothing; the lab
	// is part of the derived completion, so every award re-observes.
	if labTotal, _ := strconv.Atoi(wasmwrap.Query("[data-lab]").Dataset().Get("lab-slots")); labTotal > 0 {
		wasmwrap.Wrap(js.Global()).On("astwerk:lab", func(e wasmwrap.Event) {
			detail := e.Value().Get("detail")
			if detail.Get("quest").String() != quest {
				return
			}
			if !detail.Get("correct").Bool() {
				return
			}
			if _, awarded := save.Lab(quest, labTotal, time.Now().Unix()); !awarded {
				return
			}
			settle()
			switch {
			case detail.Get("peek").Bool():
				toast(toasts, "+5 XP — good eye!")
			case detail.Get("reveal").Bool():
				toast(toasts, "+5 XP — answer shown")
			default:
				toast(toasts, "+5 XP")
			}
			if save.Sound {
				ding()
			}
			observe()
		})
	}

	// The try-it reveals: +15 once per quest, on open.
	for _, details := range wasmwrap.QueryAll(".tryit details") {
		reactive.On(details, "toggle", func(e wasmwrap.Event) {
			if !e.Value().Get("target").Get("open").Bool() {
				return
			}
			if _, awarded := save.Bonus(quest, time.Now().Unix()); awarded {
				settle()
				toast(toasts, "+15 XP — bonus!")
				if save.Sound {
					ding()
				}
			}
		})
	}

	// The quest hub (05-quest-hub.md): the mobile sheet under the HUD's
	// Objectives button, the "you" summary, the chain's four states and the
	// objective mirrors.
	hub := wasmwrap.Query("[data-quest-hub]")
	if hub.Exists() {
		hubOpen := reactive.Bool(false)
		reactive.BindClass(hub, "open", hubOpen)
		if btn := wasmwrap.Query("[data-hub-open]"); btn.Exists() {
			reactive.BindAttr(btn, "aria-expanded", reactive.Computed(func() string {
				if hubOpen.Get() {
					return "true"
				}
				return "false"
			}))
			reactive.On(btn, "click", func(wasmwrap.Event) { hubOpen.Set(!hubOpen.Get()) })
		}
		for _, closer := range wasmwrap.QueryAll("[data-hub-close]") {
			reactive.On(closer, "click", func(wasmwrap.Event) { hubOpen.Set(false) })
		}
		// Click-outside closes the sheet; clicks inside it (or on the toggle
		// button) pass through.
		wasmwrap.Wrap(js.Global()).On("click", func(e wasmwrap.Event) {
			if !hubOpen.Get() {
				return
			}
			target := e.Value().Get("target")
			if target.Call("closest", "[data-quest-hub], [data-hub-open]").Truthy() {
				return
			}
			hubOpen.Set(false)
		})
		wasmwrap.Wrap(js.Global()).On("keydown", func(e wasmwrap.Event) {
			if e.Value().Get("key").String() == "Escape" && hubOpen.Get() {
				hubOpen.Set(false)
			}
		})

		// The "you" summary — the HUD's bindings, compact.
		reactive.BindText(hub.Find("[data-hub-rank]"), reactive.Computed(func() string {
			refresh.Get()
			return save.Rank()
		}))
		reactive.BindStyle(hub.Find("[data-hub-xpbar]"), "width", reactive.Computed(func() string {
			refresh.Get()
			from, to := save.RankProgress()
			pct := 100.0
			if to > from {
				pct = float64(save.XP-from) / float64(to-from) * 100
			}
			return fmt.Sprintf("%.0f%%", pct)
		}))
		reactive.BindText(hub.Find("[data-hub-badges]"), reactive.Computed(func() string {
			refresh.Get()
			return fmt.Sprintf("%d badges", len(save.Badges))
		}))

		// The now-playing card's unlocks line: the quest that becomes
		// current once this one completes — "the next quest" — not its
		// requirements (07-dependencies.md).
		if unlocks := hub.Find("[data-hub-unlocks]"); unlocks.Exists() {
			reactive.BindText(unlocks, reactive.Computed(func() string {
				refresh.Get()
				if next := unlockNext(save, totals, quest); next != "" {
					return "Unlocks: " + questName(totals, next)
				}
				if line, _, ok := parseQuestID(quest); ok && lastQuestID(totals, line) == quest {
					return "Next: the overview"
				}
				return ""
			}))
		}

		// The now-playing card's status: the progress bar, the done count
		// and the quiz/lab meta — the quest's state, live in the card. The
		// lab counts toward completion, so the bar and the count include it.
		if playing := hub.Find("[data-hub-playing]"); playing.Exists() {
			reactive.BindStyle(playing.Find("[data-hub-playing-bar]"), "width", reactive.Computed(func() string {
				refresh.Get()
				q := save.Quests[quest]
				qt := totals.Quests[quest]
				pct := 100.0
				if total := qt.Objectives + qt.Quizzes + qt.Lab; total > 0 {
					pct = float64(q.Objectives+q.QuizFirst+q.QuizRetry+q.Lab) / float64(total) * 100
				}
				return fmt.Sprintf("%.0f%%", pct)
			}))
			reactive.BindText(playing.Find("[data-hub-playing-count]"), reactive.Computed(func() string {
				refresh.Get()
				q := save.Quests[quest]
				qt := totals.Quests[quest]
				if q.Complete {
					return "Quest complete!"
				}
				return fmt.Sprintf("%d of %d done", q.Objectives+q.QuizFirst+q.QuizRetry+q.Lab, qt.Objectives+qt.Quizzes+qt.Lab)
			}))
			reactive.BindClass(playing, "done", reactive.Computed(func() bool {
				refresh.Get()
				return save.Quests[quest].Complete
			}))
			reactive.BindText(playing.Find("[data-hub-playing-meta]"), reactive.Computed(func() string {
				refresh.Get()
				q := save.Quests[quest]
				qt := totals.Quests[quest]
				parts := []string{fmt.Sprintf("%d/%d quizzes", q.QuizFirst+q.QuizRetry, qt.Quizzes)}
				if qt.Lab > 0 {
					parts = append(parts, fmt.Sprintf("Lab %d/%d", q.Lab, qt.Lab))
				}
				return strings.Join(parts, " · ")
			}))
		}

		// The objective mirrors: matched by data-mirror-of (the objective's
		// index), so one loop serves every mirror set and the step body's
		// boxes stay the single write point (05-quest-hub.md).
		boxes := wasmwrap.QueryAll("[data-objective]")
		for _, mirror := range wasmwrap.QueryAll("[data-objective-mirror]") {
			n, _ := strconv.Atoi(mirror.Dataset().Get("mirror-of"))
			if n < 1 || n > len(boxes) {
				continue
			}
			box := boxes[n-1]
			// Mirror → box: check the real box. Programmatic sets don't fire
			// change, so the award and the derived re-check happen here.
			reactive.On(mirror, "change", func(e wasmwrap.Event) {
				checked := e.Value().Get("target").Get("checked").Bool()
				if box.Value().Get("checked").Bool() == checked {
					return
				}
				box.Value().Set("checked", checked)
				if !checked {
					return
				}
				if _, awarded := save.Objective(quest, objectiveTotal, time.Now().Unix()); awarded {
					settle()
				}
				observe()
			})
			// Box → mirror: the hub follows the page.
			reactive.On(box, "change", func(e wasmwrap.Event) {
				mirror.Value().Set("checked", e.Value().Get("target").Get("checked").Bool())
			})
		}

		// The chain: done / current (NextAvailable) / waiting (requirements
		// unmet) — the tutorial's four dependency states (07-dependencies.md).
		// Waiting rows are hidden unless the show-all toggle is on.
		wireShowAll(ls, save, hub, hub.Find("[data-hub-toggle]"))
		bindHubChain(save, totals, refresh)

		// Each objective row is also a navigation: clicking scrolls to the
		// objective's section and pulses it.
		for _, row := range wasmwrap.QueryAll("[data-hub-obj]") {
			reactive.On(row, "click", func(e wasmwrap.Event) {
				if e.Target().Dataset().Get("objective-mirror") != "" {
					return
				}
				target := row.Find("input").Dataset().Get("scroll-to")
				if el := wasmwrap.Query(target); el.Exists() {
					el.Value().Call("scrollIntoView", map[string]any{"behavior": "smooth", "block": "center"})
					pulse(el)
				}
			})
		}
	}

	// The header's state chip re-labels from the save (08-reactive-pages.md):
	// never visited keeps the plain callout, visited shows "In progress", a
	// complete quest shows the ribbon.
	if chip := wasmwrap.Query("[data-header-state]"); chip.Exists() {
		reactive.BindText(chip, reactive.Computed(func() string {
			refresh.Get()
			q := save.Quests[quest]
			if q.Complete {
				return fmt.Sprintf("Quest complete — +%s XP earned", chip.Dataset().Get("reward"))
			}
			if q.Visited {
				return fmt.Sprintf("In progress · %d of %d", q.Objectives+q.QuizFirst+q.QuizRetry+q.Lab, objectiveTotal+quizTotal+labTotal)
			}
			return "Up to +" + chip.Dataset().Get("reward") + " XP"
		}))
		reactive.BindClass(chip, "done", reactive.Computed(func() bool {
			refresh.Get()
			return save.Quests[quest].Complete
		}))
		reactive.BindClass(chip, "in-progress", reactive.Computed(func() bool {
			refresh.Get()
			q := save.Quests[quest]
			return q.Visited && !q.Complete
		}))
		if header := wasmwrap.Query("[data-quest-header]"); header.Exists() {
			reactive.BindClass(header, "done", reactive.Computed(func() bool {
				refresh.Get()
				return save.Quests[quest].Complete
			}))
		}
	}

	// The lab chip next to the reward: the quest's slot progress.
	if labChip := wasmwrap.Query("[data-lab-chip]"); labChip.Exists() {
		reactive.BindText(labChip, reactive.Computed(func() string {
			refresh.Get()
			return fmt.Sprintf("Lab %d/%s", save.Quests[quest].Lab, labChip.Dataset().Get("lab-total"))
		}))
	}

	// The completed-quest footer panel: the modal's quiet cousin for revisits.
	if panel := wasmwrap.Query("[data-complete-panel]"); panel.Exists() {
		panel.RemoveAttr("hidden")
		reactive.BindShow(panel, reactive.Computed(func() bool {
			refresh.Get()
			return save.Quests[quest].Complete
		}))
		reactive.BindTempl(panel.Find("[data-complete-breakdown]"), func() templ.Component {
			refresh.Get()
			return breakdownList(breakdown(save.Quests[quest]))
		})
		reactive.BindTempl(panel.Find("[data-complete-next]"), func() templ.Component {
			refresh.Get()
			return panelNext(computeNext(save, totals, quest, paths))
		})
	}

	// The pager's "Next" follows the suggested quest: when the player's save
	// makes a different quest current than the literal next step, the link
	// relabels and repoints (07-dependencies.md).
	if nextLink := wasmwrap.Query(".pager .next"); nextLink.Exists() {
		want := save.NextAvailable(totals, line)
		literal := nextLink.Dataset().Get("next-quest")
		if want != "" && want != literal {
			if href, ok := paths[want]; ok {
				nextLink.SetAttr("href", href)
				if name, ok := totals.Names[want]; ok {
					if label := nextLink.Find(".pager-name"); label.Exists() {
						label.SetText(name)
					}
				}
				if span := nextLink.Find("span"); span.Exists() {
					span.SetText("Suggested next")
				}
			}
		}
	}

	// The restore pass: the page's teaching state comes from the save, so a
	// revisit reads exactly as the first play left it (08-reactive-pages.md).
	restorePage(save, quest)
	// The vanilla progress scripts read the DOM; tell them the state is in.
	js.Global().Get("window").Call("dispatchEvent", js.Global().Get("CustomEvent").New("astwerk:restore"))

	// The shell is alive: the vanilla fallback's done-marking stays out of the
	// way — the save is the single source of truth now.
	wasmwrap.Body().Class().Add("shell-alive")
}

// restorePage rewrites the page's teaching state from the save on boot:
// objectives checked, solved quizzes marked, opened try-its. It sets the exact
// DOM flags the vanilla scripts read, so the page and the game never disagree.
func restorePage(save *game.Save, quest string) {
	q := save.Quests[quest]

	// Objectives: the count restores losslessly — the first N boxes.
	boxes := wasmwrap.QueryAll("[data-objective]")
	for i, box := range boxes {
		box.Value().Set("checked", i < q.Objectives)
	}
	for _, mirror := range wasmwrap.QueryAll("[data-objective-mirror]") {
		n, _ := strconv.Atoi(mirror.Dataset().Get("mirror-of"))
		if n < 1 || n > len(boxes) {
			continue
		}
		mirror.Value().Set("checked", boxes[n-1].Value().Get("checked").Bool())
	}

	// Quizzes restore their solved state: the first QuizFirst+QuizRetry in
	// page order, the first-try/retry split shown in the feedback line. The
	// quiz script's data-done guard keeps them from being re-answered.
	solved := q.QuizFirst + q.QuizRetry
	for i, quiz := range wasmwrap.QueryAll(".quiz") {
		if i >= solved {
			break
		}
		correct, _ := strconv.Atoi(quiz.Dataset().Get("correct"))
		quiz.Dataset().Set("solved", "1")
		quiz.Dataset().Set("done", "1")
		quiz.Dataset().Set("tried", "1")
		for j, opt := range findAll(quiz, ".quiz-option") {
			opt.Value().Set("disabled", true)
			if j == correct {
				opt.Class().Add("right")
			}
		}
		if fb := quiz.Find(".quiz-feedback"); fb.Exists() {
			if i < q.QuizFirst {
				fb.SetText("Correct. (+20 XP earned)")
			} else {
				fb.SetText("Correct. (+5 XP earned)")
			}
			fb.Class().Add("right")
		}
		if why := quiz.Find(".quiz-why"); why.Exists() {
			why.Value().Set("hidden", false)
		}
		if reset := quiz.Find(".quiz-reset"); reset.Exists() {
			reset.Value().Set("hidden", false)
		}
	}

	// Try-its the player opened stay open, marked done.
	if q.Bonus {
		for _, details := range wasmwrap.QueryAll(".tryit details") {
			details.Value().Set("open", true)
			details.Class().Add("done")
			if summary := details.Find("summary"); summary.Exists() {
				summary.SetText("Opened ✓ " + summary.Text())
			}
		}
	}

	// The lab's progress restores from the count; the inputs stay neutral —
	// re-checking them pays nothing, because the ledger caps at the slot count.
	if frame := wasmwrap.Query("[data-lab]"); frame.Exists() {
		if p := frame.Find("[data-lab-progress]"); p.Exists() {
			p.SetText(fmt.Sprintf("%d/%s slots", q.Lab, frame.Dataset().Get("lab-slots")))
		}
	}
}

// wireShowAll binds one "show all quests" toggle: with the container's
// show-all class off, CSS hides the waiting quests (the ones whose
// requirements aren't met yet); the toggle reveals them — dimmed, still
// clickable, nothing is gated. The choice is player state, persisted in the
// save like the sound toggle. The returned signal lets callers (the map's
// dependency lines) read the toggle.
func wireShowAll(ls wasmwrap.Storage, save *game.Save, container, btn wasmwrap.Element) *reactive.BoolSignal {
	showAll := reactive.Bool(save.ShowAll)
	if !btn.Exists() {
		return showAll
	}
	reactive.BindClass(container, "show-all", showAll)
	reactive.BindText(btn, reactive.Computed(func() string {
		if showAll.Get() {
			return "Show unlocked only"
		}
		return "Show all quests"
	}))
	reactive.BindAttr(btn, "aria-pressed", reactive.Computed(func() string {
		if showAll.Get() {
			return "true"
		}
		return "false"
	}))
	reactive.On(btn, "click", func(wasmwrap.Event) {
		save.ShowAll = !save.ShowAll
		showAll.Set(save.ShowAll)
		saveGame(ls, save)
	})
	return showAll
}

// markLegacyDone keeps the no-wasm fallback's checklist in step with the
// save: the vanilla scripts read the legacy key, so a completion earned here
// shows up there too.
func markLegacyDone(ls wasmwrap.Storage, quest string) {
	var done []string
	if raw := ls.Get(game.LegacyKey); raw != "" {
		json.Unmarshal([]byte(raw), &done)
	}
	for _, d := range done {
		if d == quest {
			return
		}
	}
	done = append(done, quest)
	if b, err := json.Marshal(done); err == nil {
		ls.Set(game.LegacyKey, string(b))
	}
}

// pulse adds the one-shot accent ring used by the completion moment, the
// rail's and the hub's scroll targets (1.2 s, collapsed by reduced motion).
func pulse(el wasmwrap.Element) {
	el.Class().Add("pulse")
	wasmwrap.SetTimeout(func() { el.Class().Remove("pulse") }, 1200)
}

// findAll is a scoped querySelectorAll — wasmwrap only offers document-level
// QueryAll.
func findAll(el wasmwrap.Element, selector string) []wasmwrap.Element {
	list := el.Value().Call("querySelectorAll", selector)
	n := list.Get("length").Int()
	out := make([]wasmwrap.Element, 0, n)
	for i := range n {
		out = append(out, wasmwrap.Wrap(list.Index(i)))
	}
	return out
}

// breakdown lists the XP lines a quest has actually earned, in award order.
func breakdown(q game.Quest) []string {
	var lines []string
	if q.Objectives > 0 {
		lines = append(lines, fmt.Sprintf("Objectives +%d", q.Objectives*10))
	}
	if q.QuizFirst > 0 {
		lines = append(lines, fmt.Sprintf("Quizzes, first try +%d", q.QuizFirst*20))
	}
	if q.QuizRetry > 0 {
		lines = append(lines, fmt.Sprintf("Quizzes, after retry +%d", q.QuizRetry*5))
	}
	if q.Lab > 0 {
		lines = append(lines, fmt.Sprintf("Lab +%d", q.Lab*5))
	}
	if q.Bonus {
		lines = append(lines, "Try-it bonus +15")
	}
	lines = append(lines, "Quest complete +50")
	return lines
}

// breakdownList renders the modal's XP breakdown as a list.
func breakdownList(lines []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, line := range lines {
			if _, err := fmt.Fprintf(w, "<li>%s</li>", templ.EscapeString(line)); err != nil {
				return err
			}
		}
		return nil
	})
}

// badgeList renders the badges a completion just unlocked; empty when none.
func badgeList(ids []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if len(ids) == 0 {
			return nil
		}
		names := make([]string, len(ids))
		for i, id := range ids {
			names[i] = game.BadgeNames[id]
		}
		_, err := io.WriteString(w, templ.EscapeString("Badges unlocked: "+strings.Join(names, ", ")))
		return err
	})
}

// nextLink is one entry of the modal's "now available" list.
type nextLink struct {
	Label     string
	Href      string
	Suggested bool
}

// computeNext is the quest-complete moment's "what now": the newly available
// quests (whose requirements this completion just satisfied) as a list with
// the suggested one first, or the overview push after a line's boss quest
// (07-dependencies.md). The boss's list adds "All quests complete" once the
// line is done.
func computeNext(save *game.Save, t game.Totals, completed string, paths map[string]string) []nextLink {
	line, _, ok := parseQuestID(completed)
	if !ok {
		return nil
	}
	overview := paths["tutorial"]

	if last := lastQuestID(t, line); last == completed {
		links := []nextLink{{Label: "The site is live — Overview", Href: overview}}
		if save.LineDone(t, line) {
			links = append(links, nextLink{Label: "All quests complete — the line is done", Href: overview})
		}
		return links
	}

	// Newly unlocked: quests that listed `completed` among their requirements
	// and now have all of them met.
	unlocked := map[string]bool{}
	for id, reqs := range t.Deps {
		if save.Quests[id].Complete {
			continue
		}
		has := false
		for _, r := range reqs {
			if r == completed {
				has = true
			}
		}
		if has && save.RequiresMet(t, id) {
			unlocked[id] = true
		}
	}

	var out []nextLink
	if suggested := save.NextAvailable(t, line); suggested != "" && unlocked[suggested] {
		out = append(out, nextLink{Label: questName(t, suggested), Href: paths[suggested], Suggested: true})
		delete(unlocked, suggested)
	}
	for _, l := range sortedLines(t) {
		for _, id := range lineIDs(t.Lines[l], l) {
			if unlocked[id] {
				out = append(out, nextLink{Label: questName(t, id), Href: paths[id]})
			}
		}
	}
	return out
}

// questName is a quest's display name from the emitted totals, with the id as
// fallback.
func questName(t game.Totals, id string) string {
	if n, ok := t.Names[id]; ok && n != "" {
		return n
	}
	return id
}

// needsNames renders a quest's requirements as display names — the "Needs:"
// tooltips on the map's waiting nodes.
func needsNames(t game.Totals, id string) string {
	var names []string
	for _, r := range t.Deps[id] {
		names = append(names, questName(t, r))
	}
	return strings.Join(names, ", ")
}

// unlockNext is the quest that becomes current once quest completes — the
// now-playing card's "Unlocks:" line, computed as if the quest were already
// done (its dependents' requirements treat it as met). "" when nothing in the
// line becomes available; the card then points at the overview for a line's
// last quest.
func unlockNext(save *game.Save, t game.Totals, quest string) string {
	line, _, ok := parseQuestID(quest)
	if !ok {
		return ""
	}
	for _, id := range lineIDs(t.Lines[line], line) {
		if save.Quests[id].Complete {
			continue
		}
		met := true
		for _, r := range t.Deps[id] {
			if r == quest {
				continue
			}
			if !save.Quests[r].Complete {
				met = false
				break
			}
		}
		if met {
			return id
		}
	}
	return ""
}

// nextList renders the modal's "now available" list.
func nextList(links []nextLink) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if len(links) == 0 {
			return nil
		}
		if _, err := io.WriteString(w, `<p class="game-modal-next-title">Now available</p><ul class="game-modal-next-list">`); err != nil {
			return err
		}
		for _, l := range links {
			href, label := templ.EscapeString(l.Href), templ.EscapeString(l.Label)
			if l.Suggested {
				if _, err := fmt.Fprintf(w, `<li class="suggested"><a href="%s">%s</a> <span>Suggested next</span></li>`, href, label); err != nil {
					return err
				}
			} else if _, err := fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, href, label); err != nil {
				return err
			}
		}
		_, err := io.WriteString(w, `</ul>`)
		return err
	})
}

// panelNext is the completed-quest footer panel's onward link: the suggested
// next quest, or the overview after the boss.
func panelNext(links []nextLink) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if len(links) == 0 {
			return nil
		}
		l := links[0]
		_, err := fmt.Fprintf(w, `<a class="button" href="%s">Next: %s &rarr;</a>`,
			templ.EscapeString(l.Href), templ.EscapeString(l.Label))
		return err
	})
}

// lineIDs returns a line's quest ids in order — the emitted Ids, or the
// numeric range for totals without them.
func lineIDs(lt game.LineTotals, line string) []string {
	if len(lt.Ids) > 0 {
		return lt.Ids
	}
	ids := make([]string, 0, lt.Quests)
	for n := 1; n <= lt.Quests; n++ {
		ids = append(ids, game.QuestID(line, n))
	}
	return ids
}

// lastQuestID returns a line's final quest id, or "" for an empty line.
func lastQuestID(t game.Totals, line string) string {
	lt, ok := t.Lines[line]
	if !ok {
		return ""
	}
	ids := lineIDs(lt, line)
	if len(ids) == 0 {
		return ""
	}
	return ids[len(ids)-1]
}

// sortedLines returns the totals' line keys in a stable order.
func sortedLines(t game.Totals) []string {
	lines := make([]string, 0, len(t.Lines))
	for l := range t.Lines {
		lines = append(lines, l)
	}
	sort.Strings(lines)
	return lines
}

// levelMap hydrates the overview page: the line-select tabs, each line's
// node states and path fill, the summary line, the badge collection, the
// pipeline diagram and the boss callout.
func levelMap(ls wasmwrap.Storage, save *game.Save, m wasmwrap.Element) {
	totals := totalsFrom(m)
	refresh := reactive.Num(0)

	// The map is the player's state, so the plain step lists below are the
	// no-JS fallback: hide them once the shell is alive.
	wasmwrap.Body().Class().Add("shell-alive")

	reactive.BindText(m.Find("[data-game-summary]"), reactive.Computed(func() string {
		refresh.Get()
		done, total := 0, 0
		for line, lt := range totals.Lines {
			for _, id := range lineIDs(lt, line) {
				total++
				if save.Quests[id].Complete {
					done++
				}
			}
		}
		return fmt.Sprintf("%d of %d quests complete · %d XP · %d badges",
			done, total, save.XP, len(save.Badges))
	}))

	reactive.BindText(m.Find("[data-game-rank]"), reactive.Computed(func() string {
		refresh.Get()
		return save.Rank()
	}))
	reactive.BindStyle(m.Find("[data-game-rankbar]"), "width", reactive.Computed(func() string {
		refresh.Get()
		from, to := save.RankProgress()
		pct := 100.0
		if to > from {
			pct = float64(save.XP-from) / float64(to-from) * 100
		}
		return fmt.Sprintf("%.0f%%", pct)
	}))

	// Quests whose requirements aren't met are hidden by default; the toggle
	// reveals them — dimmed, still clickable (nothing is gated). The choice
	// persists in the save.
	mapShowAll := wireShowAll(ls, save, m, wasmwrap.Query("[data-map-toggle]"))

	// The nodes: done (accent glow + check), current (the line's first
	// available incomplete — NextAvailable), waiting (requirements unmet —
	// dimmed, locked, with the "Needs:" tooltip; still clickable, nothing is
	// gated), and plain.
	for _, node := range wasmwrap.QueryAll("[data-map-node]") {
		id := node.Dataset().Get("map-node")
		nodeLine, _, ok := parseQuestID(id)
		if !ok {
			continue
		}
		reactive.BindClass(node, "done", reactive.Computed(func() bool {
			refresh.Get()
			return save.Quests[id].Complete
		}))
		reactive.BindClass(node, "current", reactive.Computed(func() bool {
			refresh.Get()
			return id == save.NextAvailable(totals, nodeLine)
		}))
		reactive.BindClass(node, "waiting", reactive.Computed(func() bool {
			refresh.Get()
			q := save.Quests[id]
			return !q.Complete && !save.RequiresMet(totals, id)
		}))
		reactive.BindAttr(node, "title", reactive.Computed(func() string {
			refresh.Get()
			q := save.Quests[id]
			if !q.Complete && !save.RequiresMet(totals, id) {
				return "Needs: " + needsNames(totals, id)
			}
			return node.Dataset().Get("flavor")
		}))
	}

	// The dependency lines under waiting nodes — shown only when the waiting
	// quest itself is visible (the toggle above decides).
	for _, dep := range wasmwrap.QueryAll("[data-map-dep]") {
		id := dep.Dataset().Get("map-dep")
		reactive.BindShow(dep, reactive.Computed(func() bool {
			refresh.Get()
			q := save.Quests[id]
			return mapShowAll.Get() && !q.Complete && !save.RequiresMet(totals, id)
		}))
	}

	// The connecting path fills as its quest completes.
	for _, seg := range wasmwrap.QueryAll("[data-map-path]") {
		line, n, ok := parseQuestID(seg.Dataset().Get("map-path"))
		if !ok {
			continue
		}
		reactive.BindClass(seg, "lit", reactive.Computed(func() bool {
			refresh.Get()
			return save.Quests[game.QuestID(line, n)].Complete
		}))
	}

	for _, b := range wasmwrap.QueryAll("[data-badge]") {
		id := b.Dataset().Get("badge")
		reactive.BindClass(b, "earned", reactive.Computed(func() bool {
			refresh.Get()
			for _, have := range save.Badges {
				if have == id {
					return true
				}
			}
			return false
		}))
	}

	// The pipeline diagram lights up as the quests behind each stage complete
	// (08-reactive-pages.md). "static" needs any deploy-adjacent quest.
	stageQuests := []struct {
		name string
		ids  []string
		any  bool
	}{
		{"go", []string{"blog-1", "blog-2", "blog-3"}, false},
		{"templ", []string{"blog-9"}, false},
		{"build", []string{"blog-4", "blog-5"}, false},
		{"static", []string{"blog-7", "blog-10"}, true},
	}
	for _, stage := range wasmwrap.QueryAll("[data-stage]") {
		name := stage.Dataset().Get("stage")
		reactive.BindClass(stage, "lit", reactive.Computed(func() bool {
			refresh.Get()
			for _, s := range stageQuests {
				if s.name != name {
					continue
				}
				if s.any {
					for _, id := range s.ids {
						if save.Quests[id].Complete {
							return true
						}
					}
					return false
				}
				for _, id := range s.ids {
					if !save.Quests[id].Complete {
						return false
					}
				}
				return true
			}
			return false
		}))
	}

	// The boss callout: the boss is "open" the moment its requirements are
	// met — nothing was blocked, the acknowledgement is the point.
	if callout := wasmwrap.Query("[data-boss-callout]"); callout.Exists() {
		reactive.BindShow(callout, reactive.Computed(func() bool {
			refresh.Get()
			return save.RequiresMet(totals, "blog-11")
		}))
	}

	// The quest hub on the overview: the chain rows carry the same four
	// states, and the summary + now-playing card are dropped there.
	bindHubChain(save, totals, refresh)
}

func bindHubChain(save *game.Save, totals game.Totals, refresh *reactive.NumSignal[int]) {
	for _, row := range wasmwrap.QueryAll("[data-hub-row]") {
		id := row.Dataset().Get("hub-row")
		rowLine, _, ok := parseQuestID(id)
		if !ok {
			continue
		}
		reactive.BindClass(row, "done", reactive.Computed(func() bool {
			refresh.Get()
			return save.Quests[id].Complete
		}))
		reactive.BindClass(row, "current", reactive.Computed(func() bool {
			refresh.Get()
			return id == save.NextAvailable(totals, rowLine)
		}))
		reactive.BindClass(row, "waiting", reactive.Computed(func() bool {
			refresh.Get()
			q := save.Quests[id]
			return !q.Complete && !save.RequiresMet(totals, id)
		}))
		reactive.BindAttr(row, "title", reactive.Computed(func() string {
			refresh.Get()
			q := save.Quests[id]
			if !q.Complete && !save.RequiresMet(totals, id) {
				return "Needs: " + needsNames(totals, id)
			}
			return ""
		}))
		// The row's meta: the in-progress fraction (objectives + quizzes + lab
		// done of the quest's total — everything completion counts) and the lab
		// receipt whenever the quest has a lab — the sidebar reflects the
		// player's per-quest progress, not just the done state.
		reactive.BindText(row.Find("[data-hub-meta]"), reactive.Computed(func() string {
			refresh.Get()
			q := save.Quests[id]
			qt := totals.Quests[id]
			var parts []string
			if q.Visited && !q.Complete {
				parts = append(parts, fmt.Sprintf("%d/%d", q.Objectives+q.QuizFirst+q.QuizRetry+q.Lab, qt.Objectives+qt.Quizzes+qt.Lab))
			}
			if qt.Lab > 0 {
				parts = append(parts, fmt.Sprintf("Lab %d/%d", q.Lab, qt.Lab))
			}
			return strings.Join(parts, " · ")
		}))
	}
}

// parseQuestID splits "blog-3" into ("blog", 3); ok is false for garbage.
func parseQuestID(id string) (line string, n int, ok bool) {
	line, nStr, found := strings.Cut(id, "-")
	if !found {
		return "", 0, false
	}
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		return "", 0, false
	}
	return line, n, true
}

// toast drops one line onto the toast stack: auto-dismisses after ~3 seconds,
// dismissed instantly on click.
func toast(container wasmwrap.Element, text string) {
	if !container.Exists() {
		return
	}
	el := wasmwrap.Create("div")
	el.Class().Add("game-toast")
	el.SetText(text)
	container.Append(el)

	cancel := wasmwrap.SetTimeout(func() { el.Remove() }, 3000)
	reactive.On(el, "click", func(wasmwrap.Event) {
		cancel()
		el.Remove()
	})
}

// --- Sound: WebAudio oscillator blips, no assets. Opt-in only: the shell
// calls these only when save.Sound is on, and nothing plays before then.

func audioContext() js.Value {
	ctx := js.Global().Get("audioCtx")
	if ctx.IsUndefined() || ctx.IsNull() {
		ac := js.Global().Get("AudioContext")
		if !ac.Truthy() {
			return js.Null()
		}
		ctx = ac.New()
		js.Global().Set("audioCtx", ctx)
	}
	return ctx
}

// blip plays one sine tone starting at start seconds with the given duration.
func blip(freq float64, start, dur float64) {
	ctx := audioContext()
	if ctx.IsNull() {
		return
	}
	osc := ctx.Call("createOscillator")
	gain := ctx.Call("createGain")
	now := ctx.Get("currentTime").Float()
	osc.Call("connect", gain)
	gain.Call("connect", ctx.Get("destination"))
	osc.Set("type", "sine")
	osc.Set("frequency", freq)
	gain.Get("gain").Call("setValueAtTime", 0.001, now+start)
	gain.Get("gain").Call("exponentialRampToValueAtTime", 0.12, now+start+0.02)
	gain.Get("gain").Call("exponentialRampToValueAtTime", 0.0001, now+start+dur)
	osc.Call("start", now+start)
	osc.Call("stop", now+start+dur+0.05)
}

// ding is the badge-unlock blip.
func ding() { blip(880, 0, 0.25) }

// fanfare is the quest-complete two-note blip.
func fanfare() {
	blip(523.25, 0, 0.18)
	blip(659.25, 0.12, 0.18)
	blip(783.99, 0.24, 0.3)
}
