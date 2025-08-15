package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nats-io/nuid"
	"github.com/zcyc/idinfo/internal/types"
)

// NUIDParser handles parsing of NUID (NATS Unique Identifier) format
type NUIDParser struct{}

// NUID uses base62 encoding (0-9, A-Z, a-z) and is 22 characters long
var nuidRegex = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)

func (p *NUIDParser) Name() string {
	return "NUID"
}

func (p *NUIDParser) CanParse(input string) bool {
	// NUID is exactly 22 characters long
	if len(input) != 22 {
		return false
	}

	// Must contain only base62 characters (0-9, A-Z, a-z)
	if !nuidRegex.MatchString(input) {
		return false
	}

	// Additional heuristic: NUID typically doesn't start with 0
	// and has good character distribution
	return input[0] != '0'
}

func (p *NUIDParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid NUID format")
	}

	// NUID structure:
	// - 12 bytes (crypto random prefix) + 10 bytes (sequential counter)
	// - Total 22 characters in base62 encoding
	// - First 12 bytes are crypto-generated, last 10 bytes increment sequentially

	entropy := 132 // 22 characters * log2(62) ≈ 22 * 5.95 ≈ 131 bits

	extra := map[string]string{
		"alphabet":         "Base62 (0-9A-Za-z)",
		"length":           "22 characters",
		"format":           "NATS Unique Identifier",
		"specification":    "https://github.com/nats-io/nuid",
		"crypto_prefix":    "12 bytes crypto random",
		"sequential_part":  "10 bytes sequential",
		"performance":      "~60ns generation, 16M/sec",
		"entropy_friendly": "Yes (minimal crypto/rand usage)",
		"url_safe":         "Yes",
		"case_sensitive":   "Yes",
		"sortable":         "Partially (by prefix)",
	}

	// Convert to bytes for binary representation
	inputBytes := []byte(input)

	return &types.IDInfo{
		IDType:   "NUID (NATS Unique Identifier)",
		Standard: input,
		Size:     132, // ~132 bits of entropy
		Entropy:  &entropy,
		Hex:      fmt.Sprintf("%x", inputBytes),
		Binary:   inputBytes,
		Extra:    extra,
	}, nil
}

func (p *NUIDParser) Generate() (string, error) {
	// Generate a NUID using the official library
	return nuid.Next(), nil
}
