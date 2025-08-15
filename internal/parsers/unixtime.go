package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/zcyc/idinfo/internal/types"
)

type UnixTimeParser struct{}

var unixTimeRegex = regexp.MustCompile(`^(0|\d{10,19})$`)

func (p *UnixTimeParser) Name() string {
	return "UnixTime"
}

func (p *UnixTimeParser) CanParse(input string) bool {
	if !unixTimeRegex.MatchString(input) {
		return false
	}

	// Must be a valid integer (use uint64 for large numbers)
	_, err := strconv.ParseUint(input, 10, 64)
	return err == nil
}

func (p *UnixTimeParser) Parse(input string) (*types.IDInfo, error) {
	timestampUint, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}

	// For compatibility with time functions, try to convert to signed
	// For very large values, we may lose precision but still provide info
	timestamp := int64(timestampUint)
	if timestampUint > 9223372036854775807 {
		// For values that don't fit in int64, we'll still process them
		// but the time interpretation may be limited
		timestamp = int64(timestampUint & 0x7FFFFFFFFFFFFFFF)
	}

	// Determine the most likely unit based on the magnitude
	var t time.Time
	var unit string
	var precision string

	now := time.Now()

	// Try different interpretations
	candidates := []struct {
		time      time.Time
		unit      string
		precision string
	}{
		{time.Unix(timestamp, 0), "seconds", "second"},
		{time.UnixMilli(timestamp), "milliseconds", "millisecond"},
		{time.UnixMicro(timestamp), "microseconds", "microsecond"},
		{time.Unix(0, timestamp), "nanoseconds", "nanosecond"},
	}

	// Find the most reasonable interpretation (closest to current time)
	minDiff := int64(1<<63 - 1)
	for _, candidate := range candidates {
		if candidate.time.Year() >= 1970 && candidate.time.Year() <= 2100 {
			diff := absInt64(candidate.time.Unix() - now.Unix())
			if diff < minDiff {
				minDiff = diff
				t = candidate.time
				unit = candidate.unit
				precision = candidate.precision
			}
		}
	}

	// Default to seconds if no good match
	if t.IsZero() {
		t = time.Unix(timestamp, 0)
		unit = "seconds"
		precision = "second"
	}

	info := &types.IDInfo{
		IDType:   fmt.Sprintf("Unix timestamp (%s)", unit),
		Standard: input,
		Size:     64, // Assuming 64-bit timestamp
		Hex:      fmt.Sprintf("%016x", uint64(timestamp)),
		Binary:   []byte(fmt.Sprintf("%064b", timestamp)),
		Extra:    make(map[string]string),
	}

	info.Integer = &input
	info.DateTime = &t
	timestampStr := fmt.Sprintf("%.3f", float64(t.UnixMilli())/1000)
	info.Timestamp = &timestampStr

	// Unix timestamps have no entropy (they're deterministic)
	entropy := 0
	info.Entropy = &entropy

	// Add Unix timestamp-specific information
	info.Extra["unit"] = unit
	info.Extra["precision"] = precision
	info.Extra["epoch"] = "1970-01-01T00:00:00Z"
	info.Extra["deterministic"] = "true"

	return info, nil
}

func (p *UnixTimeParser) Generate() (string, error) {
	return fmt.Sprintf("%d", time.Now().Unix()), nil
}

func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
