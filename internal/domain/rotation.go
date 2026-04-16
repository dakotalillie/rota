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
	ID              string
	Name            string
	Cadence         RotationCadence
	ScheduledMember *Member
	Members         []Member
	Overrides       []Override
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
	Start      time.Time
	End        time.Time
	Member     *Member
	IsOverride bool
}

func (r *Rotation) Schedule(now time.Time, numWeeks int) ([]ScheduleBlock, error) {
	if r.Cadence.Weekly == nil {
		return nil, errors.New("rotation has no weekly cadence")
	}
	if len(r.Members) == 0 || numWeeks <= 0 {
		return []ScheduleBlock{}, nil
	}
	if r.ScheduledMember == nil {
		return nil, errors.New("rotation has no scheduled member")
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
		return sorted[i].Position < sorted[j].Position
	})

	currentIdx := -1
	for i, m := range sorted {
		if m.ID == r.ScheduledMember.ID {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		return nil, fmt.Errorf("scheduled member %q not found in rotation members", r.ScheduledMember.ID)
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

	result := applyOverrides(blocks, r.Overrides)
	filtered := make([]ScheduleBlock, 0, len(result))
	for _, b := range result {
		if b.End.After(now) {
			filtered = append(filtered, b)
		}
	}

	return filtered, nil
}

func (r *Rotation) NeedsAdvance(now time.Time) (bool, time.Time, error) {
	if r.ScheduledMember == nil || r.Cadence.Weekly == nil || r.ScheduledMember.BecameCurrentAt.IsZero() {
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
	if handoff.After(r.ScheduledMember.BecameCurrentAt) {
		return true, handoff, nil
	}
	return false, time.Time{}, nil
}

func (r *Rotation) NextMember() (*Member, error) {
	if r.ScheduledMember == nil {
		return nil, errors.New("rotation has no scheduled member")
	}
	if len(r.Members) == 0 {
		return nil, errors.New("rotation has no members")
	}

	sorted := make([]Member, len(r.Members))
	copy(sorted, r.Members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	for i, m := range sorted {
		if m.ID == r.ScheduledMember.ID {
			next := sorted[(i+1)%len(sorted)]
			return &next, nil
		}
	}
	return nil, fmt.Errorf("scheduled member %q not found in rotation members", r.ScheduledMember.ID)
}

func (r *Rotation) EffectiveOnCall(now time.Time) *Member {
	for _, o := range r.Overrides {
		if !now.Before(o.Start) && now.Before(o.End) {
			member := o.Member
			return &member
		}
	}

	return r.ScheduledMember
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

// ValidateOverride checks whether it is valid to create an override for the given
// member during [start, end). It returns ErrMemberNotFound if memberID is not in
// this rotation's member list, and ErrOverrideSameMember if the member is already
// the scheduled on-call for any block that overlaps the requested window.
//
// `now` must be the current time: Schedule anchors the currently-scheduled
// member to the week containing its `now` argument, so passing anything other
// than the real clock would phase-shift the computed cycle.
func (r *Rotation) ValidateOverride(memberID string, now, start, end time.Time) error {
	memberInRotation := false
	for _, m := range r.Members {
		if m.ID == memberID {
			memberInRotation = true
			break
		}
	}
	if !memberInRotation {
		return ErrMemberNotFound
	}

	// No active schedule — nothing to conflict with.
	if r.Cadence.Weekly == nil || r.ScheduledMember == nil || len(r.Members) == 0 {
		return nil
	}

	// Generate enough blocks to cover everything from `now` through `end`.
	numWeeks := int(end.Sub(now).Hours()/168) + 2
	if numWeeks < 1 {
		numWeeks = 1
	}
	blocks, err := r.Schedule(now, numWeeks)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if block.End.After(start) && block.Start.Before(end) {
			if block.Member != nil && block.Member.ID == memberID {
				return ErrOverrideSameMember
			}
		}
	}

	return nil
}

func applyOverrides(blocks []ScheduleBlock, overrides []Override) []ScheduleBlock {
	if len(overrides) == 0 {
		return blocks
	}

	sort.Slice(overrides, func(i, j int) bool {
		return overrides[i].Start.Before(overrides[j].Start)
	})

	var result []ScheduleBlock
	for _, block := range blocks {
		cur := block.Start
		for _, o := range overrides {
			if !o.End.After(cur) || !o.Start.Before(block.End) {
				continue
			}

			overStart := o.Start
			if cur.After(overStart) {
				overStart = cur
			}

			if cur.Before(overStart) {
				result = append(result, ScheduleBlock{
					Start:  cur,
					End:    overStart,
					Member: block.Member,
				})
			}

			overEnd := o.End
			if block.End.Before(overEnd) {
				overEnd = block.End
			}

			m := o.Member
			result = append(result, ScheduleBlock{
				Start:      overStart,
				End:        overEnd,
				Member:     &m,
				IsOverride: true,
			})

			cur = overEnd
			if !cur.Before(block.End) {
				break
			}
		}

		if cur.Before(block.End) {
			result = append(result, ScheduleBlock{
				Start:  cur,
				End:    block.End,
				Member: block.Member,
			})
		}
	}
	return result
}
