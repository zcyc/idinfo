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
	if color.NoColor {
		ShowCard(info)
		return
	}
	timestamp := cardTimestamp(info)
	rSpace := maxInt(43, len([]rune(timestamp)))
	border := func(char, left, middle, right string) {
		borderColor.Print(left)
		borderColor.Print(strings.Repeat(char, 10))
		borderColor.Print(middle)
		borderColor.Print(strings.Repeat(char, rSpace))
		borderColor.Println(right)
	}
	row := func(label, value string) {
		borderColor.Print("┃ ")
		labelColor.Printf("%-9s ", label)
		borderColor.Print("│ ")
		valueColor.Printf("%-*s ", rSpace, value)
		borderColor.Println("┃")
	}
	border("━", "┏━", "┯", "━━┓")
	row("ID Type", info.IDType)
	version := info.Version
	if version == "" {
		version = "-"
	}
	row("Version", version)
	border("─", "┠─", "┼", "──┨")
	row("String", truncateCardValue(info.Standard))
	if info.Integer != nil {
		row("Integer", *info.Integer)
	}
	if info.UUIDWrap != nil {
		row("UUID wrap", *info.UUIDWrap)
	}
	border("─", "┠─", "┼", "──┨")
	row("Size", sizeDescription(info))

	entropy := "-"
	if info.Size > 0 {
		value := 0
		if info.Entropy != nil {
			value = *info.Entropy
		}
		entropy = fmt.Sprintf("%d bits", value)
	}
	row("Entropy", entropy)
	row("Timestamp", timestamp)
	if info.Relative != nil {
		row("Relative", *info.Relative)
	}
	row("Node 1", truncateCardValue(pointerValue(info.Node1, "-")))
	row("Node 2", truncateCardValue(pointerValue(info.Node2, "-")))
	if info.Node3 != nil {
		row("Node 3", truncateCardValue(*info.Node3))
	}
	row("Sequence", pointerInt64Value(info.Sequence, "-"))
	border("─", "┠─", "┼", "──┨")
	for _, line := range cardBinaryLines(info.Hex) {
		borderColor.Print("┃ ")
		color.New(color.FgCyan).Printf("%-9s ", line.hex)
		borderColor.Print("│ ")
		binaryColor.Printf("%-*s ", rSpace, line.binary)
		borderColor.Println("┃")
	}
	border("━", "┗━", "┷", "━━┛")
}
