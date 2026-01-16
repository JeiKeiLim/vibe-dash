package statsview

import "time"

// BucketActivityCounts calculates activity per time bucket.
// timestamps: transition times (oldest first expected, but handles any order)
// buckets: number of time buckets (e.g., 7 for weekly)
// timeRange: total time period to cover
// now: reference time for bucket calculation (enables testing)
// Returns: slice of counts, one per bucket (oldest first)
func BucketActivityCounts(timestamps []time.Time, buckets int, timeRange time.Duration, now time.Time) []int {
	if buckets <= 0 {
		return nil
	}

	counts := make([]int, buckets)
	if len(timestamps) == 0 {
		return counts // All zeros
	}

	startTime := now.Add(-timeRange)
	bucketDuration := timeRange / time.Duration(buckets)

	for _, ts := range timestamps {
		idx := bucketIndex(ts, startTime, bucketDuration, buckets)
		counts[idx]++
	}

	return counts
}

// bucketIndex calculates which bucket a timestamp belongs to.
// Returns index clamped to [0, totalBuckets-1].
func bucketIndex(timestamp, startTime time.Time, bucketDuration time.Duration, totalBuckets int) int {
	elapsed := timestamp.Sub(startTime)
	idx := int(elapsed / bucketDuration)

	// Clamp to valid range
	if idx >= totalBuckets {
		idx = totalBuckets - 1
	}
	if idx < 0 {
		idx = 0
	}

	return idx
}
