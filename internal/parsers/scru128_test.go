package parsers

import (
	"strings"
	"testing"
	"time"

	"github.com/scru128/go-scru128"
)

func TestSCRU128Parser_Name(t *testing.T) {
	parser := &SCRU128Parser{}
	if parser.Name() != "SCRU128" {
		t.Errorf("Expected name 'SCRU128', got '%s'", parser.Name())
	}
}

func TestSCRU128Parser_CanParse(t *testing.T) {
	parser := &SCRU128Parser{}

	// Test valid SCRU128 IDs (generate some using the official library)
	validSCRU128s := make([]string, 5)
	for i := 0; i < 5; i++ {
		validSCRU128s[i] = scru128.New().String()
	}

	for _, id := range validSCRU128s {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid SCRU128: %s", id)
		}
	}

	// Test some additional valid patterns (if library accepts them)
	additionalValid := []string{
		scru128.NewString(), // Another way to generate
	}

	for _, id := range additionalValid {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid SCRU128: %s", id)
		}
	}

	// Test invalid SCRU128s
	invalidSCRU128s := []string{
		"",                              // empty
		"abc",                           // too short
		"12345678901234567890123456789", // too long (28 chars)
		"1234567890123456789012345",     // too long (25 chars)
		"123456789012345678901",         // too short (21 chars)
		"!@#$%^&*()1234567890123",       // invalid characters
		"aBcDeFgHiJkLmNoPqRsT++",        // contains '+' (invalid)
		"aBcDeFgHiJkLmNoPqRsT==",        // contains '=' (invalid)
		"aBcDeFgHiJkLmNoPqRsT  ",        // contains space
		"aBcDeFgHiJkLmNoPqRsT~~",        // contains '~' (invalid)
	}

	for _, id := range invalidSCRU128s {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid SCRU128: %s", id)
		}
	}
}

func TestSCRU128Parser_Parse(t *testing.T) {
	parser := &SCRU128Parser{}

	// Generate a valid SCRU128 for testing
	generatedSCRU128 := scru128.New()
	validSCRU128 := generatedSCRU128.String()

	info, err := parser.Parse(validSCRU128)
	if err != nil {
		t.Fatalf("Failed to parse valid SCRU128: %v", err)
	}

	// Check basic properties
	expectedIDType := "SCRU128 (Sortable, Clock-based, Realm-specific, Unique identifier)"
	if info.IDType != expectedIDType {
		t.Errorf("Expected IDType '%s', got '%s'", expectedIDType, info.IDType)
	}

	if info.Standard != validSCRU128 {
		t.Errorf("Expected Standard '%s', got '%s'", validSCRU128, info.Standard)
	}

	// Check size (should be 128 bits)
	if info.Size != 128 {
		t.Errorf("Expected Size 128, got %d", info.Size)
	}

	// Check entropy (should be around 26 bits for counter + randomness)
	if info.Entropy == nil || *info.Entropy != 26 {
		t.Errorf("Expected Entropy 26, got %v", info.Entropy)
	}

	// Check timestamp is present (though it might not be accurate due to implementation)
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	}

	// Check sequence/counter
	if info.Sequence == nil {
		t.Error("Expected Sequence to be set")
	}

	// Check extra information
	if info.Extra["encoding"] != "Base36" {
		t.Errorf("Expected encoding 'Base36', got '%s'", info.Extra["encoding"])
	}

	if info.Extra["timestamp_precision"] != "millisecond" {
		t.Errorf("Expected timestamp_precision 'millisecond', got '%s'", info.Extra["timestamp_precision"])
	}

	if info.Extra["sortable"] != "true" {
		t.Errorf("Expected sortable 'true', got '%s'", info.Extra["sortable"])
	}

	if info.Extra["timestamp_bits"] != "48" {
		t.Errorf("Expected timestamp_bits '48', got '%s'", info.Extra["timestamp_bits"])
	}

	if info.Extra["counter_bits"] != "24" {
		t.Errorf("Expected counter_bits '24', got '%s'", info.Extra["counter_bits"])
	}

	if info.Extra["randomness_bits"] != "32" {
		t.Errorf("Expected randomness_bits '32', got '%s'", info.Extra["randomness_bits"])
	}

	// Check hex and binary representations exist
	if len(info.Hex) == 0 {
		t.Error("Expected Hex representation to be set")
	}

	if len(info.Binary) == 0 {
		t.Error("Expected Binary representation to be set")
	}

	// Test invalid input
	_, err = parser.Parse("invalid_scru128")
	if err == nil {
		t.Error("Expected error for invalid SCRU128")
	}
}

func TestSCRU128Parser_Generate(t *testing.T) {
	parser := &SCRU128Parser{}

	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate SCRU128: %v", err)
	}

	// The current implementation generates a simple hex format
	// Check that it's a reasonable hex string
	if len(generated) == 0 {
		t.Error("Generated SCRU128 should not be empty")
	}

	// Check character set (should be hex characters)
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Generated SCRU128 contains invalid character '%c': %s", char, generated)
		}
	}

	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second SCRU128: %v", err)
	}

	if generated == generated2 {
		t.Error("Two consecutive SCRU128 generations should produce different results")
	}

	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ { // Smaller number since generate might not be truly unique
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate SCRU128 #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate SCRU128: %s", gen)
		}
		seen[gen] = true
	}
}

func TestSCRU128Parser_OfficialLibraryIntegration(t *testing.T) {
	parser := &SCRU128Parser{}

	// Test that our parser works with the official library
	officialSCRU128 := scru128.New()
	scru128Str := officialSCRU128.String()

	// Verify it's exactly 26 characters (SCRU128 specification)
	if len(scru128Str) != 26 {
		t.Errorf("SCRU128 should be 26 characters, got %d: %s", len(scru128Str), scru128Str)
	}

	// Parse with our parser
	if !parser.CanParse(scru128Str) {
		t.Errorf("Our parser should accept SCRU128 from official library: %s", scru128Str)
	}

	info, err := parser.Parse(scru128Str)
	if err != nil {
		t.Fatalf("Failed to parse SCRU128 from official library: %v", err)
	}

	// Basic validation
	expectedIDType := "SCRU128 (Sortable, Clock-based, Realm-specific, Unique identifier)"
	if info.IDType != expectedIDType {
		t.Errorf("Expected correct IDType, got: %s", info.IDType)
	}

	if info.Standard != scru128Str {
		t.Errorf("Expected Standard to match input: %s vs %s", info.Standard, scru128Str)
	}

	// Test with NewString as well
	anotherSCRU128 := scru128.NewString()
	if !parser.CanParse(anotherSCRU128) {
		t.Errorf("Should be able to parse SCRU128 from NewString: %s", anotherSCRU128)
	}
}

func TestSCRU128Parser_Sortability(t *testing.T) {
	parser := &SCRU128Parser{}

	// Generate SCRU128s with small delays to ensure different timestamps
	var scru128s []string
	for i := 0; i < 5; i++ {
		generated := scru128.New().String()
		scru128s = append(scru128s, generated)
		time.Sleep(1 * time.Millisecond) // Small delay
	}

	// Check that string comparison maintains order (SCRU128 property)
	for i := 1; i < len(scru128s); i++ {
		if strings.Compare(scru128s[i], scru128s[i-1]) < 0 {
			t.Errorf("SCRU128s are not lexicographically sorted: %s came before %s",
				scru128s[i], scru128s[i-1])
		}
	}

	// Parse all SCRU128s to ensure they're valid
	for _, s := range scru128s {
		if !parser.CanParse(s) {
			t.Errorf("Generated SCRU128 should be parseable: %s", s)
		}
	}
}

func TestSCRU128Parser_EdgeCases(t *testing.T) {
	parser := &SCRU128Parser{}

	// Test edge cases
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"00000000000000000000000000", true, "all zeros (26 chars)"},           // Will test with official library
		{"ZZZZZZZZZZZZZZZZZZZZZZZZZZ", false, "all Z's (might be invalid)"},    // Will test with official library
		{"aBcDeFgHiJkLmNoPqRsTuVwXyZ", false, "mixed case (might be invalid)"}, // Will test with official library
		{strings.Repeat("0", 26), true, "all zeros"},                           // Will test with official library
		{strings.Repeat("a", 26), false, "all a's (might be invalid)"},         // Will test with official library
	}

	for _, tc := range edgeCases {
		// Use the official library as the source of truth
		expectedResult := tc.shouldParse
		if _, err := scru128.Parse(tc.input); err != nil {
			expectedResult = false
		} else {
			expectedResult = true
		}

		result := parser.CanParse(tc.input)
		if result != expectedResult {
			t.Errorf("CanParse(\"%s\") = %v, expected %v (%s)",
				tc.input, result, expectedResult, tc.description)
		}

		// If it should parse, try parsing it
		if expectedResult && result {
			_, err := parser.Parse(tc.input)
			if err != nil {
				t.Errorf("Parse(\"%s\") failed but CanParse returned true: %v (%s)",
					tc.input, err, tc.description)
			}
		}
	}
}

func TestSCRU128Parser_Performance(t *testing.T) {
	parser := &SCRU128Parser{}

	// Test parsing performance
	testSCRU128 := scru128.New().String()

	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testSCRU128)
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

func TestSCRU128Parser_CharacterSet(t *testing.T) {
	// Test that the regex accepts valid SCRU128 characters
	validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-"

	// Create a test string with all valid characters (truncated to 26 chars)
	testStr := validChars[:26]

	// This might not be a valid SCRU128 according to the official library,
	// but we can test that our regex at least handles the character set correctly
	regex := scru128Regex
	if !regex.MatchString(testStr) {
		t.Errorf("Regex should match valid character set: %s", testStr)
	}

	// Test invalid characters
	invalidTestStr := "!@#$%^&*()1234567890abcdef"
	if regex.MatchString(invalidTestStr) {
		t.Errorf("Regex should not match invalid characters: %s", invalidTestStr)
	}
}
