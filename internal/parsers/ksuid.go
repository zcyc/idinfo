package parsers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"

	"github.com/segmentio/ksuid"
	"github.com/zcyc/idinfo/internal/types"
)

type KSUIDParser struct{}

var ksuidRegex = regexp.MustCompile(`^[0-9A-Za-z]{27}$`)

func (p *KSUIDParser) Name() string {
	return "KSUID"
}

func (p *KSUIDParser) CanParse(input string) bool {
	if len(input) != 27 {
		return false
	}
	if !ksuidRegex.MatchString(input) {
		return false
	}

	// Try to parse it
	_, err := ksuid.Parse(input)
	return err == nil
}

func (p *KSUIDParser) Parse(input string) (*types.IDInfo, error) {
	k, err := ksuid.Parse(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "KSUID (K-Sortable Unique Identifier)",
		Standard: k.String(),
		Size:     160, // 20 bytes
		Hex:      hex.EncodeToString(k.Bytes()),
		Binary:   k.Bytes(),
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bigInt := new(big.Int)
	bigInt.SetBytes(k.Bytes())
	intStr := bigInt.String()
	info.Integer = &intStr

	// Extract timestamp (first 4 bytes represent seconds since KSUID epoch)
	timestamp := k.Time()
	info.DateTime = &timestamp
	timestampStr := fmt.Sprintf("%.3f", float64(timestamp.Unix()))
	info.Timestamp = &timestampStr

	// KSUID has 128 bits of entropy (16 bytes of payload)
	entropy := 128
	info.Entropy = &entropy

	// Add KSUID-specific information
	info.Extra["encoding"] = "Base62"
	info.Extra["timestamp_precision"] = "second"
	info.Extra["epoch"] = "2014-05-13T16:53:20Z"
	info.Extra["sortable"] = "true"
	info.Extra["payload_bytes"] = "16"

	return info, nil
}

func (p *KSUIDParser) Generate() (string, error) {
	k := ksuid.New()
	return k.String(), nil
}
