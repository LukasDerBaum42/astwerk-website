package game

import (
	"encoding/json"
	"testing"
)

// totals describes a full two-line tutorial: the 11-quest blog (23
// objectives, 23 quizzes — quest 9 carries three — 8 try-it bonuses) and the
// 7-quest clicker (15 objectives — quest 6 carries three — 14 quizzes, 6
// bonuses: every clicker quest has a try-it except the first).
func totals() Totals {
	return Totals{Lines: map[string]LineTotals{
		"blog":    {Quests: 11, Quizzes: 23, Bonuses: 8, Objectives: 23},
		"clicker": {Quests: 7, Quizzes: 14, Bonuses: 6, Objectives: 15},
	}}
}

// blogTotals is the blog-only view — what a v1 player's save is compared
// against, so the blog-scoped badges behave exactly as before.
func blogTotals() Totals {
	return Totals{Lines: map[string]LineTotals{
		"blog": totals().Lines["blog"],
	}}
}

// objectives reports the quest's objective count per the quest table.
func objectives(line string, n int) int {
	if line == "blog" && (n == 1 || n == 2 || n == 3) {
		return 3
	}
	if line == "blog" && (n == 9 || n == 10) {
		return 1
	}
	if line == "clicker" && n == 6 {
		return 3
	}
	return 2
}

// hasBonus reports whether the quest's page carries a try-it block.
func hasBonus(line string, n int) bool {
	if line == "clicker" {
		return n != 1
	}
	return n != 1 && n != 2 && n != 7
}

// questCount is the line's quest count.
func questCount(line string) int { return totals().Lines[line].Quests }

// quizCount is the line's per-quest quiz count (two everywhere).
func quizCount() int { return 2 }

// quizCountFor is the quest's quiz count — quest 9 of the blog carries three.
func quizCountFor(line string, n int) int {
	if line == "blog" && n == 9 {
		return 3
	}
	return 2
}

// labCount is the quest's step-lab unit count per the site's labSpecs — the
// lab is part of a quest's completion.
func labCount(line string, n int) int {
	blog := []int{1, 2, 1, 2, 2, 2, 2, 2, 1, 1, 2}
	clicker := []int{2, 2, 2, 1, 1, 1, 2}
	if line == "blog" {
		return blog[n-1]
	}
	return clicker[n-1]
}

// completeQuest plays one quest from first visit to completion in one session:
// the visit, every objective, quiz and lab slot, the bonus where the page has
// one, and the completion five seconds later (so a full playthrough unlocks
// Speedrunner).
func completeQuest(s *Save, line string, n int, session int64) {
	id := QuestID(line, n)
	s.Visit(id, session)
	for range objectives(line, n) {
		s.Objective(id, objectives(line, n), session+1)
	}
	for range quizCountFor(line, n) {
		s.Quiz(id, quizCountFor(line, n), true, true, session+2)
	}
	for range labCount(line, n) {
		s.Lab(id, labCount(line, n), session+3)
	}
	if hasBonus(line, n) {
		s.Bonus(id, session+4)
	}
	s.Complete(id, session, session+5)
}

// completeLine plays a whole line to completion.
func completeLine(s *Save, line string, session int64) {
	for n := 1; n <= questCount(line); n++ {
		completeQuest(s, line, n, session+int64(n)*10)
	}
}

// Every award happens once per quest: re-awarding the same thing pays nothing.
func TestAwardsHappenOnce(t *testing.T) {
	s := NewSave()
	now := int64(1000)
	id := QuestID("blog", 1)

	if xp, _ := s.Visit(id, now); xp != 5 {
		t.Errorf("visit = %d, want 5", xp)
	}
	if xp, first := s.Visit(id, now); first || xp != 0 {
		t.Errorf("second visit = %d (first %v), want 0", xp, first)
	}
	if xp, _ := s.Objective(id, 2, now); xp != 10 {
		t.Errorf("objective = %d, want 10", xp)
	}
	if xp, _ := s.Objective(id, 2, now); xp != 10 {
		t.Errorf("second objective = %d, want 10 — one per objective", xp)
	}
	if xp, ok := s.Objective(id, 2, now); ok || xp != 0 {
		t.Errorf("third objective on a 2-objective quest = %d (ok %v), want 0", xp, ok)
	}
	if xp, _ := s.Quiz(id, 2, true, true, now); xp != 20 {
		t.Errorf("quiz first try = %d, want 20", xp)
	}
	if xp, _ := s.Quiz(id, 2, true, true, now); xp != 20 {
		t.Errorf("second first-try quiz = %d, want 20 — one per quiz", xp)
	}
	if xp, ok := s.Quiz(id, 2, true, true, now); ok || xp != 0 {
		t.Errorf("third first-try quiz = %d (ok %v), want 0 — the page has two quizzes", xp, ok)
	}
	if xp, _ := s.Quiz(id, 2, true, false, now); xp != 5 {
		t.Errorf("quiz retry = %d, want 5", xp)
	}
	if xp, _ := s.Bonus(id, now); xp != 15 {
		t.Errorf("bonus = %d, want 15", xp)
	}
	if xp, ok := s.Bonus(id, now); ok || xp != 0 {
		t.Errorf("second bonus = %d (ok %v), want 0", xp, ok)
	}
	if xp, _ := s.Complete(id, now, now); xp != 50 {
		t.Errorf("complete = %d, want 50", xp)
	}
	if xp, ok := s.Complete(id, now, now); ok || xp != 0 {
		t.Errorf("second complete = %d (ok %v), want 0", xp, ok)
	}

	want := 5 + 2*10 + 2*20 + 5 + 15 + 50
	if s.XP != want {
		t.Errorf("total XP = %d, want %d", s.XP, want)
	}
}

// Nothing is ever deducted: a wrong quiz answer pays no XP and changes nothing.
func TestWrongAnswersCostNothing(t *testing.T) {
	s := NewSave()
	id := QuestID("blog", 1)
	if xp, ok := s.Quiz(id, 2, false, true, 1000); ok || xp != 0 {
		t.Errorf("wrong first try = %d (ok %v), want 0", xp, ok)
	}
	if xp, ok := s.Quiz(id, 2, false, false, 1000); ok || xp != 0 {
		t.Errorf("wrong retry = %d (ok %v), want 0", xp, ok)
	}
	if s.XP != 0 {
		t.Errorf("XP = %d, want 0", s.XP)
	}
	q := s.quest(id)
	if q.QuizFirst != 0 || q.QuizRetry != 0 {
		t.Errorf("quiz counters moved on wrong answers: %+v", q)
	}
}

// The ceilings come from the page's own counts, so unchecking and re-ticking
// an objective (or resetting a quiz) cannot farm XP.
func TestCeilingsBlockFarming(t *testing.T) {
	s := NewSave()
	id := QuestID("blog", 1)
	for range 5 {
		s.Objective(id, 2, 1000)
		s.Quiz(id, 2, true, true, 1000)
		s.Quiz(id, 2, true, false, 1000)
	}
	// Two awards each, then silence.
	if s.XP != 2*10+2*20+2*5 {
		t.Errorf("XP = %d, want %d", s.XP, 2*10+2*20+2*5)
	}
}

// Completion is derived: the +50 fires the first time every objective, every
// quiz and the whole lab of a quest is observed done — the manual checkbox is
// gone, and the lab is part of it.
func TestDerivedCompletion(t *testing.T) {
	s := NewSave()
	id := QuestID("blog", 1)

	s.Objective(id, 2, 1000)
	s.Quiz(id, 2, true, true, 1001)
	if s.Observe(id, 2, 2, 1, 2000, 1002) {
		t.Error("completed with one objective and one quiz done")
	}
	if s.quest(id).Complete {
		t.Error("quest marked complete too early")
	}

	s.Objective(id, 2, 1003)
	if s.Observe(id, 2, 2, 1, 2000, 1004) {
		t.Error("completed with one quiz still missing")
	}
	s.Quiz(id, 2, true, true, 1005)
	if s.Observe(id, 2, 2, 1, 2000, 1006) {
		t.Error("completed with the lab still undone")
	}
	s.Lab(id, 1, 1007)
	if !s.Observe(id, 2, 2, 1, 2000, 1008) {
		t.Error("derived completion did not fire when the last quiz and the lab were done")
	}
	q := s.quest(id)
	if !q.Complete || q.CompletedSession != 2000 {
		t.Errorf("quest not recorded as complete: %+v", q)
	}
	if s.Observe(id, 2, 2, 1, 2000, 1005) {
		t.Error("completion awarded twice")
	}

	// A retry-answered quiz counts too: answered correctly is answered.
	s2 := NewSave()
	s2.Objective(QuestID("clicker", 1), 2, 1000)
	s2.Objective(QuestID("clicker", 1), 2, 1001)
	s2.Quiz(QuestID("clicker", 1), 2, true, false, 1002) // right after a retry
	s2.Quiz(QuestID("clicker", 1), 2, true, true, 1003)
	s2.Lab(QuestID("clicker", 1), 2, 1004)
	s2.Lab(QuestID("clicker", 1), 2, 1005)
	if !s2.Observe(QuestID("clicker", 1), 2, 2, 2, 1000, 1006) {
		t.Error("retry-correct quizzes and the lab must count toward derived completion")
	}
}

func TestRanks(t *testing.T) {
	cases := []struct {
		xp   int
		want string
	}{
		{0, "Novice"},
		{99, "Novice"},
		{100, "Apprentice"},
		{249, "Apprentice"},
		{250, "Builder"},
		{450, "Engineer"},
		{700, "Architect"},
		{999, "Architect"},
		{1000, "Grandmaster"},
		{5000, "Grandmaster"},
	}
	for _, c := range cases {
		s := NewSave()
		s.XP = c.xp
		if got := s.Rank(); got != c.want {
			t.Errorf("XP %d: rank %q, want %q", c.xp, got, c.want)
		}
	}
}

func TestRankProgress(t *testing.T) {
	s := NewSave()
	s.XP = 120
	from, to := s.RankProgress()
	if from != 100 || to != 250 {
		t.Errorf("XP 120: from %d to %d, want 100 to 250", from, to)
	}
	s.XP = 2000
	from, to = s.RankProgress()
	if from != 1000 || to != 1000 {
		t.Errorf("XP 2000: from %d to %d, want full bar", from, to)
	}
}

// A full two-line playthrough: 18 visits, 38 objectives, 37 first-try
// quizzes, 14 bonuses and 18 completions.
func TestFullPlaythroughXP(t *testing.T) {
	s := NewSave()
	for _, line := range []string{"blog", "clicker"} {
		completeLine(s, line, int64(1000))
	}
	// 18 visits + 38 objectives + 37 quizzes + 29 lab slots + 14 bonuses + 18 completions.
	want := 18*5 + 38*10 + 37*20 + 29*5 + 14*15 + 18*50
	if s.XP != want {
		t.Errorf("XP = %d, want %d", s.XP, want)
	}
	if s.Rank() != "Grandmaster" {
		t.Errorf("rank = %q, want Grandmaster", s.Rank())
	}
}

func hasBadge(earned []string, id string) bool {
	for _, b := range earned {
		if b == id {
			return true
		}
	}
	return false
}

func TestBadgesAllUnlock(t *testing.T) {
	s := NewSave()
	// Quest blog-1 starts in session 1000 and is finished in session 2000 —
	// the Back for More condition. Everything else is played in its own session.
	s.Visit(QuestID("blog", 1), 1000)
	s.Objective(QuestID("blog", 1), 2, 1001)
	s.Objective(QuestID("blog", 1), 2, 1002)
	s.Quiz(QuestID("blog", 1), 2, true, true, 1003)
	s.Quiz(QuestID("blog", 1), 2, true, true, 1004)
	s.Bonus(QuestID("blog", 1), 1005)
	s.Complete(QuestID("blog", 1), 2000, 2001)
	completeQuest(s, "blog", 2, 3000)
	completeQuest(s, "blog", 3, 3000)
	completeQuest(s, "blog", 4, 3000)
	completeQuest(s, "blog", 5, 3000)
	completeQuest(s, "blog", 6, 3000)
	completeQuest(s, "blog", 7, 3000)
	completeQuest(s, "blog", 8, 3000)
	completeQuest(s, "blog", 9, 3000)
	completeQuest(s, "blog", 10, 3000)
	completeQuest(s, "blog", 11, 3000)
	completeLine(s, "clicker", 4000)
	earned := s.UnlockBadges(totals())

	all := []string{
		BadgeFirstQuest, BadgeBossSlayer, BadgeQuizWhiz, BadgePerfectionist,
		BadgeBonusHunter, BadgeExplorer, BadgeSpeedrunner, BadgeBackForMore,
		BadgeCompletionist, BadgeCookieMonster,
	}
	if len(earned) != len(all) {
		t.Fatalf("earned %d badges, want %d: %v", len(earned), len(all), earned)
	}
	for _, want := range all {
		if !hasBadge(earned, want) {
			t.Errorf("missing badge %q in %v", want, earned)
		}
	}
}

// Badges must not unlock before their condition: partial playthroughs earn
// only what they actually did.
func TestBadgesRespectConditions(t *testing.T) {
	tot := totals()

	s := NewSave()
	s.Visit(QuestID("blog", 1), 1000)
	s.Complete(QuestID("blog", 1), 1000, 1121) // completion more than two minutes after the visit
	earned := s.UnlockBadges(tot)
	if len(earned) != 1 || earned[0] != BadgeFirstQuest {
		t.Errorf("quest 1 only: earned %v, want [first-quest]", earned)
	}

	s = NewSave()
	for n := 1; n <= 11; n++ {
		s.Visit(QuestID("blog", n), 1000)
	}
	if earned := s.UnlockBadges(tot); len(earned) != 1 || earned[0] != BadgeExplorer {
		t.Errorf("visits only: earned %v, want [explorer]", earned)
	}

	s = NewSave()
	for n := 1; n <= 11; n++ {
		s.Visit(QuestID("blog", n), 1000)
		s.Complete(QuestID("blog", n), 1000, 1121) // same session, no speedrun, no back-for-more
	}
	earned = s.UnlockBadges(tot)
	if len(earned) != 3 {
		t.Errorf("completed blog same session: earned %v, want [first-quest boss-slayer explorer]", earned)
	}
}

// The clicker line's badges: Cookie Monster needs every clicker quest, and
// completing only the clicker unlocks the line badges but not the blog-scoped
// ones.
func TestClickerLineBadges(t *testing.T) {
	tot := totals()

	s := NewSave()
	s.Visit(QuestID("clicker", 1), 1000)
	s.Complete(QuestID("clicker", 1), 1000, 1300) // no speedrun
	if earned := s.UnlockBadges(tot); len(earned) != 1 || earned[0] != BadgeFirstQuest {
		t.Errorf("clicker quest 1: earned %v, want [first-quest]", earned)
	}

	s = NewSave()
	completeLine(s, "clicker", 1000)
	earned := s.UnlockBadges(tot)
	if !hasBadge(earned, BadgeCookieMonster) {
		t.Errorf("full clicker line: missing cookie-monster in %v", earned)
	}
	if !hasBadge(earned, BadgeBossSlayer) {
		t.Errorf("full clicker line: missing boss-slayer in %v", earned)
	}
	if hasBadge(earned, BadgePerfectionist) || hasBadge(earned, BadgeExplorer) {
		t.Errorf("clicker-only playthrough earned blog-scoped badges: %v", earned)
	}
}

func TestSpeedrunnerNeedsTwoMinutes(t *testing.T) {
	tot := totals()
	id := QuestID("blog", 1)

	s := NewSave()
	s.Visit(id, 1000)
	s.Complete(id, 1000, 1121) // 121 seconds: over the limit
	if hasBadge(s.UnlockBadges(tot), BadgeSpeedrunner) {
		t.Error("speedrunner unlocked after more than two minutes")
	}

	s = NewSave()
	s.Visit(id, 1000)
	s.Complete(id, 1000, 1120) // exactly 120 seconds: at the limit
	if !hasBadge(s.UnlockBadges(tot), BadgeSpeedrunner) {
		t.Error("speedrunner did not unlock at exactly two minutes")
	}

	s = NewSave()
	s.Visit(id, 1000)
	s.Complete(id, 1000, 1119) // 119 seconds: within
	if !hasBadge(s.UnlockBadges(tot), BadgeSpeedrunner) {
		t.Error("speedrunner did not unlock within two minutes")
	}
}

func TestBackForMoreNeedsLaterVisit(t *testing.T) {
	id := QuestID("blog", 1)

	s := NewSave()
	// Visited in session 1000, completed in session 2000: a later visit.
	s.Visit(id, 1000)
	s.Complete(id, 2000, 2001)
	if !hasBadge(s.UnlockBadges(totals()), BadgeBackForMore) {
		t.Error("back-for-more did not unlock for a completion on a later visit")
	}

	s = NewSave()
	// Visited and completed in the same session: no badge.
	s.Visit(id, 1000)
	s.Complete(id, 1000, 1001)
	if hasBadge(s.UnlockBadges(totals()), BadgeBackForMore) {
		t.Error("back-for-more unlocked for a completion in the starting visit")
	}
}

// The blog-scoped badges must behave identically when only the blog totals
// are known — the v1 situation.
func TestBlogScopedAgainstBlogTotals(t *testing.T) {
	s := NewSave()
	s.Visit(QuestID("blog", 1), 1000)
	s.Objective(QuestID("blog", 1), 2, 1001)
	s.Objective(QuestID("blog", 1), 2, 1002)
	s.Quiz(QuestID("blog", 1), 2, true, true, 1003)
	s.Quiz(QuestID("blog", 1), 2, true, true, 1004)
	s.Bonus(QuestID("blog", 1), 1005)
	s.Complete(QuestID("blog", 1), 1000, 1300) // 300 seconds: no speedrun
	earned := s.UnlockBadges(blogTotals())
	if len(earned) != 1 || earned[0] != BadgeFirstQuest {
		t.Errorf("blog quest 1: earned %v, want [first-quest]", earned)
	}
}

// The legacy key now carries quest ids ("blog-3", "clicker-1"), with plain
// numbers still read as blog steps.
func TestMigration(t *testing.T) {
	s, ok := MigrateLegacy([]string{"1", "3", "3", "x", "99", "clicker-2"}, 11, 1000)
	if !ok {
		t.Fatal("migration reported nothing")
	}
	if s.XP != 150 {
		t.Errorf("XP = %d, want 150 (blog 1+3 and clicker 2 × 50)", s.XP)
	}
	for _, id := range []string{QuestID("blog", 1), QuestID("blog", 3), QuestID("clicker", 2)} {
		q := s.quest(id)
		if !q.Complete || !q.Visited {
			t.Errorf("quest %s not migrated: %+v", id, q)
		}
	}
	if s.quest(QuestID("blog", 2)).Complete {
		t.Error("blog quest 2 completed without being in the legacy list")
	}
	if s.Version != 2 {
		t.Errorf("version = %d, want 2", s.Version)
	}

	if _, ok := MigrateLegacy(nil, 11, 1000); ok {
		t.Error("empty legacy list reported a migration")
	}
}

// A v1 save (quests keyed by step number) migrates to per-line ids, keeping
// its XP and badges; a v2 save passes through untouched.
func TestMigrateV1(t *testing.T) {
	v1 := &Save{
		Version: 1,
		XP:      150,
		Badges:  []string{BadgeFirstQuest},
		Quests: map[string]Quest{
			"1":  {Visited: true, Complete: true, CompletedSession: 1000},
			"11": {Visited: true, Complete: true, CompletedSession: 2000},
		},
	}
	s := MigrateV1(v1)
	if s.Version != 2 {
		t.Fatalf("version = %d, want 2", s.Version)
	}
	if s.XP != 150 || len(s.Badges) != 1 {
		t.Errorf("XP/badges changed in migration: %+v", s)
	}
	if !s.quest(QuestID("blog", 1)).Complete || !s.quest(QuestID("blog", 11)).Complete {
		t.Errorf("v1 keys did not become blog ids: %+v", s.Quests)
	}
	if _, ok := s.Quests["1"]; ok {
		t.Error("old numeric key survived migration")
	}

	// A v2 save passes through.
	v2 := MigrateV1(s)
	if v2 != s {
		t.Error("MigrateV1 must not rebuild a v2 save")
	}
}

// The save is a plain JSON value; it must round-trip through the key the
// shell reads and writes.
func TestSaveRoundTrips(t *testing.T) {
	s := NewSave()
	for _, line := range []string{"blog", "clicker"} {
		completeLine(s, line, int64(1000))
	}
	s.Sound = true

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var back Save
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.XP != s.XP || back.Sound != s.Sound || len(back.Badges) != 0 {
		t.Errorf("round trip changed the save: %+v", back)
	}
	for _, line := range []string{"blog", "clicker"} {
		for n := 1; n <= questCount(line); n++ {
			if back.quest(QuestID(line, n)) != s.quest(QuestID(line, n)) {
				t.Errorf("quest %s changed in round trip", QuestID(line, n))
			}
		}
	}
}

// depTotals is the two-line totals with the dependency graph and the lines'
// quest id lists — what the pages emit (07-dependencies.md). The blog keeps
// its 11 quests here (6 and 8 still live); the graph branches after 3 and 5
// and converges on the boss.
func depTotals() Totals {
	blog := []string{}
	for n := 1; n <= 11; n++ {
		blog = append(blog, QuestID("blog", n))
	}
	clicker := []string{}
	for n := 1; n <= 7; n++ {
		clicker = append(clicker, QuestID("clicker", n))
	}
	t := totals()
	t.Lines["blog"] = func() LineTotals {
		lt := t.Lines["blog"]
		lt.Ids = blog
		return lt
	}()
	t.Lines["clicker"] = func() LineTotals {
		lt := t.Lines["clicker"]
		lt.Ids = clicker
		return lt
	}()
	t.Deps = map[string][]string{
		QuestID("blog", 2):  {QuestID("blog", 1)},
		QuestID("blog", 3):  {QuestID("blog", 2)},
		QuestID("blog", 4):  {QuestID("blog", 2), QuestID("blog", 3)},
		QuestID("blog", 5):  {QuestID("blog", 4)},
		QuestID("blog", 6):  {QuestID("blog", 5)},
		QuestID("blog", 7):  {QuestID("blog", 5)},
		QuestID("blog", 8):  {QuestID("blog", 7)},
		QuestID("blog", 9):  {QuestID("blog", 3)},
		QuestID("blog", 10): {QuestID("blog", 5)},
		QuestID("blog", 11): {QuestID("blog", 7), QuestID("blog", 9), QuestID("blog", 10)},
	}
	for n := 2; n <= 7; n++ {
		t.Deps[QuestID("clicker", n)] = []string{QuestID("clicker", n-1)}
	}
	return t
}

// NextAvailable is requirements-first: a quest whose requirements are unmet is
// waiting, not current; the first incomplete with requirements met is the
// current quest; "" when the line is done.
func TestNextAvailable(t *testing.T) {
	tot := depTotals()

	s := NewSave()
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 1) {
		t.Errorf("fresh save: next = %q, want blog-1", got)
	}

	// Completing 3 opens the branch: both 4 (content) and 9 (the spark) are
	// available; 4 comes first in line order.
	for _, n := range []int{1, 2, 3} {
		s.Complete(QuestID("blog", n), 1000, 1001)
	}
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 4) {
		t.Errorf("after quest 3: next = %q, want blog-4", got)
	}

	// After 5, quests 6, 7 and 10 are all open — 6 comes first in line order,
	// then 7, 8 (which needs 7), 10, and the 9 detour.
	s.Complete(QuestID("blog", 4), 1000, 1001)
	s.Complete(QuestID("blog", 5), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 6) {
		t.Errorf("after quest 5: next = %q, want blog-6", got)
	}

	// The boss stays waiting until its three branches are done: 7, 9 and 10
	// (8 rides on 7 but is not a boss requirement — it leaves the line later).
	s.Complete(QuestID("blog", 6), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 7) {
		t.Errorf("after quest 6: next = %q, want blog-7", got)
	}
	s.Complete(QuestID("blog", 7), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 8) {
		t.Errorf("after quest 7: next = %q, want blog-8", got)
	}
	s.Complete(QuestID("blog", 8), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 9) {
		t.Errorf("after quest 8: next = %q, want blog-9", got)
	}
	s.Complete(QuestID("blog", 9), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 10) {
		t.Errorf("after quest 9: next = %q, want blog-10", got)
	}
	s.Complete(QuestID("blog", 10), 1000, 1001)
	if got := s.NextAvailable(tot, "blog"); got != QuestID("blog", 11) {
		t.Errorf("boss open: next = %q, want blog-11", got)
	}

	// A quest completed out of order still awards and is simply not the
	// suggested next: jumping straight to blog-9 from an empty save works and
	// the suggestion ignores it (9 is waiting, not current).
	s2 := NewSave()
	s2.Complete(QuestID("blog", 9), 1000, 1001)
	if got := s2.NextAvailable(tot, "blog"); got != QuestID("blog", 1) {
		t.Errorf("out-of-order completion: next = %q, want blog-1", got)
	}

	// A complete line has no next.
	completeLine(s, "blog", 1000)
	if got := s.NextAvailable(tot, "blog"); got != "" {
		t.Errorf("done line: next = %q, want empty", got)
	}

	// The clicker line is strictly sequential.
	sc := NewSave()
	if got := sc.NextAvailable(tot, "clicker"); got != QuestID("clicker", 1) {
		t.Errorf("fresh clicker: next = %q, want clicker-1", got)
	}
	sc.Complete(QuestID("clicker", 2), 1000, 1001) // out of order: still waiting
	if got := sc.NextAvailable(tot, "clicker"); got != QuestID("clicker", 1) {
		t.Errorf("clicker with quest 2 done: next = %q, want clicker-1", got)
	}
}

// RequiresMet and LineDone report the dependency state the map and the badges
// read.
func TestRequiresMetAndLineDone(t *testing.T) {
	tot := depTotals()
	s := NewSave()
	if s.RequiresMet(tot, QuestID("blog", 11)) {
		t.Error("boss requirements met on a fresh save")
	}
	for _, n := range []int{1, 2, 3} {
		s.Complete(QuestID("blog", n), 1000, 1001)
	}
	if !s.RequiresMet(tot, QuestID("blog", 4)) {
		t.Error("quest 4 requirements not met after 1-3")
	}
	if s.LineDone(tot, "blog") {
		t.Error("blog line done after three quests")
	}
	completeLine(s, "blog", 1000)
	if !s.LineDone(tot, "blog") {
		t.Error("blog line not done after a full playthrough")
	}
	if s.LineDone(tot, "clicker") {
		t.Error("clicker line done when no clicker quest was played")
	}
}

// The step-lab award: +5 per slot, once, capped at the frame's slot count;
// wrong answers cost nothing.
func TestLabAwardsOncePerSlot(t *testing.T) {
	s := NewSave()
	id := QuestID("blog", 2)
	if xp, ok := s.Lab(id, 2, 1000); !ok || xp != 5 {
		t.Errorf("first slot = %d (ok %v), want 5", xp, ok)
	}
	if xp, ok := s.Lab(id, 2, 1001); !ok || xp != 5 {
		t.Errorf("second slot = %d (ok %v), want 5", xp, ok)
	}
	if xp, ok := s.Lab(id, 2, 1002); ok || xp != 0 {
		t.Errorf("third slot on a 2-slot frame = %d (ok %v), want 0", xp, ok)
	}
	// Peek/reset never deduct: re-running the same slots pays nothing but
	// never takes anything back.
	if s.XP != 10 {
		t.Errorf("XP = %d, want 10", s.XP)
	}
	// The lab is part of derived completion: objectives and quizzes alone no
	// longer complete a quest, the lab must be done too.
	s3 := NewSave()
	id3 := QuestID("blog", 2)
	s3.Objective(id3, 2, 2000)
	s3.Objective(id3, 2, 2001)
	s3.Quiz(id3, 2, true, true, 2002)
	s3.Quiz(id3, 2, true, true, 2003)
	if s3.Observe(id3, 2, 2, 2, 1000, 2004) {
		t.Error("objectives + quizzes without the lab must not complete a quest")
	}
	s3.Lab(id3, 2, 2005)
	if s3.Observe(id3, 2, 2, 2, 1000, 2006) {
		t.Error("one lab slot of two must not complete the quest")
	}
	s3.Lab(id3, 2, 2007)
	if !s3.Observe(id3, 2, 2, 2, 1000, 2008) {
		t.Error("the last lab slot completes the quest")
	}
	if got := s.quest(id).Lab; got != 2 {
		t.Errorf("Lab count = %d, want 2", got)
	}
}

// The totals round-trip through JSON with the new shape — the emitted
// data-game-totals is what the shell unmarshals, so Deps/Ids/Names must land.
func TestTotalsRoundTrip(t *testing.T) {
	tot := depTotals()
	tot.Names = map[string]string{QuestID("blog", 3): "3 · The Shell"}
	b, err := json.Marshal(tot)
	if err != nil {
		t.Fatal(err)
	}
	var back Totals
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Lines["blog"].Ids == nil || len(back.Lines["blog"].Ids) != 11 {
		t.Errorf("Ids did not round-trip: %v", back.Lines["blog"].Ids)
	}
	if len(back.Deps[QuestID("blog", 11)]) != 3 {
		t.Errorf("Deps did not round-trip: %v", back.Deps)
	}
	if back.Names[QuestID("blog", 3)] != "3 · The Shell" {
		t.Errorf("Names did not round-trip: %v", back.Names)
	}
}
