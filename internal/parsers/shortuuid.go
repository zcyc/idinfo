package parsers

import (
	"fmt"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/zcyc/idinfo/internal/types"
)

// ShortUUIDParser handles parsing of ShortUUID format using official SDK
type ShortUUIDParser struct{}

func (p *ShortUUIDParser) Name() string {
	return "ShortUUID"
}

func (p *ShortUUIDParser) CanParse(input string) bool {
	// ShortUUID is typically 22 characters
	if len(input) != 22 {
		return false
	}

	// Try to decode using official SDK
	_, err := shortuuid.DefaultEncoder.Decode(input)
	return err == nil
}

func (p *ShortUUIDParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	// Decode ShortUUID to standard UUID using official SDK
	uuidObj, err := shortuuid.DefaultEncoder.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("invalid ShortUUID format: %v", err)
	}
	uuid := uuidObj.String()

	entropy := 122 // 22 characters * log2(57) ≈ 22 * 5.83 ≈ 128 bits (UUID entropy)

	extra := map[string]string{
		"alphabet":       "Base57 (no ambiguous chars)",
		"length":         "22 characters",
		"format":         "Shortened UUID representation",
		"reversible":     "Yes (to UUID)",
		"original_uuid":  uuid,
		"specification":  "https://github.com/lithammer/shortuuid",
		"url_safe":       "Yes",
		"case_sensitive": "Yes",
	}

	// Convert UUID string to bytes for binary representation
	uuidBytes := []byte(uuid)

	return &types.IDInfo{
		IDType:   "ShortUUID",
		Standard: input,
		Size:     128, // Same as UUID
		Entropy:  &entropy,
		Hex:      fmt.Sprintf("%x", uuidBytes),
		Binary:   uuidBytes,
		Extra:    extra,
	}, nil
}

func (p *ShortUUIDParser) Generate() (string, error) {
	// Generate a ShortUUID using official SDK
	shortUUID := shortuuid.New()
	return shortUUID, nil
}
