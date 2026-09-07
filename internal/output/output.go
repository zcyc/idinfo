package output

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zcyc/idinfo/internal/types"
)

// ShowCard displays the ID information in a card format
func ShowCard(info *types.IDInfo) {
	timestamp := cardTimestamp(info)
	rSpace := maxInt(43, utf8.RuneCountInString(timestamp))
	border := func(top, left, middle, right string) {
		fmt.Printf("%s%s%s%s%s\n", left, strings.Repeat(top, 10), middle, strings.Repeat(top, rSpace), right)
	}
	row := func(label, value string) {
		fmt.Printf("┃ %-9s │ %-*s ┃\n", label, rSpace, value)
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
		row(line.hex, line.binary)
	}
	border("━", "┗━", "┷", "━━┛")
}

type cardBinaryLine struct{ hex, binary string }

func cardBinaryLines(value string) []cardBinaryLine {
	if value == "" {
		return []cardBinaryLine{{"No hex", "No bits (non-numeric ID)"}}
	}
	padded := value + strings.Repeat(".", (8-len(value)%8)%8)
	lines := make([]cardBinaryLine, 0, len(padded)/8)
	for start := 0; start < len(padded); start += 8 {
		group := padded[start : start+8]
		var binary strings.Builder
		for index, char := range group {
			value := "...."
			if char != '.' {
				nibble, _ := strconv.ParseUint(string(char), 16, 4)
				value = fmt.Sprintf("%04b", nibble)
			}
			binary.WriteString(value)
			binary.WriteByte(' ')
			if (index+1)%2 == 0 {
				binary.WriteByte(' ')
			}
			if (index+1)%4 == 0 {
				binary.WriteByte(' ')
			}
			if (index+1)%8 == 0 {
				binary.WriteByte(' ')
			}
		}
		hexLine := group[:4] + " " + group[4:]
		lines = append(lines, cardBinaryLine{hexLine, strings.TrimSpace(binary.String())})
	}
	return lines
}

func cardTimestamp(info *types.IDInfo) string {
	if info.Timestamp == nil {
		return "-"
	}
	timestamp := *info.Timestamp
	if dot := strings.IndexByte(timestamp, '.'); dot >= 0 {
		end := dot + 4
		if end > len(timestamp) {
			end = len(timestamp)
		}
		timestamp = timestamp[:end]
	}
	if info.DateTime == nil {
		return timestamp + " (-)"
	}
	return timestamp + " (" + cardDateTime(*info.DateTime) + ")"
}

func cardDateTime(value time.Time) string {
	value = value.UTC()
	result := value.Format("2006-01-02T15:04:05.000Z07:00")
	if value.Year() >= 10000 {
		result = "+" + result
	}
	return result
}

func truncateCardValue(value string) string {
	if utf8.RuneCountInString(value) <= 43 {
		return value
	}
	runes := []rune(value)
	return string(runes[:40]) + "..."
}

func pointerValue(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func pointerInt64Value(value *int64, fallback string) string {
	if value == nil {
		return fallback
	}
	return strconv.FormatInt(*value, 10)
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
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
	if info.Integer != nil && info.Size <= 128 {
		if value, ok := new(big.Int).SetString(*info.Integer, 10); ok {
			data := make([]byte, 16)
			bytes := value.Bytes()
			if len(bytes) > len(data) {
				bytes = bytes[len(bytes)-len(data):]
			}
			copy(data[len(data)-len(bytes):], bytes)
			offset := (128 - info.Size) / 8
			_, _ = os.Stdout.Write(data[offset:])
			return
		}
	}
	if info.Hex != "" {
		if data, err := hex.DecodeString(info.Hex); err == nil {
			_, _ = os.Stdout.Write(data)
			return
		}
	}
	if info.Binary != nil {
		_, _ = os.Stdout.Write(info.Binary)
		return
	}
	fmt.Println(info.Standard)
}

// ShowEverything displays all successful parses
func ShowEverything(results []*types.IDInfo) {
	for _, info := range results {
		ShowCard(info)
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
