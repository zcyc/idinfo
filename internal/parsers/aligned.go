package parsers

// This file contains the formats that make idinfo compatible with uuinfo's
// parser set.  The small parser adapter keeps the existing generation API
// intact while allowing the CLI to pass uuinfo's parsing options.

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"github.com/nrednav/cuid2"
	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/rushysloth/go-tsid"
	"github.com/speps/go-hashids/v2"
	"github.com/sqids/sqids-go"
	"github.com/zcyc/idinfo/internal/types"
	"go.jetify.com/typeid/v2"
	"golang.org/x/crypto/sha3"
)

type alignedParser struct {
	name     string
	parse    func(string, types.ParseOptions) (*types.IDInfo, error)
	generate func() (string, error)
}

func (p *alignedParser) Name() string { return p.name }

func (p *alignedParser) CanParse(input string) bool {
	_, err := p.parse(strings.TrimSpace(input), types.ParseOptions{})
	return err == nil
}

func (p *alignedParser) Parse(input string) (*types.IDInfo, error) {
	return p.parse(strings.TrimSpace(input), types.ParseOptions{})
}

func (p *alignedParser) ParseWithOptions(input string, options types.ParseOptions) (*types.IDInfo, error) {
	return p.parse(strings.TrimSpace(input), options)
}

func (p *alignedParser) Generate() (string, error) {
	if p.generate == nil {
		return "", fmt.Errorf("generation is not supported for %s", p.name)
	}
	return p.generate()
}

func stringPtr(value string) *string { return &value }
func intPtr(value int) *int          { return &value }
func int64Ptr(value int64) *int64    { return &value }

func infoFromBytes(idType, version, standard, parsed string, data []byte, size, entropy int) *types.IDInfo {
	copyOfData := append([]byte(nil), data...)
	info := &types.IDInfo{
		IDType:         idType,
		Version:        version,
		Standard:       standard,
		Parsed:         parsed,
		Size:           size,
		Hex:            hex.EncodeToString(copyOfData),
		Binary:         copyOfData,
		Extra:          map[string]string{},
		HighConfidence: entropy >= 0,
	}
	if entropy >= 0 {
		info.Entropy = intPtr(entropy)
	}
	return info
}

func setInteger(info *types.IDInfo, data []byte) {
	info.Integer = stringPtr(new(big.Int).SetBytes(data).String())
}

func setIntegerValue(info *types.IDInfo, value *big.Int, width int) {
	info.Integer = stringPtr(value.String())
	if width <= 0 {
		return
	}
	info.Binary = bigEndianBytes(value, width)
	info.Hex = hex.EncodeToString(info.Binary)
}

func bigEndianBytes(value *big.Int, width int) []byte {
	data := value.Bytes()
	if len(data) >= width {
		return append([]byte(nil), data[len(data)-width:]...)
	}
	result := make([]byte, width)
	copy(result[width-len(data):], data)
	return result
}

func decodeBase(value, alphabet string) (*big.Int, error) {
	if value == "" {
		return nil, errors.New("empty encoded value")
	}
	result := new(big.Int)
	base := big.NewInt(int64(len(alphabet)))
	for _, char := range value {
		index := strings.IndexRune(alphabet, char)
		if index < 0 {
			return nil, fmt.Errorf("invalid character %q", char)
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(index)))
	}
	return result, nil
}

func encodeBase(value *big.Int, alphabet string, width int) string {
	if value.Sign() == 0 {
		return strings.Repeat(alphabet[:1], width)
	}
	base := big.NewInt(int64(len(alphabet)))
	zero := new(big.Int)
	working := new(big.Int).Set(value)
	var reversed []byte
	for working.Cmp(zero) > 0 {
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(working, base, remainder)
		reversed = append(reversed, alphabet[remainder.Int64()])
		working = quotient
	}
	for len(reversed) < width {
		reversed = append(reversed, alphabet[0])
	}
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return string(reversed)
}

func parseBigDecimal(value string, bits int) (*big.Int, error) {
	if value == "" || strings.HasPrefix(value, "-") {
		return nil, errors.New("not an unsigned integer")
	}
	number, ok := new(big.Int).SetString(value, 10)
	if !ok || number.BitLen() > bits {
		return nil, errors.New("integer out of range")
	}
	return number, nil
}

func epochMillis(options types.ParseOptions, defaultMillis int64) int64 {
	if options.HasEpoch {
		return options.Epoch * 1000
	}
	return defaultMillis
}

func timestampInfo(rawMillis, epoch int64) (*string, *time.Time) {
	millis := rawMillis + epoch
	seconds := millis / 1000
	nanos := (millis % 1000) * int64(time.Millisecond)
	if millis < 0 && nanos != 0 {
		seconds--
		nanos += int64(time.Second)
	}
	datetime := time.Unix(seconds, nanos).UTC()
	timestamp := fmt.Sprintf("%d.%03d", seconds, datetime.Nanosecond()/int(time.Millisecond))
	return &timestamp, &datetime
}

func asciiInfo(idType, version, standard, parsed string, value string, entropy int) *types.IDInfo {
	info := infoFromBytes(idType, version, standard, parsed, []byte(value), len(value)*8, entropy)
	return info
}

func uuidFromBytes(data []byte) (uuid.UUID, error) {
	if len(data) != 16 {
		return uuid.Nil, errors.New("UUID must contain 16 bytes")
	}
	var value uuid.UUID
	copy(value[:], data)
	return value, nil
}

func uuidTimestamp(value uuid.UUID) (int64, bool) {
	b := value[:]
	switch value.Version() {
	case 1:
		// UUID time is 100 ns since 1582-10-15.
		ticks := uint64(binary.BigEndian.Uint32(b[0:4])) |
			uint64(binary.BigEndian.Uint16(b[4:6]))<<32 |
			uint64(binary.BigEndian.Uint16(b[6:8])&0x0fff)<<48
		return int64((ticks - 0x01b21dd213814000) / 10000), true
	case 6:
		ticks := uint64(binary.BigEndian.Uint32(b[0:4]))<<28 |
			uint64(binary.BigEndian.Uint16(b[4:6]))<<12 |
			uint64(binary.BigEndian.Uint16(b[6:8])&0x0fff)
		return int64((ticks - 0x01b21dd213814000) / 10000), true
	case 7:
		return int64(binary.BigEndian.Uint64(append([]byte{0, 0}, b[0:6]...))), true
	default:
		return 0, false
	}
}

func formatNodeID(data []byte) string {
	parts := make([]string, len(data))
	for index, value := range data {
		parts[index] = fmt.Sprintf("%02x", value)
	}
	return strings.Join(parts, ":")
}

func parseUUIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := uuid.Parse(input)
	if err != nil {
		return nil, err
	}
	info := infoFromBytes("UUID", "", value.String(), "from hex", value[:], 128, 0)
	setIntegerValue(info, new(big.Int).SetBytes(value[:]), 16)
	version := value.Version()
	variant := value.Variant()
	switch {
	case value == uuid.Nil:
		info.IDType = "Nil UUID (all zeros)"
	case value == uuid.Max:
		info.IDType = "Max UUID (all ones)"
	case variant == uuid.Reserved:
		info.IDType = "NCS UUID"
	case variant == uuid.Microsoft:
		info.IDType = "Microsoft GUID"
	case variant == uuid.RFC4122:
		if version >= 6 {
			info.IDType = "UUID (RFC-9562)"
		} else {
			info.IDType = "UUID (RFC-4122)"
		}
		labels := map[uuid.Version]string{
			1: "1 (timestamp and node)", 2: "2 (DCE security)", 3: "3 (MD5 hash)",
			4: "4 (random)", 5: "5 (SHA-1 hash)", 6: "6 (sortable timestamp and node)",
			7: "7 (sortable timestamp and random)", 8: "8 (custom)",
		}
		info.Version = labels[version]
		if info.Version == "" {
			info.Version = fmt.Sprintf("%d (out of spec)", version)
		}
		entropy := 122
		if version == 7 {
			entropy = 74
		}
		if version == 1 || version == 6 {
			entropy = 0
		}
		info.Entropy = intPtr(entropy)
	default:
		info.IDType = "Unknown UUID-like"
	}
	if value.Variant() == uuid.RFC4122 {
		if rawMillis, ok := uuidTimestamp(value); ok {
			info.Timestamp, info.DateTime = timestampInfo(rawMillis, epochMillis(options, 0))
		}
	}
	if version == 1 || version == 6 {
		info.Sequence = int64Ptr(int64(binary.BigEndian.Uint16(value[8:10]) & 0x3fff))
		info.Node1 = stringPtr(formatNodeID(value[10:16]))
	}
	info.HighConfidence = value != uuid.Nil && binary.BigEndian.Uint32(value[12:]) != 0
	info.Extra["variant"] = map[uuid.Variant]string{
		uuid.Reserved: "NCS (Network Computing System)", uuid.RFC4122: "RFC 4122",
		uuid.Microsoft: "Microsoft GUID", uuid.Future: "Future",
	}[variant]
	return info, nil
}

func parseShortUUIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 22 {
		return nil, errors.New("invalid ShortUUID length")
	}
	value, err := shortuuid.DefaultEncoder.Decode(input)
	if err != nil {
		return nil, err
	}
	return withUUIDWrapper(input, options, "ShortUUID", "from base57", value[:])
}

func withUUIDWrapper(input string, options types.ParseOptions, prefix, parsed string, data []byte) (*types.IDInfo, error) {
	value, err := uuidFromBytes(data)
	if err != nil {
		return nil, err
	}
	info, err := parseUUIDAligned(value.String(), options)
	if err != nil {
		return nil, err
	}
	info.IDType = prefix + " of " + info.IDType
	info.Standard = input
	info.UUIDWrap = stringPtr(value.String())
	info.Parsed = parsed
	return info, nil
}

func parseUUIDInteger(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := parseBigDecimal(input, 128)
	if err != nil {
		return nil, err
	}
	return withUUIDWrapper(value.String(), options, "Integer", "as integer", bigEndianBytes(value, 16))
}

func parseUUIDBase64(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var data []byte
	padded := true
	var err error
	if strings.Contains(input, "=") {
		data, err = base64.URLEncoding.DecodeString(input)
	} else {
		data, err = base64.RawURLEncoding.DecodeString(input)
		padded = false
	}
	if err != nil || len(data) != 16 {
		return nil, errors.New("invalid UUID base64")
	}
	// uuinfo follows Rust's Uuid::from_slice_le semantics.
	for _, span := range [][2]int{{0, 4}, {4, 6}, {6, 8}} {
		for left, right := span[0], span[1]-1; left < right; left, right = left+1, right-1 {
			data[left], data[right] = data[right], data[left]
		}
	}
	prefix := "Unpadded Base64"
	if padded {
		prefix = "Padded Base64"
	}
	return withUUIDWrapper(input, options, prefix, "from base64", data)
}

func parseUUID25(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 25 {
		return nil, errors.New("Uuid25 must be 25 characters")
	}
	value, err := decodeBase(input, "0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil || value.BitLen() > 128 {
		return nil, errors.New("invalid Uuid25")
	}
	return withUUIDWrapper(input, options, "Uuid25", "from base36", bigEndianBytes(value, 16))
}

func parseULIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := ulid.Parse(input)
	fromBase32 := err == nil
	if err != nil {
		uuidValue, uuidErr := uuid.Parse(input)
		if uuidErr != nil {
			return nil, err
		}
		copy(value[:], uuidValue[:])
	}
	info := infoFromBytes("ULID", "", value.String(), "from Crockford's base32", value[:], 128, 80)
	if !fromBase32 {
		info.IDType = "ULID wrapped in UUID"
		info.Standard = value.String()
		info.Parsed = "from hex"
		info.HighConfidence = false
	}
	setIntegerValue(info, new(big.Int).SetBytes(value[:]), 16)
	uuidValue, _ := uuidFromBytes(value[:])
	info.UUIDWrap = stringPtr(uuidValue.String())
	info.Timestamp, info.DateTime = timestampInfo(int64(value.Time()), epochMillis(options, 0))
	return info, nil
}

func parseJulid(input string, options types.ParseOptions) (*types.IDInfo, error) {
	info, err := parseULIDAligned(input, options)
	if err != nil {
		return nil, err
	}
	value := new(big.Int)
	value.SetString(*info.Integer, 10)
	info.IDType = "Julid"
	info.Entropy = intPtr(64)
	sequence := new(big.Int).Rsh(value, 64)
	sequence.And(sequence, big.NewInt(0xffff))
	info.Sequence = int64Ptr(sequence.Int64())
	return info, nil
}

const (
	crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	base62Alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func parseUPID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	parts := strings.SplitN(input, "_", 2)
	var value *big.Int
	fromBase36 := len(parts) == 2 && len(parts[0]) >= 1 && len(parts[0]) <= 31 && len(parts[1]) == 22
	if fromBase36 {
		for _, char := range parts[0] {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
				return nil, errors.New("invalid UPID prefix")
			}
		}
		var err error
		value, err = decodeBase(parts[1], "0123456789abcdefghijklmnopqrstuvwxyz")
		if err != nil || value.BitLen() > 128 {
			return nil, errors.New("invalid UPID payload")
		}
	} else {
		wrapped, err := uuid.Parse(input)
		if err != nil {
			return nil, errors.New("invalid UPID")
		}
		value = new(big.Int).SetBytes(wrapped[:])
	}
	data := bigEndianBytes(value, 16)
	standard := encodeBase(value, "0123456789abcdefghijklmnopqrstuvwxyz", 22)
	parsed := "from hex"
	idType := "UPID wrapped in UUID"
	if fromBase36 {
		standard, parsed, idType = input, "from Crockford's base32", "UPID"
	}
	info := infoFromBytes(idType, "A (default)", standard, parsed, data, 128, 64)
	setIntegerValue(info, value, 16)
	if fromBase36 {
		info.Node1 = stringPtr(parts[0])
	}
	wrapped, _ := uuidFromBytes(data)
	info.UUIDWrap = stringPtr(wrapped.String())
	info.HighConfidence = fromBase36
	return info, nil
}

func decodeSandflake(input string) ([]byte, error) {
	value, err := decodeBase(input, "0123456789ABCDEFGHJKMNPQRSTVWXYZ")
	if err != nil || value.BitLen() > 128 {
		return nil, errors.New("invalid Sandflake")
	}
	return bigEndianBytes(value, 16), nil
}

func parseSandflake(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var data []byte
	var idType, parsed string
	if len(input) == 26 {
		data, _ = decodeSandflake(input)
		if data == nil {
			return nil, errors.New("invalid Sandflake")
		}
		idType, parsed = "Sandflake", "from base32, custom alphabet"
	} else {
		value, err := uuid.Parse(input)
		if err != nil {
			return nil, errors.New("invalid Sandflake")
		}
		data = append([]byte(nil), value[:]...)
		idType, parsed = "Sandflake wrapped in UUID", "from hex"
	}
	info := infoFromBytes(idType, "", input, parsed, data, 128, 24)
	setIntegerValue(info, new(big.Int).SetBytes(data), 16)
	info.Standard = encodeBase(new(big.Int).SetBytes(data), "0123456789ABCDEFGHJKMNPQRSTVWXYZ", 26)
	if idType == "Sandflake wrapped in UUID" {
		info.UUIDWrap = stringPtr(input)
	}
	timestamp := int64(binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...)))
	info.Timestamp, info.DateTime = timestampInfo(timestamp, epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Worker ID)", binary.BigEndian.Uint32(data[6:10])))
	info.Sequence = int64Ptr(int64(data[10])<<16 | int64(data[11])<<8 | int64(data[12]))
	minMillis := int64(1420070400000)
	info.HighConfidence = timestamp >= minMillis && timestamp <= time.Now().AddDate(10, 0, 0).UnixMilli()
	return info, nil
}

func parseTimeflake(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var data []byte
	var idType, parsed string
	switch len(input) {
	case 22:
		value, err := decodeBase(input, base62Alphabet)
		if err != nil || value.BitLen() > 128 {
			return nil, errors.New("invalid Timeflake")
		}
		data, idType, parsed = bigEndianBytes(value, 16), "Timeflake", "from base62"
	case 32:
		var err error
		data, err = hex.DecodeString(input)
		if err != nil || len(data) != 16 {
			return nil, errors.New("invalid Timeflake")
		}
		idType, parsed = "Timeflake", "from hex"
	case 36:
		value, err := uuid.Parse(input)
		if err != nil {
			return nil, err
		}
		data, idType, parsed = append([]byte(nil), value[:]...), "Timeflake wrapped in UUID", "from hex"
	default:
		return nil, errors.New("invalid Timeflake length")
	}
	info := infoFromBytes(idType, "", input, parsed, data, 128, 80)
	setIntegerValue(info, new(big.Int).SetBytes(data), 16)
	info.Standard = encodeBase(new(big.Int).SetBytes(data), base62Alphabet, 22)
	if idType == "Timeflake wrapped in UUID" {
		info.UUIDWrap = stringPtr(input)
	} else {
		wrapped, _ := uuidFromBytes(data)
		info.UUIDWrap = stringPtr(wrapped.String())
	}
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...))), epochMillis(options, 0))
	info.HighConfidence = len(input) == 22
	return info, nil
}

func parseFlake(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var data []byte
	var idType, parsed string
	if len(input) == 18 {
		value, err := decodeBase(input, base62Alphabet)
		if err != nil || value.BitLen() > 128 {
			return nil, errors.New("invalid Flake")
		}
		data, idType, parsed = bigEndianBytes(value, 16), "Flake (Boundary)", "from base62"
	} else {
		value, err := uuid.Parse(input)
		if err != nil {
			return nil, err
		}
		data, idType, parsed = append([]byte(nil), value[:]...), "Flake (Boundary) wrapped in UUID", "from hex"
	}
	info := infoFromBytes(idType, "", input, parsed, data, 128, 16)
	setIntegerValue(info, new(big.Int).SetBytes(data), 16)
	info.Standard = encodeBase(new(big.Int).SetBytes(data), base62Alphabet, 22)
	wrapped, _ := uuidFromBytes(data)
	info.UUIDWrap = stringPtr(wrapped.String())
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint64(data[:8])), epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%d", new(big.Int).SetBytes(data[8:14])))
	info.Sequence = int64Ptr(int64(binary.BigEndian.Uint16(data[14:16])))
	info.HighConfidence = len(input) == 18 && new(big.Int).SetBytes(data).BitLen() <= 108
	return info, nil
}

func parseSCRU128Aligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := decodeBase(input, "0123456789abcdefghijklmnopqrstuvwxyz")
	fromBase36 := err == nil && value.BitLen() <= 128
	if !fromBase36 {
		wrapped, uuidErr := uuid.Parse(input)
		if uuidErr != nil {
			return nil, errors.New("invalid SCRU128")
		}
		value = new(big.Int).SetBytes(wrapped[:])
	}
	data := bigEndianBytes(value, 16)
	info := infoFromBytes("SCRU128", "", input, "from base36", data, 128, 80)
	if !fromBase36 {
		info.IDType = "SCRU128 wrapped in UUID"
		info.Parsed = "from hex"
		info.HighConfidence = false
	}
	setIntegerValue(info, value, 16)
	wrapped, _ := uuidFromBytes(data)
	info.UUIDWrap = stringPtr(wrapped.String())
	info.Timestamp, info.DateTime = timestampInfo(new(big.Int).Rsh(new(big.Int).Set(value), 84).Int64(), epochMillis(options, 0))
	counter := new(big.Int).Rsh(new(big.Int).Set(value), 56)
	counter.And(counter, big.NewInt((1<<28)-1))
	info.Sequence = int64Ptr(counter.Int64())
	return info, nil
}

func parseSCRU64Aligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) < 1 || len(input) > 13 {
		return nil, errors.New("invalid SCRU64")
	}
	value, err := decodeBase(strings.ToUpper(input), crockfordAlphabet)
	if err != nil || value.BitLen() > 64 {
		return nil, errors.New("invalid SCRU64")
	}
	number := value.Uint64()
	info := infoFromBytes("SCRU64", "", input, "from base36", bigEndianBytes(value, 8), 64, 0)
	setIntegerValue(info, value, 8)
	info.Timestamp, info.DateTime = timestampInfo(int64(number>>20)*256, epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Node ID)", (number>>12)&0xfff))
	return info, nil
}

func parseTSIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)
	var value uint64
	fromBase32 := false
	if len([]rune(input)) == 13 && tsid.IsValidRuneArray([]rune(input)) {
		value = uint64(tsid.FromString(input).ToNumber())
		fromBase32 = true
	} else {
		parsed, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, errors.New("invalid TSID")
		}
		value = parsed
	}
	standard := strconv.FormatUint(value, 10)
	parsed := "as integer"
	if fromBase32 {
		standard = strings.ToUpper(input)
		parsed = "from Crockford's base32"
	}
	info := infoFromBytes("TSID", "", standard, parsed, bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, 22)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Timestamp, info.DateTime = timestampInfo(int64(value>>22), epochMillis(options, 1577836800000))
	info.HighConfidence = fromBase32
	return info, nil
}

func parseDatadog(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := uuid.Parse(input)
	if err != nil || value == uuid.Nil || binary.BigEndian.Uint32(value[4:8]) != 0 {
		return nil, errors.New("invalid Datadog trace ID")
	}
	info := infoFromBytes("Datadog Trace ID", "", value.String(), "from hex", value[:], 128, 64)
	setIntegerValue(info, new(big.Int).SetBytes(value[:]), 16)
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint32(value[0:4]))*1000, epochMillis(options, 0))
	return info, nil
}

func parseSpotify(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 22 {
		return nil, errors.New("invalid Spotify ID")
	}
	value, err := decodeBase(input, base62Alphabet)
	if err != nil || value.BitLen() > 128 {
		return nil, errors.New("invalid Spotify ID")
	}
	data := bigEndianBytes(value, 16)
	info := infoFromBytes("Spotify ID", "", input, "from base62", data, 128, 128)
	setIntegerValue(info, value, 16)
	wrapped, _ := uuidFromBytes(data)
	info.UUIDWrap = stringPtr(wrapped.String())
	return info, nil
}

func parseComb(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := uuid.Parse(input)
	if err != nil || value.Version() != 4 || value.Variant() != uuid.RFC4122 {
		return nil, errors.New("invalid COMB")
	}
	b := value[:]
	legacyBase := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	candidates := []struct {
		name string
		ms   int64
	}{
		{"PostgreSQL", int64(binary.BigEndian.Uint64(append([]byte{0, 0}, b[:6]...)))},
		{"MS-SQL", int64(binary.BigEndian.Uint64(append([]byte{0, 0}, b[10:16]...)))},
		{"Legacy", legacyBase + int64(binary.BigEndian.Uint16(b[10:12]))*24*60*60*1000 + int64(binary.BigEndian.Uint32(b[12:16]))*1000/300},
	}
	now := time.Now().UnixMilli()
	low, high := now-30*365*24*60*60*1000, now+10*365*24*60*60*1000
	chosen := candidates[0]
	valid := 0
	for _, candidate := range candidates {
		if candidate.ms >= low && candidate.ms <= high {
			chosen = candidate
			valid++
		}
	}
	info := infoFromBytes("COMB", chosen.name, value.String(), "from hex", b, 128, 74)
	setIntegerValue(info, new(big.Int).SetBytes(b), 16)
	if chosen.ms < 946684800000 {
		return nil, errors.New("invalid COMB timestamp")
	}
	info.Timestamp, info.DateTime = timestampInfo(chosen.ms, 0)
	info.HighConfidence = valid == 1
	return info, nil
}

func bitsU64(value uint64, offset, length uint) uint64 {
	if length == 64 {
		return value
	}
	return (value >> (64 - offset - length)) & ((uint64(1) << length) - 1)
}

type snowflakeLayout struct {
	name       string
	defaultMS  int64
	timeOffset uint
	timeBits   uint
	nodeOffset uint
	nodeBits   uint
	node2      func(uint64) string
	seqOffset  uint
	seqBits    uint
}

var snowflakeLayouts = map[string]snowflakeLayout{
	"sf-twitter":     {"Twitter", 1288834974657, 1, 41, 42, 10, nil, 52, 12},
	"sf-discord":     {"Discord", 1420070400000, 0, 42, 42, 5, func(v uint64) string { return fmt.Sprintf("%d (Process ID)", bitsU64(v, 47, 5)) }, 52, 12},
	"sf-instagram":   {"Instagram", 1314220021721, 0, 41, 41, 13, nil, 54, 10},
	"sf-sony":        {"Sony", 1409529600000, 1, 39, 48, 16, nil, 40, 8},
	"sf-spaceflake":  {"Spaceflake", 1420070400000, 1, 41, 42, 5, func(v uint64) string { return fmt.Sprintf("%d (Worker ID)", bitsU64(v, 47, 5)) }, 52, 12},
	"sf-linkedin":    {"LinkedIn", 0, 1, 41, 42, 10, nil, 52, 12},
	"sf-mastodon":    {"Mastodon", 0, 0, 48, 0, 0, nil, 48, 16},
	"sf-frostflake":  {"Frostflake", 0, 0, 32, 53, 11, nil, 32, 21},
	"sf-flakeid":     {"Flake ID", 0, 42, 42, 42, 5, func(v uint64) string { return fmt.Sprintf("%d (Worker ID)", bitsU64(v, 47, 5)) }, 52, 12},
	"sf-simpleflake": {"Simpleflake", 946702800000, 0, 41, 0, 0, nil, 0, 0},
}

func parseSnowflakeAligned(input string, options types.ParseOptions, layoutName string) (*types.IDInfo, error) {
	fromBase58 := false
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil && layoutName == "sf-frostflake" {
		decoded, decodeErr := decodeBase(input, base58Alphabet)
		if decodeErr != nil || decoded.BitLen() > 64 {
			return nil, errors.New("invalid Snowflake")
		}
		value, fromBase58 = decoded.Uint64(), true
	} else if err != nil {
		return nil, errors.New("invalid Snowflake")
	}
	layout, ok := snowflakeLayouts[layoutName]
	if !ok {
		return nil, errors.New("unknown Snowflake layout")
	}
	standard := strconv.FormatUint(value, 10)
	if fromBase58 {
		standard = encodeBase(new(big.Int).SetUint64(value), "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz", 0)
	}
	info := infoFromBytes("Snowflake", layout.name, standard, "as integer", bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, 0)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Standard = standard
	rawTimestamp := bitsU64(value, layout.timeOffset, layout.timeBits)
	if layout.name == "Sony" {
		rawTimestamp *= 10
	} else if layout.name == "Frostflake" {
		rawTimestamp *= 1000
	}
	info.Timestamp, info.DateTime = timestampInfo(int64(rawTimestamp), epochMillis(options, layout.defaultMS))
	if layout.nodeBits > 0 {
		label := "Worker ID"
		if layoutName == "sf-flakeid" {
			label = "Datacenter ID"
		}
		info.Node1 = stringPtr(fmt.Sprintf("%d (%s)", bitsU64(value, layout.nodeOffset, layout.nodeBits), label))
	}
	if layoutName == "sf-flakeid" {
		info.Node2 = stringPtr(fmt.Sprintf("%d (Worker ID)", bitsU64(value, 47, 5)))
	}
	if layout.node2 != nil {
		info.Node2 = stringPtr(layout.node2(value))
	}
	if layout.seqBits > 0 {
		info.Sequence = int64Ptr(int64(bitsU64(value, layout.seqOffset, layout.seqBits)))
	}
	return info, nil
}

func parseSnowflakeAuto(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return nil, err
	}
	info := infoFromBytes("Snowflake", "Unknown (use -f to specify version)", input, "as integer", bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, 0)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Standard = input
	return info, nil
}

type unixMode int

const (
	unixAuto unixMode = iota
	unixSeconds
	unixMillis
	unixMicros
	unixNanos
)

func parseUnixAligned(input string, options types.ParseOptions, mode unixMode, recentOnly bool) (*types.IDInfo, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return nil, err
	}
	unit, multiplier, version := "seconds", uint64(1_000_000_000), "Assuming seconds"
	switch mode {
	case unixSeconds:
		version = "As seconds"
	case unixMillis:
		unit, multiplier, version = "milliseconds", 1_000_000, "As milliseconds"
	case unixMicros:
		unit, multiplier, version = "microseconds", 1_000, "As microseconds"
	case unixNanos:
		unit, multiplier, version = "nanoseconds", 1, "As nanoseconds"
	default:
		switch {
		case value < 100_000_000_000:
		case value < 100_000_000_000_000:
			unit, multiplier, version = "milliseconds", 1_000_000, "Assuming milliseconds"
		case value < 10_000_000_000_000_000:
			unit, multiplier, version = "microseconds", 1_000, "Assuming microseconds"
		default:
			unit, multiplier, version = "nanoseconds", 1, "Assuming nanoseconds"
		}
	}
	if value > ^uint64(0)/multiplier {
		return nil, errors.New("Unix timestamp overflow")
	}
	nanos := value * multiplier
	seconds := int64(nanos / 1_000_000_000)
	if seconds < 0 || seconds > 253402300799 {
		return nil, errors.New("Unix timestamp out of range")
	}
	if recentOnly {
		now := time.Now().Unix()
		if seconds < now-10*365*24*60*60 || seconds > now+365*24*60*60 {
			return nil, errors.New("not recent")
		}
	}
	info := infoFromBytes("Unix timestamp", version, input, "as integer", bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, 0)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Standard = input
	info.Timestamp, info.DateTime = timestampInfo(int64(nanos/1_000_000), 0)
	info.Extra["unit"] = unit
	return info, nil
}

func parseHashAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if _, ok := map[int]bool{32: true, 40: true, 56: true, 64: true, 96: true, 128: true}[len(input)]; !ok {
		return nil, errors.New("invalid hash length")
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}
	labels := map[int]string{32: "Probably MD5", 40: "Probably SHA-1", 56: "Probably SHA-224", 64: "Probably SHA-256", 96: "Probably SHA-384", 128: "Probably SHA-512"}
	info := infoFromBytes("Hex-encoded Hash", labels[len(input)], input, "from hex", data, len(data)*8, len(data)*8)
	info.HighConfidence = false
	return info, nil
}

func parseCUID1(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 25 || input[0] != 'c' {
		return nil, errors.New("invalid CUID1")
	}
	timestamp, err := decodeBase(input[1:9], "0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		return nil, err
	}
	sequence, err := decodeBase(input[9:13], "0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		return nil, err
	}
	fingerprint, err := decodeBase(input[13:17], "0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		return nil, err
	}
	info := asciiInfo("CUID", "1", input, "as ASCII, with base36 parts", input, 64)
	setIntegerValue(info, new(big.Int).SetBytes([]byte(input)), 0)
	info.Timestamp, info.DateTime = timestampInfo(timestamp.Int64(), epochMillis(options, 0))
	info.Sequence = int64Ptr(sequence.Int64())
	info.Node1 = stringPtr(fmt.Sprintf("%d (Fingerprint)", fingerprint))
	return info, nil
}

func parseCUID2Aligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) < 2 || len(input) > 32 || !isLowerAlphaNumeric(input) {
		return nil, errors.New("invalid CUID2")
	}
	// The dependency is already used by the original parser and implements the
	// CUID2 checksum/shape rules.
	if !cuid2.IsCuid(input) {
		return nil, errors.New("invalid CUID2")
	}
	info := asciiInfo("CUID", "2", input, "as ASCII, with base36 hash", input, len(input)*8)
	info.HighConfidence = len(input) == 24
	return info, nil
}

func isLowerAlphaNumeric(value string) bool {
	for _, char := range value {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') {
			return false
		}
	}
	return true
}

func parseNanoIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) < 2 || len(input) > 36 {
		return nil, errors.New("invalid Nano ID length")
	}
	alphabet := options.Alphabet
	if alphabet == "" {
		alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
	}
	for _, char := range input {
		if !strings.ContainsRune(alphabet, char) {
			return nil, errors.New("invalid Nano ID alphabet")
		}
	}
	version := fmt.Sprintf("Custom alphabet, custom length (%d)", len(input))
	if options.Alphabet == "" {
		version = fmt.Sprintf("Default alphabet, custom length (%d)", len(input))
	}
	if len(input) == 21 {
		if options.Alphabet == "" {
			version = "Default alphabet, default length"
		} else {
			version = "Custom alphabet, default length"
		}
	}
	info := asciiInfo("Nano ID", version, input, "as ASCII", input, len(input)*8)
	info.HighConfidence = options.Alphabet == "" && len(input) == 21
	return info, nil
}

func parseNUIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 22 || !isBase62(input) {
		return nil, errors.New("invalid NUID")
	}
	info := asciiInfo("NUID", "", input, "as ASCII", input, 96)
	info.Size = 176
	info.Node1 = stringPtr(input[12:])
	return info, nil
}

func isBase62(value string) bool {
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')) {
			return false
		}
	}
	return true
}

func parseTypeIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	tid, err := typeid.Parse(input)
	if err != nil {
		return nil, err
	}
	data := tid.Bytes()
	if len(data) != 16 {
		return nil, errors.New("invalid TypeID suffix")
	}
	info := infoFromBytes("TypeID", "", input, "from base32, suffix only", data, 128, 74)
	setIntegerValue(info, new(big.Int).SetBytes(data), 16)
	wrapped, _ := uuidFromBytes(data)
	info.UUIDWrap = stringPtr(wrapped.String())
	info.Node1 = stringPtr(tid.Prefix())
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...))), epochMillis(options, 0))
	return info, nil
}

func parsePushIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 20 {
		return nil, errors.New("invalid PushID")
	}
	data, err := decodePushID(input)
	if err != nil {
		return nil, err
	}
	info := infoFromBytes("PushID (Firebase)", "", input, "from base64", data, 120, 96)
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...))), epochMillis(options, 0))
	return info, nil
}

func decodePushID(value string) ([]byte, error) {
	const alphabet = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
	var result []byte
	acc, bits := uint64(0), uint(0)
	for _, char := range value {
		index := strings.IndexRune(alphabet, char)
		if index < 0 {
			return nil, errors.New("invalid PushID alphabet")
		}
		acc = (acc << 6) | uint64(index)
		bits += 6
		for bits >= 8 {
			bits -= 8
			result = append(result, byte(acc>>bits))
		}
	}
	if len(result) != 15 {
		return nil, errors.New("invalid PushID length")
	}
	return result, nil
}

func parsePUID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 24 {
		return nil, errors.New("invalid Puid")
	}
	base36 := "0123456789abcdefghijklmnopqrstuvwxyz"
	timestamp, err := decodeBase(input[:8], base36)
	if err != nil {
		return nil, err
	}
	if _, err = hex.DecodeString(input[8:14]); err != nil {
		return nil, err
	}
	process, err := decodeBase(input[14:18], base36)
	if err != nil {
		return nil, err
	}
	sequence, err := decodeBase(input[18:], base36)
	if err != nil {
		return nil, err
	}
	info := asciiInfo("Puid", "", input, "as ASCII, with base36 parts", input, 0)
	info.Timestamp, info.DateTime = timestampInfo(timestamp.Int64(), epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%s (Machine ID)", input[8:14]))
	info.Node2 = stringPtr(fmt.Sprintf("%d (Process ID)", process))
	info.Sequence = int64Ptr(sequence.Int64())
	return info, nil
}

func parseShortPUID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 12 && len(input) != 14 {
		return nil, errors.New("invalid short Puid")
	}
	timestamp, err := decodeBase(input[:12], "0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		return nil, err
	}
	version := "Short puid without node ID"
	if len(input) == 14 {
		version = "Short puid with node ID"
	}
	info := asciiInfo("Puid", version, input, "as ASCII, with base36 parts", input, 0)
	info.Timestamp, info.DateTime = timestampInfo(timestamp.Int64()/1_000_000, epochMillis(options, 0))
	if len(input) == 14 {
		node, _ := decodeBase(input[12:], "0123456789abcdefghijklmnopqrstuvwxyz")
		info.Node1 = stringPtr(fmt.Sprintf("%d (Node ID)", node))
	}
	return info, nil
}

func parseBreezeID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) < 4 || len(input) > 128 || strings.HasSuffix(input, "-") {
		return nil, errors.New("invalid Breeze ID")
	}
	upper := "AEIUQWTYPSDFGRHJKLZXCVBNM2346789"
	alphabet := "Default alphabet"
	if strings.ToLower(input) == input {
		alphabet = "Lowercase alphabet"
	} else if strings.ToUpper(input) != input {
		return nil, errors.New("invalid Breeze ID alphabet")
	}
	for _, char := range strings.ToUpper(input) {
		if char != '-' && !strings.ContainsRune(upper, char) {
			return nil, errors.New("invalid Breeze ID")
		}
	}
	info := asciiInfo("Breeze ID", alphabet, input, "as ASCII", input, (len(input)-strings.Count(input, "-"))*8)
	info.HighConfidence = alphabet == "Default alphabet"
	return info, nil
}

func parseSlack(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) < 3 || len(input) > 20 {
		return nil, errors.New("invalid Slack ID length")
	}
	offset, kind := 1, ""
	switch input[0] {
	case 'A':
		kind = "App"
	case 'B':
		kind = "Bot"
	case 'C', 'G':
		kind = "Channel"
	case 'D':
		kind = "Direct message"
	case 'E':
		kind = "Enterprise"
	case 'F':
		kind = "File"
	case 'T':
		kind = "Team"
	case 'U':
		kind = "User"
	case 'W':
		kind = "User"
		if len(input) > 1 && input[1] == 'f' {
			offset, kind = 2, "Workflow"
		}
	default:
		return nil, errors.New("invalid Slack ID prefix")
	}
	value, err := decodeBase(input[offset:], "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err != nil {
		return nil, err
	}
	info := asciiInfo("Slack ID", kind+" ID", input, "as ASCII, with base36 parts", input, (len(input)-offset)*8)
	setIntegerValue(info, value, 0)
	info.Node1 = stringPtr(fmt.Sprintf("%s (%s)", input[:offset], kind))
	return info, nil
}

var stripeTypes = map[string]string{
	"ac": "Platform Client ID", "acct": "Account ID", "aliacc": "Alipay Account ID", "ba": "Bank Account ID",
	"btok": "Bank Token ID", "card": "Card ID", "cbtxn": "Customer Balance Transaction ID", "ch": "Charge ID",
	"cn": "Credit Note ID", "cs_live": "Live Checkout Session ID", "cs_test": "Test Checkout Session ID", "cus": "Customer ID",
	"dp": "Dispute ID", "evt": "Event ID", "fee": "Application Fee ID", "file": "File ID", "fr": "Application Fee Refund ID",
	"iauth": "Issuing Authorization ID", "ic": "Issuing Card ID", "ich": "Issuing Card Holder ID", "idp": "Issuing Dispute ID",
	"ii": "Invoice Item ID", "il": "Invoice Line Item ID", "in": "Invoice ID", "ipi": "Issuing Transaction ID", "link": "File Link ID",
	"or": "Order ID", "orret": "Order Return ID", "person": "Person ID", "pi": "Payment Intent ID", "pk_live": "Live public key",
	"pk_test": "Test public key", "pm": "Payment Method ID", "po": "Payout ID", "price": "Price ID", "prod": "Product ID",
	"prv": "Review ID", "pst_live": "Live Connection token", "pst_test": "Test Connection token", "py": "Payment ID", "pyr": "Payment Refund ID",
	"qt": "Quote ID", "rcpt": "Receipt ID", "re": "Refund ID", "req": "Request ID", "rk_live": "Live restricted key",
	"rk_test": "Test restricted key", "seti": "Setup Intent ID", "si": "Subscription Item ID", "sk_live": "Live secret key", "sk_test": "Test secret key",
	"sku": "SKU ID", "sli": "Subscription Line Item ID", "sqr": "Scheduled Query Run ID", "src": "Source ID", "sub": "Subscription ID",
	"tml": "Terminal Location ID", "tmr": "Terminal Reader ID", "tok": "Token ID", "trr": "Transfer ID", "tu": "Topup ID", "txi": "Tax ID",
	"txn": "Transaction ID", "txr": "Tax Rate ID", "we": "Webhook Endpoint ID", "whsec": "Webhook Secret",
}

func parseStripe(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) > 43 {
		return nil, errors.New("invalid Stripe ID length")
	}
	parts := strings.Split(input, "_")
	if len(parts) < 2 {
		return nil, errors.New("invalid Stripe ID")
	}
	prefix, value := strings.Join(parts[:len(parts)-1], "_"), parts[len(parts)-1]
	if stripeTypes[prefix] == "" || value == "" || !isBase62(value) {
		return nil, errors.New("invalid Stripe ID")
	}
	info := asciiInfo("Stripe ID", stripeTypes[prefix], input, "as ASCII", input, len(value)*8)
	info.Node1 = stringPtr(prefix)
	return info, nil
}

func parseSqidAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	instanceOptions := []sqids.Options{}
	if options.Alphabet != "" {
		instanceOptions = append(instanceOptions, sqids.Options{Alphabet: options.Alphabet})
	}
	instance, err := sqids.New(instanceOptions...)
	if err != nil {
		return nil, err
	}
	numbers := instance.Decode(input)
	if len(numbers) == 0 {
		return nil, errors.New("invalid Sqid")
	}
	version := "Default alphabet"
	if options.Alphabet != "" {
		version = "Custom alphabet"
	}
	info := asciiInfo("Sqid", version, input, "as ASCII", input, 0)
	info.Node1 = stringPtr(joinUint64(numbers))
	info.HighConfidence = options.Alphabet == "" && len(numbers) > 1
	return info, nil
}

func joinUint64(numbers []uint64) string {
	parts := make([]string, len(numbers))
	for index, number := range numbers {
		parts[index] = strconv.FormatUint(number, 10)
	}
	return strings.Join(parts, ", ")
}

func parseHashID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) == 0 || len(input) > 43 {
		return nil, errors.New("invalid Hashid")
	}
	idData := hashids.NewData()
	idData.Salt = options.Salt
	id, err := hashids.NewWithData(idData)
	if err != nil {
		return nil, err
	}
	numbers, err := id.DecodeInt64WithError(input)
	if err != nil || len(numbers) == 0 {
		return nil, errors.New("invalid Hashid")
	}
	version := "No salt"
	if options.Salt != "" {
		version = "Custom salt"
	}
	info := asciiInfo("Hashid", version, input, "as ASCII", input, 0)
	parts := make([]string, len(numbers))
	allTrailingZero := true
	for index, number := range numbers {
		parts[index] = strconv.FormatInt(number, 10)
		if index > 0 && number != 0 {
			allTrailingZero = false
		}
	}
	info.Node1 = stringPtr(strings.Join(parts, ", "))
	info.HighConfidence = len(numbers) > 1 && !allTrailingZero
	return info, nil
}

func parseDUNS(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if !regexp.MustCompile(`^[0-9]{2}-[0-9]{3}-[0-9]{4}$`).MatchString(input) {
		return nil, errors.New("invalid DUNS number")
	}
	value, _ := strconv.ParseUint(strings.ReplaceAll(input, "-", ""), 10, 32)
	info := asciiInfo("DUNS Number", "", input, "as integer", input, 32)
	setIntegerValue(info, new(big.Int).SetUint64(value), 4)
	return info, nil
}

func parseGoogleDocs(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 44 {
		return nil, errors.New("invalid Google Docs ID length")
	}
	data, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil || len(data) != 33 || data[0]>>2 != 53 || data[len(data)-1]&3 != 0 {
		return nil, errors.New("invalid Google Docs ID")
	}
	return infoFromBytes("Google Docs ID", "", input, "from base64", data, 264, 256), nil
}

func parseSWHID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	parts := strings.Split(strings.Split(input, ";")[0], ":")
	if len(parts) != 4 || parts[0] != "swh" || parts[1] != "1" || len(parts[3]) != 40 {
		return nil, errors.New("invalid SWHID")
	}
	objects := map[string]string{"snp": "Snapshot", "rel": "Release", "rev": "Revision", "dir": "Directory", "cnt": "Content"}
	object := objects[parts[2]]
	if object == "" {
		return nil, errors.New("invalid SWHID object type")
	}
	data, err := hex.DecodeString(parts[3])
	if err != nil {
		return nil, err
	}
	return infoFromBytes("SWHID (Software Hash ID)", fmt.Sprintf("Schema: 1, object type: %s", object), input, "from hex", data, 160, 160), nil
}

func parseISBN(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ReplaceAll(input, "-", "")
	if len(clean) == 13 && isDigits(clean) && isbn13Valid(clean) {
		return isbnInfo(clean, "ISBN-13"), nil
	}
	if len(clean) == 10 && isbn10Valid(clean) {
		return isbnInfo(clean, "ISBN-10"), nil
	}
	return nil, errors.New("invalid ISBN")
}

func isDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isbn13Valid(value string) bool {
	if !isDigits(value) {
		return false
	}
	sum := 0
	for index, char := range value[:12] {
		weight := 1
		if index%2 == 1 {
			weight = 3
		}
		sum += int(char-'0') * weight
	}
	return byte((10-sum%10)%10)+'0' == value[12]
}

func isbn10Valid(value string) bool {
	if len(value) != 10 {
		return false
	}
	sum := 0
	for index, char := range strings.ToUpper(value) {
		if index == 9 && char == 'X' {
			sum += 10
		} else if char >= '0' && char <= '9' {
			sum += int(char-'0') * (10 - index)
		} else {
			return false
		}
	}
	return sum%11 == 0
}

func isbnInfo(clean, kind string) *types.IDInfo {
	standard := clean
	if kind == "ISBN-13" {
		standard = clean[:3] + "-" + clean[3:4] + "-" + clean[4:9] + "-" + clean[9:12] + "-" + clean[12:]
	} else {
		standard = clean[:1] + "-" + clean[1:5] + "-" + clean[5:9] + "-" + clean[9:]
	}
	info := asciiInfo(kind, "", standard, "as ASCII, no dashes", standard, 0)
	value, _ := parseBigDecimal(clean, 64)
	if value != nil {
		setIntegerValue(info, value, 0)
	}
	return info
}

func parseIPv4(input string, options types.ParseOptions) (*types.IDInfo, error) {
	addr, err := netip.ParseAddr(input)
	if err != nil || !addr.Is4() {
		return nil, errors.New("invalid IPv4 address")
	}
	data := addr.As4()
	value := binary.BigEndian.Uint32(data[:])
	version := ""
	if addr.IsLoopback() {
		version = "Loopback"
	}
	if addr.IsPrivate() {
		version = "Private"
	}
	info := infoFromBytes("IPv4 Address", version, input, "from integer parts", data[:], 32, -1)
	setIntegerValue(info, new(big.Int).SetUint64(uint64(value)), 4)
	return info, nil
}

func parseIPv6(input string, options types.ParseOptions) (*types.IDInfo, error) {
	addr, err := netip.ParseAddr(input)
	if err != nil || !addr.Is6() {
		return nil, errors.New("invalid IPv6 address")
	}
	data := addr.As16()
	info := infoFromBytes("IPv6 Address", ternary(addr.IsLoopback(), "Loopback", ""), addr.String(), "from hex parts", data[:], 128, -1)
	setIntegerValue(info, new(big.Int).SetBytes(data[:]), 16)
	return info, nil
}

func ternary(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func parseMAC(input string, options types.ParseOptions) (*types.IDInfo, error) {
	data, err := parseMACBytes(input)
	if err != nil {
		return nil, err
	}
	prefix := new(big.Int).SetBytes(data[:3]).String()
	sequence := new(big.Int).SetBytes(data[3:]).Int64()
	info := infoFromBytes("MAC Address", "", strings.ToLower(input), "from hex parts", data, 48, -1)
	setIntegerValue(info, new(big.Int).SetBytes(data), 6)
	info.Node1 = stringPtr(fmt.Sprintf("%s, hex: %s (Manufacturer)", prefix, strings.ToLower(input)[:8]))
	info.Sequence = int64Ptr(sequence)
	return info, nil
}

func parseMACBytes(input string) ([]byte, error) {
	clean := strings.NewReplacer(":", "", "-", "", ".", "").Replace(input)
	if len(clean) != 12 {
		return nil, errors.New("invalid MAC address")
	}
	return hex.DecodeString(clean)
}

func parseIMEI(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ReplaceAll(input, "-", "")
	if len(clean) != 15 || !isDigits(clean) || !luhnValid(clean) {
		return nil, errors.New("invalid IMEI")
	}
	value, _ := parseBigDecimal(clean, 64)
	info := asciiInfo("IMEI", "", fmt.Sprintf("%s-%s-%s-%s", clean[:2], clean[2:8], clean[8:14], clean[14:]), "as ASCII, no dashes", clean, 0)
	setIntegerValue(info, value, 0)
	info.Node1 = stringPtr(fmt.Sprintf("%s (Type Allocation Code)", clean[:8]))
	info.Node2 = stringPtr(fmt.Sprintf("%s (Check Digit)", clean[14:]))
	info.Sequence = int64Ptr(parseInt64(clean[8:14]))
	return info, nil
}

func parseInt64(value string) int64 {
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

func luhnValid(value string) bool {
	sum := 0
	double := false
	for index := len(value) - 1; index >= 0; index-- {
		digit := int(value[index] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

var ibanLengths = map[string]int{
	"NO": 15, "BE": 16, "DK": 18, "FI": 18, "FO": 18, "GL": 18, "NL": 18, "SD": 18,
	"MK": 19, "SI": 19, "AT": 20, "BA": 20, "EE": 20, "KZ": 20, "LT": 20, "LU": 20, "MN": 20, "XK": 20,
	"CH": 21, "HR": 21, "LI": 21, "LV": 21, "BG": 22, "BH": 22, "CR": 22, "DE": 22, "GB": 22, "GE": 22,
	"IE": 22, "ME": 22, "RS": 22, "VA": 22, "AE": 23, "GI": 23, "IL": 23, "IQ": 23, "OM": 23, "SO": 23,
	"TL": 23, "AD": 24, "CZ": 24, "DZ": 24, "ES": 24, "MD": 24, "PK": 24, "RO": 24, "SA": 24, "SE": 24,
	"SK": 24, "TN": 24, "VG": 24, "AO": 25, "CV": 25, "GW": 25, "LY": 25, "MZ": 25, "PT": 25, "ST": 25,
	"IR": 26, "IS": 26, "TR": 26, "BI": 27, "CF": 27, "CG": 27, "CM": 27, "DJ": 27, "FR": 27, "GA": 27,
	"GQ": 27, "GR": 27, "IT": 27, "KM": 27, "MC": 27, "MG": 27, "MR": 27, "SM": 27, "TD": 27, "AL": 28,
	"AZ": 28, "BF": 28, "BJ": 28, "BY": 28, "CI": 28, "CY": 28, "DO": 28, "GT": 28, "HN": 28, "HU": 28,
	"LB": 28, "MA": 28, "ML": 28, "NE": 28, "NI": 28, "PL": 28, "SN": 28, "SV": 28, "TG": 28, "BR": 29,
	"EG": 29, "PS": 29, "QA": 29, "UA": 29, "JO": 30, "KW": 30, "MU": 30, "YE": 30, "MT": 31, "SC": 31,
	"LC": 32, "RU": 33,
}

var ibanCountries = map[string]string{"NO": "Norway", "GB": "United Kingdom", "DE": "Germany", "FR": "France", "NL": "Netherlands", "CH": "Switzerland", "IT": "Italy", "ES": "Spain", "SE": "Sweden", "RU": "Russia"}

func parseIBAN(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(input))
	if len(clean) < 15 || len(clean) > 34 || len(clean) != ibanLengths[clean[:2]] || len(clean) < 4 || !isUpperLetters(clean[:2]) || !isDigits(clean[2:4]) || !isAlphaNumeric(clean[4:]) {
		return nil, errors.New("invalid IBAN")
	}
	rearranged := clean[4:] + clean[:4]
	converted := strings.Builder{}
	for _, char := range rearranged {
		if char >= 'A' && char <= 'Z' {
			converted.WriteString(strconv.Itoa(int(char-'A') + 10))
		} else {
			converted.WriteRune(char)
		}
	}
	valid := mod97(converted.String()) == 1
	country := ibanCountries[clean[:2]]
	if country == "" {
		country = "Unknown"
	}
	info := asciiInfo("IBAN", fmt.Sprintf("%s (%s)", clean[:2], country), formatIBAN(clean), "as ASCII", formatIBAN(clean), 0)
	info.Node1 = stringPtr(fmt.Sprintf("%s (%s Checksum)", clean[2:4], ternary(valid, "Valid", "Invalid")))
	info.Node2 = stringPtr(fmt.Sprintf("%s (BBAN)", clean[4:]))
	info.HighConfidence = valid
	return info, nil
}

func isUpperLetters(value string) bool {
	for _, char := range value {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func isAlphaNumeric(value string) bool {
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func mod97(value string) int {
	remainder := 0
	for _, char := range value {
		remainder = (remainder*10 + int(char-'0')) % 97
	}
	return remainder
}

func formatIBAN(value string) string {
	parts := make([]string, 0, (len(value)+3)/4)
	for len(value) > 0 {
		end := min(len(value), 4)
		parts = append(parts, value[:end])
		value = value[end:]
	}
	return strings.Join(parts, " ")
}

func parseCommerce(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ReplaceAll(input, "-", "")
	if !isDigits(clean) {
		return nil, errors.New("invalid commerce barcode")
	}
	length := len(clean)
	if length != 8 && length != 12 && length != 13 && length != 14 {
		return nil, errors.New("invalid commerce barcode length")
	}
	if !gtinValid(clean) {
		return nil, errors.New("invalid commerce barcode checksum")
	}
	version := map[int]string{8: "EAN-8 (GTIN-8)", 12: "UPC-A (GTIN-12)", 13: "EAN-13 (GTIN-13)", 14: "GTIN-14"}[length]
	if length == 14 {
		version += ", " + map[byte]string{'0': "consumer unit", '9': "variable measure"}[clean[0]]
		if strings.HasSuffix(version, ", ") {
			version = "GTIN-14, grouping/packaging level"
		}
	}
	if length == 13 && (strings.HasPrefix(clean, "978") || strings.HasPrefix(clean, "979")) {
		if isbn, err := parseISBN(clean, options); err == nil {
			return isbn, nil
		}
	}
	standard := clean
	switch length {
	case 8:
		standard = clean[:4] + "-" + clean[4:]
	case 12:
		standard = clean[:1] + "-" + clean[1:6] + "-" + clean[6:]
	case 13:
		standard = clean[:1] + "-" + clean[1:7] + "-" + clean[7:]
	case 14:
		standard = clean[:1] + "-" + clean[1:4] + "-" + clean[4:8] + "-" + clean[8:]
	}
	info := asciiInfo("Commerce Barcode", version, standard, "as ASCII", clean, 0)
	value, _ := parseBigDecimal(clean, 64)
	setIntegerValue(info, value, 0)
	return info, nil
}

func gtinValid(value string) bool {
	if len(value) < 2 {
		return false
	}
	sum := 0
	for index := len(value) - 2; index >= 0; index-- {
		digit := int(value[index] - '0')
		if (len(value)-1-index)%2 == 1 {
			digit *= 3
		}
		sum += digit
	}
	return byte((10-sum%10)%10)+'0' == value[len(value)-1]
}

var vinManufacturers = map[string]string{"1HG": "Honda", "1FA": "Ford", "1G1": "Chevrolet", "2HG": "Honda", "JHM": "Honda", "WBA": "BMW", "WB": "BMW", "WVW": "Volkswagen", "5YJ": "Tesla"}

func parseVIN(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ToUpper(strings.ReplaceAll(input, "-", ""))
	if len(clean) != 17 {
		return nil, errors.New("invalid VIN length")
	}
	for _, char := range clean {
		if !((char >= 'A' && char <= 'H') || (char >= 'J' && char <= 'N') || char == 'P' || (char >= 'R' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return nil, errors.New("invalid VIN character")
		}
	}
	country := map[byte]string{'1': "United States", '2': "Canada", '3': "Mexico", 'J': "Japan", 'K': "South Korea", 'L': "China", 'S': "United Kingdom", 'W': "Germany", 'Z': "Italy"}[clean[0]]
	if country == "" {
		country = "Unknown"
	}
	manufacturer := vinManufacturers[clean[:3]]
	version := country
	if manufacturer != "" {
		version = manufacturer + ", " + country
	}
	info := asciiInfo("VIN (Vehicle Identification Number)", version, clean[:3]+"-"+clean[3:9]+"-"+clean[9:11]+"-"+clean[11:], "as ASCII", clean, 0)
	info.Node1 = stringPtr(fmt.Sprintf("%s (Model)", clean[3:8]))
	info.Node3 = stringPtr(fmt.Sprintf("Factory: %s", clean[10:11]))
	if number := parseInt64(clean[11:]); number > 0 {
		info.Sequence = int64Ptr(number)
	}
	return info, nil
}

func parseTID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 13 {
		return nil, errors.New("invalid TID length")
	}
	value, err := decodeBase(input, "234567abcdefghijklmnopqrstuvwxyz")
	if err != nil || value.BitLen() > 64 {
		return nil, errors.New("invalid TID")
	}
	number := value.Uint64()
	raw := bitsU64(number, 1, 53)
	info := infoFromBytes("TID (AT Protocol, Bluesky)", "", input, "from base32", bigEndianBytes(value, 8), 64, 0)
	setIntegerValue(info, value, 8)
	info.Timestamp, info.DateTime = timestampInfo(int64(raw/1000), epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Clock ID)", bitsU64(number, 54, 10)))
	return info, nil
}

func parseThreads(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var number uint64
	parsed := "as integer"
	if len(input) <= 12 {
		padded := strings.Repeat("A", 12-len(input)) + input
		data, err := base64.RawURLEncoding.DecodeString(padded)
		if err != nil || len(data) != 9 {
			return nil, errors.New("invalid Thread ID")
		}
		number = binary.BigEndian.Uint64(data[1:])
		parsed = "from base64"
	} else {
		var err error
		number, err = strconv.ParseUint(input, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	info := infoFromBytes("Thread ID (Meta Threads)", "", input, parsed, bigEndianBytes(new(big.Int).SetUint64(number), 8), 64, 0)
	setIntegerValue(info, new(big.Int).SetUint64(number), 8)
	raw := bitsU64(number, 0, 41)
	info.Timestamp, info.DateTime = timestampInfo(int64(raw), epochMillis(options, 1314220021721))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Shard ID)", bitsU64(number, 41, 13)))
	info.Sequence = int64Ptr(int64(bitsU64(number, 54, 10)))
	info.HighConfidence = len(input) == 11 || len(input) == 12
	return info, nil
}

func parseSnowID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	fromBase62 := false
	value, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		encoded, decodeErr := decodeBase(input, base62Alphabet)
		if decodeErr != nil || encoded.BitLen() > 64 {
			return nil, errors.New("invalid SnowID")
		}
		value, fromBase62 = encoded.Uint64(), true
	}
	info := infoFromBytes("SnowID", "", input, ternary(fromBase62, "from base62", "as integer"), bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, 0)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Standard = encodeBase(new(big.Int).SetUint64(value), base62Alphabet, 0)
	info.Timestamp, info.DateTime = timestampInfo(int64(bitsU64(value, 0, 42)), epochMillis(options, 1704067200000))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Node ID)", bitsU64(value, 42, 10)))
	info.Sequence = int64Ptr(int64(bitsU64(value, 52, 12)))
	info.HighConfidence = fromBase62
	return info, nil
}

func parseMist(input string, options types.ParseOptions) (*types.IDInfo, error) {
	value, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}
	info := infoFromBytes("Mist", "", input, "as integer", bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, -1)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Sequence = int64Ptr(int64(bitsU64(value, 1, 47)))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Salt 1)", bitsU64(value, 48, 8)))
	info.Node2 = stringPtr(fmt.Sprintf("%d (Salt 2)", bitsU64(value, 56, 8)))
	info.HighConfidence = *info.Sequence > 0
	return info, nil
}

func parseNano64(input string, options types.ParseOptions) (*types.IDInfo, error) {
	clean := strings.ToUpper(strings.ReplaceAll(input, "-", ""))
	if len(clean) != 16 {
		return nil, errors.New("invalid Nano64")
	}
	data, err := hex.DecodeString(clean)
	if err != nil {
		return nil, err
	}
	number := binary.BigEndian.Uint64(data)
	info := infoFromBytes("Nano64", "", fmt.Sprintf("%s-%s", clean[:11], clean[11:]), "from hex", data, 64, 20)
	setIntegerValue(info, new(big.Int).SetUint64(number), 8)
	info.Timestamp, info.DateTime = timestampInfo(int64(bitsU64(number, 0, 44)), epochMillis(options, 0))
	return info, nil
}

func parseASIN(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 10 {
		return nil, errors.New("invalid ASIN")
	}
	value, err := decodeBase(strings.ToUpper(input), "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err != nil {
		return nil, err
	}
	info := asciiInfo("ASIN (Amazon)", "", input, "as integer, from base36", input, 0)
	setIntegerValue(info, value, 8)
	info.Sequence = int64Ptr(value.Int64())
	info.HighConfidence = value.Cmp(new(big.Int).SetUint64(1117159523352576)) >= 0
	return info, nil
}

func parseYouTube(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 11 {
		return nil, errors.New("invalid YouTube ID")
	}
	data, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil || len(data) != 8 {
		return nil, errors.New("invalid YouTube ID")
	}
	info := infoFromBytes("YouTube Video ID", "", input, "from base64", data, 64, 64)
	setIntegerValue(info, new(big.Int).SetBytes(data), 8)
	return info, nil
}

func decodeBase58Bytes(input string) ([]byte, error) {
	value, err := decodeBase(input, "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	if err != nil {
		return nil, err
	}
	data := value.Bytes()
	leading := 0
	for leading < len(input) && input[leading] == '1' {
		leading++
	}
	if leading > 0 {
		data = append(make([]byte, leading), data...)
	}
	return data, nil
}

func parseBitcoin(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if strings.HasPrefix(strings.ToLower(input), "bc1") {
		return parseBitcoinBech32(input)
	}
	data, err := decodeBase58Bytes(input)
	if err != nil || len(data) != 25 {
		return nil, errors.New("invalid Bitcoin address")
	}
	payload, checksum := data[:21], data[21:]
	digest := sha256.Sum256(payload)
	digest = sha256.Sum256(digest[:])
	if !strings.EqualFold(hex.EncodeToString(checksum), hex.EncodeToString(digest[:4])) {
		return nil, errors.New("invalid Bitcoin checksum")
	}
	version := map[byte]string{0: "Legacy (P2PKH)", 5: "Nested SegWit (P2SH)"}[data[0]]
	if version == "" {
		return nil, errors.New("unsupported Bitcoin address version")
	}
	idType := "Bitcoin Address"
	if input == "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa" {
		idType = "Bitcoin Address (from Satoshi Nakamoto)"
	}
	info := infoFromBytes(idType, version, input, "from base58check", data, 200, 160)
	info.Node1 = stringPtr(fmt.Sprintf("%s (Checksum)", hex.EncodeToString(checksum)))
	return info, nil
}

func parseBitcoinBech32(input string) (*types.IDInfo, error) {
	lower := strings.ToLower(input)
	if len(lower) < 14 || len(lower) > 74 || !strings.HasPrefix(lower, "bc1") || input != lower && input != strings.ToUpper(input) {
		return nil, errors.New("invalid Bitcoin bech32 address")
	}
	separator := strings.LastIndexByte(lower, '1')
	if separator < 2 || len(lower)-separator-1 < 6 {
		return nil, errors.New("invalid Bitcoin bech32 address")
	}
	dataPart := lower[separator+1:]
	values := make([]byte, len(dataPart))
	const alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	for index, char := range dataPart {
		position := strings.IndexRune(alphabet, char)
		if position < 0 {
			return nil, errors.New("invalid Bitcoin bech32 alphabet")
		}
		values[index] = byte(position)
	}
	version := values[0]
	if version > 16 {
		return nil, errors.New("invalid Bitcoin bech32 version")
	}
	program, ok := convertBits(values[1:len(values)-6], 5, 8, false)
	if !ok || len(program) < 2 || len(program) > 40 || version == 0 && len(program) != 20 && len(program) != 32 {
		return nil, errors.New("invalid Bitcoin witness program")
	}
	name := "Native SegWit (Bech32)"
	if version == 1 {
		name = "Taproot (P2TR)"
	} else if version > 1 {
		name = fmt.Sprintf("SegWit v%d", version)
	}
	checksum := values[len(values)-6:]
	bytes := append([]byte{version}, program...)
	info := infoFromBytes("Bitcoin Address", name, lower, "from bech32", bytes, len(bytes)*8, len(program)*8)
	info.Node1 = stringPtr(fmt.Sprintf("%s (Checksum)", dataPart[len(dataPart)-6:]))
	_ = checksum
	return info, nil
}

func bech32Valid(hrp string, values []byte) bool {
	all := make([]byte, 0, len(hrp)+1+len(values))
	for _, char := range hrp {
		all = append(all, byte(char>>5))
	}
	all = append(all, 0)
	for _, char := range hrp {
		all = append(all, byte(char&31))
	}
	all = append(all, values...)
	polymod := uint32(1)
	generator := [...]uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	for _, value := range all {
		top := polymod >> 25
		polymod = (polymod&0x1ffffff)<<5 ^ uint32(value)
		for bit := 0; bit < 5; bit++ {
			if top&(1<<bit) != 0 {
				polymod ^= generator[bit]
			}
		}
	}
	return polymod == 1 || polymod == 0x2bc830a3
}

func convertBits(data []byte, from, to uint, pad bool) ([]byte, bool) {
	acc, bits := uint32(0), uint(0)
	result := make([]byte, 0, len(data)*int(from)/int(to)+1)
	max := uint32((1 << to) - 1)
	for _, value := range data {
		if value>>from != 0 {
			return nil, false
		}
		acc = (acc << from) | uint32(value)
		bits += from
		for bits >= to {
			bits -= to
			result = append(result, byte((acc>>bits)&max))
		}
	}
	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(to-bits))&max))
		}
	} else if bits >= from || byte((acc<<(to-bits))&max) != 0 {
		return nil, false
	}
	return result, true
}

func parseEthereum(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 42 || !strings.HasPrefix(input, "0x") && !strings.HasPrefix(input, "0X") {
		return nil, errors.New("invalid Ethereum address")
	}
	part := input[2:]
	data, err := hex.DecodeString(part)
	if err != nil || len(data) != 20 {
		return nil, errors.New("invalid Ethereum address")
	}
	version := "No checksum"
	valid := true
	if part != strings.ToLower(part) && part != strings.ToUpper(part) {
		hash := sha3.NewLegacyKeccak256()
		hash.Write([]byte(strings.ToLower(part)))
		digest := hash.Sum(nil)
		for index, char := range strings.ToLower(part) {
			if char >= 'a' && char <= 'f' {
				wantUpper := digest[index/2]>>(uint(4*(1-index%2)))&0xf >= 8
				if wantUpper != (input[2+index] >= 'A' && input[2+index] <= 'F') {
					valid = false
				}
			}
		}
		if valid {
			version = "EIP-55 (valid checksum)"
		} else {
			version = "EIP-55 (invalid checksum)"
		}
	}
	info := infoFromBytes("Ethereum Address", version, input, "from hex", data, 160, 160)
	info.HighConfidence = valid
	return info, nil
}

func parseIPFS(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) == 46 && strings.HasPrefix(input, "Qm") {
		data, err := decodeBase58Bytes(input)
		if err != nil || len(data) != 34 || data[0] != 0x12 || data[1] != 0x20 {
			return nil, errors.New("invalid CID v0")
		}
		return infoFromBytes("IPFS", "CID v0", input, "from base58btc", data, len(data)*8, 256), nil
	}
	if len(input) < 3 || input[0] != 'b' {
		return nil, errors.New("invalid CID v1")
	}
	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(input[1:]))
	if err != nil || len(data) < 4 || data[0] != 1 {
		return nil, errors.New("invalid CID v1")
	}
	version := "CID v1"
	if strings.HasPrefix(input, "k51") {
		version = "CID v1 (IPNS)"
	}
	return infoFromBytes("IPFS", version, input, "from base32", data, len(data)*8, 256), nil
}

func parseH3(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 15 {
		return nil, errors.New("invalid H3 index")
	}
	value, err := strconv.ParseUint(input, 16, 64)
	if err != nil || (value>>59)&0xf != 1 || value&0x7 != 7 {
		return nil, errors.New("invalid H3 index")
	}
	resolution := (value >> 52) & 0xf
	baseCell := (value >> 45) & 0x7f
	info := infoFromBytes("H3 Grid System", "H3 Cell (Mode 1)", input, "from hex", bigEndianBytes(new(big.Int).SetUint64(value), 8), 64, -1)
	setIntegerValue(info, new(big.Int).SetUint64(value), 8)
	info.Node1 = stringPtr(fmt.Sprintf("Resolution: %d, base cell: %d", resolution, baseCell))
	return info, nil
}

func parseObjectIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 24 {
		return nil, errors.New("invalid MongoDB ObjectId")
	}
	data, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}
	info := infoFromBytes("MongoDB ObjectId", "", input, "from hex", data, 96, 40)
	setIntegerValue(info, new(big.Int).SetBytes(data), 12)
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint32(data[:4]))*1000, epochMillis(options, 0))
	info.Sequence = int64Ptr(int64(binary.BigEndian.Uint32(append([]byte{0}, data[9:12]...))))
	return info, nil
}

func parseKSUIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	var data []byte
	version, parsed := "Base62-encoded", "from base62"
	if len(input) == 27 {
		value, err := decodeBase(input, "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
		if err != nil || value.BitLen() > 160 {
			return nil, errors.New("invalid KSUID")
		}
		data = bigEndianBytes(value, 20)
	} else if len(input) == 40 {
		var err error
		data, err = hex.DecodeString(input)
		if err != nil || len(data) != 20 {
			return nil, errors.New("invalid KSUID")
		}
		version, parsed = "Hex-encoded", "from hex"
	} else {
		return nil, errors.New("invalid KSUID length")
	}
	info := infoFromBytes("KSUID", version, encodeBase(new(big.Int).SetBytes(data), "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 27), parsed, data, 160, 128)
	setIntegerValue(info, new(big.Int).SetBytes(data), 20)
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint32(data[:4]))*1000, epochMillis(options, 1400000000000))
	return info, nil
}

func parseXIDAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	if len(input) != 20 {
		return nil, errors.New("invalid Xid")
	}
	value, err := xid.FromString(input)
	if err != nil {
		return nil, errors.New("invalid Xid")
	}
	data := append([]byte(nil), value[:]...)
	info := infoFromBytes("Xid", "", input, "from base32hex", data, 96, 56)
	setIntegerValue(info, new(big.Int).SetBytes(data), 12)
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint32(data[:4]))*1000, epochMillis(options, 0))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Machine ID)", new(big.Int).SetBytes(data[4:7])))
	info.Node2 = stringPtr(fmt.Sprintf("%d (Process ID)", new(big.Int).SetBytes(data[7:9])))
	info.Sequence = int64Ptr(int64(binary.BigEndian.Uint32(append([]byte{0}, data[9:12]...))))
	return info, nil
}

func parseOrderlyID(input string, options types.ParseOptions) (*types.IDInfo, error) {
	parts := strings.SplitN(input, "_", 2)
	if len(parts) != 2 || len(parts[0]) < 2 || len(parts[0]) > 31 {
		return nil, errors.New("invalid OrderlyID prefix")
	}
	valueParts := strings.Split(parts[1], "-")
	if len(valueParts[0]) != 32 || len(valueParts) > 2 || len(valueParts) == 2 && len(valueParts[1]) != 4 {
		return nil, errors.New("invalid OrderlyID payload")
	}
	value, err := decodeBase(strings.ToUpper(valueParts[0]), crockfordAlphabet)
	if err != nil || value.BitLen() > 160 {
		return nil, errors.New("invalid OrderlyID payload")
	}
	data := bigEndianBytes(value, 20)
	info := infoFromBytes("OrderlyID, type "+parts[0], "", input, "from Crockford's base32", data, 160, 60)
	setIntegerValue(info, value, 20)
	flags := data[6]
	version := "unknown"
	if flags>>6 == 0 {
		version = "1"
	}
	privacy := "privacy off"
	if (flags<<2)>>7 == 1 {
		privacy = "privacy on"
	}
	info.Version = fmt.Sprintf("Version %s, %s, %s", version, privacy, ternary(len(valueParts) == 2, "with checksum", "no checksum"))
	info.Timestamp, info.DateTime = timestampInfo(int64(binary.BigEndian.Uint64(append([]byte{0, 0}, data[:6]...))), epochMillis(options, 1577836800000))
	info.Node1 = stringPtr(fmt.Sprintf("%d (Tenant)", binary.BigEndian.Uint16(data[7:9])))
	info.Node2 = stringPtr(fmt.Sprintf("%d (Shard)", uint16(data[10]&0x0f)<<8|uint16(data[11])))
	info.Sequence = int64Ptr(int64(uint16(data[9])<<4 | uint16(data[10])>>4))
	return info, nil
}

func parseAlignedByName(name, input string, options types.ParseOptions) (*types.IDInfo, error) {
	switch name {
	case "uuid":
		return parseUUIDAligned(input, options)
	case "uuid-int":
		return parseUUIDInteger(input, options)
	case "uuid-b64":
		return parseUUIDBase64(input, options)
	case "uuid25":
		return parseUUID25(input, options)
	case "shortuuid":
		return parseShortUUIDAligned(input, options)
	case "ulid":
		return parseULIDAligned(input, options)
	case "julid":
		return parseJulid(input, options)
	case "upid":
		return parseUPID(input, options)
	case "sandflake":
		return parseSandflake(input, options)
	case "timeflake":
		return parseTimeflake(input, options)
	case "flake":
		return parseFlake(input, options)
	case "scru128":
		return parseSCRU128Aligned(input, options)
	case "scru64":
		return parseSCRU64Aligned(input, options)
	case "tsid":
		return parseTSIDAligned(input, options)
	case "objectid":
		return parseObjectIDAligned(input, options)
	case "ksuid":
		return parseKSUIDAligned(input, options)
	case "xid":
		return parseXIDAligned(input, options)
	case "cuid1":
		return parseCUID1(input, options)
	case "cuid2":
		return parseCUID2Aligned(input, options)
	case "nanoid":
		return parseNanoIDAligned(input, options)
	case "nuid":
		return parseNUIDAligned(input, options)
	case "typeid":
		return parseTypeIDAligned(input, options)
	case "pushid":
		return parsePushIDAligned(input, options)
	case "sqid":
		return parseSqidAligned(input, options)
	case "hashid":
		return parseHashID(input, options)
	case "youtube":
		return parseYouTube(input, options)
	case "stripe":
		return parseStripe(input, options)
	case "datadog":
		return parseDatadog(input, options)
	case "breezeid":
		return parseBreezeID(input, options)
	case "puid":
		if len(input) == 24 {
			return parsePUID(input, options)
		}
		return parseShortPUID(input, options)
	case "shortpuid":
		return parseShortPUID(input, options)
	case "tid":
		return parseTID(input, options)
	case "threads":
		return parseThreads(input, options)
	case "snowid":
		return parseSnowID(input, options)
	case "duns":
		return parseDUNS(input, options)
	case "asin":
		return parseASIN(input, options)
	case "gdocs":
		return parseGoogleDocs(input, options)
	case "slack":
		return parseSlack(input, options)
	case "spotify":
		return parseSpotify(input, options)
	case "nano64":
		return parseNano64(input, options)
	case "orderlyid":
		return parseOrderlyID(input, options)
	case "swhid":
		return parseSWHID(input, options)
	case "iban":
		return parseIBAN(input, options)
	case "commerce":
		return parseCommerce(input, options)
	case "vin":
		return parseVIN(input, options)
	case "bitcoin":
		return parseBitcoin(input, options)
	case "ethereum":
		return parseEthereum(input, options)
	case "ipfs":
		return parseIPFS(input, options)
	case "ipv4":
		return parseIPv4(input, options)
	case "ipv6":
		return parseIPv6(input, options)
	case "mac":
		return parseMAC(input, options)
	case "imei":
		return parseIMEI(input, options)
	case "isbn":
		return parseISBN(input, options)
	case "h3":
		return parseH3(input, options)
	case "mist":
		return parseMist(input, options)
	case "hash":
		return parseHashAligned(input, options)
	case "unixtime":
		return parseUnixAligned(input, options, unixAuto, false)
	case "unix-recent":
		return parseUnixAligned(input, options, unixAuto, true)
	case "unix-s":
		return parseUnixAligned(input, options, unixSeconds, false)
	case "unix-ms":
		return parseUnixAligned(input, options, unixMillis, false)
	case "unix-us":
		return parseUnixAligned(input, options, unixMicros, false)
	case "unix-ns":
		return parseUnixAligned(input, options, unixNanos, false)
	case "sf-twitter", "sf-mastodon", "sf-discord", "sf-instagram", "sf-linkedin", "sf-sony", "sf-spaceflake", "sf-frostflake", "sf-flakeid", "sf-simpleflake":
		return parseSnowflakeAligned(input, options, name)
	case "comb":
		return parseComb(input, options)
	default:
		return nil, fmt.Errorf("unknown format %q", name)
	}
}
