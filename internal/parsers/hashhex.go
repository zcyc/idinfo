package parsers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/zcyc/idinfo/internal/types"
)

type HashHexParser struct{}

var hashHexRegex = regexp.MustCompile(`^[0-9a-fA-F]+$`)

// Common hash lengths in hex characters
var commonHashLengths = map[int]string{
	32:  "MD5",
	40:  "SHA-1",
	56:  "SHA-224",
	64:  "SHA-256",
	96:  "SHA-384",
	128: "SHA-512",
}

func (p *HashHexParser) Name() string {
	return "HashHex"
}

func (p *HashHexParser) CanParse(input string) bool {
	// Must be hex string
	if !hashHexRegex.MatchString(input) {
		return false
	}

	// Should be at least 8 hex characters (32 bits)
	if len(input) < 8 {
		return false
	}

	// Should be even length (full bytes)
	if len(input)%2 != 0 {
		return false
	}

	return true
}

func (p *HashHexParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.ToLower(input)

	// Try to decode hex
	bytes, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}

	// Determine hash type based on length
	hashType := "Unknown Hash"
	if knownType, exists := commonHashLengths[len(input)]; exists {
		hashType = knownType
	} else {
		hashType = fmt.Sprintf("Hash (%d bits)", len(bytes)*8)
	}

	info := &types.IDInfo{
		IDType:   fmt.Sprintf("Hex-encoded %s", hashType),
		Standard: strings.ToUpper(input),
		Size:     len(bytes) * 8,
		Hex:      input,
		Binary:   bytes,
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bigInt := new(big.Int)
	bigInt.SetBytes(bytes)
	intStr := bigInt.String()
	info.Integer = &intStr

	// Hash has full entropy (assuming it's a proper hash)
	entropy := len(bytes) * 8
	info.Entropy = &entropy

	// Add hash-specific information
	info.Extra["encoding"] = "hexadecimal"
	info.Extra["byte_length"] = fmt.Sprintf("%d", len(bytes))
	info.Extra["deterministic"] = "depends on hash function"

	if knownType, exists := commonHashLengths[len(input)]; exists {
		info.Extra["probable_algorithm"] = knownType

		// Add algorithm-specific information
		switch knownType {
		case "MD5":
			info.Extra["cryptographic_strength"] = "broken (collisions found)"
			info.Extra["recommended_use"] = "checksums only"
		case "SHA-1":
			info.Extra["cryptographic_strength"] = "weak (collisions found)"
			info.Extra["recommended_use"] = "deprecated for security"
		case "SHA-256", "SHA-224":
			info.Extra["cryptographic_strength"] = "strong"
			info.Extra["recommended_use"] = "cryptographic applications"
		case "SHA-384", "SHA-512":
			info.Extra["cryptographic_strength"] = "very strong"
			info.Extra["recommended_use"] = "high-security applications"
		}
	}

	return info, nil
}

func (p *HashHexParser) Generate() (string, error) {
	// Generate a random 32-byte (256-bit) hex string
	bytes := make([]byte, 32)
	for i := range bytes {
		bytes[i] = byte(i * 7 % 256) // Simple deterministic pattern for demo
	}
	return hex.EncodeToString(bytes), nil
}
