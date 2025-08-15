package parsers

import (
	"strings"
	"testing"
)

func TestHashHexParser_Name(t *testing.T) {
	parser := &HashHexParser{}
	if parser.Name() != "HashHex" {
		t.Errorf("Expected name 'HashHex', got '%s'", parser.Name())
	}
}

func TestHashHexParser_CanParse(t *testing.T) {
	parser := &HashHexParser{}
	
	// Test valid hex hashes
	validHashes := []string{
		"d41d8cd98f00b204e9800998ecf8427e",                                                         // MD5 (32 hex chars)
		"da39a3ee5e6b4b0d3255bfef95601890afd80709",                                                 // SHA-1 (40 hex chars) 
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",                         // SHA-256 (64 hex chars)
		"38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b", // SHA-384 (96 hex chars)
		"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", // SHA-512 (128 hex chars)
		"d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35",                         // SHA-256 of "hello"
		"12345678",                                                                                 // Short hash (8 chars)
		"ABCDEF1234567890",                                                                         // Mixed case
		"0123456789abcdef",                                                                         // 16 hex chars
	}
	
	for _, hash := range validHashes {
		if !parser.CanParse(hash) {
			t.Errorf("Expected to parse valid hex hash: %s", hash)
		}
	}
	
	// Test invalid hashes
	invalidHashes := []string{
		"",                    // empty
		"123",                 // odd length
		"12345",               // odd length  
		"1234567",             // too short (7 chars)
		"xyz123",              // contains non-hex characters
		"12g45678",            // contains 'g' (invalid hex)
		"hello world",         // not hex at all
		"12-34-56-78",         // contains hyphens
		"12 34 56 78",         // contains spaces
		"!@#$%^&*",            // special characters
	}
	
	for _, hash := range invalidHashes {
		if parser.CanParse(hash) {
			t.Errorf("Expected to reject invalid hex hash: %s", hash)
		}
	}
}

func TestHashHexParser_Parse(t *testing.T) {
	parser := &HashHexParser{}
	
	// Test MD5 hash
	md5Hash := "d41d8cd98f00b204e9800998ecf8427e"
	info, err := parser.Parse(md5Hash)
	if err != nil {
		t.Fatalf("Failed to parse MD5 hash: %v", err)
	}
	
	if info.IDType != "Hex-encoded MD5" {
		t.Errorf("Expected IDType 'Hex-encoded MD5', got '%s'", info.IDType)
	}
	
	if info.Standard != strings.ToUpper(md5Hash) {
		t.Errorf("Expected Standard '%s', got '%s'", strings.ToUpper(md5Hash), info.Standard)
	}
	
	if info.Size != 128 { // MD5 is 128 bits
		t.Errorf("Expected Size 128, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy != 128 {
		t.Errorf("Expected Entropy 128, got %v", info.Entropy)
	}
	
	if info.Extra["probable_algorithm"] != "MD5" {
		t.Errorf("Expected probable_algorithm 'MD5', got '%s'", info.Extra["probable_algorithm"])
	}
	
	if info.Extra["cryptographic_strength"] != "broken (collisions found)" {
		t.Errorf("Expected cryptographic_strength warning for MD5, got '%s'", info.Extra["cryptographic_strength"])
	}
	
	// Test SHA-256 hash
	sha256Hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	info, err = parser.Parse(sha256Hash)
	if err != nil {
		t.Fatalf("Failed to parse SHA-256 hash: %v", err)
	}
	
	if info.IDType != "Hex-encoded SHA-256" {
		t.Errorf("Expected IDType 'Hex-encoded SHA-256', got '%s'", info.IDType)
	}
	
	if info.Size != 256 { // SHA-256 is 256 bits
		t.Errorf("Expected Size 256, got %d", info.Size)
	}
	
	if info.Extra["probable_algorithm"] != "SHA-256" {
		t.Errorf("Expected probable_algorithm 'SHA-256', got '%s'", info.Extra["probable_algorithm"])
	}
	
	if info.Extra["cryptographic_strength"] != "strong" {
		t.Errorf("Expected cryptographic_strength 'strong' for SHA-256, got '%s'", info.Extra["cryptographic_strength"])
	}
	
	// Test unknown length hash
	unknownHash := "123456789abcdef0"
	info, err = parser.Parse(unknownHash)
	if err != nil {
		t.Fatalf("Failed to parse unknown hash: %v", err)
	}
	
	if !strings.Contains(info.IDType, "Hash (64 bits)") {
		t.Errorf("Expected IDType to contain 'Hash (64 bits)', got '%s'", info.IDType)
	}
	
	// Test case insensitivity
	mixedCaseHash := "D41D8CD98F00B204E9800998ECF8427E"
	info, err = parser.Parse(mixedCaseHash)
	if err != nil {
		t.Fatalf("Failed to parse mixed case hash: %v", err)
	}
	
	if info.Standard != mixedCaseHash { // Should be converted to uppercase
		t.Errorf("Expected Standard to be uppercase, got '%s'", info.Standard)
	}
	
	// Test invalid input
	_, err = parser.Parse("invalid_hash")
	if err == nil {
		t.Error("Expected error for invalid hash")
	}
}

func TestHashHexParser_Generate(t *testing.T) {
	parser := &HashHexParser{}
	
	// Test generation
	generated, err := parser.Generate()
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	
	// Check if generated hash can be parsed
	if !parser.CanParse(generated) {
		t.Errorf("Generated hash is not valid: %s", generated)
	}
	
	// Check length (should be 64 hex chars for SHA-256)
	if len(generated) != 64 {
		t.Errorf("Generated hash should be 64 characters, got %d: %s", len(generated), generated)
	}
	
	// Check character set (hex only)
	for _, char := range generated {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Generated hash contains invalid character '%c': %s", char, generated)
		}
	}
	
	// Test parsing the generated hash
	info, err := parser.Parse(generated)
	if err != nil {
		t.Errorf("Failed to parse generated hash: %v", err)
	}
	
	if info.IDType != "Hex-encoded SHA-256" {
		t.Errorf("Expected generated hash to be identified as SHA-256, got: %s", info.IDType)
	}
}

func TestHashHexParser_AllHashTypes(t *testing.T) {
	parser := &HashHexParser{}
	
	// Test all common hash types
	hashTests := []struct {
		name     string
		length   int
		expected string
	}{
		{"MD5", 32, "MD5"},
		{"SHA-1", 40, "SHA-1"},
		{"SHA-224", 56, "SHA-224"},
		{"SHA-256", 64, "SHA-256"},
		{"SHA-384", 96, "SHA-384"},
		{"SHA-512", 128, "SHA-512"},
	}
	
	for _, test := range hashTests {
		// Create a hex string of the required length
		hexString := strings.Repeat("a", test.length)
		
		info, err := parser.Parse(hexString)
		if err != nil {
			t.Errorf("Failed to parse %s hash: %v", test.name, err)
			continue
		}
		
		expectedIDType := "Hex-encoded " + test.expected
		if info.IDType != expectedIDType {
			t.Errorf("Expected IDType '%s', got '%s'", expectedIDType, info.IDType)
		}
		
		expectedSize := test.length * 4 // Each hex char is 4 bits
		if info.Size != expectedSize {
			t.Errorf("Expected Size %d for %s, got %d", expectedSize, test.name, info.Size)
		}
		
		if info.Extra["probable_algorithm"] != test.expected {
			t.Errorf("Expected probable_algorithm '%s', got '%s'", test.expected, info.Extra["probable_algorithm"])
		}
	}
}

func TestHashHexParser_ExtraFields(t *testing.T) {
	parser := &HashHexParser{}
	
	hash := "d41d8cd98f00b204e9800998ecf8427e" // MD5
	info, err := parser.Parse(hash)
	if err != nil {
		t.Fatalf("Failed to parse hash: %v", err)
	}
	
	// Check common extra fields
	if info.Extra["encoding"] != "hexadecimal" {
		t.Errorf("Expected encoding 'hexadecimal', got '%s'", info.Extra["encoding"])
	}
	
	if info.Extra["byte_length"] != "16" { // MD5 is 16 bytes
		t.Errorf("Expected byte_length '16', got '%s'", info.Extra["byte_length"])
	}
	
	if info.Extra["deterministic"] != "depends on hash function" {
		t.Errorf("Expected deterministic info, got '%s'", info.Extra["deterministic"])
	}
	
	// Check MD5-specific fields
	if info.Extra["recommended_use"] != "checksums only" {
		t.Errorf("Expected MD5 recommendation, got '%s'", info.Extra["recommended_use"])
	}
}

func TestHashHexParser_BinaryConversion(t *testing.T) {
	parser := &HashHexParser{}
	
	// Test with known hex string
	hexString := "48656c6c6f" // "Hello" in hex
	info, err := parser.Parse(hexString)
	if err != nil {
		t.Fatalf("Failed to parse hex string: %v", err)
	}
	
	// Check binary conversion
	expectedBytes := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}
	if len(info.Binary) != len(expectedBytes) {
		t.Errorf("Expected %d bytes, got %d", len(expectedBytes), len(info.Binary))
	}
	
	for i, expected := range expectedBytes {
		if i < len(info.Binary) && info.Binary[i] != expected {
			t.Errorf("Binary mismatch at index %d: expected %02x, got %02x", i, expected, info.Binary[i])
		}
	}
	
	// Check hex representation
	if info.Hex != strings.ToLower(hexString) {
		t.Errorf("Expected hex '%s', got '%s'", strings.ToLower(hexString), info.Hex)
	}
}

func TestHashHexParser_Performance(t *testing.T) {
	parser := &HashHexParser{}
	
	testHash := "d41d8cd98f00b204e9800998ecf8427e"
	
	// Test parsing performance
	for i := 0; i < 1000; i++ {
		_, err := parser.Parse(testHash)
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