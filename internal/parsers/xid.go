package parsers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"

	"github.com/rs/xid"
	"github.com/zcyc/idinfo/internal/types"
)

type XidParser struct{}

var xidRegex = regexp.MustCompile(`^[0-9a-v]{20}$`)

func (p *XidParser) Name() string {
	return "Xid"
}

func (p *XidParser) CanParse(input string) bool {
	if len(input) != 20 {
		return false
	}
	if !xidRegex.MatchString(input) {
		return false
	}

	// Try to parse it
	_, err := xid.FromString(input)
	return err == nil
}

func (p *XidParser) Parse(input string) (*types.IDInfo, error) {
	x, err := xid.FromString(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "Xid (globally unique sortable id)",
		Standard: x.String(),
		Size:     96, // 12 bytes
		Hex:      hex.EncodeToString(x.Bytes()),
		Binary:   x.Bytes(),
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bigInt := new(big.Int)
	bigInt.SetBytes(x.Bytes())
	intStr := bigInt.String()
	info.Integer = &intStr

	// Extract timestamp (first 4 bytes)
	timestamp := x.Time()
	info.DateTime = &timestamp
	timestampStr := fmt.Sprintf("%.3f", float64(timestamp.Unix()))
	info.Timestamp = &timestampStr

	// Xid has limited entropy due to predictable components
	entropy := 56 // 7 bytes: 3 bytes machine + 2 bytes process + 3 bytes counter
	info.Entropy = &entropy

	// Extract components from raw bytes
	bytes := x.Bytes()

	// Machine identifier (bytes 4-6)
	machineId := hex.EncodeToString(bytes[4:7])
	info.Node1 = &machineId

	// Process identifier (bytes 7-8)
	processId := hex.EncodeToString(bytes[7:9])
	info.Node2 = &processId

	// Counter (bytes 9-11)
	counter := int64(bytes[9])<<16 | int64(bytes[10])<<8 | int64(bytes[11])
	info.Sequence = &counter

	// Add Xid-specific information
	info.Extra["encoding"] = "Base32 (Crockford)"
	info.Extra["timestamp_precision"] = "second"
	info.Extra["machine_bytes"] = machineId
	info.Extra["process_bytes"] = processId
	info.Extra["counter_value"] = fmt.Sprintf("%d", counter)
	info.Extra["sortable"] = "true"

	return info, nil
}

func (p *XidParser) Generate() (string, error) {
	x := xid.New()
	return x.String(), nil
}
