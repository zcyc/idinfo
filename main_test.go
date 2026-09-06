package main

import (
	"reflect"
	"testing"
)

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
	wantShortAttached := []string{"-f", "puid", "aeby6ob5sso4zd"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantShortAttached) {
		t.Fatalf("normalizeIDArgs() mishandled an attached short value: %#v", got)
	}

	args = []string{"-f=pushid", "-OFrJ24CPTXLcIPPjvh3"}
	wantEquals := []string{"-f=pushid", "--", "-OFrJ24CPTXLcIPPjvh3"}
	if got := normalizeIDArgs(args); !reflect.DeepEqual(got, wantEquals) {
		t.Fatalf("normalizeIDArgs() mishandled an equals-form short value: %#v", got)
	}
}
