package parsers

import (
	"testing"
	"time"
)

func TestULIDParser_CanParse(t *testing.T) {
	parser := &ULIDParser{}
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"01BX5ZZKBKACTAV9WEVGEMMVRZ", true},
		{"not-a-ulid", false},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA", false}, // too short
		{"01ARZ3NDEKTSV4RRFFQ69G5FAVV", false}, // too long
		{"", false},
	}
	
	for _, test := range tests {
		result := parser.CanParse(test.input)
		if result != test.expected {
			t.Errorf("CanParse(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestULIDParser_Parse(t *testing.T) {
	parser := &ULIDParser{}
	
	info, err := parser.Parse("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if info.IDType != "ULID (Universally Unique Lexicographically Sortable Identifier)" {
		t.Errorf("Expected ULID type, got %q", info.IDType)
	}
	
	if info.Size != 128 {
		t.Errorf("Expected size 128, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy != 80 {
		t.Errorf("Expected entropy 80, got %v", info.Entropy)
	}
	
	if info.DateTime == nil {
		t.Errorf("Expected timestamp to be present")
	} else {
		// Check if timestamp is reasonable (not too far in past/future)
		now := time.Now()
		diff := now.Sub(*info.DateTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > 50*365*24*time.Hour { // 50 years
			t.Errorf("Timestamp seems unreasonable: %v", info.DateTime)
		}
	}
}

func TestULIDParser_Generate(t *testing.T) {
	parser := &ULIDParser{}
	
	ulid, err := parser.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if !parser.CanParse(ulid) {
		t.Errorf("Generated ULID %q cannot be parsed by the same parser", ulid)
	}
}