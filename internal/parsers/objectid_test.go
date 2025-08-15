package parsers

import (
	"testing"
	"time"
)

func TestObjectIDParser_CanParse(t *testing.T) {
	parser := &ObjectIDParser{}
	
	tests := []struct {
		input    string
		expected bool
	}{
		{"507f1f77bcf86cd799439011", true},
		{"5f8a7b2d3c4e5f6a7b8c9d0e", true},
		{"not-an-objectid", false},
		{"507f1f77bcf86cd79943901", false}, // too short
		{"507f1f77bcf86cd799439011a", false}, // too long
		{"507f1f77bcf86cd799439g11", false}, // invalid hex
		{"", false},
	}
	
	for _, test := range tests {
		result := parser.CanParse(test.input)
		if result != test.expected {
			t.Errorf("CanParse(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestObjectIDParser_Parse(t *testing.T) {
	parser := &ObjectIDParser{}
	
	info, err := parser.Parse("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if info.IDType != "MongoDB ObjectId" {
		t.Errorf("Expected MongoDB ObjectId type, got %q", info.IDType)
	}
	
	if info.Size != 96 {
		t.Errorf("Expected size 96, got %d", info.Size)
	}
	
	if info.Entropy == nil || *info.Entropy != 40 {
		t.Errorf("Expected entropy 40, got %v", info.Entropy)
	}
	
	if info.DateTime == nil {
		t.Errorf("Expected timestamp to be present")
	}
	
	if info.Node1 == nil {
		t.Errorf("Expected Node1 (machine identifier) to be present")
	}
	
	if info.Node2 == nil {
		t.Errorf("Expected Node2 (process identifier) to be present")
	}
	
	if info.Sequence == nil {
		t.Errorf("Expected Sequence (counter) to be present")
	}
}

func TestObjectIDParser_Generate(t *testing.T) {
	parser := &ObjectIDParser{}
	
	oid, err := parser.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	
	if !parser.CanParse(oid) {
		t.Errorf("Generated ObjectId %q cannot be parsed by the same parser", oid)
	}
	
	// Parse the generated ObjectId and check that timestamp is recent
	info, err := parser.Parse(oid)
	if err != nil {
		t.Fatalf("Parse of generated ObjectId failed: %v", err)
	}
	
	if info.DateTime == nil {
		t.Fatalf("Generated ObjectId should have a timestamp")
	}
	
	now := time.Now()
	diff := now.Sub(*info.DateTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Minute {
		t.Errorf("Generated ObjectId timestamp is too far from current time: %v", diff)
	}
}