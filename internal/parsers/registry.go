package parsers

import (
	"strings"

	"github.com/zcyc/idinfo/internal/types"
)

// Registry manages all ID parsers
type Registry struct {
	parsers []types.IDParser
}

// NewRegistry creates a new parser registry with all parsers registered
func NewRegistry() *Registry {
	return &Registry{
		parsers: []types.IDParser{
			&alignedParser{name: "uuid-int", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUUIDInteger(input, options)
			}},
			&alignedParser{name: "uuid-b64", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUUIDBase64(input, options)
			}},
			&alignedParser{name: "uuid25", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUUID25(input, options)
			}},
			&alignedParser{name: "julid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseJulid(input, options)
			}},
			&alignedParser{name: "upid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUPID(input, options)
			}},
			&alignedParser{name: "sandflake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSandflake(input, options)
			}},
			&alignedParser{name: "timeflake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseTimeflake(input, options)
			}},
			&alignedParser{name: "flake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseFlake(input, options)
			}},
			&alignedParser{name: "scru64", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSCRU64Aligned(input, options)
			}},
			&alignedParser{name: "scru128", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSCRU128Aligned(input, options)
			}, generate: generateSCRU128},
			&alignedParser{name: "datadog", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseDatadog(input, options)
			}},
			&alignedParser{name: "spotify", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSpotify(input, options)
			}},
			&alignedParser{name: "cuid1", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseCUID1(input, options)
			}},
			&alignedParser{name: "cuid2", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseCUID2Aligned(input, options)
			}},
			&alignedParser{name: "hashid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseHashID(input, options)
			}},
			&alignedParser{name: "youtube", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseYouTube(input, options)
			}},
			&alignedParser{name: "stripe", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseStripe(input, options)
			}},
			&alignedParser{name: "breezeid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseBreezeID(input, options)
			}},
			&alignedParser{name: "puid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseAlignedByName("puid", input, options)
			}},
			&alignedParser{name: "tid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) { return parseTID(input, options) }},
			&alignedParser{name: "threads", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseThreads(input, options)
			}},
			&alignedParser{name: "snowid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowID(input, options)
			}},
			&alignedParser{name: "mist", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseMist(input, options)
			}},
			&alignedParser{name: "comb", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseComb(input, options)
			}},
			&alignedParser{name: "duns", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseDUNS(input, options)
			}},
			&alignedParser{name: "asin", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseASIN(input, options)
			}},
			&alignedParser{name: "gdocs", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseGoogleDocs(input, options)
			}},
			&alignedParser{name: "slack", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSlack(input, options)
			}},
			&alignedParser{name: "nano64", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseNano64(input, options)
			}},
			&alignedParser{name: "orderlyid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseOrderlyID(input, options)
			}},
			&alignedParser{name: "swhid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSWHID(input, options)
			}},
			&alignedParser{name: "iban", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseIBAN(input, options)
			}},
			&alignedParser{name: "commerce", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseCommerce(input, options)
			}},
			&alignedParser{name: "vin", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) { return parseVIN(input, options) }},
			&alignedParser{name: "bitcoin", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseBitcoin(input, options)
			}},
			&alignedParser{name: "ethereum", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseEthereum(input, options)
			}},
			&alignedParser{name: "ipfs", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseIPFS(input, options)
			}},
			&alignedParser{name: "ipv4", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseIPv4(input, options)
			}},
			&alignedParser{name: "ipv6", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseIPv6(input, options)
			}},
			&alignedParser{name: "mac", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) { return parseMAC(input, options) }},
			&alignedParser{name: "imei", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseIMEI(input, options)
			}},
			&alignedParser{name: "isbn", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseISBN(input, options)
			}},
			&alignedParser{name: "h3", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) { return parseH3(input, options) }},
			&alignedParser{name: "hash", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseHashAligned(input, options)
			}},
			&alignedParser{name: "unixtime", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixAuto, false)
			}},
			&alignedParser{name: "unix-s", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixSeconds, false)
			}},
			&alignedParser{name: "unix-ms", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixMillis, false)
			}},
			&alignedParser{name: "unix-us", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixMicros, false)
			}},
			&alignedParser{name: "unix-ns", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixNanos, false)
			}},
			&alignedParser{name: "unix-recent", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseUnixAligned(input, options, unixAuto, true)
			}},
			&alignedParser{name: "sf-twitter", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-twitter")
			}},
			&alignedParser{name: "sf-mastodon", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-mastodon")
			}},
			&alignedParser{name: "sf-discord", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-discord")
			}},
			&alignedParser{name: "sf-instagram", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-instagram")
			}},
			&alignedParser{name: "sf-linkedin", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-linkedin")
			}},
			&alignedParser{name: "sf-sony", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-sony")
			}},
			&alignedParser{name: "sf-spaceflake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-spaceflake")
			}},
			&alignedParser{name: "sf-frostflake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-frostflake")
			}},
			&alignedParser{name: "sf-flakeid", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-flakeid")
			}},
			&alignedParser{name: "sf-simpleflake", parse: func(input string, options types.ParseOptions) (*types.IDInfo, error) {
				return parseSnowflakeAligned(input, options, "sf-simpleflake")
			}},
			&UUIDParser{},
			&ULIDParser{},
			&ObjectIDParser{},
			&KSUIDParser{},
			&XidParser{},
			&CUIDParser{},
			&SCRU128Parser{},
			&TSIDParser{},
			&TypeIDParser{},    // Moved before NanoID to get priority
			&NUIDParser{},      // NATS Unique Identifier - moved before ShortUUID
			&ShortUUIDParser{}, // Moved before Sqids to get priority
			&SqidsParser{},     // Moved before NanoID to get priority
			&NanoIDParser{},
			&SnowflakeParserWrapper{},
			&UnixTimeParser{},
			&HashHexParser{},
			&Base58Parser{},
			&PushIDParser{},
			&Base32Parser{},
		},
	}
}

// GetParser returns a parser by name
func (r *Registry) GetParser(name string) types.IDParser {
	for _, parser := range r.parsers {
		if strings.EqualFold(parser.Name(), name) {
			return parser
		}
	}
	return nil
}

// GetAllParsers returns all registered parsers
func (r *Registry) GetAllParsers() []types.IDParser {
	return r.parsers
}

// GetAvailableParsers returns the names of all available parsers
func (r *Registry) GetAvailableParsers() []string {
	var names []string
	for _, parser := range r.parsers {
		names = append(names, parser.Name())
	}
	return names
}

// Global registry of all parsers
var globalRegistry = NewRegistry()

// ParseID attempts to parse an ID using all registered parsers
func ParseID(input string, forceFormat string) []*types.IDInfo {
	return globalRegistry.ParseID(input, forceFormat)
}

// ParseID attempts to parse an ID using all registered parsers in the registry
func (r *Registry) ParseID(input string, forceFormat string) []*types.IDInfo {
	input = strings.TrimSpace(input)
	var results []*types.IDInfo

	if forceFormat != "" {
		// If a specific format is forced, try only that parser
		for _, parser := range r.parsers {
			if matchesForceFormat(parser.Name(), forceFormat) {
				if info, err := parser.Parse(input); err == nil {
					results = append(results, info)
				}
			}
		}
	} else {
		// Try all parsers and collect successful results
		for _, parser := range r.parsers {
			if parser.CanParse(input) {
				if info, err := parser.Parse(input); err == nil {
					results = append(results, info)
				}
			}
		}
	}

	return results
}

// GetParser returns a parser by name (global function for backward compatibility)
func GetParser(name string) types.IDParser {
	return globalRegistry.GetParser(name)
}

// GetAllParsers returns all registered parsers (global function for backward compatibility)
func GetAllParsers() []types.IDParser {
	return globalRegistry.GetAllParsers()
}

// matchesForceFormat checks if a parser name matches the forced format
func matchesForceFormat(parserName, forceFormat string) bool {
	parserName = strings.ToLower(parserName)
	forceFormat = strings.ToLower(forceFormat)

	// Handle aliases and variations
	aliases := map[string][]string{
		"uuid":           {"uuid", "guid"},
		"uuid-int":       {"uuid-int", "uuid-integer", "integer"},
		"uuid-b64":       {"uuid-b64", "uuid-base64", "base64"},
		"uuid25":         {"uuid25", "uuid-25"},
		"ulid":           {"ulid"},
		"julid":          {"julid"},
		"upid":           {"upid"},
		"sandflake":      {"sandflake"},
		"timeflake":      {"timeflake"},
		"flake":          {"flake"},
		"objectid":       {"objectid", "mongodb", "bson"},
		"ksuid":          {"ksuid"},
		"xid":            {"xid"},
		"cuid":           {"cuid", "cuid1", "cuid2"},
		"scru128":        {"scru128", "scru"},
		"scru64":         {"scru64"},
		"tsid":           {"tsid"},
		"nuid":           {"nuid", "nats-uid", "nats-id"},
		"nanoid":         {"nanoid", "nano-id", "nano_id"},
		"snowflake":      {"snowflake", "sf"},
		"sf-twitter":     {"sf-twitter", "twitter"},
		"sf-mastodon":    {"sf-mastodon", "mastodon"},
		"sf-discord":     {"sf-discord", "discord"},
		"sf-instagram":   {"sf-instagram", "instagram"},
		"sf-linkedin":    {"sf-linkedin", "linkedin"},
		"sf-sony":        {"sf-sony", "sony"},
		"sf-spaceflake":  {"sf-spaceflake", "spaceflake"},
		"sf-frostflake":  {"sf-frostflake", "frostflake"},
		"sf-flakeid":     {"sf-flakeid", "flakeid"},
		"sf-simpleflake": {"sf-simpleflake", "simpleflake"},
		"unixtime":       {"unixtime", "unix", "timestamp"},
		"unix-s":         {"unix-s", "unix-seconds"},
		"unix-ms":        {"unix-ms", "unix-milliseconds"},
		"unix-us":        {"unix-us", "unix-microseconds"},
		"unix-ns":        {"unix-ns", "unix-nanoseconds"},
		"hashhex":        {"hashhex", "hash", "hex"},
		"base58":         {"base58", "b58", "bitcoin"},
		"pushid":         {"pushid", "push-id", "firebase"},
		"base32":         {"base32", "b32"},

		"shortuuid": {"shortuuid", "short-uuid", "suuid"},
		"sqids":     {"sqids", "sqid"},
		"typeid":    {"typeid", "type-id"},
		"datadog":   {"datadog", "traceid"},
		"spotify":   {"spotify"},
		"hashid":    {"hashid", "hashids"},
		"youtube":   {"youtube", "youtube-id"},
		"stripe":    {"stripe"},
		"breezeid":  {"breezeid", "breeze-id"},
		"puid":      {"puid"},
		"tid":       {"tid"},
		"threads":   {"threads", "thread-id"},
		"snowid":    {"snowid"},
		"duns":      {"duns"},
		"asin":      {"asin"},
		"gdocs":     {"gdocs", "google-docs"},
		"slack":     {"slack"},
		"nano64":    {"nano64"},
		"orderlyid": {"orderlyid", "orderly-id"},
		"swhid":     {"swhid"},
		"iban":      {"iban"},
		"commerce":  {"commerce", "barcode", "gtin"},
		"vin":       {"vin"},
		"bitcoin":   {"bitcoin", "btc"},
		"ethereum":  {"ethereum", "eth"},
		"ipfs":      {"ipfs", "cid"},
		"ipv4":      {"ipv4"},
		"ipv6":      {"ipv6"},
		"mac":       {"mac", "mac-address"},
		"imei":      {"imei"},
		"isbn":      {"isbn"},
		"h3":        {"h3"},
		"mist":      {"mist"},
	}

	for canonicalName, aliasList := range aliases {
		if parserName == canonicalName {
			for _, alias := range aliasList {
				if alias == forceFormat {
					return true
				}
			}
		}
	}

	return parserName == forceFormat
}
