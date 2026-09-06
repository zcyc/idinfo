package output

import (
	"fmt"
	"strings"

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

	// Rust's card always includes the optional fields as "-".
	version := info.Version
	if version == "" {
		version = "-"
	}
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "Version")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", version)
	borderColor.Println("┃")

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

	if info.UUIDWrap != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "UUID wrap")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.UUIDWrap)
		borderColor.Println("┃")
	}

	borderColor.Println("┠───────────┼─────────────────────────────────────────────┨")

	// Size and entropy
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "Size")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", sizeDescription(info))
	borderColor.Println("┃")

	entropy := "-"
	if info.Size > 0 {
		value := 0
		if info.Entropy != nil {
			value = *info.Entropy
		}
		entropy = fmt.Sprintf("%d bits", value)
	}
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "Entropy")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", entropy)
	borderColor.Println("┃")

	// Timestamp
	timeStr := "-"
	if info.Timestamp != nil {
		timeStr = *info.Timestamp
		if info.DateTime != nil {
			timeStr = fmt.Sprintf("%s (%s)", timeStr, info.DateTime.UTC().Format("2006-01-02T15:04:05.000Z07:00"))
		}
	}
	if len(timeStr) > 43 {
		timeStr = timeStr[:40] + "..."
	}
	borderColor.Print("┃ ")
	labelColor.Printf("%-9s ", "Timestamp")
	borderColor.Print("│ ")
	valueColor.Printf("%-43s ", timeStr)
	borderColor.Println("┃")
	if info.Relative != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Relative")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.Relative)
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
	if info.Node3 != nil {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", "Node 3")
		borderColor.Print("│ ")
		valueColor.Printf("%-43s ", *info.Node3)
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
