// Package game is the tutorial game's pure-Go core: the save schema, the XP
// ledger and the badge rules. It has no build tags and no DOM, so it runs on
// the host — go test ./scripts/game — and compiles into tutorial_game.wasm
// unchanged.
//
// The ledger's one law is "no punishment": every award happens once, nothing
// is ever deducted, and a wrong quiz answer costs nothing — it just pays no
// XP.
package game

import (
	"strconv"
	"strings"
)

// Quest is one quest's player state. Quests are keyed by id — "blog-3" for
// the blog line's third quest, "clicker-1" for the clicker's first — so the
// lines share one save without colliding.
type Quest struct {
	Visited     bool  `json:"visited"`     // first visit, +5, once
	Objectives  int   `json:"objectives"`  // checkpoints done, +10 each, once
	QuizFirst   int   `json:"quizFirst"`   // quizzes right first try, +20 each, once
	QuizRetry   int   `json:"quizRetry"`   // quizzes right after a retry, +5 each, once
	Bonus       bool  `json:"bonus"`       // try-it opened, +15, once
	Lab         int   `json:"lab"`         // step-lab slots solved, +5 each, once
	Complete    bool  `json:"complete"`    // derived: every objective and quiz done, +50, once
	VisitedAt   int64 `json:"visitedAt"`   // unix seconds the first visit started
	CompletedAt int64 `json:"completedAt"` // unix seconds the completion happened
	// CompletedSession is the unix time of the page load the completion
	// happened in — what separates "completed in the visit that started it"
	// from "completed on a later visit" (the Back for More badge).
	CompletedSession int64 `json:"completedSession,omitempty"`
}

// Save is the whole save game: a plain JSON value under
// localStorage["astwerk-tutorial-save"], guarded by Version for migrations.
type Save struct {
	Version int              `json:"v"`
	XP      int              `json:"xp"`
	Badges  []string         `json:"badges"`
	Quests  map[string]Quest `json:"quests"`
	Sound   bool             `json:"sound"`
	ShowAll bool             `json:"showAll"` // the map/hub reveal waiting quests
	SavedAt int64            `json:"savedAt"`
}

// NewSave returns the empty save for a new player.
func NewSave() *Save {
	return &Save{Version: 2, Quests: map[string]Quest{}}
}

// QuestID is the per-line quest key: "blog-3", "clicker-1". Lines and
// numbers are the stable API between the pages, the shell and the ledger.
func QuestID(line string, n int) string { return line + "-" + strconv.Itoa(n) }

func (s *Save) quest(id string) Quest { return s.Quests[id] }

func (s *Save) setQuest(id string, q Quest) { s.Quests[id] = q }

// award adds xp and stamps the save; every award goes through it.
func (s *Save) award(xp int, now int64) int {
	s.XP += xp
	s.SavedAt = now
	return xp
}

// Visit records a quest visit and awards the once-per-quest +5. first is true
// when this visit is the first.
func (s *Save) Visit(id string, now int64) (xp int, first bool) {
	q := s.quest(id)
	if q.Visited {
		return 0, false
	}
	q.Visited = true
	q.VisitedAt = now
	s.setQuest(id, q)
	return s.award(5, now), true
}

// Objective awards +10 for one more checkpoint of quest. total is the quest's
// objective count from the page, so re-ticking an already-awarded objective —
// or unchecking and re-checking one — can never over-award.
func (s *Save) Objective(id string, total int, now int64) (xp int, awarded bool) {
	q := s.quest(id)
	if q.Objectives >= total {
		return 0, false
	}
	q.Objectives++
	s.setQuest(id, q)
	return s.award(10, now), true
}

// Quiz scores one quiz answer: correct first try +20, correct after a retry
// +5, wrong nothing. The ceilings are the quest's quiz count from the page,
// so a quiz reset can't be farmed past the page's own quizzes.
func (s *Save) Quiz(id string, total int, correct, firstTry bool, now int64) (xp int, awarded bool) {
	if !correct {
		return 0, false
	}
	q := s.quest(id)
	if firstTry {
		if q.QuizFirst >= total {
			return 0, false
		}
		q.QuizFirst++
		s.setQuest(id, q)
		return s.award(20, now), true
	}
	if q.QuizRetry >= total {
		return 0, false
	}
	q.QuizRetry++
	s.setQuest(id, q)
	return s.award(5, now), true
}

// Bonus awards the once-per-quest +15 for opening a try-it reveal.
func (s *Save) Bonus(id string, now int64) (xp int, awarded bool) {
	q := s.quest(id)
	if q.Bonus {
		return 0, false
	}
	q.Bonus = true
	s.setQuest(id, q)
	return s.award(15, now), true
}

// Lab awards +5 for one more solved step-lab slot of quest. total is the
// quest's slot count from the frame (data-lab-slots), so re-checking a slot —
// or peeking and resetting — can never over-award.
func (s *Save) Lab(id string, total int, now int64) (xp int, awarded bool) {
	q := s.quest(id)
	if q.Lab >= total {
		return 0, false
	}
	q.Lab++
	s.setQuest(id, q)
	return s.award(5, now), true
}

// Complete marks the quest done and awards +50, once. sessionStart is the
// unix time of the page load the completion happened in, for the Back for
// More badge.
func (s *Save) Complete(id string, sessionStart, now int64) (xp int, awarded bool) {
	q := s.quest(id)
	if q.Complete {
		return 0, false
	}
	q.Complete = true
	q.CompletedAt = now
	q.CompletedSession = sessionStart
	s.setQuest(id, q)
	return s.award(50, now), true
}

// Observe re-evaluates the quest's derived completion: every objective ticked,
// every quiz answered correctly and the whole step lab solved. It awards the
// +50 the first time that state is observed — the manual "complete this step"
// checkbox is gone, so completion comes from the page's own mechanics. The
// lab is part of it (06-step-lab.md): a quest is done when its practice is
// done too, and the reveal-on-check keeps the lab never blocking. sessionStart
// is the unix time of the page load the observation happened in, for the Back
// for More badge.
func (s *Save) Observe(id string, objectives, quizzes, lab int, sessionStart, now int64) bool {
	q := s.quest(id)
	if q.Complete || q.Objectives < objectives || q.QuizFirst+q.QuizRetry < quizzes || q.Lab < lab {
		return false
	}
	_, awarded := s.Complete(id, sessionStart, now)
	return awarded
}

// Badge ids, stable API between the ledger, the shell and the pages' badge
// collection.
const (
	BadgeFirstQuest    = "first-quest"    // any line's quest 1
	BadgeBossSlayer    = "boss-slayer"    // any line's final quest
	BadgeQuizWhiz      = "quiz-whiz"      // 3 quizzes first try, any line
	BadgePerfectionist = "perfectionist"  // every blog quiz first try
	BadgeBonusHunter   = "bonus-hunter"   // every blog try-it opened
	BadgeExplorer      = "explorer"       // every blog quest visited
	BadgeSpeedrunner   = "speedrunner"    // any quest done within 2 minutes of visiting
	BadgeBackForMore   = "back-for-more"  // any quest completed on a later visit
	BadgeCompletionist = "completionist"  // every objective, quiz, bonus and quest of a line
	BadgeCookieMonster = "cookie-monster" // every clicker quest complete
)

// Totals is what a full playthrough requires, per quest line, so the badge
// rules, the map's per-line progress and the dependency graph can compare.
// The pages emit these at build time; the shell unmarshals them into this
// shape. Deps is the quest dependency graph (07-dependencies.md): quest id →
// the quest ids that must be complete before it is "available".
// RecommendedAfter is line-level guidance (line id → quest id). Names maps
// quest ids to their display labels ("3 · The Shell") for tooltips and the
// modal's lists.
type Totals struct {
	Lines            map[string]LineTotals
	Deps             map[string][]string
	RecommendedAfter map[string]string
	Names            map[string]string
	Quests           map[string]QuestMeta
}

// LineTotals is one quest line's game-relevant counts.
type LineTotals struct {
	Quests     int
	Quizzes    int
	Bonuses    int
	Objectives int
	// Ids is the line's quest ids in order — the ledger iterates this list,
	// never a numeric range, so a line's gaps (the moved quests) stay honest.
	// Empty falls back to 1..Quests for hand-built totals.
	Ids []string
}

// QuestMeta is one quest's teachable content counts — the hub's per-row
// progress fraction and its lab receipt read them.
type QuestMeta struct {
	Objectives int
	Quizzes    int
	Lab        int
}

// BadgeNames maps badge ids to their display names — shared by the shell's
// toasts and the overview's badge collection.
var BadgeNames = map[string]string{
	BadgeFirstQuest:    "First Quest",
	BadgeBossSlayer:    "Boss Slayer",
	BadgeQuizWhiz:      "Quiz Whiz",
	BadgePerfectionist: "Perfectionist",
	BadgeBonusHunter:   "Bonus Hunter",
	BadgeExplorer:      "Explorer",
	BadgeSpeedrunner:   "Speedrunner",
	BadgeBackForMore:   "Back for More",
	BadgeCompletionist: "Completionist",
	BadgeCookieMonster: "Cookie Monster",
}

// Rank is a cosmetic XP threshold, shown in the HUD.
type Rank struct {
	Threshold int
	Name      string
}

// Ranks in order; the highest threshold the XP passes is the player's rank.
var Ranks = []Rank{
	{0, "Novice"},
	{100, "Apprentice"},
	{250, "Builder"},
	{450, "Engineer"},
	{700, "Architect"},
	{1000, "Grandmaster"},
}

// Rank returns the player's rank name for their current XP.
func (s *Save) Rank() string {
	name := Ranks[0].Name
	for _, r := range Ranks {
		if s.XP >= r.Threshold {
			name = r.Name
		}
	}
	return name
}

// RankProgress returns the current rank's threshold and the next one's — the
// HUD bar's from/to. At the top rank the next threshold is the current one,
// so the bar reads full.
func (s *Save) RankProgress() (from, to int) {
	from = Ranks[0].Threshold
	to = Ranks[1].Threshold
	for i, r := range Ranks {
		if s.XP >= r.Threshold {
			from = r.Threshold
			if i+1 < len(Ranks) {
				to = Ranks[i+1].Threshold
			} else {
				to = r.Threshold
			}
		}
	}
	return from, to
}

// UnlockBadges evaluates every badge against the save and returns the ones
// that are new — the toasts' payload. It is called after every award, so a
// badge unlocks the moment its condition is met.
func (s *Save) UnlockBadges(t Totals) []string {
	var earned []string
	have := map[string]bool{}
	for _, b := range s.Badges {
		have[b] = true
	}
	for _, b := range s.evaluate(t) {
		if have[b] {
			continue
		}
		have[b] = true
		s.Badges = append(s.Badges, b)
		earned = append(earned, b)
	}
	return earned
}

// lineOrder returns the totals' line keys in a stable order (sorted), so
// badge evaluation and the tests are deterministic.
func lineOrder(t Totals) []string {
	lines := make([]string, 0, len(t.Lines))
	for line := range t.Lines {
		lines = append(lines, line)
	}
	for i := 1; i < len(lines); i++ {
		for j := i; j > 0 && lines[j] < lines[j-1]; j-- {
			lines[j], lines[j-1] = lines[j-1], lines[j]
		}
	}
	return lines
}

// evaluate lists the badge ids the save currently qualifies for, in a stable
// order.
func (s *Save) evaluate(t Totals) []string {
	var out []string
	var firstTry int
	var speedrun, backForMore bool

	for _, line := range lineOrder(t) {
		lt := t.Lines[line]
		ids := questIDs(lt, line)
		objectives, quizzes, bonuses, done, visits := 0, 0, 0, 0, 0
		for _, id := range ids {
			q := s.quest(id)
			objectives += q.Objectives
			quizzes += q.QuizFirst + q.QuizRetry
			bonuses += boolInt(q.Bonus)
			done += boolInt(q.Complete)
			visits += boolInt(q.Visited)
			firstTry += q.QuizFirst
			if q.Complete && q.VisitedAt != 0 && q.CompletedAt-q.VisitedAt <= 120 {
				speedrun = true
			}
			if q.Complete && q.VisitedAt != 0 && q.VisitedAt < q.CompletedSession {
				backForMore = true
			}
		}
		if len(ids) > 0 && s.quest(ids[0]).Complete {
			out = append(out, BadgeFirstQuest)
		}
		if len(ids) > 0 && s.quest(ids[len(ids)-1]).Complete {
			out = append(out, BadgeBossSlayer)
		}
		if lt.Objectives > 0 && objectives >= lt.Objectives && quizzes >= lt.Quizzes &&
			bonuses >= lt.Bonuses && done >= len(ids) {
			out = append(out, BadgeCompletionist)
		}
		if line == "clicker" && done >= len(ids) {
			out = append(out, BadgeCookieMonster)
		}
	}

	// The global four stay blog-scoped — the flagship line defines "the whole
	// tutorial"; the clicker's own line badge is Cookie Monster above.
	if blog, ok := t.Lines["blog"]; ok {
		quizFirst, bonuses, visits := 0, 0, 0
		for _, id := range questIDs(blog, "blog") {
			q := s.quest(id)
			quizFirst += q.QuizFirst
			bonuses += boolInt(q.Bonus)
			visits += boolInt(q.Visited)
		}
		if blog.Quizzes > 0 && quizFirst >= blog.Quizzes {
			out = append(out, BadgePerfectionist)
		}
		if blog.Bonuses > 0 && bonuses >= blog.Bonuses {
			out = append(out, BadgeBonusHunter)
		}
		if visits >= blog.Quests {
			out = append(out, BadgeExplorer)
		}
	}
	if firstTry >= 3 {
		out = append(out, BadgeQuizWhiz)
	}
	if speedrun {
		out = append(out, BadgeSpeedrunner)
	}
	if backForMore {
		out = append(out, BadgeBackForMore)
	}
	return out
}

// questIDs returns a line's quest ids in order: the emitted Ids when present
// (so the ledger honours a line's gaps), else the numeric range 1..Quests for
// hand-built totals.
func questIDs(lt LineTotals, line string) []string {
	if len(lt.Ids) > 0 {
		return lt.Ids
	}
	ids := make([]string, 0, lt.Quests)
	for n := 1; n <= lt.Quests; n++ {
		ids = append(ids, QuestID(line, n))
	}
	return ids
}

// RequiresMet reports whether every requirement of quest id is complete — the
// difference between a quest that is "available" and one that is waiting.
func (s *Save) RequiresMet(t Totals, id string) bool {
	for _, req := range t.Deps[id] {
		if !s.quest(req).Complete {
			return false
		}
	}
	return true
}

// NextAvailable returns the first incomplete quest of a line whose
// requirements are met — the "current quest" the map pulses and the modal
// pushes. "" when the line is complete or has no quests.
func (s *Save) NextAvailable(t Totals, line string) string {
	lt, ok := t.Lines[line]
	if !ok {
		return ""
	}
	for _, id := range questIDs(lt, line) {
		if s.quest(id).Complete {
			continue
		}
		if s.RequiresMet(t, id) {
			return id
		}
	}
	return ""
}

// LineDone reports whether every quest of a line is complete.
func (s *Save) LineDone(t Totals, line string) bool {
	lt, ok := t.Lines[line]
	if !ok {
		return false
	}
	for _, id := range questIDs(lt, line) {
		if !s.quest(id).Complete {
			return false
		}
	}
	return len(questIDs(lt, line)) > 0
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// LegacyKey is the localStorage key the pre-game vanilla progress script
// writes: an array of step markers ("blog-3", or plain numbers from before
// the quest lines). The migration reads it once; the key is left alone
// afterwards, because the no-wasm fallback still writes it.
const LegacyKey = "astwerk-tutorial-steps"

// MigrateLegacy builds a save from the legacy progress array: every completed
// step becomes a completed quest with its +50 completion XP (and a visit,
// since completing implies visiting). A plain number is a blog step ("3"
// becomes "blog-3"); a "line-N" marker passes through validated. The boolean
// reports whether anything was migrated.
func MigrateLegacy(steps []string, blogQuests int, now int64) (*Save, bool) {
	s := NewSave()
	migrated := false
	for _, step := range steps {
		id, ok := legacyID(step, blogQuests)
		if !ok {
			continue
		}
		q := s.quest(id)
		if q.Complete {
			continue
		}
		q.Visited = true
		q.VisitedAt = now
		q.Complete = true
		q.CompletedAt = now
		q.CompletedSession = now
		s.setQuest(id, q)
		s.award(50, now)
		migrated = true
	}
	return s, migrated
}

// legacyID maps one legacy step marker to a quest id. A plain number must be
// within the blog's quest range; a marker with a dash must be "line-N" with
// N ≥ 1. Anything else is garbage and skipped.
func legacyID(step string, blogQuests int) (string, bool) {
	if n, err := strconv.Atoi(step); err == nil {
		if n < 1 || n > blogQuests {
			return "", false
		}
		return QuestID("blog", n), true
	}
	line, nStr, ok := strings.Cut(step, "-")
	if !ok || line == "" {
		return "", false
	}
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		return "", false
	}
	return line + "-" + nStr, true
}

// MigrateV1 rewrites a v1 save into v2: quest keys "1".."11" become
// "blog-1".."blog-11". XP, badges and the rest carry over unchanged; a
// v2 save passes through untouched.
func MigrateV1(s *Save) *Save {
	if s.Version >= 2 {
		return s
	}
	next := map[string]Quest{}
	for k, q := range s.Quests {
		if n, err := strconv.Atoi(k); err == nil {
			next[QuestID("blog", n)] = q
			continue
		}
		next[k] = q
	}
	s.Quests = next
	s.Version = 2
	return s
}
