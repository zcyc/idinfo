package parsers

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"regexp"
	"strings"

	"github.com/zcyc/idinfo/internal/types"
)

// Base32Parser handles parsing of Base32 encoded IDs
type Base32Parser struct{}

func (p *Base32Parser) Name() string {
	return "Base32"
}

func (p *Base32Parser) CanParse(input string) bool {
	// Base32 IDs are typically 8-64 characters long and may have padding
	if len(input) < 8 || len(input) > 64 {
		return false
	}

	// Check if all characters are valid Base32
	validChars := regexp.MustCompile(`^[A-Z2-7]+=*$`)
	if !validChars.MatchString(strings.ToUpper(input)) {
		return false
	}

	// Length should be multiple of 8 (with padding) or follow Base32 rules
	upperInput := strings.ToUpper(input)

	// Try to decode to validate
	_, err := base32.StdEncoding.DecodeString(upperInput)
	if err != nil {
		// Try without padding
		_, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(upperInput, "="))
		if err != nil {
			return false
		}
	}

	return true
}

func (p *Base32Parser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(strings.ToUpper(input))

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid Base32 format")
	}

	// Try to decode
	decoded, err := base32.StdEncoding.DecodeString(input)
	if err != nil {
		// Try without padding
		decoded, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(input, "="))
		if err != nil {
			return nil, fmt.Errorf("failed to decode Base32: %v", err)
		}
	}

	// Calculate entropy
	entropy := int(float64(len(strings.TrimRight(input, "="))) * 5.0) // log2(32) = 5

	extra := map[string]string{
		"alphabet":     "Base32 (A-Z, 2-7)",
		"decoded_size": fmt.Sprintf("%d bytes", len(decoded)),
		"padding":      fmt.Sprintf("%d characters", strings.Count(input, "=")),
		"encoding":     "RFC 4648 Base32",
	}

	// Try to detect possible content types based on decoded size
	switch len(decoded) {
	case 16:
		extra["possible_type"] = "128-bit identifier (UUID size)"
	case 20:
		extra["possible_type"] = "160-bit hash (SHA-1 size)"
	case 32:
		extra["possible_type"] = "256-bit hash (SHA-256 size)"
	case 48:
		extra["possible_type"] = "384-bit hash (SHA-384 size)"
	case 64:
		extra["possible_type"] = "512-bit hash (SHA-512 size)"
	default:
		if len(decoded) >= 8 && len(decoded) <= 12 {
			extra["possible_type"] = "Short identifier"
		}
	}

	return &types.IDInfo{
		IDType:   "Base32",
		Standard: input,
		Size:     len(decoded) * 8,
		Entropy:  &entropy,
		Hex:      fmt.Sprintf("%x", decoded),
		Binary:   decoded,
		Extra:    extra,
	}, nil
}

func (p *Base32Parser) Generate() (string, error) {
	// Generate 20 random bytes (common for identifiers)
	randomBytes := make([]byte, 20)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Encode as Base32
	encoded := base32.StdEncoding.EncodeToString(randomBytes)
	return encoded, nil
}
