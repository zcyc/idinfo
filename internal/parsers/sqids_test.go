package parsers

import (
	"strings"
	"testing"
)

func TestSqidsParser_Name(t *testing.T) {
	parser := &SqidsParser{}
	if parser.Name() != "Sqids" {
		t.Errorf("Expected name 'Sqids', got '%s'", parser.Name())
	}
}

func TestSqidsParser_CanParse(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test valid Sqids (based on official SDK behavior)
	validIDs := []string{
		"86Rf07",         // Example from Sqids docs
		"a1B2c3D4",       // Mixed case and numbers
		"xyz123ABC",      // Mixed case and numbers
		"A5K7M9",         // With uppercase and numbers
		"abcdefgh",       // All lowercase (valid in official SDK)
		"ABCDEFGH",       // All uppercase (valid in official SDK)
		"12345678",       // All numbers (valid in official SDK)
		"i1D7jfxUL",      // Our generated ID
		"bMUkgbEfVquwOI", // Long valid ID
	}
	
	for _, id := range validIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid Sqids ID: %s", id)
		}
	}
	
	// Test invalid IDs (based on official SDK behavior)
	invalidIDs := []string{
		"",               // empty
		"a",              // single character (invalid in official SDK)
		"A",              // single character (invalid in official SDK)
		"1",              // single character (invalid in official SDK)
		"Z",              // single character (invalid in official SDK)
		"hello@world",    // invalid @ character
		"test_id",        // invalid _ character  
		"test-id",        // invalid - character
		"test.id",        // invalid . character
		"test id",        // contains space
		"hello world",    // contains space
		"test#id",        // invalid # character
	}
	
	for _, id := range invalidIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid Sqids ID: %s", id)
		}
	}
}

func TestSqidsParser_Parse(t *testing.T) {
	parser := &SqidsParser{}
	
	validID := "86Rf07"
	
	info, err := parser.Parse(validID)
	if err != nil {
		t.Fatalf("Failed to parse valid Sqids ID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "Sqids" {
		t.Errorf("Expected IDType 'Sqids', got '%s'", info.IDType)
	}
	
	if info.Standard != validID {
		t.Errorf("Expected Standard '%s', got '%s'", validID, info.Standard)
	}
	
	// Check that size and entropy are calculated
	if info.Size <= 0 {
		t.Errorf("Expected positive size, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy <= 0 {
		t.Errorf("Expected positive entropy, got %v", info.Entropy)
	}
	
	// Check extra information should include canonical form
	if _, exists := info.Extra["canonical"]; !exists {
		t.Error("Expected canonical form in extra information")
	}
	
	if _, exists := info.Extra["numbers"]; !exists {
		t.Error("Expected decoded numbers in extra information")
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid@id")
	if err == nil {
		t.Error("Expected error for invalid Sqids ID")
	}
	
	// Test whitespace handling
	info, err = parser.Parse("  " + validID + "  ")
	if err != nil {
		t.Errorf("Failed to parse Sqids ID with whitespace: %v", err)
	}
	if info.Standard != validID {
		t.Errorf("Expected trimmed input, got '%s'", info.Standard)
	}
}

func TestSqidsParser_Generate(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test generation - our implementation uses fixed numbers [42, 123, 7890]
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Sqids ID: %v", err)
	}
	
	// Should generate a consistent ID since we use fixed numbers
	expectedID := "i1D7jfxUL" // This is what [42, 123, 7890] encodes to
	if generated != expectedID {
		t.Errorf("Expected consistent generated ID '%s', got '%s'", expectedID, generated)
	}
	
	// Check if generated ID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated Sqids ID should be parseable: %s", generated)
	}
	
	// Test that multiple generations produce the same result (since we use fixed numbers)
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second Sqids ID: %v", err)
	}
	
	if generated != generated2 {
		t.Errorf("Generated Sqids IDs should be consistent with fixed numbers: %s vs %s", generated, generated2)
	}
	
	// Verify the generated ID decodes to our expected numbers
	info, err := parser.Parse(generated)
	if err != nil {
		t.Fatalf("Failed to parse generated ID: %v", err)
	}
	
	expectedNumbers := "[42 123 7890]"
	if info.Extra["numbers"] != expectedNumbers {
		t.Errorf("Generated ID should decode to %s, got %s", expectedNumbers, info.Extra["numbers"])
	}
}

func TestSqidsParser_AlphabetValidation(t *testing.T) {
	parser := &SqidsParser{}
	
	// Sqids default alphabet: abcdefghijkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ0123456789
	// Note: excludes 'l', 'I', 'o', 'O' to avoid confusion
	validChars := "abcdefghijkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ0123456789"
	for _, char := range validChars {
		// Create a multi-character ID to avoid single-char rejection
		testID := "ab" + string(char)
		if !parser.CanParse(testID) {
			t.Errorf("Should parse ID with valid character '%c': %s", char, testID)
		}
	}
	
	// Test invalid characters
	invalidChars := "_-@#$%^&*()+=[]{}|\\:;\"'<>,.?/" // special chars only
	for _, char := range invalidChars {
		// Use multi-character to test just the invalid char
		testID := "ab" + string(char) + "cd"
		if parser.CanParse(testID) {
			t.Errorf("Should not parse ID with invalid character '%c': %s", char, testID)
		}
	}
}

func TestSqidsParser_LengthValidation(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test minimum length - single characters are invalid in official SDK
	invalidSingle := []string{
		"a", "A", "1", "Z",
	}
	
	for _, id := range invalidSingle {
		if parser.CanParse(id) {
			t.Errorf("Should not parse single character Sqids ID: %s", id)
		}
	}
	
	// Test valid multi-character IDs
	validMulti := []string{
		"ab", "AB", "12", "aB", "a1", "1a",
	}
	
	for _, id := range validMulti {
		if !parser.CanParse(id) {
			t.Errorf("Should parse multi-character Sqids ID: %s", id)
		}
	}
	
	// Test reasonable length limits - very long IDs might be rejected
	reasonableLengthID := "ab" + strings.Repeat("c", 50) // 52 chars total
	if !parser.CanParse(reasonableLengthID) {
		t.Error("Should parse reasonably long Sqids ID")
	}
}

func TestSqidsParser_EdgeCases(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test edge cases for full coverage
	testCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"", false, "empty string"},
		{"a", false, "single character"},
		{"ab", true, "two characters"},
		{strings.Repeat("a", 255), true, "very long but valid chars"},
		{"a b", false, "contains space"},
		{"a\tb", false, "contains tab"},
		{"a\nb", false, "contains newline"},
		{"a-b", false, "contains hyphen"},
		{"a_b", false, "contains underscore"},
		{"a.b", false, "contains period"},
		{"a@b", false, "contains at symbol"},
		{"a#b", false, "contains hash"},
		{"a$b", false, "contains dollar"},
		{"a%b", false, "contains percent"},
		{"a^b", false, "contains caret"},
		{"a&b", false, "contains ampersand"},
		{"a*b", false, "contains asterisk"},
		{"a(b", false, "contains parenthesis"},
		{"a+b", false, "contains plus"},
		{"a=b", false, "contains equals"},
		{"a[b", false, "contains bracket"},
		{"a{b", false, "contains brace"},
		{"a|b", false, "contains pipe"},
		{"a\\b", false, "contains backslash"},
		{"a:b", false, "contains colon"},
		{"a;b", false, "contains semicolon"},
		{"a\"b", false, "contains quote"},
		{"a'b", false, "contains apostrophe"},
		{"a<b", false, "contains less than"},
		{"a>b", false, "contains greater than"},
		{"a,b", false, "contains comma"},
		{"a?b", false, "contains question mark"},
		{"a/b", false, "contains slash"},
		{"alb", true, "contains 'l' (valid in official SDK)"},
		{"aIb", true, "contains 'I' (valid in official SDK)"},
		{"aob", true, "contains 'o' (valid in official SDK)"},
		{"aOb", true, "contains 'O' (valid in official SDK)"},
	}
	
	for _, tc := range testCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(%s) = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
	}
}

func TestSqidsParser_CharacterComposition(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test different character compositions - in official SDK, all are valid if using valid chars
	testCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"abcdefgh", true, "lowercase letters only (valid in official SDK)"},
		{"ABCDEFGH", true, "uppercase letters only (valid in official SDK)"},
		{"12345678", true, "numbers only (valid in official SDK)"},
		{"a1B2c3D4", true, "mixed case and numbers"},
		{"abcd1234", true, "lowercase and numbers"},
		{"ABCD1234", true, "uppercase and numbers"},
		{"aAbBcCdD", true, "alternating case"},
		{"987654321", true, "all numbers (valid in official SDK)"},
		{"aBcDeFgH", true, "alternating case letters"},
	}
	
	for _, tc := range testCases {
		result := parser.CanParse(tc.input)
		if result != tc.shouldParse {
			t.Errorf("CanParse(%s) = %v, expected %v (%s)", 
				tc.input, result, tc.shouldParse, tc.description)
		}
	}
}

func TestSqidsParser_OfficialExamples(t *testing.T) {
	parser := &SqidsParser{}
	
	// Test with known examples from the official SDK
	knownValidIDs := []string{
		"86Rf07",     // [1, 2, 3]
		"i1D7jfxUL",  // [42, 123, 7890] - our generated ID
	}
	
	for _, id := range knownValidIDs {
		if !parser.CanParse(id) {
			t.Errorf("Should parse known valid Sqids ID: %s", id)
		}
		
		info, err := parser.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse known valid ID %s: %v", id, err)
		}
		
		// Check that we get canonical form and numbers
		if _, exists := info.Extra["canonical"]; !exists {
			t.Errorf("Missing canonical form for ID %s", id)
		}
		
		if _, exists := info.Extra["numbers"]; !exists {
			t.Errorf("Missing decoded numbers for ID %s", id)
		}
	}
}