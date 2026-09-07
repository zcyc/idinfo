package parsers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zcyc/idinfo/internal/types"
)

var alignedFormats = []string{
	"uuid", "uuid-b64", "uuid25", "shortuuid", "uuid-int", "ulid", "julid", "sandflake", "upid", "mongodb", "ksuid", "xid",
	"scru128", "scru64", "timeflake", "flake", "tsid", "nuid", "typeid", "pushid", "orderlyid", "threads", "snowid", "nano64",
	"sqid", "hashid", "youtube", "stripe", "datadog", "snowflake", "unix", "hash", "ipfs", "breezeid", "puid", "shortpuid",
	"ipv4", "ipv6", "mac", "isbn10", "tid", "duns", "asin", "nanoid", "cuid1", "cuid2", "h3", "imei", "gdocs", "slack",
	"spotify", "swhid", "iban", "bitcoin", "ethereum", "commerce", "vin", "mist", "comb",
}

func parseExisting(name, input string) (*types.IDInfo, error) {
	var parser types.IDParser
	switch name {
	case "base58":
		parser = &Base58Parser{}
	case "base32":
		parser = &Base32Parser{}
	default:
		return nil, fmt.Errorf("unknown legacy format %q", name)
	}
	return parser.Parse(input)
}

func parseNameWithOptions(name, input string, options types.ParseOptions) (*types.IDInfo, error) {
	if name == "base58" || name == "base32" {
		return parseExisting(name, input)
	}
	if name == "snowflake" {
		return parseSnowflakeAuto(input, options)
	}
	if name == "isbn10" {
		info, err := parseISBN(input, options)
		if err != nil || info.IDType != "ISBN-10" {
			return nil, errors.New("invalid ISBN-10")
		}
		return info, nil
	}
	if name == "shortpuid" {
		return parseShortPUID(input, options)
	}
	return parseAlignedByName(name, input, options)
}

func tryAligned(options types.ParseOptions, input string, names ...string) (*types.IDInfo, bool) {
	for _, name := range names {
		if info, err := parseNameWithOptions(name, input, options); err == nil {
			return info, true
		}
	}
	return nil, false
}

func autoDetectAligned(input string, options types.ParseOptions) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)
	if info, ok := tryAligned(options, input, "iban"); ok {
		return info, nil
	}
	if _, err := parseBigDecimal(input, 128); err == nil {
		if info, ok := tryAligned(options, input, "isbn", "imei", "unix-recent", "snowflake", "uuid-int"); ok {
			return info, nil
		}
	} else {
		var names []string
		switch len([]rune(input)) {
		case 62:
			names = []string{"bitcoin"}
		case 56, 64, 96, 128:
			names = []string{"hash"}
		case 44:
			names = []string{"gdocs"}
		case 42:
			names = []string{"ethereum", "bitcoin"}
		case 40:
			names = []string{"ksuid"}
		case 34:
			names = []string{"bitcoin"}
		case 32, 36:
			names = []string{"datadog", "uuid"}
		case 27:
			names = []string{"upid", "ksuid"}
		case 26:
			if info, ok := tryAligned(options, input, "julid"); ok && info.Sequence != nil && *info.Sequence < 128 {
				return info, nil
			}
			names = []string{"ulid"}
		case 25:
			names = []string{"cuid1", "scru128"}
		case 24:
			names = []string{"mongodb", "puid", "uuid-b64"}
		case 22:
			names = []string{"shortuuid", "timeflake", "uuid-b64", "nuid", "spotify"}
		case 21:
			names = []string{"nanoid"}
		case 20:
			names = []string{"xid", "stripe", "pushid"}
		case 18:
			names = []string{"flake"}
		case 17:
			names = []string{"vin", "nano64"}
		case 16:
			names = []string{"nano64"}
		case 15:
			names = []string{"h3"}
		case 14:
			names = []string{"shortpuid"}
		case 13:
			names = []string{"tid", "tsid"}
		case 12:
			names = []string{"scru64", "shortpuid"}
		case 11:
			names = []string{"slack", "youtube", "snowid"}
		case 10:
			names = []string{"asin", "snowid"}
		}
		if info, ok := tryAligned(options, input, names...); ok {
			return info, nil
		}
		if info, ok := tryAligned(options, input, "vin", "orderlyid", "isbn", "commerce", "typeid", "ipfs", "stripe", "breezeid", "ipv4", "ipv6", "mac", "cuid2", "sqid", "snowid", "duns", "threads", "imei", "hashid", "nanoid", "swhid"); ok {
			return info, nil
		}
	}
	return nil, errors.New("unknown ID type")
}

func ParseIDWithOptions(input, force string, options types.ParseOptions) []*types.IDInfo {
	input = strings.TrimSpace(input)
	if force != "" {
		name := strings.ToLower(strings.TrimSpace(force))
		if name == "puid" {
			if info, err := parsePUIDAny(input, options); err == nil {
				return []*types.IDInfo{info}
			}
		}
		if info, err := parseNameWithOptions(name, input, options); err == nil {
			return []*types.IDInfo{info}
		}
		return nil
	}
	info, err := autoDetectAligned(input, options)
	if err != nil {
		return nil
	}
	return []*types.IDInfo{info}
}

func ParseAllWithOptions(input string, options types.ParseOptions) []*types.IDInfo {
	input = strings.TrimSpace(input)
	var results []*types.IDInfo
	for _, name := range alignedFormats {
		info, err := parseNameWithOptions(name, input, options)
		if err == nil && info.HighConfidence {
			results = append(results, info)
		}
	}
	return results
}

func ParseTimesWithOptions(input string, options types.ParseOptions) []*types.IDInfo {
	input = strings.TrimSpace(input)
	var results []*types.IDInfo
	for _, name := range alignedFormats {
		if info, err := parseNameWithOptions(name, input, options); err == nil && info.DateTime != nil {
			results = append(results, info)
		}
	}
	if _, err := parseBigDecimal(input, 64); err == nil {
		for _, name := range []string{"sf-twitter", "sf-discord", "sf-instagram", "sf-sony", "sf-spaceflake", "sf-linkedin", "sf-mastodon", "sf-frostflake", "sf-flakeid", "sf-simpleflake"} {
			if info, err := parseNameWithOptions(name, input, options); err == nil && info.DateTime != nil {
				results = append(results, info)
			}
		}
	}
	return results
}
