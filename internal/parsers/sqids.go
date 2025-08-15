package parsers

import (
	"fmt"
	"strings"

	"github.com/sqids/sqids-go"
	"github.com/zcyc/idinfo/internal/types"
)

// SqidsParser handles parsing of Sqids format (successor to Hashids)
type SqidsParser struct{}

func (p *SqidsParser) Name() string {
	return "Sqids"
}

func (p *SqidsParser) CanParse(input string) bool {
	// Empty strings are invalid
	if len(input) == 0 {
		return false
	}

	// Try to decode with default Sqids configuration to validate
	s, err := sqids.New()
	if err != nil {
		return false
	}

	// If it decodes and returns non-empty result, it's valid
	numbers := s.Decode(input)
	return len(numbers) > 0
}

func (p *SqidsParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid Sqids format")
	}

	// Create default Sqids instance
	s, err := sqids.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Sqids instance: %v", err)
	}

	// Decode the Sqids
	numbers := s.Decode(input)
	if len(numbers) == 0 {
		return nil, fmt.Errorf("failed to decode Sqids: no numbers decoded")
	}

	// Calculate entropy based on default alphabet size and length
	defaultAlphabetSize := 62 // Default Sqids alphabet size
	entropy := int(float64(len(input)) * sqidsLogBase2Float(float64(defaultAlphabetSize)))

	extra := map[string]string{
		"alphabet":       "URL-safe (no profanity)",
		"numbers":        fmt.Sprintf("%v", numbers),
		"reversible":     "Yes (to number array)",
		"format":         "Sqids (Hashids successor)",
		"anti_profanity": "Yes",
		"url_safe":       "Yes",
	}

	// Re-encode to check canonical form
	canonical, err := s.Encode(numbers)
	if err == nil {
		extra["canonical"] = canonical
		if canonical != input {
			extra["is_canonical"] = "No"
		} else {
			extra["is_canonical"] = "Yes"
		}
	} else {
		extra["canonical"] = input
		extra["is_canonical"] = "Yes"
	}

	// Estimate original number range
	if len(numbers) > 0 {
		maxNum := uint64(0)
		for _, num := range numbers {
			if num > maxNum {
				maxNum = num
			}
		}

		if maxNum < 1000 {
			extra["number_range"] = "Small (< 1K)"
		} else if maxNum < 1000000 {
			extra["number_range"] = "Medium (< 1M)"
		} else {
			extra["number_range"] = "Large (>= 1M)"
		}
	}

	return &types.IDInfo{
		IDType:   "Sqids",
		Standard: input,
		Size:     len(input) * 6, // Rough estimation based on ~6 bits per character
		Entropy:  &entropy,
		Hex:      fmt.Sprintf("%x", []byte(input)),
		Binary:   []byte(input),
		Extra:    extra,
	}, nil
}

func (p *SqidsParser) Generate() (string, error) {
	// Create default Sqids instance
	s, err := sqids.New()
	if err != nil {
		return "", fmt.Errorf("failed to create Sqids instance: %v", err)
	}

	// Generate some random numbers to encode
	numbers := []uint64{
		uint64(42),   // Fixed number for demo
		uint64(123),  // Another number
		uint64(7890), // Third number
	}

	// Encode the numbers
	encoded, err := s.Encode(numbers)
	if err != nil {
		return "", fmt.Errorf("failed to encode Sqids: %v", err)
	}

	return encoded, nil
}

// sqidsLogBase2Float calculates log base 2 of a float
func sqidsLogBase2Float(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return 3.321928 * sqidsLogBase10Float(x) // log2(x) = log10(x) / log10(2)
}

// sqidsLogBase10Float calculates log base 10 of a float (simplified)
func sqidsLogBase10Float(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Simplified log10 calculation
	count := 0.0
	for x >= 10 {
		x /= 10
		count++
	}
	return count + (x-1)/2.3 // Rough approximation
}
