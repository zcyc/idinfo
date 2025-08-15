package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zcyc/idinfo/internal/output"
	"github.com/zcyc/idinfo/internal/parsers"
)

func main() {
	var (
		forceFormat  = flag.String("f", "", "Force parsing as specific format")
		outputFormat = flag.String("o", "card", "Output format (card, short, json, binary)")
		everything   = flag.Bool("e", false, "Show all possible format interpretations")
		compare      = flag.Bool("compare", false, "Compare timestamps from different formats")
		generate     = flag.String("g", "", "Generate ID of specified format")
		colorOutput  = flag.Bool("color", true, "Enable colored output")
		help         = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Handle ID generation
	if *generate != "" {
		handleGeneration(*generate)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please provide an ID to parse\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <ID>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Try '%s --help' for more information.\n", os.Args[0])
		os.Exit(1)
	}

	var input string
	if args[0] == "-" {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please ensure valid input is provided via pipe.\n")
			os.Exit(1)
		}
	} else {
		input = args[0]
	}

	if input == "" {
		fmt.Fprintf(os.Stderr, "Error: Empty input provided\n")
		fmt.Fprintf(os.Stderr, "Please provide a valid ID to parse.\n")
		os.Exit(1)
	}

	// Parse the ID
	results := parsers.ParseID(input, *forceFormat)

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Unable to parse ID '%s'\n", input)
		if *forceFormat != "" {
			fmt.Fprintf(os.Stderr, "The ID cannot be parsed as format '%s'.\n", *forceFormat)
			fmt.Fprintf(os.Stderr, "Try without the -f flag for auto-detection.\n")
		} else {
			fmt.Fprintf(os.Stderr, "The ID format is not recognized or supported.\n")
			fmt.Fprintf(os.Stderr, "Supported formats: UUID, ULID, ObjectId, KSUID, Xid, CUID, SCRU128, TSID, NUID, NanoID, Snowflake, UnixTime, HashHex, Base58, PushID, Base32, ShortUUID, Sqids, TypeID\n")
			fmt.Fprintf(os.Stderr, "Try using -f to force a specific format.\n")
		}
		os.Exit(1)
	}

	// Handle different output modes
	if *everything {
		output.ShowEverything(results)
		return
	}

	if *compare {
		output.ShowComparison(results)
		return
	}

	// Show the best match (first result)
	result := results[0]

	switch *outputFormat {
	case "card":
		if *colorOutput {
			output.ShowCardColored(result)
		} else {
			output.ShowCard(result)
		}
	case "short":
		output.ShowShort(result)
	case "json":
		jsonOutput, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
			fmt.Fprintf(os.Stderr, "This is likely due to invalid data in the parsed result.\n")
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	case "binary":
		output.ShowBinary(result)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown output format '%s'\n", *outputFormat)
		fmt.Fprintf(os.Stderr, "Supported formats: card, short, json, binary\n")
		os.Exit(1)
	}
}

func handleGeneration(format string) {
	registry := parsers.NewRegistry()
	parser := registry.GetParser(format)
	if parser == nil {
		fmt.Fprintf(os.Stderr, "Error: Unsupported format '%s'\n", format)
		fmt.Fprintf(os.Stderr, "Supported formats: ")
		parserNames := registry.GetAvailableParsers()
		for i, name := range parserNames {
			if i > 0 {
				fmt.Fprintf(os.Stderr, ", ")
			}
			fmt.Fprintf(os.Stderr, "%s", name)
		}
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	id, err := parser.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", format, err)
		os.Exit(1)
	}

	fmt.Println(id)
}

func showHelp() {
	fmt.Print(`idinfo: ID Information Tool

USAGE:
    idinfo [OPTIONS] <ID>
    idinfo [OPTIONS] -
    idinfo -g <FORMAT>

OPTIONS:
    -f <FORMAT>     Force parsing as specific format
                    Available formats: uuid, ulid, objectid, ksuid, xid, cuid,
                    scru128, tsid, nuid, nanoid, snowflake, base58, pushid,
                    base32, shortuuid,
                    sqids, typeid, etc.
    -o <OUTPUT>     Output format (card, short, json, binary) [default: card]
    -e              Show all possible format interpretations
    -g <FORMAT>     Generate new ID of specified format
    --color         Enable colored output [default: true]
    --compare       Compare timestamps from different format interpretations
    --help          Show this help message

EXAMPLES:
    Parse ID:
      idinfo 01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa
      idinfo -f uuid 01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa
      idinfo -o json 01HVZ7JKJJ8M9K9M9M9M9M9M9M
      echo "01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa" | idinfo -

    Generate ID:
      idinfo -g uuid
      idinfo -g ulid
      idinfo -g objectid

SUPPORTED ID FORMATS:
    - UUID (v1-v8), ULID, MongoDB ObjectId
    - KSUID, Xid, CUID2, SCRU128, TSID, NUID
    - Snowflake variants (Twitter, Discord, etc.)
    - NanoID, Firebase PushID
    - Base58 (Bitcoin-style), Base32, Unix timestamps
    - Hex-encoded hashes (MD5, SHA-1, SHA-256, etc.)
    - ShortUUID
    - Sqids, TypeID (typed identifiers)
`)
}
