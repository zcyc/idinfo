package parsers

import (
	"errors"
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

var canonicalForceFormats = []string{
	"uuid", "shortuuid", "uuid-int", "uuid-b64", "uuid25", "ulid", "sandflake", "julid", "upid", "comb", "timeflake", "flake",
	"scru128", "scru64", "mongodb", "ksuid", "xid", "cuid1", "cuid2", "nanoid", "tsid", "sqid", "hashid", "youtube", "stripe",
	"datadog", "nuid", "typeid", "breezeid", "puid", "pushid", "tid", "threads", "duns", "asin", "snowid", "gdocs", "slack",
	"spotify", "nano64", "orderlyid", "swhid", "iban", "commerce", "vin", "bitcoin", "ethereum", "sf-twitter", "sf-mastodon",
	"sf-discord", "sf-instagram", "sf-linkedin", "sf-sony", "sf-spaceflake", "sf-frostflake", "sf-flakeid", "sf-simpleflake",
	"mist", "unix", "unix-s", "unix-ms", "unix-us", "unix-ns", "hash", "ipfs", "ipv4", "ipv6", "mac", "imei", "isbn", "h3",
}

func CanonicalForceFormats() []string {
	return append([]string(nil), canonicalForceFormats...)
}

func IsCanonicalForceFormat(name string) bool {
	for _, format := range canonicalForceFormats {
		if format == name {
			return true
		}
	}
	return false
}

func parseNameWithOptions(name, input string, options types.ParseOptions) (*types.IDInfo, error) {
	var (
		info *types.IDInfo
		err  error
	)
	if name == "snowflake" {
		info, err = parseSnowflakeAuto(input, options)
	} else if name == "isbn10" {
		info, err := parseISBN(input, options)
		if err != nil || info.IDType != "ISBN-10" {
			return nil, errors.New("invalid ISBN-10")
		}
		return colorize(info), nil
	} else if name == "shortpuid" {
		info, err = parseShortPUID(input, options)
	} else {
		info, err = parseAlignedByName(name, input, options)
	}
	if err != nil {
		return nil, err
	}
	return colorize(info), nil
}

func colorize(info *types.IDInfo) *types.IDInfo {
	if info == nil || info.Size <= 0 || info.Hex == "" {
		return info
	}
	info.ColorMap = colorMapFor(info)
	return info
}

func colorRun(code byte, length int) string {
	return strings.Repeat(string(code), length)
}

func colorMap(parts ...string) string {
	return strings.Join(parts, "")
}

func colorMapFor(info *types.IDInfo) string {
	id, version := info.IDType, info.Version
	defaultMap := colorRun('2', info.Size)
	switch {
	case strings.Contains(id, "UUID") || id == "NCS UUID" || id == "Microsoft GUID" || id == "Unknown UUID-like":
		if strings.HasPrefix(version, "1 ") || strings.HasPrefix(version, "6 ") {
			return colorMap(colorRun('3', 48), colorRun('1', 4), colorRun('3', 12), colorRun('0', 2), colorRun('6', 16), colorRun('4', 48))
		}
		if strings.HasPrefix(version, "7 ") {
			return colorMap(colorRun('3', 48), colorRun('1', 4), colorRun('2', 12), colorRun('0', 2), colorRun('2', 62))
		}
		if id == "NCS UUID" || id == "Microsoft GUID" || id == "Unknown UUID-like" {
			return colorMap(colorRun('2', 64), colorRun('0', 2), colorRun('2', 62))
		}
		return colorMap(colorRun('2', 48), colorRun('1', 4), colorRun('2', 12), colorRun('0', 2), colorRun('2', 62))
	case id == "ULID" || id == "ULID wrapped in UUID":
		return colorMap(colorRun('3', 48), colorRun('2', 80))
	case id == "Julid":
		return colorMap(colorRun('3', 48), colorRun('6', 16), colorRun('2', 64))
	case strings.HasPrefix(id, "UPID"):
		return colorMap(colorRun('3', 40), colorRun('2', 64), colorRun('4', 20), colorRun('1', 4))
	case strings.HasPrefix(id, "Sandflake"):
		return colorMap(colorRun('3', 48), colorRun('4', 32), colorRun('6', 24), colorRun('2', 24))
	case strings.HasPrefix(id, "Timeflake"):
		return colorMap(colorRun('3', 48), colorRun('2', 80))
	case strings.HasPrefix(id, "Flake"):
		return colorMap(colorRun('3', 64), colorRun('4', 48), colorRun('6', 16))
	case strings.HasPrefix(id, "SCRU128"):
		return colorMap(colorRun('3', 48), colorRun('2', 80))
	case id == "SCRU64":
		return colorMap(colorRun('0', 2), colorRun('3', 38), colorRun('4', 24))
	case id == "TSID":
		return colorMap(colorRun('3', 42), colorRun('2', 22))
	case id == "MongoDB ObjectId":
		return colorMap(colorRun('3', 32), colorRun('2', 40), colorRun('6', 24))
	case id == "KSUID":
		return colorMap(colorRun('3', 32), colorRun('2', 128))
	case id == "Xid":
		return colorMap(colorRun('3', 32), colorRun('4', 24), colorRun('5', 16), colorRun('6', 24))
	case id == "COMB":
		return combColorMap(version)
	case id == "CUID" && version == "1":
		return colorMap(colorRun('1', 8), colorRun('3', 64), colorRun('6', 32), colorRun('4', 32), colorRun('2', 64))
	case id == "TypeID":
		return colorMap(colorRun('3', 48), colorRun('1', 4), colorRun('2', 12), colorRun('0', 2), colorRun('2', 62))
	case id == "PushID (Firebase)":
		return colorMap(colorRun('3', 48), colorRun('2', 72))
	case id == "Datadog Trace ID":
		return colorMap(colorRun('3', 32), colorRun('0', 32), colorRun('2', 64))
	case id == "Thread ID (Meta Threads)":
		return colorMap(colorRun('3', 41), colorRun('4', 13), colorRun('6', 10))
	case id == "SnowID":
		return colorMap(colorRun('3', 42), colorRun('4', 10), colorRun('6', 12))
	case id == "Nano64":
		return colorMap(colorRun('3', 44), colorRun('2', 20))
	case id == "NUID":
		return colorMap(colorRun('2', 96), colorRun('4', 80))
	case id == "Puid":
		switch info.Size {
		case 192:
			return colorMap(colorRun('3', 64), colorRun('4', 48), colorRun('5', 32), colorRun('6', 48))
		case 112:
			return colorMap(colorRun('3', 96), colorRun('4', 16))
		case 96:
			return colorRun('3', 96)
		}
	case id == "TID (AT Protocol, Bluesky)":
		return colorMap(colorRun('0', 1), colorRun('3', 53), colorRun('4', 10))
	case id == "ASIN (Amazon)":
		return colorRun('6', info.Size)
	case id == "Google Docs ID":
		return colorMap(colorRun('0', 6), colorRun('2', 256), colorRun('0', 2))
	case id == "IPv4 Address" || id == "IPv6 Address":
		return colorRun('0', info.Size)
	case id == "MAC Address":
		return colorMap(colorRun('4', 24), colorRun('6', 24))
	case id == "IMEI":
		return colorMap(colorRun('4', 64), colorRun('6', 48), colorRun('5', 8))
	case id == "IBAN":
		clean := strings.ReplaceAll(info.Standard, " ", "")
		return colorMap(colorRun('1', 16), colorRun('4', 16), colorRun('5', (len(clean)-4)*8))
	case id == "Commerce Barcode":
		switch {
		case strings.HasPrefix(version, "EAN-8"):
			return colorMap(colorRun('4', 24), colorRun('2', 32), colorRun('0', 8))
		case strings.HasPrefix(version, "UPC-A"):
			return colorMap(colorRun('4', 8), colorRun('2', 80), colorRun('0', 8))
		case strings.HasPrefix(version, "EAN-13"):
			return colorMap(colorRun('4', 24), colorRun('2', 72), colorRun('0', 8))
		case strings.HasPrefix(version, "GTIN-14"):
			return colorMap(colorRun('1', 8), colorRun('4', 24), colorRun('2', 72), colorRun('0', 8))
		}
	case id == "ISBN-10" || id == "ISBN-13":
		return isbnColorMap(info.Standard)
	case id == "VIN (Vehicle Identification Number)":
		return colorMap(colorRun('1', 24), colorRun('4', 40), colorRun('0', 8), colorRun('5', 8), colorRun('7', 8), colorRun('6', 48))
	case id == "Stripe ID":
		if separator := strings.LastIndexByte(info.Standard, '_'); separator >= 0 {
			prefixBits := separator * 8
			return colorMap(colorRun('4', prefixBits), colorRun('0', 8), colorRun('2', info.Size-prefixBits-8))
		}
	case id == "Slack ID":
		offset := 1
		if strings.HasPrefix(info.Standard, "Wf") {
			offset = 2
		}
		return colorMap(colorRun('4', offset*8), colorRun('2', (len(info.Standard)-offset)*8))
	case id == "Breeze ID":
		var result strings.Builder
		for _, char := range info.Standard {
			if char == '-' {
				result.WriteString(colorRun('0', 8))
			} else {
				result.WriteString(colorRun('2', 8))
			}
		}
		return result.String()
	case id == "Hashid" || id == "Sqid":
		return colorRun('4', info.Size)
	case id == "H3 Grid System":
		return colorMap(colorRun('0', 1), colorRun('1', 4), colorRun('0', 3), colorRun('4', 11), colorRun('5', 45))
	case id == "Mist":
		return colorMap(colorRun('0', 1), colorRun('6', 47), colorRun('4', 8), colorRun('5', 8))
	case id == "Unix timestamp":
		return colorRun('3', info.Size)
	case id == "DUNS Number" || id == "SWHID (Software Hash ID)" || id == "Ethereum Address":
		return defaultMap
	case strings.HasPrefix(id, "OrderlyID"):
		return colorMap(colorRun('3', 48), colorRun('1', 3), colorRun('0', 5), colorRun('4', 16), colorRun('6', 12), colorRun('5', 16), colorRun('2', 60))
	case id == "Bitcoin Address" || id == "Bitcoin Address (from Satoshi Nakamoto)":
		if strings.Contains(info.Parsed, "bech32") {
			entropy := 0
			if info.Entropy != nil {
				entropy = *info.Entropy
			}
			return colorMap(colorRun('1', 8), colorRun('2', entropy), colorRun('4', info.Size-8-entropy))
		}
		return colorMap(colorRun('1', 8), colorRun('2', 160), colorRun('4', 32))
	case id == "IPFS":
		entropy := 0
		if info.Entropy != nil {
			entropy = *info.Entropy
		}
		return colorMap(colorRun('1', info.Size-entropy), colorRun('2', entropy))
	case id == "Snowflake":
		return snowflakeColorMap(version)
	}
	return defaultMap
}

func combColorMap(version string) string {
	parts := colorMap(colorRun('3', 48), colorRun('2', 80))
	if version != "PostgreSQL" {
		parts = colorMap(colorRun('2', 80), colorRun('3', 48))
	}
	colors := []byte(parts)
	for _, index := range []int{48, 49, 50, 51, 64, 65} {
		colors[index] = '0'
	}
	return string(colors)
}

func isbnColorMap(standard string) string {
	parts := strings.Split(standard, "-")
	if len(parts) < 4 {
		return ""
	}
	if len(parts) == 5 {
		return colorMap(colorRun('4', (len(parts[0])+len(parts[2]))*8), colorRun('5', len(parts[2])*8), colorRun('6', len(parts[3])*8), colorRun('0', 8))
	}
	return colorMap(colorRun('4', len(parts[0])*8), colorRun('5', len(parts[1])*8), colorRun('6', len(parts[2])*8), colorRun('0', 8))
}

func snowflakeColorMap(version string) string {
	switch version {
	case "Twitter", "LinkedIn":
		return colorMap(colorRun('0', 1), colorRun('3', 41), colorRun('4', 10), colorRun('6', 12))
	case "Discord":
		return colorMap(colorRun('3', 42), colorRun('4', 5), colorRun('5', 5), colorRun('6', 12))
	case "Instagram":
		return colorMap(colorRun('3', 41), colorRun('4', 13), colorRun('6', 10))
	case "Sony":
		return colorMap(colorRun('0', 1), colorRun('3', 39), colorRun('6', 8), colorRun('4', 16))
	case "Spaceflake":
		return colorMap(colorRun('0', 1), colorRun('3', 41), colorRun('4', 5), colorRun('5', 5), colorRun('6', 12))
	case "Mastodon":
		return colorMap(colorRun('3', 48), colorRun('6', 16))
	case "Frostflake":
		return colorMap(colorRun('3', 32), colorRun('6', 21), colorRun('4', 11))
	case "Flake ID":
		return colorMap(colorRun('3', 42), colorRun('4', 5), colorRun('5', 5), colorRun('6', 12))
	case "Simpleflake":
		return colorMap(colorRun('3', 41), colorRun('2', 23))
	default:
		return colorRun('0', 64)
	}
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
	if force != "" {
		if force == "puid" {
			if info, err := parsePUIDAny(input, options); err == nil {
				return []*types.IDInfo{colorize(info)}
			}
		}
		if info, err := parseNameWithOptions(force, input, options); err == nil {
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
