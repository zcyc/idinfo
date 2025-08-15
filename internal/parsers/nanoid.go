package parsers

import (
	"encoding/hex"
	"fmt"
	"math"
	"regexp"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/zcyc/idinfo/internal/types"
)

type NanoIDParser struct{}

// NanoID default alphabet
const nanoIDAlphabet = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var nanoIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func (p *NanoIDParser) Name() string {
	return "NanoID"
}

func (p *NanoIDParser) CanParse(input string) bool {
	// NanoID typical length is 21, but can vary
	if len(input) < 6 || len(input) > 255 {
		return false
	}

	// Check if all characters are from the NanoID alphabet
	if !nanoIDRegex.MatchString(input) {
		return false
	}

	// Check if characters are from the default alphabet
	for _, char := range input {
		if !containsChar(nanoIDAlphabet, char) {
			return false
		}
	}

	return true
}

func (p *NanoIDParser) Parse(input string) (*types.IDInfo, error) {
	// Validate input first
	if !p.CanParse(input) {
		return nil, fmt.Errorf("invalid NanoID format: %s", input)
	}

	info := &types.IDInfo{
		IDType:   "Nano ID",
		Standard: input,
		Size:     len(input) * 8, // Approximate bit size
		Extra:    make(map[string]string),
	}

	// Convert to hex representation (approximation)
	hexStr := ""
	for _, char := range input {
		index := indexOfChar(nanoIDAlphabet, char)
		if index == -1 {
			return nil, fmt.Errorf("invalid character in NanoID: %c", char)
		}
		hexStr += fmt.Sprintf("%02x", index)
	}
	info.Hex = hexStr

	// Convert hex to binary
	if hexBytes, err := hex.DecodeString(hexStr); err == nil {
		info.Binary = hexBytes
	}

	// Calculate entropy
	alphabetSize := len(nanoIDAlphabet)
	entropy := int(math.Ceil(float64(len(input)) * math.Log2(float64(alphabetSize))))
	info.Entropy = &entropy

	// Add NanoID-specific information
	info.Extra["alphabet"] = nanoIDAlphabet
	info.Extra["alphabet_size"] = fmt.Sprintf("%d", alphabetSize)
	info.Extra["length"] = fmt.Sprintf("%d", len(input))
	info.Extra["url_safe"] = "true"
	info.Extra["collision_resistant"] = "true"

	// Calculate collision probability approximation
	if len(input) == 21 {
		info.Extra["collision_probability"] = "~1% in 4 years (1 ID/hour)"
	}

	return info, nil
}

func (p *NanoIDParser) Generate() (string, error) {
	return gonanoid.New()
}

// Helper functions
func containsChar(str string, char rune) bool {
	for _, c := range str {
		if c == char {
			return true
		}
	}
	return false
}

func indexOfChar(str string, char rune) int {
	for i, c := range str {
		if c == char {
			return i
		}
	}
	return -1
}
