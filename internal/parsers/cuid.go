package parsers

import (
	"fmt"
	"regexp"

	"github.com/zcyc/idinfo/internal/types"

	"github.com/nrednav/cuid2"
)

type CUIDParser struct{}

// CUID2 uses lowercase letters and digits with variable length
var cuid2Regex = regexp.MustCompile(`^[a-z0-9]+$`)

func (p *CUIDParser) Name() string {
	return "CUID"
}

func (p *CUIDParser) CanParse(input string) bool {
	// CUID2 typically generates IDs between 4-32 characters
	if len(input) < 4 || len(input) > 32 {
		return false
	}

	// Must contain only lowercase letters and digits
	if !cuid2Regex.MatchString(input) {
		return false
	}

	// Use the official library to validate
	return cuid2.IsCuid(input)
}

func (p *CUIDParser) Parse(input string) (*types.IDInfo, error) {
	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid CUID format")
	}

	info := &types.IDInfo{
		IDType:   "CUID v2 (Collision-resistant Unique Identifier)",
		Standard: input,
		Size:     len(input) * 6, // ~5.2 bits per character in practice
		Extra:    make(map[string]string),
	}

	// Convert to hex representation
	hexStr := ""
	for _, char := range input {
		if char >= '0' && char <= '9' {
			hexStr += fmt.Sprintf("%x", char-'0')
		} else if char >= 'a' && char <= 'z' {
			hexStr += fmt.Sprintf("%x", char-'a'+10)
		}
	}
	info.Hex = hexStr

	// Convert hex to binary - only if valid hex length
	if len(hexStr)%2 == 0 {
		if hexBytes := make([]byte, len(hexStr)/2); len(hexBytes) > 0 {
			for i := 0; i < len(hexStr); i += 2 {
				var byteVal byte
				if hexStr[i] >= '0' && hexStr[i] <= '9' {
					byteVal = (hexStr[i] - '0') << 4
				} else {
					byteVal = (hexStr[i] - 'a' + 10) << 4
				}
				if hexStr[i+1] >= '0' && hexStr[i+1] <= '9' {
					byteVal |= hexStr[i+1] - '0'
				} else {
					byteVal |= hexStr[i+1] - 'a' + 10
				}
				hexBytes[i/2] = byteVal
			}
			info.Binary = hexBytes
		}
	}

	// CUID2 has high entropy due to cryptographically secure random generation
	entropy := int(float64(len(input)) * 5.2) // More accurate bits per character
	info.Entropy = &entropy

	// Add CUID2-specific information
	info.Extra["version"] = "2"
	info.Extra["encoding"] = "Base36 (lowercase)"
	info.Extra["collision_resistant"] = "Yes"
	info.Extra["cryptographically_secure"] = "Yes"
	info.Extra["url_safe"] = "Yes"
	info.Extra["length"] = fmt.Sprintf("%d", len(input))
	info.Extra["alphabet"] = "abcdefghijklmnopqrstuvwxyz0123456789"
	info.Extra["alphabet_size"] = "36"

	return info, nil
}

func (p *CUIDParser) Generate() (string, error) {
	// Use the official CUID2 library to generate an ID
	return cuid2.Generate(), nil
}
