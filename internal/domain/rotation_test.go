package domain_test

import (
	"testing"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

func newWeeklyRotation(day, t, tz string, members []domain.Member, scheduled *domain.Member) domain.Rotation {
	return domain.Rotation{
		ID:   "rot_test",
		Name: "Test Rotation",
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      day,
				Time:     t,
				TimeZone: tz,
			},
		},
		Members:         members,
		ScheduledMember: scheduled,
	}
}

var (
	alice   = domain.Member{ID: "m1", Order: 1, User: domain.User{Name: "Alice"}}
	bob     = domain.Member{ID: "m2", Order: 2, User: domain.User{Name: "Bob"}}
	charlie = domain.Member{ID: "m3", Order: 3, User: domain.User{Name: "Charlie"}}
)

// Monday 2024-01-08 10:00 UTC — a Monday, after the 09:00 handoff
var fixedNow = time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)

func TestSchedule_BasicCycle(t *testing.T) {
	members := []domain.Member{alice, bob, charlie}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &bob)

	blocks, err := r.Schedule(fixedNow, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}

	// Current period started 2024-01-08 09:00 UTC (today, before fixedNow)
	periodStart := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)

	expectations := []struct {
		memberID string
		start    time.Time
	}{
		{"m2", periodStart},                   // Bob
		{"m3", periodStart.AddDate(0, 0, 7)},  // Charlie
		{"m1", periodStart.AddDate(0, 0, 14)}, // Alice (wrap)
	}

	for i, exp := range expectations {
		b := blocks[i]
		if b.Member.ID != exp.memberID {
			t.Errorf("block %d: expected member %s, got %s", i, exp.memberID, b.Member.ID)
		}
		if !b.Start.Equal(exp.start) {
			t.Errorf("block %d: expected start %v, got %v", i, exp.start, b.Start)
		}
		expectedEnd := exp.start.AddDate(0, 0, 7)
		if !b.End.Equal(expectedEnd) {
			t.Errorf("block %d: expected end %v, got %v", i, expectedEnd, b.End)
		}
	}
}

func TestSchedule_WrapAround(t *testing.T) {
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &bob)

	blocks, err := r.Schedule(fixedNow, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantIDs := []string{"m2", "m1", "m2", "m1"} // Bob, Alice, Bob, Alice
	for i, id := range wantIDs {
		if blocks[i].Member.ID != id {
			t.Errorf("block %d: expected %s, got %s", i, id, blocks[i].Member.ID)
		}
	}
}

func TestSchedule_BeforeHandoffTime(t *testing.T) {
	// now is Monday 2024-01-08 08:00 UTC — before the 09:00 handoff
	// so current period started the PREVIOUS Monday: 2024-01-01 09:00 UTC
	now := time.Date(2024, 1, 8, 8, 0, 0, 0, time.UTC)
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &alice)

	blocks, err := r.Schedule(now, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	if !blocks[0].Start.Equal(expectedStart) {
		t.Errorf("expected period start %v, got %v", expectedStart, blocks[0].Start)
	}
	if blocks[0].Member.ID != "m1" {
		t.Errorf("expected Alice, got %s", blocks[0].Member.ID)
	}
}

func TestSchedule_ExactlyAtHandoffTime(t *testing.T) {
	// now equals the handoff time exactly — current period starts now
	now := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &bob)

	blocks, err := r.Schedule(now, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !blocks[0].Start.Equal(now) {
		t.Errorf("expected start %v, got %v", now, blocks[0].Start)
	}
}

func TestSchedule_ZeroWeeks(t *testing.T) {
	members := []domain.Member{alice}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &alice)

	blocks, err := r.Schedule(fixedNow, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected empty slice, got %d blocks", len(blocks))
	}
}

func TestSchedule_NoMembers(t *testing.T) {
	r := newWeeklyRotation("Monday", "09:00", "UTC", nil, nil)
	r.ScheduledMember = &alice // ScheduledMember set but Members list is empty

	blocks, err := r.Schedule(fixedNow, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected empty slice, got %d blocks", len(blocks))
	}
}

func TestSchedule_NoScheduledMember(t *testing.T) {
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, nil)

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error when ScheduledMember is nil")
	}
}

func TestSchedule_ScheduledMemberNotInMembers(t *testing.T) {
	outsider := domain.Member{ID: "m99", Order: 99, User: domain.User{Name: "Outsider"}}
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &outsider)

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error when scheduled member is not in members list")
	}
}

func TestSchedule_NilWeeklyCadence(t *testing.T) {
	r := domain.Rotation{
		Members:         []domain.Member{alice},
		ScheduledMember: &alice,
	}

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error when weekly cadence is nil")
	}
}

func TestSchedule_InvalidTimezone(t *testing.T) {
	members := []domain.Member{alice}
	r := newWeeklyRotation("Monday", "09:00", "Not/ATimezone", members, &alice)

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error for invalid timezone")
	}
}

func TestSchedule_DSTTimezone(t *testing.T) {
	// America/New_York clocks spring forward 2024-03-10 02:00 → 03:00
	// Use a rotation that crosses this boundary.
	// now: Monday 2024-03-04 10:00 ET (before DST change)
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Date(2024, 3, 4, 10, 0, 0, 0, loc)

	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "America/New_York", members, &alice)

	blocks, err := r.Schedule(now, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Block 0: 2024-03-04 09:00 ET → 2024-03-11 09:00 ET
	// Block 1: 2024-03-11 09:00 ET → 2024-03-18 09:00 ET (after DST spring-forward)
	// Both blocks should still span exactly 7 calendar days (even though duration differs).
	for i, b := range blocks {
		// Compare calendar dates in UTC to avoid DST-skewed durations.
		startY, startM, startD := b.Start.In(loc).Date()
		endY, endM, endD := b.End.In(loc).Date()
		startDate := time.Date(startY, startM, startD, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(endY, endM, endD, 0, 0, 0, 0, time.UTC)
		if !endDate.Equal(startDate.AddDate(0, 0, 7)) {
			t.Errorf("block %d: expected end date 7 calendar days after start, start=%v end=%v", i, startDate, endDate)
		}
	}

	// Handoff time should be 09:00 in local time for both blocks
	for i, b := range blocks {
		h, m, _ := b.Start.In(loc).Clock()
		if h != 9 || m != 0 {
			t.Errorf("block %d: expected start time 09:00, got %02d:%02d", i, h, m)
		}
	}
}

func TestSchedule_OverridePartialBlock(t *testing.T) {
	// Override covers Wednesday–Friday within a Mon–Mon block.
	// Expect three sub-blocks: Mon→Wed (alice), Wed→Fri (charlie, override), Fri→Mon (alice).
	members := []domain.Member{alice, bob, charlie}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &alice)

	// Override: Wed 2024-01-10 09:00 → Fri 2024-01-12 09:00
	override := domain.Override{
		ID:     "ovr_1",
		Member: charlie,
		Start:  time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
		End:    time.Date(2024, 1, 12, 9, 0, 0, 0, time.UTC),
	}

	r.Overrides = []domain.Override{override}
	blocks, err := r.Schedule(fixedNow, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}

	periodStart := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)

	// Block 0: regular alice Mon→Wed
	if blocks[0].Member.ID != "m1" || blocks[0].IsOverride {
		t.Errorf("block 0: expected alice (regular), got member=%s override=%v", blocks[0].Member.ID, blocks[0].IsOverride)
	}
	if !blocks[0].Start.Equal(periodStart) || !blocks[0].End.Equal(override.Start) {
		t.Errorf("block 0 times wrong: %v – %v", blocks[0].Start, blocks[0].End)
	}

	// Block 1: override charlie Wed→Fri
	if blocks[1].Member.ID != "m3" || !blocks[1].IsOverride {
		t.Errorf("block 1: expected charlie (override), got member=%s override=%v", blocks[1].Member.ID, blocks[1].IsOverride)
	}
	if !blocks[1].Start.Equal(override.Start) || !blocks[1].End.Equal(override.End) {
		t.Errorf("block 1 times wrong: %v – %v", blocks[1].Start, blocks[1].End)
	}

	// Block 2: regular alice Fri→Mon
	if blocks[2].Member.ID != "m1" || blocks[2].IsOverride {
		t.Errorf("block 2: expected alice (regular), got member=%s override=%v", blocks[2].Member.ID, blocks[2].IsOverride)
	}
	if !blocks[2].Start.Equal(override.End) || !blocks[2].End.Equal(periodStart.AddDate(0, 0, 7)) {
		t.Errorf("block 2 times wrong: %v – %v", blocks[2].Start, blocks[2].End)
	}
}

func TestEffectiveOnCall_UsesActiveOverride(t *testing.T) {
	r := newWeeklyRotation("Monday", "09:00", "UTC", []domain.Member{alice, bob}, &alice)
	r.Overrides = []domain.Override{
		{
			ID:     "ovr_1",
			Member: bob,
			Start:  fixedNow.Add(-time.Hour),
			End:    fixedNow.Add(time.Hour),
		},
	}

	got := r.EffectiveOnCall(fixedNow)

	if got == nil {
		t.Fatal("expected effective on-call member")
	}
	if got.ID != bob.ID {
		t.Fatalf("expected override member %s, got %s", bob.ID, got.ID)
	}
}

func TestEffectiveOnCall_FallsBackToScheduledMember(t *testing.T) {
	r := newWeeklyRotation("Monday", "09:00", "UTC", []domain.Member{alice, bob}, &alice)
	r.Overrides = []domain.Override{
		{
			ID:     "ovr_1",
			Member: bob,
			Start:  fixedNow.Add(time.Hour),
			End:    fixedNow.Add(2 * time.Hour),
		},
	}

	got := r.EffectiveOnCall(fixedNow)

	if got == nil {
		t.Fatal("expected effective on-call member")
	}
	if got.ID != alice.ID {
		t.Fatalf("expected scheduled member %s, got %s", alice.ID, got.ID)
	}
}

func TestSchedule_OverrideFullBlock(t *testing.T) {
	// Override covers the entire first weekly block.
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &alice)

	periodStart := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)
	override := domain.Override{
		ID:     "ovr_1",
		Member: charlie,
		Start:  periodStart,
		End:    periodStart.AddDate(0, 0, 7),
	}

	r.Overrides = []domain.Override{override}
	blocks, err := r.Schedule(fixedNow, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	// Block 0: charlie override for the full first week
	if blocks[0].Member.ID != "m3" || !blocks[0].IsOverride {
		t.Errorf("block 0: expected charlie (override), got member=%s override=%v", blocks[0].Member.ID, blocks[0].IsOverride)
	}

	// Block 1: regular bob for second week
	if blocks[1].Member.ID != "m2" || blocks[1].IsOverride {
		t.Errorf("block 1: expected bob (regular), got member=%s override=%v", blocks[1].Member.ID, blocks[1].IsOverride)
	}
}

func TestSchedule_OverrideSpansMultipleBlocks(t *testing.T) {
	// Override spans from mid-week-1 into mid-week-2.
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &alice)

	periodStart := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)
	// Override: Wed Jan 10 → Wed Jan 17 (spans into bob's week)
	override := domain.Override{
		ID:     "ovr_1",
		Member: charlie,
		Start:  time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC),
		End:    time.Date(2024, 1, 17, 9, 0, 0, 0, time.UTC),
	}

	r.Overrides = []domain.Override{override}
	blocks, err := r.Schedule(fixedNow, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect: alice(Mon-Wed), charlie(Wed-Mon), charlie(Mon-Wed), bob(Wed-Mon)
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}

	// Block 0: alice Mon→Wed
	if blocks[0].Member.ID != "m1" || blocks[0].IsOverride {
		t.Errorf("block 0 wrong: member=%s override=%v", blocks[0].Member.ID, blocks[0].IsOverride)
	}
	if !blocks[0].Start.Equal(periodStart) || !blocks[0].End.Equal(override.Start) {
		t.Errorf("block 0 times wrong: %v – %v", blocks[0].Start, blocks[0].End)
	}

	// Block 1: charlie override Mon→Mon (rest of week 1)
	if blocks[1].Member.ID != "m3" || !blocks[1].IsOverride {
		t.Errorf("block 1 wrong: member=%s override=%v", blocks[1].Member.ID, blocks[1].IsOverride)
	}
	if !blocks[1].Start.Equal(override.Start) || !blocks[1].End.Equal(periodStart.AddDate(0, 0, 7)) {
		t.Errorf("block 1 times wrong: %v – %v", blocks[1].Start, blocks[1].End)
	}

	// Block 2: charlie override Mon→Wed (start of week 2)
	if blocks[2].Member.ID != "m3" || !blocks[2].IsOverride {
		t.Errorf("block 2 wrong: member=%s override=%v", blocks[2].Member.ID, blocks[2].IsOverride)
	}
	if !blocks[2].Start.Equal(periodStart.AddDate(0, 0, 7)) || !blocks[2].End.Equal(override.End) {
		t.Errorf("block 2 times wrong: %v – %v", blocks[2].Start, blocks[2].End)
	}

	// Block 3: bob regular Wed→Mon
	if blocks[3].Member.ID != "m2" || blocks[3].IsOverride {
		t.Errorf("block 3 wrong: member=%s override=%v", blocks[3].Member.ID, blocks[3].IsOverride)
	}
	if !blocks[3].Start.Equal(override.End) || !blocks[3].End.Equal(periodStart.AddDate(0, 0, 14)) {
		t.Errorf("block 3 times wrong: %v – %v", blocks[3].Start, blocks[3].End)
	}
}
