package domain

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Rotation struct {
	ID            string
	Name          string
	Cadence       RotationCadence
	CurrentMember *Member
	Members       []Member
}

type RotationCadence struct {
	Weekly *RotationCadenceWeekly
}

type RotationCadenceWeekly struct {
	Day      string
	Time     string
	TimeZone string
}

type ScheduleBlock struct {
	Start  time.Time
	End    time.Time
	Member *Member
}

func (r *Rotation) Schedule(now time.Time, numWeeks int) ([]ScheduleBlock, error) {
	if r.Cadence.Weekly == nil {
		return nil, errors.New("rotation has no weekly cadence")
	}
	if len(r.Members) == 0 || numWeeks <= 0 {
		return []ScheduleBlock{}, nil
	}
	if r.CurrentMember == nil {
		return nil, errors.New("rotation has no current member")
	}

	loc, err := time.LoadLocation(r.Cadence.Weekly.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", r.Cadence.Weekly.TimeZone, err)
	}

	weekday, err := parseWeekday(r.Cadence.Weekly.Day)
	if err != nil {
		return nil, err
	}

	hour, minute, err := parseHandoffTime(r.Cadence.Weekly.Time)
	if err != nil {
		return nil, err
	}

	sorted := make([]Member, len(r.Members))
	copy(sorted, r.Members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order < sorted[j].Order
	})

	currentIdx := -1
	for i, m := range sorted {
		if m.ID == r.CurrentMember.ID {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		return nil, fmt.Errorf("current member %q not found in rotation members", r.CurrentMember.ID)
	}

	periodStart := mostRecentHandoff(now, weekday, hour, minute, loc)

	blocks := make([]ScheduleBlock, numWeeks)
	for i := range numWeeks {
		start := periodStart.AddDate(0, 0, 7*i)
		end := periodStart.AddDate(0, 0, 7*(i+1))
		member := sorted[(currentIdx+i)%len(sorted)]
		blocks[i] = ScheduleBlock{
			Start:  start,
			End:    end,
			Member: &member,
		}
	}

	return blocks, nil
}

func (r *Rotation) NeedsAdvance(now time.Time) (bool, time.Time, error) {
	if r.CurrentMember == nil || r.Cadence.Weekly == nil || r.CurrentMember.BecameCurrentAt.IsZero() {
		return false, time.Time{}, nil
	}

	loc, err := time.LoadLocation(r.Cadence.Weekly.TimeZone)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("invalid timezone %q: %w", r.Cadence.Weekly.TimeZone, err)
	}

	weekday, err := parseWeekday(r.Cadence.Weekly.Day)
	if err != nil {
		return false, time.Time{}, err
	}

	hour, minute, err := parseHandoffTime(r.Cadence.Weekly.Time)
	if err != nil {
		return false, time.Time{}, err
	}

	handoff := mostRecentHandoff(now, weekday, hour, minute, loc)
	if handoff.After(r.CurrentMember.BecameCurrentAt) {
		return true, handoff, nil
	}
	return false, time.Time{}, nil
}

func (r *Rotation) NextMember() (*Member, error) {
	if r.CurrentMember == nil {
		return nil, errors.New("rotation has no current member")
	}
	if len(r.Members) == 0 {
		return nil, errors.New("rotation has no members")
	}

	sorted := make([]Member, len(r.Members))
	copy(sorted, r.Members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order < sorted[j].Order
	})

	for i, m := range sorted {
		if m.ID == r.CurrentMember.ID {
			next := sorted[(i+1)%len(sorted)]
			return &next, nil
		}
	}
	return nil, fmt.Errorf("current member %q not found in rotation members", r.CurrentMember.ID)
}

func mostRecentHandoff(now time.Time, weekday time.Weekday, hour, minute int, loc *time.Location) time.Time {
	nowInLoc := now.In(loc)
	y, m, d := nowInLoc.Date()
	// Candidate: the handoff time on the most recent occurrence of weekday
	daysBack := (int(nowInLoc.Weekday()) - int(weekday) + 7) % 7
	candidate := time.Date(y, m, d-daysBack, hour, minute, 0, 0, loc)
	// If candidate is still in the future (same weekday, but time hasn't passed yet), go back 7 days
	if candidate.After(nowInLoc) {
		candidate = candidate.AddDate(0, 0, -7)
	}
	return candidate
}

func parseWeekday(s string) (time.Weekday, error) {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	w, ok := days[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid weekday %q", s)
	}
	return w, nil
}

func parseHandoffTime(s string) (hour, minute int, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time %q: expected HH:MM", s)
	}
	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid time %q: bad hour", s)
	}
	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid time %q: bad minute", s)
	}
	return hour, minute, nil
}
