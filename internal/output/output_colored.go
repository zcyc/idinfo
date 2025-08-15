package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/zcyc/idinfo/internal/types"

	"github.com/fatih/color"
)

// Color definitions
var (
	headerColor = color.New(color.FgCyan, color.Bold)
	labelColor  = color.New(color.FgWhite, color.Bold)
	valueColor  = color.New(color.FgGreen)
	binaryColor = color.New(color.FgYellow)
	borderColor = color.New(color.FgBlue)
)

// ShowCardColored displays the ID information in a colorful card format
func ShowCardColored(info *types.IDInfo) {
	// Create the card
	borderColor.Println("┏━━━━━━━━━━━┯━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")

	// ID Type
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "ID Type")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", info.IDType)
	borderColor.Println("┃")

	// Version (if available)
	if info.Version != "" {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Version")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", info.Version)
		borderColor.Println("┃")
	}

	borderColor.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Standard representation
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "String")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", info.Standard)
	borderColor.Println("┃")

	// Integer representation
	if info.Integer != nil {
		intStr := *info.Integer
		if len(intStr) > 43 {
			intStr = intStr[:40] + "..."
		}
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Integer")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", intStr)
		borderColor.Println("┃")
	}

	// Additional representations
	if info.ShortUUID != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "ShortUUID")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.ShortUUID)
		borderColor.Println("┃")
	}
	if info.Base64 != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Base64")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.Base64)
		borderColor.Println("┃")
	}

	borderColor.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Size and entropy
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "Size")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", fmt.Sprintf("%d bits", info.Size))
	borderColor.Println("┃")

	if info.Entropy != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Entropy")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", fmt.Sprintf("%d bits", *info.Entropy))
		borderColor.Println("┃")
	}

	// Timestamp
	if info.DateTime != nil {
		timeStr := info.DateTime.Format(time.RFC3339)
		if info.Timestamp != nil {
			timeStr = fmt.Sprintf("%s (%s)", *info.Timestamp, timeStr)
		}
		if len(timeStr) > 43 {
			timeStr = timeStr[:40] + "..."
		}
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Timestamp")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", timeStr)
		borderColor.Println("┃")
	}

	// Node information
	if info.Node1 != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Node 1")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.Node1)
		borderColor.Println("┃")
	} else {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Node 1")
		borderColor.Print("│ ")
		color.New(color.FgHiBlack).Printf("%-43s ", "-")
		borderColor.Println("┃")
	}

	if info.Node2 != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Node 2")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.Node2)
		borderColor.Println("┃")
	} else {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Node 2")
		borderColor.Print("│ ")
		color.New(color.FgHiBlack).Printf("%-43s ", "-")
		borderColor.Println("┃")
	}

	// Sequence
	if info.Sequence != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Sequence")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", fmt.Sprintf("%d", *info.Sequence))
		borderColor.Println("┃")
	} else {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Sequence")
		borderColor.Print("│ ")
		color.New(color.FgHiBlack).Printf("%-43s ", "-")
		borderColor.Println("┃")
	}

	borderColor.Println("┠───────────┼─────────────────────────────────────────────┨")

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

				borderColor.Print("┃ ")
				color.New(color.FgCyan).Printf("%-9s ", group)
				borderColor.Print("│ ")
				binaryColor.Printf("%-43s ", binaryStr)
				borderColor.Println("┃")
			}
		}
	}

	borderColor.Println("┗━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
}
