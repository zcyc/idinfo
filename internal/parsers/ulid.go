package parsers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/zcyc/idinfo/internal/types"
)

type ULIDParser struct{}

var ulidRegex = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)

func (p *ULIDParser) Name() string {
	return "ULID"
}

func (p *ULIDParser) CanParse(input string) bool {
	if len(input) != 26 {
		return false
	}
	return ulidRegex.MatchString(input)
}

func (p *ULIDParser) Parse(input string) (*types.IDInfo, error) {
	u, err := ulid.Parse(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "ULID (Universally Unique Lexicographically Sortable Identifier)",
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

	// Extract timestamp (first 48 bits / 6 bytes)
	timestampMs := int64(u.Time())
	t := time.UnixMilli(timestampMs)
	info.DateTime = &t
	timestampStr := fmt.Sprintf("%.3f", float64(timestampMs)/1000)
	info.Timestamp = &timestampStr

	// ULID has 80 bits of entropy (10 bytes)
	entropy := 80
	info.Entropy = &entropy

	// Add ULID-specific information
	info.Extra["encoding"] = "Crockford Base32"
	info.Extra["timestamp_precision"] = "millisecond"
	info.Extra["sortable"] = "true"

	return info, nil
}

func (p *ULIDParser) Generate() (string, error) {
	u := ulid.Make()
	return u.String(), nil
}
