package types

import "time"

// IDInfo represents the parsed information from an ID
type IDInfo struct {
	IDType    string            `json:"id_type"`
	Version   string            `json:"version,omitempty"`
	Standard  string            `json:"standard"`
	Integer   *string           `json:"integer,omitempty"`
	ShortUUID *string           `json:"short_uuid,omitempty"`
	Base64    *string           `json:"base64,omitempty"`
	UUIDWrap  *string           `json:"uuid_wrap,omitempty"`
	Size      int               `json:"size"`
	Entropy   *int              `json:"entropy,omitempty"`
	DateTime  *time.Time        `json:"datetime,omitempty"`
	Timestamp *string           `json:"timestamp,omitempty"`
	Sequence  *int64            `json:"sequence,omitempty"`
	Node1     *string           `json:"node1,omitempty"`
	Node2     *string           `json:"node2,omitempty"`
	Hex       string            `json:"hex"`
	Binary    []byte            `json:"-"`
	Extra     map[string]string `json:"extra,omitempty"`
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
	FormatObjectID  IDFormat = "objectid"
	FormatKSUID     IDFormat = "ksuid"
	FormatXid       IDFormat = "xid"
	FormatCUID2     IDFormat = "cuid2"
	FormatSCRU128   IDFormat = "scru128"
	FormatSCRU64    IDFormat = "scru64"
	FormatSnowflake IDFormat = "snowflake"
	FormatTSID      IDFormat = "tsid"
	FormatNUID      IDFormat = "nuid"
	FormatNanoID    IDFormat = "nanoid"
	FormatUnixTime  IDFormat = "unixtime"
	FormatHashHex   IDFormat = "hashhex"
	FormatBase58    IDFormat = "base58"
	FormatPushID    IDFormat = "pushid"
	FormatBase32    IDFormat = "base32"
	FormatShortUUID IDFormat = "shortuuid"
	FormatSqids     IDFormat = "sqids"
	FormatTypeID    IDFormat = "typeid"
)
