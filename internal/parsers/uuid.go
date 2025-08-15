package parsers

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zcyc/idinfo/internal/types"
)

type UUIDParser struct{}

func (p *UUIDParser) Name() string {
	return "UUID"
}

func (p *UUIDParser) CanParse(input string) bool {
	// Remove hyphens and check if it's a valid hex string of correct length
	cleaned := strings.ReplaceAll(input, "-", "")
	if len(cleaned) != 32 {
		return false
	}

	// Check if it's valid hex
	_, err := hex.DecodeString(cleaned)
	if err != nil {
		return false
	}

	// Try to parse as UUID
	_, err = uuid.Parse(input)
	return err == nil
}

func (p *UUIDParser) Parse(input string) (*types.IDInfo, error) {
	u, err := uuid.Parse(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "UUID (RFC-9562)",
		Standard: u.String(),
		Size:     128,
		Hex:      hex.EncodeToString(u[:]),
		Binary:   u[:],
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bigInt := new(big.Int)
	bigInt.SetBytes(u[:])
	intStr := bigInt.String()
	info.Integer = &intStr

	// Generate Base64 representation
	base64Str := base64.StdEncoding.EncodeToString(u[:])
	info.Base64 = &base64Str

	// Determine version and extract version-specific information
	version := u.Version()
	variant := u.Variant()

	switch version {
	case 1:
		info.Version = "1 (timestamp and MAC address)"
		timestamp := int64(u.Time())
		clockSeq := u.ClockSequence()
		node := u.NodeID()
		t := time.Unix(0, timestamp)
		info.DateTime = &t
		timestampStr := fmt.Sprintf("%.3f", float64(timestamp)/1e9)
		info.Timestamp = &timestampStr

		seq := int64(clockSeq)
		info.Sequence = &seq
		nodeStr := hex.EncodeToString(node[:])
		info.Node1 = &nodeStr
		entropy := 14 // 14 bits for clock sequence
		info.Entropy = &entropy

	case 2:
		info.Version = "2 (DCE security)"
		entropy := 62 // Varies, but approximately
		info.Entropy = &entropy

	case 3:
		info.Version = "3 (namespace name based with MD5)"
		entropy := 122 // Most bits are hash
		info.Entropy = &entropy

	case 4:
		info.Version = "4 (random)"
		entropy := 122 // 122 bits of randomness
		info.Entropy = &entropy

	case 5:
		info.Version = "5 (namespace name based with SHA-1)"
		entropy := 122 // Most bits are hash
		info.Entropy = &entropy

	case 6:
		info.Version = "6 (reordered timestamp and MAC address)"
		entropy := 14
		info.Entropy = &entropy

	case 7:
		info.Version = "7 (sortable timestamp and random)"
		// Extract timestamp from first 48 bits
		timestampMs := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
		t := time.Unix(timestampMs/1000, (timestampMs%1000)*1e6)
		info.DateTime = &t
		timestampStr := fmt.Sprintf("%.3f", float64(timestampMs)/1000)
		info.Timestamp = &timestampStr
		entropy := 74 // 74 bits of randomness
		info.Entropy = &entropy

	case 8:
		info.Version = "8 (custom)"
		entropy := 122 // Varies
		info.Entropy = &entropy

	default:
		if u == uuid.Nil {
			info.Version = "Nil UUID"
		} else if u == uuid.Max {
			info.Version = "Max UUID"
		} else {
			info.Version = fmt.Sprintf("Unknown version %d", version)
		}
	}

	// Add variant information
	switch variant {
	case uuid.Reserved:
		info.Extra["variant"] = "NCS (Network Computing System)"
	case uuid.RFC4122:
		info.Extra["variant"] = "RFC 4122"
	case uuid.Microsoft:
		info.Extra["variant"] = "Microsoft GUID"
	case uuid.Future:
		info.Extra["variant"] = "Future"
	default:
		info.Extra["variant"] = "Unknown"
	}

	return info, nil
}

func (p *UUIDParser) Generate() (string, error) {
	// Generate a UUID v4 (random)
	u := uuid.New()
	return u.String(), nil
}
