package parsers

import (
	"testing"
)

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
	
	uuid, err := parser.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if !parser.CanParse(uuid) {
		t.Errorf("Generated UUID %q cannot be parsed by the same parser", uuid)
	}
}