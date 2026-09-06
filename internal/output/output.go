package output

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/zcyc/idinfo/internal/types"
)

// ShowCard displays the ID information in a card format
func ShowCard(info *types.IDInfo) {
	// Create the card
	fmt.Println("┏━━━━━━━━━━━┯━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")

	// ID Type
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "ID Type", info.IDType)

	// Rust's card always includes the optional fields as "-".
	version := info.Version
	if version == "" {
		version = "-"
	}
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "Version", version)

	fmt.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Standard representation
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "String", info.Standard)
	// Integer representation
	if info.Integer != nil {
		intStr := *info.Integer
		if len(intStr) > 43 {
			intStr = intStr[:40] + "..."
		}
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Integer", intStr)
	}
	if info.UUIDWrap != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "UUID wrap", *info.UUIDWrap)
	}

	fmt.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Size and entropy
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "Size", sizeDescription(info))
	entropy := "-"
	if info.Size > 0 {
		value := 0
		if info.Entropy != nil {
			value = *info.Entropy
		}
		entropy = fmt.Sprintf("%d bits", value)
	}
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "Entropy", entropy)

	// Timestamp
	timeStr := "-"
	if info.Timestamp != nil {
		timeStr = *info.Timestamp
		if info.DateTime != nil {
			timeStr = fmt.Sprintf("%s (%s)", timeStr, info.DateTime.UTC().Format("2006-01-02T15:04:05.000Z07:00"))
		}
	}
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "Timestamp", timeStr)
	if info.Relative != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Relative", *info.Relative)
	}

	// Node information
	if info.Node1 != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Node 1", *info.Node1)
	} else {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Node 1", "-")
	}

	if info.Node2 != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Node 2", *info.Node2)
	} else {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Node 2", "-")
	}
	if info.Node3 != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Node 3", *info.Node3)
	}

	// Sequence
	if info.Sequence != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Sequence", fmt.Sprintf("%d", *info.Sequence))
	} else {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Sequence", "-")
	}

	fmt.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Show hex and binary representation
	hex := info.Hex
	if len(hex) > 0 {
		// Format hex in groups of 4 characters
		var hexGroups []string
		for i := 0; i < len(hex); i += 8 {
			end := i + 8
			if end > len(hex) {
				end = len(hex)
			}
			group := hex[i:end]
			// Split into 4-char chunks
			var subgroups []string
			for j := 0; j < len(group); j += 4 {
				subEnd := j + 4
				if subEnd > len(group) {
					subEnd = len(group)
				}
				subgroups = append(subgroups, group[j:subEnd])
			}
			hexGroups = append(hexGroups, strings.Join(subgroups, " "))
		}

		// Show hex and binary
		for i, group := range hexGroups {
			if i < len(hexGroups) {
				// Convert to binary
				binaryStr := ""
				for _, char := range strings.ReplaceAll(group, " ", "") {
					if char >= '0' && char <= '9' {
						val := int(char - '0')
						binaryStr += fmt.Sprintf("%04b ", val)
					} else if char >= 'a' && char <= 'f' {
						val := int(char - 'a' + 10)
						binaryStr += fmt.Sprintf("%04b ", val)
					} else if char >= 'A' && char <= 'F' {
						val := int(char - 'A' + 10)
						binaryStr += fmt.Sprintf("%04b ", val)
					}
				}
				binaryStr = strings.TrimSpace(binaryStr)

				fmt.Printf("┃ %-9s │ %-43s ┃\n", group, binaryStr)
			}
		}
	}

	fmt.Println("┗━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
}

func sizeDescription(info *types.IDInfo) string {
	if info.Size == 0 {
		return "-"
	}
	if info.Parsed != "" {
		return fmt.Sprintf("%d bits (%s)", info.Size, info.Parsed)
	}
	return fmt.Sprintf("%d bits", info.Size)
}

// ShowShort displays a short one-line summary
func ShowShort(info *types.IDInfo) {
	if info.Version != "" {
		fmt.Printf("ID Type: %s, version: %s.\n", info.IDType, info.Version)
	} else {
		fmt.Printf("ID Type: %s.\n", info.IDType)
	}
}

// ShowBinary outputs the raw binary representation
func ShowBinary(info *types.IDInfo) {
	if info.Binary != nil {
		os.Stdout.Write(info.Binary)
	}
}

// ShowEverything displays all successful parses
func ShowEverything(results []*types.IDInfo) {
	fmt.Printf("Successfully parsed as %d different formats:\n\n", len(results))

	for i, info := range results {
		fmt.Printf("=== Format %d: %s ===\n", i+1, info.IDType)
		ShowCard(info)
		fmt.Println()
	}
}

// ShowComparison shows timestamps from different formats sorted by date
func ShowComparison(results []*types.IDInfo) {
	type timestampInfo struct {
		format    string
		timestamp time.Time
	}

	var timestamps []timestampInfo

	for _, info := range results {
		if info.DateTime != nil {
			format := info.IDType
			if info.Version != "" {
				format += ": " + info.Version
			}
			timestamps = append(timestamps, timestampInfo{
				format:    format,
				timestamp: *info.DateTime,
			})
		}
	}
	if len(timestamps) == 0 {
		fmt.Println("This ID is not valid in any time-aware format.")
		return
	}

	now := time.Now().UTC()
	timestamps = append(timestamps, timestampInfo{format: "--- Now ---", timestamp: now})

	// Sort by timestamp
	sort.SliceStable(timestamps, func(i, j int) bool {
		return timestamps[i].timestamp.Before(timestamps[j].timestamp)
	})

	fmt.Println("Date/times of the valid IDs parsed as:")

	for _, ts := range timestamps {
		fmt.Printf("- %s %s\n", ts.timestamp.UTC().Format("2006-01-02T15:04:05.000Z07:00"), ts.format)
	}
}
