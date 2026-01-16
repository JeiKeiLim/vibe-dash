package statsview

import (
	"testing"
	"time"
)

func TestDefaultDateRange(t *testing.T) {
	dr := DefaultDateRange()
	if dr.Preset != DateRange30Days {
		t.Errorf("DefaultDateRange() = %v, want DateRange30Days", dr.Preset)
	}
}

func TestDateRange_Next_Cycles(t *testing.T) {
	tests := []struct {
		name   string
		start  DateRangePreset
		expect DateRangePreset
	}{
		{"7d to 30d", DateRange7Days, DateRange30Days},
		{"30d to 90d", DateRange30Days, DateRange90Days},
		{"90d to 1y", DateRange90Days, DateRange1Year},
		{"1y to All", DateRange1Year, DateRangeAllTime},
		{"All wraps to 7d", DateRangeAllTime, DateRange7Days},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DateRange{Preset: tt.start}
			got := dr.Next()
			if got.Preset != tt.expect {
				t.Errorf("Next() = %v, want %v", got.Preset, tt.expect)
			}
		})
	}
}

func TestDateRange_Prev_Cycles(t *testing.T) {
	tests := []struct {
		name   string
		start  DateRangePreset
		expect DateRangePreset
	}{
		{"30d to 7d", DateRange30Days, DateRange7Days},
		{"90d to 30d", DateRange90Days, DateRange30Days},
		{"1y to 90d", DateRange1Year, DateRange90Days},
		{"All to 1y", DateRangeAllTime, DateRange1Year},
		{"7d wraps to All", DateRange7Days, DateRangeAllTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DateRange{Preset: tt.start}
			got := dr.Prev()
			if got.Preset != tt.expect {
				t.Errorf("Prev() = %v, want %v", got.Preset, tt.expect)
			}
		})
	}
}

func TestDateRange_Since_Calculations(t *testing.T) {
	// Use a fixed "now" for predictable tests
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		preset   DateRangePreset
		expected time.Time
		isZero   bool
	}{
		{"7 days", DateRange7Days, now.Add(-7 * 24 * time.Hour), false},
		{"30 days", DateRange30Days, now.Add(-30 * 24 * time.Hour), false},
		{"90 days", DateRange90Days, now.Add(-90 * 24 * time.Hour), false},
		{"1 year", DateRange1Year, now.Add(-365 * 24 * time.Hour), false},
		{"All time returns zero", DateRangeAllTime, time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := DateRange{Preset: tt.preset}
			got := dr.SinceFrom(now)

			if tt.isZero {
				if !got.IsZero() {
					t.Errorf("SinceFrom() = %v, want zero time", got)
				}
			} else {
				if !got.Equal(tt.expected) {
					t.Errorf("SinceFrom() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestDateRange_Labels(t *testing.T) {
	tests := []struct {
		preset         DateRangePreset
		label          string
		headerLabel    string
		breakdownLabel string
	}{
		{DateRange7Days, "7d", "Activity (7d)", "Last 7 days"},
		{DateRange30Days, "30d", "Activity (30d)", "Last 30 days"},
		{DateRange90Days, "90d", "Activity (90d)", "Last 90 days"},
		{DateRange1Year, "1y", "Activity (1y)", "Last year"},
		{DateRangeAllTime, "All", "Activity (All)", "All time"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			dr := DateRange{Preset: tt.preset}

			if got := dr.Label(); got != tt.label {
				t.Errorf("Label() = %q, want %q", got, tt.label)
			}
			if got := dr.HeaderLabel(); got != tt.headerLabel {
				t.Errorf("HeaderLabel() = %q, want %q", got, tt.headerLabel)
			}
			if got := dr.BreakdownLabel(); got != tt.breakdownLabel {
				t.Errorf("BreakdownLabel() = %q, want %q", got, tt.breakdownLabel)
			}
		})
	}
}

func TestDateRange_Since_UsesTimeNow(t *testing.T) {
	// Test that Since() returns a time close to time.Now() - duration
	dr := DateRange{Preset: DateRange7Days}
	before := time.Now().Add(-7 * 24 * time.Hour)
	got := dr.Since()
	after := time.Now().Add(-7 * 24 * time.Hour)

	// The returned time should be between before and after (allowing for execution time)
	if got.Before(before.Add(-time.Second)) || got.After(after.Add(time.Second)) {
		t.Errorf("Since() returned unexpected time: got %v, expected between %v and %v", got, before, after)
	}
}

func TestDateRange_AllTime_Since_ReturnsZero(t *testing.T) {
	dr := DateRange{Preset: DateRangeAllTime}
	got := dr.Since()
	if !got.IsZero() {
		t.Errorf("Since() for AllTime = %v, want zero time", got)
	}
}

func TestDateRange_Duration(t *testing.T) {
	tests := []struct {
		preset   DateRangePreset
		expected time.Duration
	}{
		{DateRange7Days, 7 * 24 * time.Hour},
		{DateRange30Days, 30 * 24 * time.Hour},
		{DateRange90Days, 90 * 24 * time.Hour},
		{DateRange1Year, 365 * 24 * time.Hour},
		{DateRangeAllTime, 0}, // Special case: caller handles
	}
	for _, tt := range tests {
		t.Run(tt.expected.String(), func(t *testing.T) {
			dr := DateRange{Preset: tt.preset}
			got := dr.Duration()
			if got != tt.expected {
				t.Errorf("Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDateRange_InvalidPreset_Defaults(t *testing.T) {
	// Test that invalid preset values fall back to 30-day defaults
	dr := DateRange{Preset: DateRangePreset(99)}

	// Since should return 30 days ago
	now := time.Now()
	since := dr.SinceFrom(now)
	expected := now.Add(-30 * 24 * time.Hour)
	if !since.Equal(expected) {
		t.Errorf("Invalid preset SinceFrom() = %v, want %v", since, expected)
	}

	// Label should return "30d"
	if label := dr.Label(); label != "30d" {
		t.Errorf("Invalid preset Label() = %q, want %q", label, "30d")
	}

	// BreakdownLabel should return "Last 30 days"
	if bl := dr.BreakdownLabel(); bl != "Last 30 days" {
		t.Errorf("Invalid preset BreakdownLabel() = %q, want %q", bl, "Last 30 days")
	}

	// Duration should return 30 days
	if d := dr.Duration(); d != 30*24*time.Hour {
		t.Errorf("Invalid preset Duration() = %v, want %v", d, 30*24*time.Hour)
	}
}
