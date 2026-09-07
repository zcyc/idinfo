package output

import (
	"fmt"
	"strings"

	"github.com/zcyc/idinfo/internal/types"

	"github.com/fatih/color"
)

// Color definitions
var (
	versionColor   = color.New(color.FgYellow)
	entropyColor   = color.New(color.FgGreen)
	timestampColor = color.New(color.FgCyan)
	node1Color     = color.New(color.FgMagenta)
	node2Color     = color.New(color.FgRed)
	node3Color     = color.RGB(208, 135, 112)
	sequenceColor  = color.New(color.FgBlue)
	nowColor       = color.New(color.FgYellow)
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
		fmt.Print(left)
		fmt.Print(strings.Repeat(char, 10))
		fmt.Print(middle)
		fmt.Print(strings.Repeat(char, rSpace))
		fmt.Println(right)
	}
	row := func(label, value string) {
		fmt.Print("┃ ")
		printColoredLabel(label)
		fmt.Printf(" │ %-*s ┃\n", rSpace, value)
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
	plainLines := cardBinaryLines(info.Hex)
	for index, line := range coloredCardBinaryLines(info.Hex, info.ColorMap) {
		width := len([]rune(plainLines[index].binary))
		fmt.Printf("┃ %-9s │ %s%s ┃\n", line.hex, line.binary, strings.Repeat(" ", maxInt(0, rSpace-width)))
	}
	border("━", "┗━", "┷", "━━┛")
}

func printColoredLabel(label string) {
	switch label {
	case "Version":
		versionColor.Printf("%-9s", label)
	case "Entropy":
		entropyColor.Printf("%-9s", label)
	case "Timestamp", "Relative":
		timestampColor.Printf("%-9s", label)
	case "Node 1":
		node1Color.Printf("%-9s", label)
	case "Node 2":
		node2Color.Printf("%-9s", label)
	case "Node 3":
		node3Color.Printf("%-9s", label)
	case "Sequence":
		sequenceColor.Printf("%-9s", label)
	default:
		fmt.Printf("%-9s", label)
	}
}

func coloredCardBinaryLines(value, colorMap string) []cardBinaryLine {
	lines := cardBinaryLines(value)
	if colorMap == "" {
		return lines
	}

	bitIndex := 0
	for lineIndex := range lines {
		var rendered strings.Builder
		for _, char := range lines[lineIndex].binary {
			if char != '0' && char != '1' {
				rendered.WriteRune(char)
				continue
			}
			code := byte('0')
			if bitIndex < len(colorMap) {
				code = colorMap[bitIndex]
			}
			rendered.WriteString(colorBinaryBit(char, code))
			bitIndex++
		}
		lines[lineIndex].binary = rendered.String()
	}
	return lines
}

func colorBinaryBit(bit rune, code byte) string {
	value := string(bit)
	switch code {
	case '1':
		return versionColor.Sprint(value)
	case '2':
		return entropyColor.Sprint(value)
	case '3':
		return timestampColor.Sprint(value)
	case '4':
		return node1Color.Sprint(value)
	case '5':
		return node2Color.Sprint(value)
	case '6':
		return sequenceColor.Sprint(value)
	case '7':
		return node3Color.Sprint(value)
	default:
		return value
	}
}
