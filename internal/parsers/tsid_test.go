package parsers

import (
	"fmt"
	"testing"
	"time"

	"github.com/rushysloth/go-tsid"
)

func TestTSIDParser_Name(t *testing.T) {
	parser := &TSIDParser{}
	if parser.Name() != "TSID" {
		t.Errorf("Expected name 'TSID', got '%s'", parser.Name())
	}
}

func TestTSIDParser_CanParse(t *testing.T) {
	parser := &TSIDParser{}
	
	// Test valid TSIDs (generate some using the official library)
	validTSIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		validTSIDs[i] = tsid.Fast().ToString()
	}
	
	for _, id := range validTSIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid TSID: %s", id)
		}
	}
	
	// Test case-insensitive parsing
	lowerCaseTsid := "01226n0640j7k"
	if !parser.CanParse(lowerCaseTsid) {
		t.Errorf("Expected to parse lowercase TSID: %s", lowerCaseTsid)
	}
	
	// Test invalid TSIDs
	invalidTSIDs := []string{
		"",                        // empty
		"abc",                     // too short
		"1234567890123456789012",  // too long
		"123456789012",            // too short (12 chars)
		"12345678901234",          // too long (14 chars)
		"!@#$%^&*()123",           // invalid characters
		"abcdefghijklU",           // contains 'U' (invalid in Crockford)
		"abcdefghijkl!",           // contains '!' (invalid character)
		"01226N0640J7K-",          // contains hyphen
		"01226N 0640J7K",          // contains space
	}
	
	for _, id := range invalidTSIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid TSID: %s", id)
		}
	}
}

func TestTSIDParser_Parse(t *testing.T) {
	parser := &TSIDParser{}
	
	// Generate a valid TSID for testing
	generatedTsid := tsid.Fast()
	validTSID := generatedTsid.ToString()
	
	info, err := parser.Parse(validTSID)
	if err != nil {
		t.Fatalf("Failed to parse valid TSID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "TSID (Time-Sorted Unique Identifier)" {
		t.Errorf("Expected IDType 'TSID (Time-Sorted Unique Identifier)', got '%s'", info.IDType)
	}
	
	if info.Standard != validTSID {
		t.Errorf("Expected Standard '%s', got '%s'", validTSID, info.Standard)
	}
	
	// Check size (should be 64 bits)
	if info.Size != 64 {
		t.Errorf("Expected Size 64, got %d", info.Size)
	}
	
	// Check entropy (should be 22 bits)
	if info.Entropy == nil || *info.Entropy != 22 {
		t.Errorf("Expected Entropy 22, got %v", info.Entropy)
	}
	
	// Check timestamp is present and reasonable (within last year to next year)
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		now := time.Now()
		if info.DateTime.Before(now.AddDate(-1, 0, 0)) || info.DateTime.After(now.AddDate(1, 0, 0)) {
			t.Errorf("TSID timestamp seems unreasonable: %v", info.DateTime)
		}
	}
	
	// Check integer representation
	if info.Integer == nil {
		t.Error("Expected Integer to be set")
	}
	
	// Check extra information
	if info.Extra["encoding"] != "Crockford Base32" {
		t.Errorf("Expected encoding info, got '%s'", info.Extra["encoding"])
	}
	
	if info.Extra["length"] != "13 characters" {
		t.Errorf("Expected length info, got '%s'", info.Extra["length"])
	}
	
	if info.Extra["sortable"] != "Yes (by generation time)" {
		t.Errorf("Expected sortable info, got '%s'", info.Extra["sortable"])
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_id")
	if err == nil {
		t.Error("Expected error for invalid TSID")
	}
	
	// Test whitespace handling
	info, err = parser.Parse("  " + validTSID + "  ")
	if err != nil {
		t.Errorf("Failed to parse TSID with whitespace: %v", err)
	}
	if info.Standard != validTSID {
		t.Errorf("Expected trimmed input, got '%s'", info.Standard)
	}
}

func TestTSIDParser_Generate(t *testing.T) {
	parser := &TSIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate TSID: %v", err)
	}
	
	// Check if generated TSID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated TSID is not valid: %s", generated)
	}
	
	// Check length
	if len(generated) != 13 {
		t.Errorf("Generated TSID should be 13 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (Crockford Base32)
	validChars := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	for _, char := range generated {
		found := false
		for _, validChar := range validChars {
			if char == validChar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Generated TSID contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second TSID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive TSID generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate TSID #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate TSID: %s", gen)
		}
		seen[gen] = true
	}
}

func TestTSIDParser_Performance(t *testing.T) {
	parser := &TSIDParser{}
	
	// Test parsing performance
	testTsid := tsid.Fast().ToString()
	
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testTsid)
		if err != nil {
			t.Fatalf("Parse failed during performance test: %v", err)
		}
	}
	parseTime := time.Since(start)
	
	// Test generation performance
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_, err := parser.Generate()
		if err != nil {
			t.Fatalf("Generate failed during performance test: %v", err)
		}
	}
	generateTime := time.Since(start)
	
	t.Logf("Parse performance: %v for 1000 operations", parseTime)
	t.Logf("Generate performance: %v for 1000 operations", generateTime)
	
	// Reasonable performance expectations (adjust as needed)
	if parseTime > time.Second {
		t.Errorf("Parse performance too slow: %v", parseTime)
	}
	if generateTime > time.Second {
		t.Errorf("Generate performance too slow: %v", generateTime)
	}
}

func TestTSIDParser_RealWorldExamples(t *testing.T) {
	parser := &TSIDParser{}
	
	// Generate TSIDs with different factory configurations
	// Default factory
	defaultTsid := tsid.Fast().ToString()
	if !parser.CanParse(defaultTsid) {
		t.Errorf("Failed to parse default TSID: %s", defaultTsid)
	}
	
	// Try parsing and ensure all components are extracted properly
	info, err := parser.Parse(defaultTsid)
	if err != nil {
		t.Fatalf("Failed to parse default TSID: %v", err)
	}
	
	// Verify timestamp is reasonable (should be very recent)
	now := time.Now()
	if info.DateTime == nil {
		t.Error("DateTime should be set")
	} else if info.DateTime.Before(now.Add(-time.Minute)) || info.DateTime.After(now.Add(time.Minute)) {
		t.Errorf("TSID timestamp should be recent, got: %v", info.DateTime)
	}
	
	// Verify hex and binary representations are consistent
	if len(info.Hex) != 16 { // 64 bits = 16 hex chars
		t.Errorf("Expected 16 hex characters, got %d: %s", len(info.Hex), info.Hex)
	}
	
	if len(info.Binary) != 8 { // 64 bits = 8 bytes
		t.Errorf("Expected 8 binary bytes, got %d", len(info.Binary))
	}
}

func TestTSIDParser_EdgeCases(t *testing.T) {
	parser := &TSIDParser{}
	
	// Test edge cases
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"0000000000000", true, "all zeros (minimum value)"},
		{"ZZZZZZZZZZZZ", false, "too many Z's (would overflow)"},
		{"01226N0640J7K", true, "typical TSID format"},
		{"01226n0640j7k", true, "lowercase should work"},
		{"01226N0640j7k", true, "mixed case should work"},
		{"01226N0640J7I", true, "contains I (valid in Crockford - maps to 1)"},
		{"01226N0640J7L", true, "contains L (valid in Crockford - maps to 1)"},
		{"01226N0640J7O", true, "contains O (valid in Crockford - maps to 0)"},
		{"01226N0640J7U", false, "contains U (invalid in Crockford)"},
	}
	
	for _, tc := range edgeCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
		
		// If it should parse, try parsing it
		if tc.shouldParse && result {
			_, err := parser.Parse(tc.input)
			if err != nil {
				t.Errorf("Parse(\"%s\") failed but CanParse returned true: %v (%s)", 
					tc.input, err, tc.description)
			}
		}
	}
}

func TestTSIDParser_OfficialLibraryIntegration(t *testing.T) {
	parser := &TSIDParser{}
	
	// Test that our parser works with various TSID factory configurations
	// This tests integration with the official library
	
	// Test with default factory
	defaultTsid := tsid.Fast()
	tsidStr := defaultTsid.ToString()
	
	// Parse with our parser
	info, err := parser.Parse(tsidStr)
	if err != nil {
		t.Fatalf("Failed to parse TSID from official library: %v", err)
	}
	
	// Verify the timestamp matches
	officialMillis := defaultTsid.GetUnixMillis()
	if info.DateTime == nil {
		t.Error("DateTime should be set")
	} else {
		parsedMillis := info.DateTime.UnixMilli()
		if parsedMillis != officialMillis {
			t.Errorf("Timestamp mismatch: official=%d, parsed=%d", officialMillis, parsedMillis)
		}
	}
	
	// Verify the number representation matches
	officialNumber := defaultTsid.ToNumber()
	if info.Integer == nil {
		t.Error("Integer should be set")
	} else {
		expectedStr := fmt.Sprintf("%d", officialNumber)
		if *info.Integer != expectedStr {
			t.Errorf("Number mismatch: official=%d, parsed=%s", officialNumber, *info.Integer)
		}
	}
}