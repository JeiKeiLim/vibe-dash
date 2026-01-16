package statsview

// SparklineChars are Unicode block elements for sparkline visualization (ascending height)
var SparklineChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline generates a sparkline string from activity counts.
// Input: counts slice (one value per time bucket)
// Output: sparkline string of len(counts) characters
func RenderSparkline(counts []int) string {
	if len(counts) == 0 {
		return ""
	}

	// Find max value for normalization
	max := 0
	for _, c := range counts {
		if c > max {
			max = c
		}
	}

	result := make([]rune, len(counts))
	levels := len(SparklineChars) // 8 levels

	if max == 0 {
		// All zeros - return lowest level
		for i := range result {
			result[i] = SparklineChars[0]
		}
		return string(result)
	}

	// Normalize counts to 0-(levels-1) range
	for i, c := range counts {
		// Scale to 0-(levels-1) range
		level := (c * (levels - 1)) / max
		result[i] = SparklineChars[level]
	}

	return string(result)
}
