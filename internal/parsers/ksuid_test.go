package parsers

import (
	"strings"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
)

func TestKSUIDParser_Name(t *testing.T) {
	parser := &KSUIDParser{}
	if parser.Name() != "KSUID" {
		t.Errorf("Expected name 'KSUID', got '%s'", parser.Name())
	}
}

func TestKSUIDParser_CanParse(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Test valid KSUIDs (generate some using the official library)
	validKSUIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		validKSUIDs[i] = ksuid.New().String()
	}
	
	for _, id := range validKSUIDs {
		if !parser.CanParse(id) {
			t.Errorf("Expected to parse valid KSUID: %s", id)
		}
	}
	
	// Test some known valid KSUID patterns
	additionalValid := []string{
		"0o5Fs0EELR0fUjHjbCnEtdUwQe3", // Example KSUID
		"0ujtsYcgvSTl8PAuAdqWYSMnLOv", // Another example
		"0ujsszwN8NRY24YaXiTIE2VWDTS", // Another example
	}
	
	for _, id := range additionalValid {
		// Only test if the official library considers it valid
		if _, err := ksuid.Parse(id); err == nil {
			if !parser.CanParse(id) {
				t.Errorf("Expected to parse valid KSUID: %s", id)
			}
		}
	}
	
	// Test invalid KSUIDs
	invalidKSUIDs := []string{
		"",                          // empty
		"abc",                       // too short
		"12345678901234567890123456789", // too long (28 chars)
		"1234567890123456789012345",  // too short (26 chars)
		"!@#$%^&*()1234567890123456", // invalid characters
		"0000000000000000000000000",  // 27 chars but invalid format
		"ZZZZZZZZZZZZZZZZZZZZZZZZZZ", // valid length but invalid content
		"0o5Fs0EELR0fUjHjbCnEtdUwQe", // 26 chars (too short)
		"0o5Fs0EELR0fUjHjbCnEtdUwQe33", // 28 chars (too long)
	}
	
	for _, id := range invalidKSUIDs {
		if parser.CanParse(id) {
			t.Errorf("Expected to reject invalid KSUID: %s", id)
		}
	}
}

func TestKSUIDParser_Parse(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Generate a valid KSUID for testing
	generatedKSUID := ksuid.New()
	validKSUID := generatedKSUID.String()
	
	info, err := parser.Parse(validKSUID)
	if err != nil {
		t.Fatalf("Failed to parse valid KSUID: %v", err)
	}
	
	// Check basic properties
	if info.IDType != "KSUID (K-Sortable Unique Identifier)" {
		t.Errorf("Expected IDType 'KSUID (K-Sortable Unique Identifier)', got '%s'", info.IDType)
	}
	
	if info.Standard != validKSUID {
		t.Errorf("Expected Standard '%s', got '%s'", validKSUID, info.Standard)
	}
	
	// Check size (should be 160 bits - 20 bytes)
	if info.Size != 160 {
		t.Errorf("Expected Size 160, got %d", info.Size)
	}
	
	// Check entropy (should be 128 bits - 16 bytes payload)
	if info.Entropy == nil || *info.Entropy != 128 {
		t.Errorf("Expected Entropy 128, got %v", info.Entropy)
	}
	
	// Check timestamp is present and reasonable (should be recent for generated KSUID)
	if info.DateTime == nil {
		t.Error("Expected DateTime to be set")
	} else {
		now := time.Now()
		if info.DateTime.Before(now.AddDate(-1, 0, 0)) || info.DateTime.After(now.Add(time.Minute)) {
			t.Errorf("KSUID timestamp should be recent, got: %v", info.DateTime)
		}
	}
	
	// Check integer representation
	if info.Integer == nil {
		t.Error("Expected Integer to be set")
	}
	
	// Check extra information
	if info.Extra["encoding"] != "Base62" {
		t.Errorf("Expected encoding 'Base62', got '%s'", info.Extra["encoding"])
	}
	
	if info.Extra["timestamp_precision"] != "second" {
		t.Errorf("Expected timestamp_precision 'second', got '%s'", info.Extra["timestamp_precision"])
	}
	
	if info.Extra["epoch"] != "2014-05-13T16:53:20Z" {
		t.Errorf("Expected epoch '2014-05-13T16:53:20Z', got '%s'", info.Extra["epoch"])
	}
	
	if info.Extra["sortable"] != "true" {
		t.Errorf("Expected sortable 'true', got '%s'", info.Extra["sortable"])
	}
	
	if info.Extra["payload_bytes"] != "16" {
		t.Errorf("Expected payload_bytes '16', got '%s'", info.Extra["payload_bytes"])
	}
	
	// Check hex and binary representations
	if len(info.Hex) != 40 { // 20 bytes = 40 hex chars
		t.Errorf("Expected Hex length 40, got %d: %s", len(info.Hex), info.Hex)
	}
	
	if len(info.Binary) != 20 { // 160 bits = 20 bytes
		t.Errorf("Expected Binary length 20, got %d", len(info.Binary))
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_ksuid")
	if err == nil {
		t.Error("Expected error for invalid KSUID")
	}
}

func TestKSUIDParser_Generate(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate KSUID: %v", err)
	}
	
	// Check if generated KSUID can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated KSUID is not valid: %s", generated)
	}
	
	// Check length
	if len(generated) != 27 {
		t.Errorf("Generated KSUID should be 27 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (Base62: 0-9, A-Z, a-z)
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || 
		     (char >= 'A' && char <= 'Z') || 
		     (char >= 'a' && char <= 'z')) {
			t.Errorf("Generated KSUID contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test that multiple generations produce different results
	generated2, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate second KSUID: %v", err)
	}
	
	if generated == generated2 {
		t.Error("Two consecutive KSUID generations should produce different results")
	}
	
	// Test multiple generations for uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gen, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate KSUID #%d: %v", i, err)
		}
		if seen[gen] {
			t.Errorf("Generated duplicate KSUID: %s", gen)
		}
		seen[gen] = true
	}
}

func TestKSUIDParser_Performance(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Test parsing performance
	testKsuid := ksuid.New().String()
	
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testKsuid)
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

func TestKSUIDParser_OfficialLibraryIntegration(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Test that our parser works with the official library
	officialKSUID := ksuid.New()
	ksuidStr := officialKSUID.String()
	
	// Parse with our parser
	info, err := parser.Parse(ksuidStr)
	if err != nil {
		t.Fatalf("Failed to parse KSUID from official library: %v", err)
	}
	
	// Verify the timestamp matches
	officialTime := officialKSUID.Time()
	if info.DateTime == nil {
		t.Error("DateTime should be set")
	} else if !info.DateTime.Equal(officialTime) {
		t.Errorf("Timestamp mismatch: official=%v, parsed=%v", officialTime, info.DateTime)
	}
	
	// Verify the bytes match
	officialBytes := officialKSUID.Bytes()
	if len(info.Binary) != len(officialBytes) {
		t.Errorf("Binary length mismatch: official=%d, parsed=%d", len(officialBytes), len(info.Binary))
	}
	
	for i, expected := range officialBytes {
		if i < len(info.Binary) && info.Binary[i] != expected {
			t.Errorf("Binary mismatch at index %d: expected %02x, got %02x", i, expected, info.Binary[i])
		}
	}
}

func TestKSUIDParser_Sortability(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Generate KSUIDs with delays to ensure different timestamps
	// KSUID timestamps have second precision, so we need at least 1 second
	var ksuids []string
	for i := 0; i < 3; i++ { // Reduce count for faster test
		generated, err := parser.Generate()
		if err != nil {
			t.Fatalf("Failed to generate KSUID: %v", err)
		}
		ksuids = append(ksuids, generated)
		if i < 2 { // Don't sleep after the last one
			time.Sleep(1100 * time.Millisecond) // Ensure different seconds
		}
	}
	
	// Parse all KSUIDs and check that timestamps are in order
	var timestamps []time.Time
	for _, k := range ksuids {
		info, err := parser.Parse(k)
		if err != nil {
			t.Fatalf("Failed to parse KSUID: %v", err)
		}
		if info.DateTime != nil {
			timestamps = append(timestamps, *info.DateTime)
		}
	}
	
	// Check that timestamps are in ascending order (or equal)
	for i := 1; i < len(timestamps); i++ {
		if timestamps[i].Before(timestamps[i-1]) {
			t.Errorf("KSUIDs are not properly sorted by timestamp: %v came before %v", 
				timestamps[i], timestamps[i-1])
		}
	}
	
	// Check that string comparison also maintains order
	for i := 1; i < len(ksuids); i++ {
		if strings.Compare(ksuids[i], ksuids[i-1]) < 0 {
			t.Errorf("KSUIDs are not lexicographically sorted: %s came before %s", 
				ksuids[i], ksuids[i-1])
		}
	}
}

func TestKSUIDParser_EdgeCases(t *testing.T) {
	parser := &KSUIDParser{}
	
	// Test with known KSUID values (if any)
	knownKSUIDs := []string{
		"0o5Fs0EELR0fUjHjbCnEtdUwQe3", // Example from documentation
	}
	
	for _, k := range knownKSUIDs {
		// Only test if the official library accepts it
		if _, err := ksuid.Parse(k); err == nil {
			if !parser.CanParse(k) {
				t.Errorf("Should be able to parse known KSUID: %s", k)
			}
			
			info, err := parser.Parse(k)
			if err != nil {
				t.Errorf("Failed to parse known KSUID: %v", err)
			} else {
				// Basic validation
				if info.IDType != "KSUID (K-Sortable Unique Identifier)" {
					t.Errorf("Expected correct IDType for known KSUID")
				}
			}
		}
	}
}