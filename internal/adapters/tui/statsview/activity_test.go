package statsview

import (
	"testing"
	"time"
)

func TestBucketActivityCounts_NoTimestamps(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	result := BucketActivityCounts(nil, 7, 30*24*time.Hour, now)

	if len(result) != 7 {
		t.Errorf("expected 7 buckets, got %d", len(result))
	}

	for i, c := range result {
		if c != 0 {
			t.Errorf("bucket %d: expected 0, got %d", i, c)
		}
	}

	// Also test with empty slice
	result = BucketActivityCounts([]time.Time{}, 7, 30*24*time.Hour, now)
	if len(result) != 7 {
		t.Errorf("expected 7 buckets for empty slice, got %d", len(result))
	}
}

func TestBucketActivityCounts_EvenDistribution(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timeRange := 14 * 24 * time.Hour // 14 days
	buckets := 7                     // 2 days per bucket

	// Create one event per bucket (at the midpoint of each bucket)
	timestamps := make([]time.Time, 7)
	bucketDuration := timeRange / time.Duration(buckets) // 2 days

	for i := 0; i < 7; i++ {
		// Place timestamp in each bucket (oldest first)
		// Bucket 0 = oldest (start of range)
		offset := time.Duration(i)*bucketDuration + bucketDuration/2
		timestamps[i] = now.Add(-timeRange + offset)
	}

	result := BucketActivityCounts(timestamps, buckets, timeRange, now)

	if len(result) != 7 {
		t.Fatalf("expected 7 buckets, got %d", len(result))
	}

	// Each bucket should have exactly 1 event
	for i, c := range result {
		if c != 1 {
			t.Errorf("bucket %d: expected 1, got %d", i, c)
		}
	}
}

func TestBucketActivityCounts_AllInOneBucket(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timeRange := 7 * 24 * time.Hour
	buckets := 7

	// All events in the most recent bucket (last 24 hours)
	timestamps := []time.Time{
		now.Add(-1 * time.Hour),
		now.Add(-2 * time.Hour),
		now.Add(-3 * time.Hour),
		now.Add(-12 * time.Hour),
		now.Add(-23 * time.Hour),
	}

	result := BucketActivityCounts(timestamps, buckets, timeRange, now)

	// All 5 events should be in the last bucket (most recent)
	lastBucket := result[buckets-1]
	if lastBucket != 5 {
		t.Errorf("expected 5 events in last bucket, got %d", lastBucket)
	}

	// All other buckets should be empty
	for i := 0; i < buckets-1; i++ {
		if result[i] != 0 {
			t.Errorf("bucket %d: expected 0, got %d", i, result[i])
		}
	}
}

func TestBucketActivityCounts_TimestampsOutOfRange(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timeRange := 7 * 24 * time.Hour
	buckets := 7

	// Timestamps before the time range
	oldTimestamp := now.Add(-30 * 24 * time.Hour) // 30 days ago
	timestamps := []time.Time{oldTimestamp}

	result := BucketActivityCounts(timestamps, buckets, timeRange, now)

	// Old timestamp should be clamped to bucket 0
	if result[0] != 1 {
		t.Errorf("expected old timestamp in bucket 0, got bucket[0]=%d", result[0])
	}

	// Future timestamp (shouldn't happen but handle gracefully)
	futureTimestamp := now.Add(24 * time.Hour)
	timestamps = []time.Time{futureTimestamp}

	result = BucketActivityCounts(timestamps, buckets, timeRange, now)

	// Future timestamp should be clamped to last bucket
	if result[buckets-1] != 1 {
		t.Errorf("expected future timestamp in last bucket, got bucket[%d]=%d",
			buckets-1, result[buckets-1])
	}
}

func TestBucketActivityCounts_ZeroBuckets(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timestamps := []time.Time{now.Add(-1 * time.Hour)}

	result := BucketActivityCounts(timestamps, 0, 7*24*time.Hour, now)

	// Should return nil or empty for invalid input
	if result != nil {
		t.Errorf("expected nil for 0 buckets, got %v", result)
	}
}

func TestBucketActivityCounts_UnsortedTimestamps(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timeRange := 14 * 24 * time.Hour
	buckets := 7

	// Timestamps in random order
	timestamps := []time.Time{
		now.Add(-3 * 24 * time.Hour),  // recent
		now.Add(-10 * 24 * time.Hour), // older
		now.Add(-1 * 24 * time.Hour),  // most recent
		now.Add(-7 * 24 * time.Hour),  // middle
	}

	result := BucketActivityCounts(timestamps, buckets, timeRange, now)

	// Total events should equal number of timestamps
	total := 0
	for _, c := range result {
		total += c
	}
	if total != 4 {
		t.Errorf("expected 4 total events, got %d", total)
	}
}

func TestBucketActivityCounts_SingleBucket(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	timeRange := 30 * 24 * time.Hour

	timestamps := []time.Time{
		now.Add(-1 * 24 * time.Hour),
		now.Add(-15 * 24 * time.Hour),
		now.Add(-29 * 24 * time.Hour),
	}

	result := BucketActivityCounts(timestamps, 1, timeRange, now)

	if len(result) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result))
	}
	if result[0] != 3 {
		t.Errorf("expected 3 events in single bucket, got %d", result[0])
	}
}

// Story 16.6: Tests for CalculateTimeRangeFromTimestamps

func TestCalculateTimeRangeFromTimestamps_EmptySlice(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)
	result := CalculateTimeRangeFromTimestamps(nil, now)

	// Should return 1 day as default
	if result != 24*time.Hour {
		t.Errorf("Expected 1 day for nil timestamps, got %v", result)
	}

	result = CalculateTimeRangeFromTimestamps([]time.Time{}, now)
	if result != 24*time.Hour {
		t.Errorf("Expected 1 day for empty timestamps, got %v", result)
	}
}

func TestCalculateTimeRangeFromTimestamps_FindsEarliest(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	// Timestamps in random order
	timestamps := []time.Time{
		now.Add(-10 * 24 * time.Hour), // 10 days ago
		now.Add(-5 * 24 * time.Hour),  // 5 days ago
		now.Add(-30 * 24 * time.Hour), // 30 days ago (earliest)
		now.Add(-1 * 24 * time.Hour),  // 1 day ago
	}

	result := CalculateTimeRangeFromTimestamps(timestamps, now)

	// Should be 30 days (from earliest to now)
	expected := 30 * 24 * time.Hour
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCalculateTimeRangeFromTimestamps_MinimumOneDay(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	// Timestamps very close to now
	timestamps := []time.Time{
		now.Add(-1 * time.Hour),
		now.Add(-30 * time.Minute),
	}

	result := CalculateTimeRangeFromTimestamps(timestamps, now)

	// Should return minimum 1 day
	if result != 24*time.Hour {
		t.Errorf("Expected minimum 1 day, got %v", result)
	}
}

func TestCalculateTimeRangeFromTimestamps_SingleTimestamp(t *testing.T) {
	now := time.Date(2026, 1, 16, 12, 0, 0, 0, time.UTC)

	timestamps := []time.Time{
		now.Add(-15 * 24 * time.Hour), // 15 days ago
	}

	result := CalculateTimeRangeFromTimestamps(timestamps, now)

	expected := 15 * 24 * time.Hour
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
