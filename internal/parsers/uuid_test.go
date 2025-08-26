package parsers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Add the generateUUIDWithVersion function for testing
// This is a copy of the function from main.go to make it testable
func generateUUIDWithVersion(version string) (string, error) {
	switch version {
	case "v1":
		// UUID v1: timestamp and MAC address
		u, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
		return u.String(), nil

	case "v3":
		// UUID v3: namespace name based with MD5
		// Use a default namespace (DNS) and a default name for command-line usage
		u := uuid.NewMD5(uuid.NameSpaceDNS, []byte("idinfo-generated"))
		return u.String(), nil

	case "v4":
		// UUID v4: random (default)
		u := uuid.New()
		return u.String(), nil

	case "v5":
		// UUID v5: namespace name based with SHA-1
		// Use a default namespace (DNS) and a default name for command-line usage
		u := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("idinfo-generated"))
		return u.String(), nil

	case "v6":
		// UUID v6: reordered timestamp and MAC address
		u, err := uuid.NewV6()
		if err != nil {
			return "", err
		}
		return u.String(), nil

	case "v7":
		// UUID v7: sortable timestamp and random
		u, err := uuid.NewV7()
		if err != nil {
			return "", err
		}
		return u.String(), nil

	default:
		return "", fmt.Errorf("unsupported UUID version '%s'. Supported versions: v1, v3, v4, v5, v6, v7", version)
	}
}

func TestUUIDParser_CanParse(t *testing.T) {
	parser := &UUIDParser{}
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400e29b41d4a716446655440000", true},
		{"01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa", true},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
		{"", false},
	}
	
	for _, test := range tests {
		result := parser.CanParse(test.input)
		if result != test.expected {
			t.Errorf("CanParse(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestUUIDParser_Parse(t *testing.T) {
	parser := &UUIDParser{}
	
	// Test UUID v4
	info, err := parser.Parse("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if info.IDType != "UUID (RFC-9562)" {
		t.Errorf("Expected ID type 'UUID (RFC-9562)', got %q", info.IDType)
	}
	
	if info.Version != "4 (random)" {
		t.Errorf("Expected version '4 (random)', got %q", info.Version)
	}
	
	if info.Size != 128 {
		t.Errorf("Expected size 128, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy != 122 {
		t.Errorf("Expected entropy 122, got %v", info.Entropy)
	}
}

func TestUUIDParser_Generate(t *testing.T) {
	parser := &UUIDParser{}
	
	uuidStr, err := parser.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if !parser.CanParse(uuidStr) {
		t.Errorf("Generated UUID %q cannot be parsed by the same parser", uuidStr)
	}
	
	// Verify it generates UUID v4 by default
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		t.Fatalf("Failed to parse generated UUID: %v", err)
	}
	
	if parsed.Version() != 4 {
		t.Errorf("Expected UUID v4, got v%d", parsed.Version())
	}
}

func TestUUIDVersionGeneration(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectedVer uuid.Version
		shouldError bool
	}{
		{"UUID v1", "v1", 1, false},
		{"UUID v3", "v3", 3, false},
		{"UUID v4", "v4", 4, false},
		{"UUID v5", "v5", 5, false},
		{"UUID v6", "v6", 6, false},
		{"UUID v7", "v7", 7, false},
		{"Invalid version", "v8", 0, true},
		{"Invalid format", "v99", 0, true},
		{"Empty version", "", 0, true},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := generateUUIDWithVersion(test.version)
			
			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error for version %s, but got none", test.version)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error for version %s: %v", test.version, err)
			}
			
			// Parse the generated UUID
			parsed, err := uuid.Parse(result)
			if err != nil {
				t.Fatalf("Failed to parse generated UUID %s: %v", result, err)
			}
			
			// Verify the version
			if parsed.Version() != test.expectedVer {
				t.Errorf("Expected UUID v%d, got v%d for result %s", test.expectedVer, parsed.Version(), result)
			}
			
			// Verify it's a valid UUID format
			if !strings.Contains(result, "-") {
				t.Errorf("Generated UUID %s doesn't contain hyphens", result)
			}
			
			// Verify length
			if len(result) != 36 {
				t.Errorf("Generated UUID %s has wrong length: expected 36, got %d", result, len(result))
			}
			
			// Test that our parser can parse it
			parser := &UUIDParser{}
			if !parser.CanParse(result) {
				t.Errorf("UUIDParser cannot parse generated UUID %s", result)
			}
		})
	}
}

func TestUUIDConsistentGeneration(t *testing.T) {
	// Test that v3 and v5 generate consistent results with same inputs
	v3_1, err := generateUUIDWithVersion("v3")
	if err != nil {
		t.Fatalf("Failed to generate UUID v3: %v", err)
	}
	
	v3_2, err := generateUUIDWithVersion("v3")
	if err != nil {
		t.Fatalf("Failed to generate UUID v3: %v", err)
	}
	
	// v3 should be consistent (same namespace + name = same UUID)
	if v3_1 != v3_2 {
		t.Errorf("UUID v3 should be deterministic, got %s and %s", v3_1, v3_2)
	}
	
	v5_1, err := generateUUIDWithVersion("v5")
	if err != nil {
		t.Fatalf("Failed to generate UUID v5: %v", err)
	}
	
	v5_2, err := generateUUIDWithVersion("v5")
	if err != nil {
		t.Fatalf("Failed to generate UUID v5: %v", err)
	}
	
	// v5 should be consistent (same namespace + name = same UUID)
	if v5_1 != v5_2 {
		t.Errorf("UUID v5 should be deterministic, got %s and %s", v5_1, v5_2)
	}
	
	// v3 and v5 should be different (different hash algorithms)
	if v3_1 == v5_1 {
		t.Errorf("UUID v3 and v5 should be different, both got %s", v3_1)
	}
}

func TestUUIDRandomGeneration(t *testing.T) {
	// Test that v1, v4, v6, v7 generate different results each time
	randomVersions := []string{"v1", "v4", "v6", "v7"}
	
	for _, version := range randomVersions {
		t.Run("Random_"+version, func(t *testing.T) {
			uuid1, err := generateUUIDWithVersion(version)
			if err != nil {
				t.Fatalf("Failed to generate UUID %s: %v", version, err)
			}
			
			uuid2, err := generateUUIDWithVersion(version)
			if err != nil {
				t.Fatalf("Failed to generate UUID %s: %v", version, err)
			}
			
			// These should be different (very high probability)
			if uuid1 == uuid2 {
				t.Errorf("UUID %s should generate different values, got same: %s", version, uuid1)
			}
		})
	}
}