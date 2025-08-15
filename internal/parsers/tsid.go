package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rushysloth/go-tsid"
	"github.com/zcyc/idinfo/internal/types"
)

type TSIDParser struct{}

// TSID is a 13-character string using Crockford Base32
// Valid chars: 0-9, A-H, J-K, M-N, P-T, V-Z (case insensitive) + ambiguous I,L,O
var tsidRegex = regexp.MustCompile(`^[0-9A-HJ-KM-NP-TV-ZILOa-hj-km-np-tv-zilo]{13}$`)

func (p *TSIDParser) Name() string {
	return "TSID"
}

func (p *TSIDParser) CanParse(input string) bool {
	input = strings.TrimSpace(input)

	// TSID should be exactly 13 characters
	if len(input) != 13 {
		return false
	}

	// Check against Crockford Base32 alphabet (case-insensitive)
	return tsidRegex.MatchString(strings.ToUpper(input))
}

func (p *TSIDParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid TSID format")
	}

	// Parse TSID using official library
	parsedTsid := tsid.FromString(input)
	if parsedTsid == nil {
		return nil, fmt.Errorf("failed to parse TSID: invalid format")
	}

	// Get the number representation
	number := parsedTsid.ToNumber()

	// Extract timestamp (Unix milliseconds)
	unixMillis := parsedTsid.GetUnixMillis()
	timestamp := time.UnixMilli(unixMillis)
	timestampStr := fmt.Sprintf("%.3f", float64(unixMillis)/1000)

	// Convert to binary representation
	binary := make([]byte, 8)
	for i := 0; i < 8; i++ {
		binary[7-i] = byte(number >> (i * 8))
	}

	// Integer representation as string
	intStr := strconv.FormatInt(number, 10)

	// Extract components based on TSID specification
	// Time component: 42 bits (shifted right by 22)
	timeComponent := number >> 22

	// Random component: 22 bits (masked)
	randomComponent := number & 0x3FFFFF

	// The random component contains node + counter
	// We'll show it as a single random value since the official library
	// doesn't expose node/counter breakdown in parsing
	randomStr := fmt.Sprintf("%d", randomComponent)

	// TSID entropy calculation: 22 bits of randomness
	entropy := 22

	extra := map[string]string{
		"encoding":            "Crockford Base32",
		"length":              "13 characters",
		"format":              "Time-Sorted Unique Identifier",
		"specification":       "https://github.com/rushysloth/go-tsid",
		"timestamp_precision": "millisecond",
		"epoch":               "2020-01-01T00:00:00Z (default)",
		"sortable":            "Yes (by generation time)",
		"structure":           "42-bit timestamp + 22-bit random",
		"timestamp_bits":      "42",
		"random_bits":         "22",
		"random_value":        randomStr,
		"time_component":      fmt.Sprintf("%d", timeComponent),
		"url_safe":            "Yes",
		"case_insensitive":    "Yes",
		"collision_resistant": "Up to 4M IDs per millisecond",
		"storage_efficiency":  "64-bit integer or 13-char string",
	}

	return &types.IDInfo{
		IDType:    "TSID (Time-Sorted Unique Identifier)",
		Standard:  strings.ToUpper(input), // Normalize to uppercase
		Size:      64,                     // 64-bit identifier
		Entropy:   &entropy,
		Hex:       fmt.Sprintf("%016x", number),
		Binary:    binary,
		Integer:   &intStr,
		DateTime:  &timestamp,
		Timestamp: &timestampStr,
		Node1:     &randomStr, // Random component (includes node + counter)
		Extra:     extra,
	}, nil
}

func (p *TSIDParser) Generate() (string, error) {
	// Generate a TSID using the official library's fast method
	generatedTsid := tsid.Fast()
	return generatedTsid.ToString(), nil
}
