package parsers

import (
	"fmt"
	"strings"

	"github.com/zcyc/idinfo/internal/types"
	"go.jetify.com/typeid/v2"
)

// TypeIDParser handles parsing of TypeID format using official SDK
type TypeIDParser struct{}

func (p *TypeIDParser) Name() string {
	return "TypeID"
}

func (p *TypeIDParser) CanParse(input string) bool {
	// Basic format check: must have underscore and reasonable length
	if len(input) < 27 || !strings.Contains(input, "_") {
		return false
	}

	// Use official SDK to validate
	_, err := typeid.Parse(input)
	return err == nil
}

func (p *TypeIDParser) Parse(input string) (*types.IDInfo, error) {
	input = strings.TrimSpace(input)

	// Parse using official SDK
	tid, err := typeid.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid TypeID format: %v", err)
	}

	// Extract components
	typePrefix := tid.Prefix()
	suffix := tid.Suffix()

	// Extract UUID from the TypeID
	uuidStr := tid.UUID()

	// TypeID doesn't expose timestamp extraction directly
	// We'll use a placeholder for now
	timestampStr := "N/A"

	entropy := 128 // TypeID has same entropy as ULID (128 bits)

	extra := map[string]string{
		"type_prefix":   typePrefix,
		"suffix":        suffix,
		"uuid":          uuidStr,
		"format":        "TypeID (type prefix + ULID)",
		"specification": "https://github.com/jetify-com/typeid",
		"alphabet":      "Crockford Base32",
		"timestamp_ms":  timestampStr,
		"sortable":      "Yes (chronologically sortable)",
		"url_safe":      "Yes",
	}

	// Add information about the type
	commonTypes := map[string]string{
		"user":     "User Account",
		"org":      "Organization",
		"post":     "Post/Article",
		"comment":  "Comment",
		"product":  "Product",
		"order":    "Order",
		"payment":  "Payment",
		"invoice":  "Invoice",
		"session":  "Session",
		"token":    "Token",
		"file":     "File Upload",
		"event":    "Event",
		"task":     "Task",
		"project":  "Project",
		"customer": "Customer",
		"account":  "Account",
		"document": "Document",
		"message":  "Message",
	}

	if description, exists := commonTypes[typePrefix]; exists {
		extra["type_description"] = description
	}

	// Convert UUID string to bytes
	uuidBytes := []byte(uuidStr)

	return &types.IDInfo{
		IDType:    "TypeID",
		Standard:  input,
		Size:      128, // Same as ULID
		Entropy:   &entropy,
		Timestamp: &timestampStr,
		DateTime:  nil, // TypeID doesn't expose timestamp extraction
		Hex:       fmt.Sprintf("%x", uuidBytes),
		Binary:    uuidBytes,
		Extra:     extra,
	}, nil
}

func (p *TypeIDParser) Generate() (string, error) {
	// Generate a TypeID with "demo" prefix using official SDK
	tid, err := typeid.Generate("demo")
	if err != nil {
		return "", fmt.Errorf("failed to generate TypeID: %v", err)
	}

	return tid.String(), nil
}
