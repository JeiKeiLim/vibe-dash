package statsview

import "time"

// DateRangePreset represents a predefined time period for metrics filtering.
type DateRangePreset int

const (
	DateRange7Days  DateRangePreset = iota
	DateRange30Days                 // Default
	DateRange90Days
	DateRange1Year
	DateRangeAllTime
)

// presetCount is the total number of presets (for cycling).
const presetCount = 5

// DateRange holds the selected preset and provides helper methods.
type DateRange struct {
	Preset DateRangePreset
}

// DefaultDateRange returns the default 30-day range.
func DefaultDateRange() DateRange {
	return DateRange{Preset: DateRange30Days}
}

// Next returns the next preset (wraps around).
func (d DateRange) Next() DateRange {
	next := (int(d.Preset) + 1) % presetCount
	return DateRange{Preset: DateRangePreset(next)}
}

// Prev returns the previous preset (wraps around).
func (d DateRange) Prev() DateRange {
	prev := (int(d.Preset) - 1 + presetCount) % presetCount
	return DateRange{Preset: DateRangePreset(prev)}
}

// Since returns the start time for this range using time.Now().
// For AllTime, returns time.Time{} (zero value = no filter).
func (d DateRange) Since() time.Time {
	return d.SinceFrom(time.Now())
}

// SinceFrom returns the start time for this range from a given reference time.
// For AllTime, returns time.Time{} (zero value = no filter).
func (d DateRange) SinceFrom(now time.Time) time.Time {
	switch d.Preset {
	case DateRange7Days:
		return now.Add(-7 * 24 * time.Hour)
	case DateRange30Days:
		return now.Add(-30 * 24 * time.Hour)
	case DateRange90Days:
		return now.Add(-90 * 24 * time.Hour)
	case DateRange1Year:
		return now.Add(-365 * 24 * time.Hour)
	case DateRangeAllTime:
		return time.Time{} // Zero value = no filter
	default:
		return now.Add(-30 * 24 * time.Hour) // Fallback to 30 days
	}
}

// Label returns human-readable label (e.g., "7d", "30d", "All").
func (d DateRange) Label() string {
	switch d.Preset {
	case DateRange7Days:
		return "7d"
	case DateRange30Days:
		return "30d"
	case DateRange90Days:
		return "90d"
	case DateRange1Year:
		return "1y"
	case DateRangeAllTime:
		return "All"
	default:
		return "30d"
	}
}

// HeaderLabel returns label for column header (e.g., "Activity (7d)").
func (d DateRange) HeaderLabel() string {
	return "Activity (" + d.Label() + ")"
}

// BreakdownLabel returns label for breakdown view (e.g., "Last 7 days").
func (d DateRange) BreakdownLabel() string {
	switch d.Preset {
	case DateRange7Days:
		return "Last 7 days"
	case DateRange30Days:
		return "Last 30 days"
	case DateRange90Days:
		return "Last 90 days"
	case DateRange1Year:
		return "Last year"
	case DateRangeAllTime:
		return "All time"
	default:
		return "Last 30 days"
	}
}

// Duration returns the time.Duration for this preset.
// For AllTime, returns 0 (caller should handle specially).
func (d DateRange) Duration() time.Duration {
	switch d.Preset {
	case DateRange7Days:
		return 7 * 24 * time.Hour
	case DateRange30Days:
		return 30 * 24 * time.Hour
	case DateRange90Days:
		return 90 * 24 * time.Hour
	case DateRange1Year:
		return 365 * 24 * time.Hour
	case DateRangeAllTime:
		return 0 // Special case - caller handles
	default:
		return 30 * 24 * time.Hour
	}
}
