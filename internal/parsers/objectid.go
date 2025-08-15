package parsers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"

	"github.com/zcyc/idinfo/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ObjectIDParser struct{}

var objectIdRegex = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

func (p *ObjectIDParser) Name() string {
	return "ObjectID"
}

func (p *ObjectIDParser) CanParse(input string) bool {
	if len(input) != 24 {
		return false
	}
	return objectIdRegex.MatchString(input)
}

func (p *ObjectIDParser) Parse(input string) (*types.IDInfo, error) {
	oid, err := primitive.ObjectIDFromHex(input)
	if err != nil {
		return nil, err
	}

	info := &types.IDInfo{
		IDType:   "MongoDB ObjectId",
		Standard: oid.Hex(),
		Size:     96,
		Hex:      oid.Hex(),
		Binary:   oid[:],
		Extra:    make(map[string]string),
	}

	// Convert to integer representation
	bigInt := new(big.Int)
	bigInt.SetBytes(oid[:])
	intStr := bigInt.String()
	info.Integer = &intStr

	// Extract timestamp (first 4 bytes)
	timestamp := oid.Timestamp()
	info.DateTime = &timestamp
	timestampStr := fmt.Sprintf("%.3f", float64(timestamp.Unix()))
	info.Timestamp = &timestampStr

	// ObjectId has limited entropy due to predictable components
	entropy := 40 // Approximately 40 bits of entropy (5 bytes random + counter)
	info.Entropy = &entropy

	// Extract components
	bytes := oid[:]

	// Machine identifier (bytes 4-6)
	machineId := hex.EncodeToString(bytes[4:7])
	info.Node1 = &machineId

	// Process identifier (bytes 7-8)
	processId := hex.EncodeToString(bytes[7:9])
	info.Node2 = &processId

	// Counter (bytes 9-11)
	counter := int64(bytes[9])<<16 | int64(bytes[10])<<8 | int64(bytes[11])
	info.Sequence = &counter

	// Add ObjectId-specific information
	info.Extra["timestamp_precision"] = "second"
	info.Extra["machine_bytes"] = machineId
	info.Extra["process_bytes"] = processId
	info.Extra["counter_value"] = fmt.Sprintf("%d", counter)

	return info, nil
}

func (p *ObjectIDParser) Generate() (string, error) {
	oid := primitive.NewObjectID()
	return oid.Hex(), nil
}
