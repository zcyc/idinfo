package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// IDInfo represents the parsed information from an ID
type IDInfo struct {
	IDType         string            `json:"id_type"`
	Version        string            `json:"version,omitempty"`
	Standard       string            `json:"standard"`
	Integer        *string           `json:"integer,omitempty"`
	ShortUUID      *string           `json:"short_uuid,omitempty"`
	Base64         *string           `json:"base64,omitempty"`
	UUIDWrap       *string           `json:"uuid_wrap,omitempty"`
	Parsed         string            `json:"parsed,omitempty"`
	Size           int               `json:"size"`
	Entropy        *int              `json:"entropy,omitempty"`
	DateTime       *time.Time        `json:"datetime,omitempty"`
	Timestamp      *string           `json:"timestamp,omitempty"`
	Relative       *string           `json:"relative_time,omitempty"`
	Sequence       *int64            `json:"sequence,omitempty"`
	Node1          *string           `json:"node1,omitempty"`
	Node2          *string           `json:"node2,omitempty"`
	Node3          *string           `json:"node3,omitempty"`
	Hex            string            `json:"hex"`
	Binary         []byte            `json:"-"`
	ColorMap       string            `json:"-"`
	Extra          map[string]string `json:"extra,omitempty"`
	HighConfidence bool              `json:"-"`
}

// MarshalJSON keeps the public JSON shape identical to uuinfo. Internal
// parsing uses strings for arbitrary-width integers and time.Time for sorting.
func (info IDInfo) MarshalJSON() ([]byte, error) {
	integer := json.RawMessage("null")
	if info.Integer != nil {
		value, ok := new(big.Int).SetString(*info.Integer, 10)
		if !ok || value.Sign() < 0 {
			return nil, fmt.Errorf("invalid integer %q", *info.Integer)
		}
		integer = json.RawMessage(value.String())
	}
	var version, parsed, datetime, hexValue *string
	if info.Version != "" {
		version = &info.Version
	}
	if info.Parsed != "" {
		parsed = &info.Parsed
	}
	if info.DateTime != nil {
		layout := "2006-01-02T15:04:05.000Z07:00"
		if info.Timestamp != nil {
			if dot := strings.LastIndexByte(*info.Timestamp, '.'); dot >= 0 && len(*info.Timestamp)-dot-1 == 9 {
				layout = "2006-01-02T15:04:05.000000000Z07:00"
			}
		}
		dateTime := info.DateTime.UTC()
		value := dateTime.Format(layout)
		if dateTime.Year() >= 10000 {
			value = "+" + value
		}
		datetime = &value
	} else if info.Timestamp != nil {
		value := "Invalid"
		datetime = &value
	}
	if info.Hex != "" {
		hexValue = &info.Hex
	}
	entropy := 0
	if info.Entropy != nil {
		entropy = *info.Entropy
	}
	return json.Marshal(struct {
		IDType    string          `json:"id_type"`
		Version   *string         `json:"version"`
		Standard  string          `json:"standard"`
		Integer   json.RawMessage `json:"integer"`
		UUIDWrap  *string         `json:"uuid_wrap"`
		Parsed    *string         `json:"parsed"`
		Size      int             `json:"size"`
		Entropy   int             `json:"entropy"`
		DateTime  *string         `json:"datetime"`
		Timestamp *string         `json:"timestamp"`
		Relative  *string         `json:"relative_time"`
		Sequence  *int64          `json:"sequence"`
		Node1     *string         `json:"node1"`
		Node2     *string         `json:"node2"`
		Node3     *string         `json:"node3"`
		Hex       *string         `json:"hex"`
	}{
		IDType: info.IDType, Version: version, Standard: info.Standard, Integer: integer,
		UUIDWrap: info.UUIDWrap, Parsed: parsed, Size: info.Size, Entropy: entropy,
		DateTime: datetime, Timestamp: info.Timestamp, Relative: info.Relative,
		Sequence: info.Sequence, Node1: info.Node1, Node2: info.Node2, Node3: info.Node3, Hex: hexValue,
	})
}

// ParseOptions are the command-line options that affect format parsing.
type ParseOptions struct {
	Alphabet string
	Salt     string
	Epoch    uint64
	HasEpoch bool
}

// IDParser interface for all ID parsers
type IDParser interface {
	Name() string
	CanParse(input string) bool
	Parse(input string) (*IDInfo, error)
	Generate() (string, error)
}

// IDFormat represents different ID format types
type IDFormat string

const (
	FormatUUID      IDFormat = "uuid"
	FormatULID      IDFormat = "ulid"
	FormatObjectID  IDFormat = "mongodb"
	FormatKSUID     IDFormat = "ksuid"
	FormatXid       IDFormat = "xid"
	FormatCUID2     IDFormat = "cuid2"
	FormatSCRU128   IDFormat = "scru128"
	FormatSCRU64    IDFormat = "scru64"
	FormatSnowflake IDFormat = "snowflake"
	FormatTSID      IDFormat = "tsid"
	FormatNUID      IDFormat = "nuid"
	FormatNanoID    IDFormat = "nanoid"
	FormatUnixTime  IDFormat = "unix"
	FormatHashHex   IDFormat = "hash"
	FormatPushID    IDFormat = "pushid"
	FormatShortUUID IDFormat = "shortuuid"
	FormatSqids     IDFormat = "sqid"
	FormatTypeID    IDFormat = "typeid"
)
