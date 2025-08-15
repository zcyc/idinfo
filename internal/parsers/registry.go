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
		"uuid":      {"uuid", "guid"},
		"ulid":      {"ulid"},
		"objectid":  {"objectid", "mongodb", "bson"},
		"ksuid":     {"ksuid"},
		"xid":       {"xid"},
		"cuid":      {"cuid", "cuid2"},
		"scru128":   {"scru128", "scru"},
		"tsid":      {"tsid"},
		"nuid":      {"nuid", "nats-uid", "nats-id"},
		"nanoid":    {"nanoid", "nano-id", "nano_id"},
		"snowflake": {"snowflake", "sf", "sf-twitter", "sf-discord", "twitter", "discord"},
		"unixtime":  {"unixtime", "unix", "timestamp"},
		"hashhex":   {"hashhex", "hash", "hex"},
		"base58":    {"base58", "b58", "bitcoin"},
		"pushid":    {"pushid", "push-id", "firebase"},
		"base32":    {"base32", "b32"},

		"shortuuid": {"shortuuid", "short-uuid", "suuid"},
		"sqids":     {"sqids", "sqid"},
		"typeid":    {"typeid", "type-id"},
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
