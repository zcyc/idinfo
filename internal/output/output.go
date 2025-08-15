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

	// Version (if available)
	if info.Version != "" {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Version", info.Version)
	}

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

	// Additional representations
	if info.ShortUUID != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "ShortUUID", *info.ShortUUID)
	}
	if info.Base64 != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Base64", *info.Base64)
	}

	fmt.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Size and entropy
	fmt.Printf("┃ %-9s │ %-43s ┃\n", "Size", fmt.Sprintf("%d bits", info.Size))
	if info.Entropy != nil {
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Entropy", fmt.Sprintf("%d bits", *info.Entropy))
	}

	// Timestamp
	if info.DateTime != nil {
		timeStr := info.DateTime.Format(time.RFC3339)
		if info.Timestamp != nil {
			timeStr = fmt.Sprintf("%s (%s)", *info.Timestamp, timeStr)
		}
		fmt.Printf("┃ %-9s │ %-43s ┃\n", "Timestamp", timeStr)
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
		future    bool
	}

	var timestamps []timestampInfo
	now := time.Now()

	for _, info := range results {
		if info.DateTime != nil {
			timestamps = append(timestamps, timestampInfo{
				format:    info.IDType,
				timestamp: *info.DateTime,
				future:    info.DateTime.After(now),
			})
		}
	}

	// Sort by timestamp
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].timestamp.Before(timestamps[j].timestamp)
	})

	fmt.Println("Date/times of the valid IDs parsed as:")

	for _, ts := range timestamps {
		prefix := "- "
		suffix := ""
		if ts.future {
			suffix = " (future)"
		}

		// Check if this is around now
		diff := now.Sub(ts.timestamp)
		if diff < time.Minute && diff > -time.Minute {
			prefix = "- "
			suffix = " --- Now ---"
		}

		fmt.Printf("%s%s %s%s\n",
			prefix,
			ts.timestamp.Format(time.RFC3339),
			ts.format,
			suffix)
	}
}
