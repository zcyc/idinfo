package parsers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/zcyc/idinfo/internal/types"
)

// Base58 alphabet (Bitcoin style - no 0, O, I, l to avoid confusion)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Base58Parser handles parsing of Base58 encoded IDs
type Base58Parser struct{}

func (p *Base58Parser) Name() string {
	return "Base58"
}

func (p *Base58Parser) CanParse(input string) bool {
	// Base58 IDs are typically 8-60 characters long
	if len(input) < 8 || len(input) > 60 {
		return false
	}

	// Check if all characters are in the Base58 alphabet
	validChars := regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)
	if !validChars.MatchString(input) {
		return false
	}

	// Additional heuristic: should not be all numbers or all letters
	allNumbers := regexp.MustCompile(`^[123456789]+$`)
	allLetters := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)

	if allNumbers.MatchString(input) && len(input) < 15 {
		return false // Likely a regular number
	}
	if allLetters.MatchString(input) && len(input) < 10 {
		return false // Likely a word
	}

	return true
}

func (p *Base58Parser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid Base58 format")
	}

	// Decode Base58 to get the original bytes
	decoded, err := p.decodeBase58(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Base58: %v", err)
	}

	// Calculate entropy
	entropy := int(float64(len(input)) * 5.858) // log2(58) ≈ 5.858

	extra := map[string]string{
		"alphabet":     "Base58 (Bitcoin style)",
		"decoded_size": fmt.Sprintf("%d bytes", len(decoded)),
		"encoding":     "Base58",
	}

	// Try to detect if it might be a Bitcoin address or similar
	if len(decoded) == 25 && decoded[0] == 0x00 {
		extra["possible_type"] = "Bitcoin P2PKH Address"
	} else if len(decoded) == 25 && decoded[0] == 0x05 {
		extra["possible_type"] = "Bitcoin P2SH Address"
	} else if len(decoded) >= 32 {
		extra["possible_type"] = "Hash or Key"
	}

	return &types.IDInfo{
		IDType:   "Base58",
		Standard: input,
		Size:     len(decoded) * 8,
		Entropy:  &entropy,
		Hex:      fmt.Sprintf("%x", decoded),
		Binary:   decoded,
		Extra:    extra,
	}, nil
}

func (p *Base58Parser) Generate() (string, error) {
	// Generate 32 random bytes and encode as Base58
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	encoded := p.encodeBase58(randomBytes)
	return encoded, nil
}

// decodeBase58 decodes a Base58 string to bytes
func (p *Base58Parser) decodeBase58(s string) ([]byte, error) {
	// Convert string to big integer
	num := big.NewInt(0)
	base := big.NewInt(58)

	for _, char := range s {
		pos := strings.IndexRune(base58Alphabet, char)
		if pos == -1 {
			return nil, fmt.Errorf("invalid character in Base58 string: %c", char)
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(pos)))
	}

	// Convert to bytes
	decoded := num.Bytes()

	// Handle leading zeros
	for i, char := range s {
		if char != '1' {
			break
		}
		if i == len(s)-1 {
			return []byte{0}, nil
		}
		decoded = append([]byte{0}, decoded...)
	}

	return decoded, nil
}

// encodeBase58 encodes bytes as Base58 string
func (p *Base58Parser) encodeBase58(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Convert bytes to big integer
	num := new(big.Int).SetBytes(data)
	base := big.NewInt(58)
	zero := big.NewInt(0)

	var result []byte
	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		result = append([]byte{base58Alphabet[mod.Int64()]}, result...)
	}

	// Handle leading zeros
	for _, b := range data {
		if b != 0 {
			break
		}
		result = append([]byte{'1'}, result...)
	}

	return string(result)
}
