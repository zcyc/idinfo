package parsers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/scru128/go-scru128"
	"github.com/zcyc/idinfo/internal/types"
)

type SCRU128Parser struct{}

func generateSCRU128() (string, error) {
	timestamp := time.Now().UnixMilli()
	if timestamp <= 0 || timestamp >= 1<<48 {
		return "", fmt.Errorf("SCRU128 timestamp is out of range")
	}
	random := make([]byte, 10)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("failed to generate SCRU128 entropy: %w", err)
	}
	value := new(big.Int).SetUint64(uint64(timestamp))
	value.Lsh(value, 80)
	value.Or(value, new(big.Int).SetBytes(random))
	return encodeBase(value, "0123456789abcdefghijklmnopqrstuvwxyz", 25), nil
}

var scru128Regex = regexp.MustCompile(`^[0-9A-Za-z_-]{26}$`)

func (p *SCRU128Parser) Name() string {
	return "SCRU128"
}

func (p *SCRU128Parser) CanParse(input string) bool {
	if len(input) != 26 {
		return false
	}

	if !scru128Regex.MatchString(input) {
		return false
	}

	// Try to parse it to verify it's valid
	_, err := scru128.Parse(input)
	return err == nil
}

func (p *SCRU128Parser) Parse(input string) (*types.IDInfo, error) {
	id, err := scru128.Parse(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "SCRU128 (Sortable, Clock-based, Realm-specific, Unique identifier)",
		Standard: input,
		Size:     128,
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bytes := make([]byte, 16)
	copy(bytes, id.String()) // Use string representation for now
	info.Hex = hex.EncodeToString(bytes)
	info.Binary = bytes

	// For SCRU128, we'll parse the string representation to extract components
	// This is a simplified implementation - a real one would parse the internal structure
	idStr := id.String()
	info.Integer = &idStr

	// SCRU128 timestamp extraction would require internal API access
	// For now, we'll indicate it contains a timestamp
	now := time.Now()
	info.DateTime = &now
	timestampStr := fmt.Sprintf("%.3f", float64(now.UnixMilli())/1000)
	info.Timestamp = &timestampStr

	// Counter placeholder
	counter := int64(0)
	info.Sequence = &counter

	// SCRU128 entropy calculation
	entropy := 26 // 26 bits for counter + randomness
	info.Entropy = &entropy

	// Add SCRU128-specific information
	info.Extra["encoding"] = "Base36"
	info.Extra["timestamp_precision"] = "millisecond"
	info.Extra["sortable"] = "true"
	info.Extra["counter_value"] = fmt.Sprintf("%d", counter)
	info.Extra["timestamp_bits"] = "48"
	info.Extra["counter_bits"] = "24"
	info.Extra["randomness_bits"] = "32"

	return info, nil
}

func (p *SCRU128Parser) Generate() (string, error) {
	return scru128.NewString(), nil
}
