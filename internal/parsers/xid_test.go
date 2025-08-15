package parsers

import (
	"strings"
	"testing"
	"time"

	"github.com/rs/xid"
)

func TestXidParser_Name(t *testing.T) {
	parser := &XidParser{}
	if parser.Name() != "Xid" {
		t.Errorf("Expected name 'Xid', got '%s'", parser.Name())
	}
}

func TestXidParser_CanParse(t *testing.T) {
	parser := &XidParser{}
	
	// Test valid XIDs (generate some using the official library)
	validXids := make([]string, 5)
	for i := 0; i < 5; i++ {
		validXids[i] = xid.New().String()
	}
	
	for _, id := range validXids {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid Xid: %s", id)
		}
	}
	
	// Test some additional valid Xid patterns (using Base32 Crockford alphabet)
	additionalValid := []string{
		"9m4e2mr0ui3e8a215n4g", // Example Xid format
		"b50vv9dq2bo8b2s5pkf0", // Another example
		"c0123456789abcdefghj", // Valid Base32 chars
	}
	
	for _, id := range additionalValid {
		// Only test if the official library considers it valid
		if _, err := xid.FromString(id); err == nil {
			if !parser.CanParse(id) {
				t.Errorf("Expected to parse valid Xid: %s", id)
			}
		}
	}
	
	// Test invalid XIDs
	invalidXids := []string{
		"",                       // empty
		"abc",                    // too short
		"12345678901234567890123456789", // too long (28 chars)
		"1234567890123456789",    // too short (19 chars)
		"123456789012345678901",  // too long (21 chars)
		"!@#$%^&*()1234567890",   // invalid characters
		"WXYZ1234567890123456",   // contains W, X, Y, Z (invalid in Crockford Base32)
		"9m4e2mr0ui3e8a215n4G",   // uppercase G (might be invalid - Crockford is case insensitive but normalized)
		"9m4e2mr0ui3e8a215n4 ",   // contains space
		"9m4e2mr0ui3e8a215n4-",   // contains hyphen
	}
	
	for _, id := range invalidXids {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid Xid: %s", id)
		}
	}
}

func TestXidParser_Parse(t *testing.T) {
	parser := &XidParser{}
	
	// Generate a valid Xid for testing
	generatedXid := xid.New()
	validXid := generatedXid.String()
	
	info, err := parser.Parse(validXid)
	if err != nil {
		t.Fatalf("Failed to parse valid Xid: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "Xid (globally unique sortable id)" {
		t.Errorf("Expected IDType 'Xid (globally unique sortable id)', got '%s'", info.IDType)
	}
	
	if info.Standard != validXid {
		t.Errorf("Expected Standard '%s', got '%s'", validXid, info.Standard)
	}
	
	// Check size (should be 96 bits - 12 bytes)
	if info.Size != 96 {
		t.Errorf("Expected Size 96, got %d", info.Size)
	}
	
	// Check entropy (should be 56 bits)
	if info.Entropy == nil || *info.Entropy != 56 {
		t.Errorf("Expected Entropy 56, got %v", info.Entropy)
	}
	
	// Check timestamp is present and reasonable (should be recent for generated Xid)
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		now := time.Now()
		if info.DateTime.Before(now.Add(-time.Minute)) || info.DateTime.After(now.Add(time.Minute)) {
			t.Errorf("Xid timestamp should be recent, got: %v", info.DateTime)
		}
	}
	
	// Check integer representation
	if info.Integer == nil {
		t.Error("Expected Integer to be set")
	}
	
	// Check node information
	if info.Node1 == nil {
		t.Error("Expected Node1 (machine ID) to be set")
	} else {
		// Machine ID should be 6 hex characters (3 bytes)
		if len(*info.Node1) != 6 {
			t.Errorf("Expected Node1 length 6, got %d: %s", len(*info.Node1), *info.Node1)
		}
	}
	
	if info.Node2 == nil {
		t.Error("Expected Node2 (process ID) to be set")
	} else {
		// Process ID should be 4 hex characters (2 bytes)
		if len(*info.Node2) != 4 {
			t.Errorf("Expected Node2 length 4, got %d: %s", len(*info.Node2), *info.Node2)
		}
	}
	
	// Check sequence/counter
	if info.Sequence == nil {
		t.Error("Expected Sequence to be set")
	} else {
		// Counter should be a reasonable value (0 to 16777215 for 3 bytes)
		if *info.Sequence < 0 || *info.Sequence > 16777215 {
			t.Errorf("Expected Sequence to be 0-16777215, got %d", *info.Sequence)
		}
	}
	
	// Check extra information
	if info.Extra["encoding"] != "Base32 (Crockford)" {
		t.Errorf("Expected encoding 'Base32 (Crockford)', got '%s'", info.Extra["encoding"])
	}
	
	if info.Extra["timestamp_precision"] != "second" {
		t.Errorf("Expected timestamp_precision 'second', got '%s'", info.Extra["timestamp_precision"])
	}
	
	if info.Extra["sortable"] != "true" {
		t.Errorf("Expected sortable 'true', got '%s'", info.Extra["sortable"])
	}
	
	// Check hex and binary representations
	if len(info.Hex) != 24 { // 12 bytes = 24 hex chars
		t.Errorf("Expected Hex length 24, got %d: %s", len(info.Hex), info.Hex)
	}
	
	if len(info.Binary) != 12 { // 96 bits = 12 bytes
		t.Errorf("Expected Binary length 12, got %d", len(info.Binary))
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_xid")
	if err == nil {
		t.Error("Expected error for invalid Xid")
	}
}

func TestXidParser_Generate(t *testing.T) {
	parser := &XidParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Xid: %v", err)
	}
	
	// Check if generated Xid can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated Xid is not valid: %s", generated)
	}
	
	// Check length
	if len(generated) != 20 {
		t.Errorf("Generated Xid should be 20 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (Base32 Crockford: 0-9, a-v)
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'v')) {
			t.Errorf("Generated Xid contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second Xid: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive Xid generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate Xid #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate Xid: %s", gen)
		}
		seen[gen] = true
	}
}

func TestXidParser_Performance(t *testing.T) {
	parser := &XidParser{}
	
	// Test parsing performance
	testXid := xid.New().String()
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testXid)
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

func TestXidParser_OfficialLibraryIntegration(t *testing.T) {
	parser := &XidParser{}
	
	// Test that our parser works with the official library
	officialXid := xid.New()
	xidStr := officialXid.String()
	
	// Parse with our parser
	info, err := parser.Parse(xidStr)
	if err != nil {
		t.Fatalf("Failed to parse Xid from official library: %v", err)
	}
	
	// Verify the timestamp matches
	officialTime := officialXid.Time()
	if info.DateTime == nil {
		t.Error("DateTime should be set")
	} else if !info.DateTime.Equal(officialTime) {
		t.Errorf("Timestamp mismatch: official=%v, parsed=%v", officialTime, info.DateTime)
	}
	
	// Verify the bytes match
	officialBytes := officialXid.Bytes()
	if len(info.Binary) != len(officialBytes) {
		t.Errorf("Binary length mismatch: official=%d, parsed=%d", len(officialBytes), len(info.Binary))
	}
	
	for i, expected := range officialBytes {
		if i < len(info.Binary) && info.Binary[i] != expected {
			t.Errorf("Binary mismatch at index %d: expected %02x, got %02x", i, expected, info.Binary[i])
		}
	}
}

func TestXidParser_Sortability(t *testing.T) {
	parser := &XidParser{}
	
	// Generate Xids with small delays to ensure different timestamps
	var xids []string
	for i := 0; i < 5; i++ {
		generated, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate Xid: %v", err)
		}
		xids = append(xids, generated)
		time.Sleep(1 * time.Second) // Xid has second precision
	}
	
	// Parse all Xids and check that timestamps are in order
	var timestamps []time.Time
	for _, x := range xids {
		info, err := parser.Parse(x)
		if err != nil {
			t.Fatalf("Failed to parse Xid: %v", err)
		}
		if info.DateTime != nil {
			timestamps = append(timestamps, *info.DateTime)
		}
	}
	
	// Check that timestamps are in ascending order (or equal)
	for i := 1; i < len(timestamps); i++ {
		if timestamps[i].Before(timestamps[i-1]) {
			t.Errorf("Xids are not properly sorted by timestamp: %v came before %v", 
				timestamps[i], timestamps[i-1])
		}
	}
	
	// Check that string comparison also maintains order
	for i := 1; i < len(xids); i++ {
		if strings.Compare(xids[i], xids[i-1]) < 0 {
			t.Errorf("Xids are not lexicographically sorted: %s came before %s", 
				xids[i], xids[i-1])
		}
	}
}

func TestXidParser_ComponentExtraction(t *testing.T) {
	parser := &XidParser{}
	
	// Generate an Xid and check component extraction
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate Xid: %v", err)
	}
	
	info, err := parser.Parse(generated)
	if err != nil {
		t.Fatalf("Failed to parse generated Xid: %v", err)
	}
	
	// Check that machine ID is extracted
	if info.Node1 == nil {
		t.Error("Machine ID should be extracted")
	} else {
		machineId := *info.Node1
		if len(machineId) != 6 {
			t.Errorf("Machine ID should be 6 hex characters, got %d: %s", len(machineId), machineId)
		}
		// Should be valid hex
		for _, char := range machineId {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("Machine ID should be hex, got character '%c': %s", char, machineId)
			}
		}
	}
	
	// Check that process ID is extracted
	if info.Node2 == nil {
		t.Error("Process ID should be extracted")
	} else {
		processId := *info.Node2
		if len(processId) != 4 {
			t.Errorf("Process ID should be 4 hex characters, got %d: %s", len(processId), processId)
		}
		// Should be valid hex
		for _, char := range processId {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("Process ID should be hex, got character '%c': %s", char, processId)
			}
		}
	}
	
	// Check that counter is reasonable
	if info.Sequence == nil {
		t.Error("Counter should be extracted")
	} else {
		counter := *info.Sequence
		if counter < 0 || counter > 16777215 { // 2^24 - 1
			t.Errorf("Counter should be 0-16777215, got %d", counter)
		}
	}
}

func TestXidParser_EdgeCases(t *testing.T) {
	parser := &XidParser{}
	
	// Test edge cases with the official library
	// We'll generate some Xids and test edge cases around them
	testXid := xid.New().String()
	
	edgeCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{testXid, true, "valid generated Xid"},
		{strings.ToUpper(testXid), false, "uppercase (might be invalid)"},
		{testXid[:19], false, "truncated (19 chars)"},
		{testXid + "x", false, "extended (21 chars)"},
		{"00000000000000000000", true, "all zeros (if valid)"}, // Will test with official library
		{"vvvvvvvvvvvvvvvvvvvv", false, "all v's (might be invalid)"},
	}
	
	for _, tc := range edgeCases {
		// Use the official library as the source of truth for validity
		expectedResult := tc.shouldParse
		if tc.input != testXid { // Don't re-test the known good one
			if _, err := xid.FromString(tc.input); err != nil {
				expectedResult = false
			} else {
				expectedResult = true
			}
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

func TestXidParser_Alphabet(t *testing.T) {
	// Test that the regex matches the correct Crockford Base32 alphabet for Xid
	// Xid uses 0-9, a-v (20 characters total in the string representation)
	validChars := "0123456789abcdefghijklmnopqrstuv"
	
	// Test that each character is individually accepted in a 20-char string
	for _, char := range validChars {
		testStr := strings.Repeat(string(char), 20)
		if !xidRegex.MatchString(testStr) {
			t.Errorf("Regex should match valid Xid character '%c'", char)
		}
	}
	
	// Test invalid characters
	invalidChars := "wxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|;:'\",.<>/?"
	for _, char := range invalidChars {
		testStr := strings.Repeat(string(char), 20)
		if xidRegex.MatchString(testStr) {
			t.Errorf("Regex should not match invalid Xid character '%c'", char)
		}
	}
}