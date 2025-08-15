package parsers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/zcyc/idinfo/internal/types"
)

// PushID alphabet used by Firebase
const pushIDAlphabet = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

// PushIDParser handles parsing of Firebase PushID format
type PushIDParser struct{}

func (p *PushIDParser) Name() string {
	return "PushID"
}

func (p *PushIDParser) CanParse(input string) bool {
	// PushIDs are exactly 20 characters long
	if len(input) != 20 {
		return false
	}

	// Check if all characters are in the PushID alphabet
	validChars := regexp.MustCompile(`^[-0-9A-Z_a-z]+$`)
	if !validChars.MatchString(input) {
		return false
	}

	// Additional heuristic: check character distribution
	// PushIDs start with timestamp-based characters, so first 8 chars tend to have patterns
	return true
}

func (p *PushIDParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid PushID format")
	}

	// Extract timestamp from first 8 characters
	timestamp, err := p.extractTimestamp(input[:8])
	var datetime *time.Time
	var timestampStr *string

	if err == nil {
		dt := time.Unix(timestamp/1000, (timestamp%1000)*1000000)
		datetime = &dt
		ts := fmt.Sprintf("%d", timestamp)
		timestampStr = &ts
	}

	// Calculate entropy - PushIDs have high entropy in the random part
	entropy := 120 // 20 characters * 6 bits per character (64-character alphabet)

	extra := map[string]string{
		"alphabet": "Firebase PushID (64 chars)",
		"length":   "20 characters",
		"format":   "8 chars timestamp + 12 chars random",
	}

	if datetime != nil {
		extra["timestamp_part"] = input[:8]
		extra["random_part"] = input[8:]
	}

	return &types.IDInfo{
		IDType:    "Firebase PushID",
		Standard:  input,
		Size:      120, // 20 characters * 6 bits
		Entropy:   &entropy,
		DateTime:  datetime,
		Timestamp: timestampStr,
		Hex:       fmt.Sprintf("%x", []byte(input)),
		Binary:    []byte(input),
		Extra:     extra,
	}, nil
}

func (p *PushIDParser) Generate() (string, error) {
	// Current timestamp in milliseconds
	now := time.Now().UnixNano() / 1000000

	// Encode timestamp to first 8 characters
	timestampPart := p.encodeTimestamp(now)
	if len(timestampPart) > 8 {
		timestampPart = timestampPart[len(timestampPart)-8:] // Take last 8 chars
	}
	for len(timestampPart) < 8 {
		timestampPart = string(pushIDAlphabet[1]) + timestampPart // Pad with '0' instead of '-'
	}

	// Ensure first character is not '-' to avoid command line issues
	if timestampPart[0] == '-' {
		timestampPart = "0" + timestampPart[1:]
	}

	// Generate 12 random characters
	randomPart := make([]byte, 12)
	for i := 0; i < 12; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pushIDAlphabet))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %v", err)
		}
		randomPart[i] = pushIDAlphabet[n.Int64()]
	}

	return timestampPart + string(randomPart), nil
}

// extractTimestamp attempts to decode timestamp from PushID prefix
func (p *PushIDParser) extractTimestamp(timestampPart string) (int64, error) {
	// Convert base64-like encoding back to timestamp
	var timestamp int64 = 0

	for _, char := range timestampPart {
		pos := strings.IndexRune(pushIDAlphabet, char)
		if pos == -1 {
			return 0, fmt.Errorf("invalid character in timestamp part")
		}
		timestamp = timestamp*64 + int64(pos)
	}

	// Validate that timestamp is reasonable (after 2000, before 2100)
	year2000 := int64(946684800000)  // 2000-01-01 in milliseconds
	year2100 := int64(4102444800000) // 2100-01-01 in milliseconds

	if timestamp < year2000 || timestamp > year2100 {
		return 0, fmt.Errorf("timestamp out of reasonable range")
	}

	return timestamp, nil
}

// encodeTimestamp encodes timestamp to PushID format
func (p *PushIDParser) encodeTimestamp(timestamp int64) string {
	var result []byte

	for timestamp > 0 {
		result = append([]byte{pushIDAlphabet[timestamp%64]}, result...)
		timestamp /= 64
	}

	return string(result)
}
