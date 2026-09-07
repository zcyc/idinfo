package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadStdin(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "first line", input: " 550e8400-e29b-41d4-a716-446655440000 \nignored", want: " 550e8400-e29b-41d4-a716-446655440000 "},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := readStdin(strings.NewReader(test.input))
			if err != nil {
				t.Fatalf("readStdin() error = %v", err)
			}
			if got != test.want {
				t.Errorf("readStdin() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNormalizeIDArgs(t *testing.T) {
	args := []string{"-f", "pushid", "-OFrJ24CPTXLcIPPjvh3"}
	want := []string{"-f", "pushid", "--", "-OFrJ24CPTXLcIPPjvh3"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeIDArgs() = %#v, want %#v", got, want)
	}

	args = []string{"01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, args) {
		t.Fatalf("normalizeIDArgs() changed a normal ID: %#v", got)
	}

	args = []string{"-e", "127.0.0.1"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, args) {
		t.Fatalf("normalizeIDArgs() changed a known boolean flag: %#v", got)
	}

	args = []string{"--force", "pushid", "-OFrJ24CPTXLcIPPjvh3"}
	wantLong := []string{"--force", "pushid", "--", "-OFrJ24CPTXLcIPPjvh3"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantLong) {
		t.Fatalf("normalizeIDArgs() mishandled a long value flag: %#v", got)
	}

	args = []string{"-fpuid", "aeby6ob5sso4zd"}
	wantUnsupportedShortValue := []string{"--", "-fpuid", "aeby6ob5sso4zd"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantUnsupportedShortValue) {
		t.Fatalf("normalizeIDArgs() accepted an unsupported attached short value: %#v", got)
	}

	args = []string{"-f=pushid", "aeby6ob5sso4zd"}
	wantEquals := []string{"--", "-f=pushid", "aeby6ob5sso4zd"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantEquals) {
		t.Fatalf("normalizeIDArgs() mishandled an equals-form short value: %#v", got)
	}

	args = []string{"550e8400-e29b-41d4-a716-446655440000", "-o", "json"}
	wantTrailing := []string{"-o", "json", "550e8400-e29b-41d4-a716-446655440000"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantTrailing) {
		t.Fatalf("normalizeIDArgs() did not move trailing options: %#v", got)
	}

	args = []string{"-er", "550e8400-e29b-41d4-a716-446655440000"}
	wantCluster := []string{"-e", "-r", "550e8400-e29b-41d4-a716-446655440000"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantCluster) {
		t.Fatalf("normalizeIDArgs() mishandled a boolean cluster: %#v", got)
	}

	args = []string{"550e8400-e29b-41d4-a716-446655440000", "-ojson"}
	wantAttachedCluster := []string{"-o", "json", "550e8400-e29b-41d4-a716-446655440000"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantAttachedCluster) {
		t.Fatalf("normalizeIDArgs() mishandled an attached value: %#v", got)
	}
}

func TestRelativeDuration(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{0, "in a few seconds"},
		{44, "in a few seconds"},
		{45, "in a minute"},
		{90, "in 2 minutes"},
		{45 * 60, "in an hour"},
		{22 * 60 * 60, "in a day"},
		{26 * 24 * 60 * 60, "in a month"},
		{-44, "a few seconds ago"},
		{-90, "2 minutes ago"},
		{-24 * 60 * 60, "a day ago"},
	}
	for _, test := range tests {
		if got := relativeDuration(test.seconds); got != test.want {
			t.Errorf("relativeDuration(%d) = %q, want %q", test.seconds, got, test.want)
		}
	}
}
