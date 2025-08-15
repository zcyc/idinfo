package parsers

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSnowflakeParser_Name(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	if parser.Name() != "Snowflake" {
		t.Errorf("Expected name 'Snowflake', got '%s'", parser.Name())
	}
}

func TestSnowflakeParser_CanParse(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test valid Snowflake IDs (numeric strings of 10-19 digits)
	validSnowflakes := []string{
		"1234567890",                // 10 digits
		"1234567890123456789",       // 19 digits (max)
		"175928847299117063",        // Twitter-style Snowflake
		"266241948824334336",        // Discord Snowflake
		"1125899906842624",          // Another valid format
		"9223372036854775807",       // Max int64
		"1000000000000000000",       // 19 digits
	}
	
	for _, id := range validSnowflakes {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid Snowflake: %s", id)
		}
	}
	
	// Test invalid Snowflakes
	invalidSnowflakes := []string{
		"",                          // empty
		"123456789",                 // too short (9 digits)
		"12345678901234567890123",   // too long (20+ digits)
		"abc123456789",              // contains letters
		"123456789.0",               // contains decimal point
		"123456789-",                // contains hyphen
		"-123456789",                // negative number
		"12 3456789",                // contains space
		"123456789a",                // ends with letter
		"0x123456789",               // hex prefix
		"12345678901234567890123456789", // way too long
	}
	
	for _, id := range invalidSnowflakes {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid Snowflake: %s", id)
		}
	}
}

func TestSnowflakeParser_Parse(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test with a known Twitter Snowflake format
	twitterSnowflake := "175928847299117063"
	
	info, err := parser.Parse(twitterSnowflake)
	if err != nil {
		t.Fatalf("Failed to parse valid Snowflake: %v", err)
	}
	
	// Check basic properties
	if !contains(info.IDType, "Snowflake") {
		t.Errorf("Expected IDType to contain 'Snowflake', got '%s'", info.IDType)
	}
	
	if info.Standard != twitterSnowflake {
		t.Errorf("Expected Standard '%s', got '%s'", twitterSnowflake, info.Standard)
	}
	
	// Check size (should be 64 bits)
	if info.Size != 64 {
		t.Errorf("Expected Size 64, got %d", info.Size)
	}
	
	// Check integer representation
	if info.Integer == nil || *info.Integer != twitterSnowflake {
		t.Errorf("Expected Integer '%s', got %v", twitterSnowflake, info.Integer)
	}
	
	// Check timestamp is present and reasonable
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		// Snowflake timestamps should be between 2010 and now
		minTime := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
		maxTime := time.Now().Add(time.Hour) // Allow some future time for clock skew
		if info.DateTime.Before(minTime) || info.DateTime.After(maxTime) {
			t.Errorf("Snowflake timestamp seems unreasonable: %v", info.DateTime)
		}
	}
	
	// Check node and sequence
	if info.Node1 == nil {
		t.Error("Expected Node1 to be set")
	}
	
	if info.Sequence == nil {
		t.Error("Expected Sequence to be set")
	}
	
	// Check entropy (should be reasonable for node + sequence bits)
	if info.Entropy == nil {
		t.Error("Expected Entropy to be set")
	} else {
		// Entropy should be between 10-30 bits typically
		if *info.Entropy < 10 || *info.Entropy > 30 {
			t.Errorf("Expected reasonable entropy (10-30), got %d", *info.Entropy)
		}
	}
	
	// Check extra information (library info instead of variant)
	if info.Extra["library"] == "" {
		t.Error("Expected library to be set")
	}
	
	if info.Extra["epoch"] == "" {
		t.Error("Expected epoch to be set")
	}
	
	if info.Extra["timestamp_bits"] == "" {
		t.Error("Expected timestamp_bits to be set")
	}
	
	if info.Extra["node_bits"] == "" {
		t.Error("Expected node_bits to be set")
	}
	
	if info.Extra["sequence_bits"] == "" {
		t.Error("Expected sequence_bits to be set")
	}
	
	// Check hex representation
	if len(info.Hex) != 16 { // 64 bits = 16 hex chars
		t.Errorf("Expected Hex length 16, got %d: %s", len(info.Hex), info.Hex)
	}
	
	// Check binary representation
	if len(info.Binary) == 0 {
		t.Error("Expected Binary representation to be set")
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_snowflake")
	if err == nil {
		t.Error("Expected error for invalid Snowflake")
	}
}

func TestSnowflakeParser_Generate(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Snowflake: %v", err)
	}
	
	// Check if generated Snowflake can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated Snowflake is not valid: %s", generated)
	}
	
	// Check it's a numeric string
	_, err = strconv.ParseInt(generated, 10, 64)
	if err != nil {
		t.Errorf("Generated Snowflake should be a valid integer: %s", generated)
	}
	
	// Check length is reasonable (10-19 digits)
	if len(generated) < 10 || len(generated) > 19 {
		t.Errorf("Generated Snowflake length should be 10-19 digits, got %d: %s", len(generated), generated)
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second Snowflake: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive Snowflake generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate Snowflake #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate Snowflake: %s", gen)
		}
		seen[gen] = true
	}
}

func TestSnowflakeParser_VariantDetection(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test different variants by generating IDs and checking they're detected properly
	// Note: New snowflake library uses a single epoch
	testCases := []struct {
		name        string
		epoch       int64
		expectedVar string
	}{
		{"Snowflake", 1288834974657, "Snowflake"}, // Twitter epoch from library
	}
	
	for _, tc := range testCases {
		// Generate a simple Snowflake-style ID for this epoch
		now := time.Now().UnixMilli()
		if now > tc.epoch {
			timestamp := now - tc.epoch
			nodeId := int64(1)
			sequence := int64(0)
			
			// Use Twitter-style bit layout for simplicity
			id := (timestamp << 22) | (nodeId << 12) | sequence
			idStr := strconv.FormatInt(id, 10)
			
			info, err := parser.Parse(idStr)
			if err != nil {
				t.Errorf("Failed to parse %s Snowflake: %v", tc.name, err)
				continue
			}
			
			// Check that library info is present
			if info.Extra["library"] == "" {
				t.Errorf("Should have library info for %s Snowflake", tc.name)
			}
			
			// Check that the epoch info is present
			if info.Extra["epoch"] == "" {
				t.Errorf("Should have epoch for %s Snowflake", tc.name)
			}
		}
	}
}

func TestSnowflakeParser_TimestampExtraction(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test with recently generated Snowflake
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Snowflake: %v", err)
	}
	
	info, err := parser.Parse(generated)
	if err != nil {
		t.Fatalf("Failed to parse generated Snowflake: %v", err)
	}
	
	// Timestamp should be very recent (within last minute)
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		now := time.Now()
		diff := now.Sub(*info.DateTime)
		if diff < 0 || diff > time.Minute {
			t.Errorf("Generated Snowflake timestamp should be recent, got: %v (diff: %v)", info.DateTime, diff)
		}
	}
}

func TestSnowflakeParser_EdgeCases(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test edge cases
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"1000000000", true, "minimum 10 digits"},
		{"999999999", false, "9 digits (too short)"},
		{"9223372036854775807", true, "max int64"},
		{"9999999999999999999", true, "19 digits (max length)"},
		{"10000000000000000000", false, "20 digits (too long)"},
		{"0000000001000000000", true, "with leading zeros conceptually but as string"},
		{"1000000000", true, "exactly 10 digits"},
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

func TestSnowflakeParser_Components(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test that we can extract node and sequence correctly
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Snowflake: %v", err)
	}
	
	info, err := parser.Parse(generated)
	if err != nil {
		t.Fatalf("Failed to parse generated Snowflake: %v", err)
	}
	
	// Check that node and sequence are reasonable values
	if info.Node1 != nil {
		nodeStr := *info.Node1
		if nodeStr == "" {
			t.Error("Node should not be empty")
		}
		// Node should be a reasonable number (0-1023 for Twitter format)
		if node, err := strconv.ParseInt(nodeStr, 10, 64); err == nil {
			if node < 0 || node > 8191 { // Allow for different formats
				t.Errorf("Node value seems unreasonable: %d", node)
			}
		}
	}
	
	if info.Sequence != nil {
		sequence := *info.Sequence
		if sequence < 0 || sequence > 4095 { // Allow for different formats
			t.Errorf("Sequence value seems unreasonable: %d", sequence)
		}
	}
}

func TestSnowflakeParser_Performance(t *testing.T) {
	parser := &SnowflakeParserWrapper{}
	
	// Test parsing performance
	testSnowflake := "175928847299117063"
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testSnowflake)
		if err != nil {
			t.Fatalf("Parse failed during performance test: %v", err)
		}
	}
	
	// Test generation performance
	for i := 0; i < 1000; i++ {
		_, err := parser.Generate()
		if err != nil {
			t.Fatalf("Generate failed during performance test: %v", err)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}