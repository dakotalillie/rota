package domain_test

import (
	"testing"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

func newWeeklyRotation(day, t, tz string, members []domain.Member, current *domain.Member) domain.Rotation {
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
		Members:       members,
		CurrentMember: current,
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
	r.CurrentMember = &alice // CurrentMember set but Members list is empty

	blocks, err := r.Schedule(fixedNow, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected empty slice, got %d blocks", len(blocks))
	}
}

func TestSchedule_NoCurrentMember(t *testing.T) {
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, nil)

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error when CurrentMember is nil")
	}
}

func TestSchedule_CurrentMemberNotInMembers(t *testing.T) {
	outsider := domain.Member{ID: "m99", Order: 99, User: domain.User{Name: "Outsider"}}
	members := []domain.Member{alice, bob}
	r := newWeeklyRotation("Monday", "09:00", "UTC", members, &outsider)

	_, err := r.Schedule(fixedNow, 3)
	if err == nil {
		t.Error("expected error when current member is not in members list")
	}
}

func TestSchedule_NilWeeklyCadence(t *testing.T) {
	r := domain.Rotation{
		Members:       []domain.Member{alice},
		CurrentMember: &alice,
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
