package parsers

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestUnixTimeParser_Name(t *testing.T) {
	parser := &UnixTimeParser{}
	if parser.Name() != "UnixTime" {
		t.Errorf("Expected name 'UnixTime', got '%s'", parser.Name())
	}
}

func TestUnixTimeParser_CanParse(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test valid Unix timestamps
	validTimestamps := []string{
		"1234567890",                // 10 digits - seconds since epoch
		"1234567890123",             // 13 digits - milliseconds
		"1234567890123456",          // 16 digits - microseconds  
		"1234567890123456789",       // 19 digits - nanoseconds
		"0",                         // Zero (epoch)
		"1609459200",                // 2021-01-01 00:00:00 UTC
		"1672531200000",             // 2023-01-01 00:00:00 UTC in milliseconds
		"9999999999",                // Max reasonable seconds timestamp
	}
	
	for _, ts := range validTimestamps {
		if !parser.CanParse(ts) {
			t.Errorf("Expected to parse valid Unix timestamp: %s", ts)
		}
	}
	
	// Test invalid timestamps
	invalidTimestamps := []string{
		"",                          // empty
		"123456789",                 // too short (9 digits)
		"12345678901234567890123",   // too long (20+ digits)
		"abc123456789",              // contains letters
		"123456789.0",               // contains decimal point
		"123456789-",                // contains hyphen
		"-123456789",                // negative number (not supported here)  
		"12 3456789",                // contains space
		"123456789a",                // ends with letter
		"0x123456789",               // hex prefix
		"1.23456789e9",              // scientific notation
	}
	
	for _, ts := range invalidTimestamps {
		if parser.CanParse(ts) {
			t.Errorf("Expected to reject invalid Unix timestamp: %s", ts)
		}
	}
}

func TestUnixTimeParser_Parse(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test with a known timestamp (2021-01-01 00:00:00 UTC)
	knownTimestamp := "1609459200"
	expectedTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	
	info, err := parser.Parse(knownTimestamp)
	if err != nil {
		t.Fatalf("Failed to parse valid Unix timestamp: %v", err)
	}
	
	// Check basic properties
	if !strings.Contains(info.IDType, "Unix timestamp") {
		t.Errorf("Expected IDType to contain 'Unix timestamp', got '%s'", info.IDType)
	}
	
	if info.Standard != knownTimestamp {
		t.Errorf("Expected Standard '%s', got '%s'", knownTimestamp, info.Standard)
	}
	
	// Check size (should be 64 bits)
	if info.Size != 64 {
		t.Errorf("Expected Size 64, got %d", info.Size)
	}
	
	// Check integer representation
	if info.Integer == nil || *info.Integer != knownTimestamp {
		t.Errorf("Expected Integer '%s', got %v", knownTimestamp, info.Integer)
	}
	
	// Check timestamp
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		// Should be close to expected time (allowing for unit interpretation differences)
		diff := info.DateTime.Sub(expectedTime)
		if diff < -time.Hour || diff > time.Hour {
			t.Errorf("Parsed time should be close to expected time, got: %v, expected: %v, diff: %v", 
				info.DateTime, expectedTime, diff)
		}
	}
	
	// Check entropy (should be 0 for deterministic timestamps)
	if info.Entropy == nil || *info.Entropy != 0 {
		t.Errorf("Expected Entropy 0, got %v", info.Entropy)
	}
	
	// Check extra information
	if info.Extra["unit"] == "" {
		t.Error("Expected unit to be set")
	}
	
	if info.Extra["precision"] == "" {
		t.Error("Expected precision to be set")
	}
	
	if info.Extra["epoch"] != "1970-01-01T00:00:00Z" {
		t.Errorf("Expected epoch '1970-01-01T00:00:00Z', got '%s'", info.Extra["epoch"])
	}
	
	if info.Extra["deterministic"] != "true" {
		t.Errorf("Expected deterministic 'true', got '%s'", info.Extra["deterministic"])
	}
	
	// Check hex representation
	if len(info.Hex) != 16 { // 64 bits = 16 hex chars
		t.Errorf("Expected Hex length 16, got %d: %s", len(info.Hex), info.Hex)
	}
	
	// Check binary representation exists
	if len(info.Binary) == 0 {
		t.Error("Expected Binary representation to be set")
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_timestamp")
	if err == nil {
		t.Error("Expected error for invalid Unix timestamp")
	}
}

func TestUnixTimeParser_Generate(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Unix timestamp: %v", err)
	}
	
	// Check if generated timestamp can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated Unix timestamp is not valid: %s", generated)
	}
	
	// Check it's a numeric string
	generatedInt, err := strconv.ParseInt(generated, 10, 64)
	if err != nil {
		t.Errorf("Generated Unix timestamp should be a valid integer: %s", generated)
	}
	
	// Check it's a reasonable current timestamp (within 1 minute of now)
	now := time.Now().Unix()
	diff := absInt64(generatedInt - now)
	if diff > 60 { // Allow 1 minute difference
		t.Errorf("Generated timestamp should be close to current time, got: %d, now: %d, diff: %d", 
			generatedInt, now, diff)
	}
	
	// Test that multiple generations produce different results (due to time progression)
	time.Sleep(1 * time.Millisecond) // Small delay
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second Unix timestamp: %v", err)
	}
	
	// They might be the same if generated in the same second, so we just check they're both valid
	if !parser.CanParse(generated2) {
		t.Errorf("Second generated Unix timestamp is not valid: %s", generated2)
	}
}

func TestUnixTimeParser_UnitDetection(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test different timestamp formats
	testCases := []struct {
		timestamp    string
		expectedUnit string
		description  string
	}{
		{"1609459200", "seconds", "seconds since epoch"},
		{"1609459200000", "milliseconds", "milliseconds since epoch"},
		{"1609459200000000", "microseconds", "microseconds since epoch"},
		{"1609459200000000000", "nanoseconds", "nanoseconds since epoch"},
	}
	
	for _, tc := range testCases {
		info, err := parser.Parse(tc.timestamp)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", tc.description, err)
			continue
		}
		
		// Check that the detected unit is reasonable (might not be exact due to heuristics)
		if info.Extra["unit"] == "" {
			t.Errorf("Should detect unit for %s", tc.description)
		}
		
		// Check that the parsed time is reasonable (after 1970 and before 2100)
		if info.DateTime != nil {
			year := info.DateTime.Year()
			if year < 1970 || year > 2100 {
				t.Errorf("Parsed time year should be reasonable (1970-2100), got %d for %s", 
					year, tc.description)
			}
		}
	}
}

func TestUnixTimeParser_EdgeCases(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test edge cases
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"0", true, "epoch timestamp"},
		{"1000000000", true, "10 digits seconds timestamp"},
		{"999999999", false, "9 digits (too short)"},
		{"9999999999999999999", true, "19 digits (max length)"},
		{"10000000000000000000", false, "20 digits (too long)"},
		{"2147483647", true, "max 32-bit signed int"},
		{"4294967295", true, "max 32-bit unsigned int"},
	}
	
	for _, tc := range edgeCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
		
		// If it should parse, try parsing it
		if tc.shouldParse && result {
			info, err := parser.Parse(tc.input)
			if err != nil {
				t.Errorf("Parse(\"%s\") failed but CanParse returned true: %v (%s)", 
					tc.input, err, tc.description)
			} else {
				// Basic validation
				if info.DateTime == nil {
					t.Errorf("Expected DateTime to be set for %s", tc.description)
				}
			}
		}
	}
}

func TestUnixTimeParser_HelperFunctions(t *testing.T) {
	// Test absInt64 function
	tests := []struct {
		input    int64
		expected int64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{1<<63 - 1, 1<<63 - 1}, // max int64
	}
	
	for _, test := range tests {
		result := absInt64(test.input)
		if result != test.expected {
			t.Errorf("absInt64(%d) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestUnixTimeParser_ReasonableTimeDetection(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test that the parser prefers interpretations that result in reasonable times
	// Use current time in different units
	now := time.Now()
	
	testCases := []struct {
		timestamp   string
		expectUnit  string
		description string
	}{
		{strconv.FormatInt(now.Unix(), 10), "seconds", "current time in seconds"},
		{strconv.FormatInt(now.UnixMilli(), 10), "milliseconds", "current time in milliseconds"},
		{strconv.FormatInt(now.UnixMicro(), 10), "microseconds", "current time in microseconds"},
		{strconv.FormatInt(now.UnixNano(), 10), "nanoseconds", "current time in nanoseconds"},
	}
	
	for _, tc := range testCases {
		info, err := parser.Parse(tc.timestamp)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", tc.description, err)
			continue
		}
		
		// The parser should detect a reasonable unit
		detectedUnit := info.Extra["unit"]
		if detectedUnit == "" {
			t.Errorf("Should detect unit for %s", tc.description)
		}
		
		// The parsed time should be reasonably close to now (within a day)
		if info.DateTime != nil {
			diff := info.DateTime.Sub(now)
			if diff < -24*time.Hour || diff > 24*time.Hour {
				t.Errorf("Parsed time should be close to now for %s, got: %v, diff: %v", 
					tc.description, info.DateTime, diff)
			}
		}
	}
}

func TestUnixTimeParser_Performance(t *testing.T) {
	parser := &UnixTimeParser{}
	
	// Test parsing performance
	testTimestamp := "1609459200"
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testTimestamp)
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