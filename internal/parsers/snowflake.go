package parsers

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/zcyc/idinfo/internal/types"
)

type SnowflakeParser struct {
	node *snowflake.Node
}

var snowflakeRegex = regexp.MustCompile(`^\d{10,19}$`)

func NewSnowflakeParser() *SnowflakeParser {
	// Create a snowflake node with node ID 1
	// In production, this should be unique per instance
	node, err := snowflake.NewNode(1)
	if err != nil {
		// If we can't create a node, we'll still allow parsing but not generation
		return &SnowflakeParser{node: nil}
	}
	return &SnowflakeParser{node: node}
}

func (p *SnowflakeParser) Name() string {
	return "Snowflake"
}

func (p *SnowflakeParser) CanParse(input string) bool {
	if !snowflakeRegex.MatchString(input) {
		return false
	}

	// Must be a valid integer (check uint64 for large numbers)
	_, err := strconv.ParseUint(input, 10, 64)
	return err == nil
}

func (p *SnowflakeParser) Parse(input string) (*types.IDInfo, error) {
	// Parse the input as a uint64 first
	id, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return nil, err
	}

	// Try to parse as a snowflake ID
	snowflakeId := snowflake.ID(id)

	// Extract information from the snowflake
	timestampMs := snowflakeId.Time()
	timestamp := time.UnixMilli(timestampMs)
	nodeId := snowflakeId.Node()
	step := snowflakeId.Step()

	info := &types.IDInfo{
		IDType:   "Snowflake",
		Standard: input,
		Size:     64,
		Hex:      fmt.Sprintf("%016x", id),
		Binary:   new(big.Int).SetUint64(id).Bytes(),
		Extra:    make(map[string]string),
	}

	// Set integer representation
	info.Integer = &input

	// Set timestamp information
	info.DateTime = &timestamp
	timestampStr := fmt.Sprintf("%.3f", float64(timestamp.UnixMilli())/1000)
	info.Timestamp = &timestampStr

	// Set node and sequence information
	nodeStr := fmt.Sprintf("%d", nodeId)
	info.Node1 = &nodeStr
	sequence := int64(step)
	info.Sequence = &sequence

	// Calculate entropy (10 bits for node + 12 bits for sequence = 22 bits)
	entropy := 22
	info.Entropy = &entropy

	// Add Snowflake-specific information
	info.Extra["epoch"] = time.UnixMilli(snowflake.Epoch).Format(time.RFC3339)
	info.Extra["timestamp_bits"] = "41"
	info.Extra["node_bits"] = "10"
	info.Extra["sequence_bits"] = "12"
	info.Extra["node_id"] = fmt.Sprintf("%d", nodeId)
	info.Extra["sequence_number"] = fmt.Sprintf("%d", step)
	info.Extra["library"] = "github.com/bwmarrin/snowflake"

	return info, nil
}

func (p *SnowflakeParser) Generate() (string, error) {
	if p.node == nil {
		// Try to create a node if we don't have one
		var err error
		p.node, err = snowflake.NewNode(1)
		if err != nil {
			return "", fmt.Errorf("failed to create snowflake node: %v", err)
		}
	}

	// Generate a new snowflake ID
	id := p.node.Generate()
	return id.String(), nil
}

// Package-level variable for the parser instance
var defaultSnowflakeParser *SnowflakeParser

func init() {
	defaultSnowflakeParser = NewSnowflakeParser()
}

// For compatibility with the existing registry system, we provide these wrapper functions
type SnowflakeParserWrapper struct{}

func (p *SnowflakeParserWrapper) Name() string {
	return defaultSnowflakeParser.Name()
}

func (p *SnowflakeParserWrapper) CanParse(input string) bool {
	return defaultSnowflakeParser.CanParse(input)
}

func (p *SnowflakeParserWrapper) Parse(input string) (*types.IDInfo, error) {
	return defaultSnowflakeParser.Parse(input)
}

func (p *SnowflakeParserWrapper) Generate() (string, error) {
	return defaultSnowflakeParser.Generate()
}
