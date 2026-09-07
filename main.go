package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zcyc/idinfo/internal/output"
	"github.com/zcyc/idinfo/internal/parsers"
	"github.com/zcyc/idinfo/internal/types"
)

func main() {
	var (
		forceFormat  = flag.String("f", "", "Force parsing as specific format")
		outputFormat = flag.String("o", "card", "Output format (card, short, json, binary)")
		everything   = flag.Bool("e", false, "Show all possible format interpretations")
		compare      = flag.Bool("compare", false, "Compare timestamps from different formats")
		compareShort = flag.Bool("c", false, "Compare timestamps from different formats")
		generate     = flag.String("g", "", "Generate ID of specified format")
		alphabet     = flag.String("a", "", "Custom alphabet for Sqids and Nano ID")
		relative     = flag.Bool("r", false, "Show relative time if available")
		salt         = flag.String("salt", "", "Custom salt for Hashids")
		epoch        = flag.Uint64("epoch", 0, "Override epoch (seconds since 1970-01-01 UTC)")
		version      = flag.Bool("version", false, "Show version")
		help         = flag.Bool("help", false, "Show help")
	)
	flag.CommandLine.StringVar(forceFormat, "force", *forceFormat, "Force parsing as specific format")
	flag.CommandLine.StringVar(outputFormat, "output", *outputFormat, "Output format (card, short, json, binary)")
	flag.CommandLine.BoolVar(everything, "everything", *everything, "Show all possible format interpretations")
	flag.CommandLine.StringVar(alphabet, "alphabet", *alphabet, "Custom alphabet for Sqids and Nano ID")
	flag.CommandLine.BoolVar(relative, "relative", *relative, "Show relative time if available")
	flag.CommandLine.StringVar(generate, "generate", *generate, "Generate ID of specified format")
	flag.CommandLine.BoolVar(help, "h", *help, "Show help")
	flag.CommandLine.BoolVar(version, "V", *version, "Show version")
	flag.CommandLine.Parse(normalizeIDArgs(os.Args[1:]))

	if *help {
		showHelp()
		return
	}
	if *version {
		fmt.Println("idinfo 0.1.0")
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

	compareMode := *compare || *compareShort
	var input string
	if args[0] == "-" && !compareMode {
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

	options := types.ParseOptions{Alphabet: *alphabet, Salt: *salt}
	flag.Visit(func(value *flag.Flag) {
		if value.Name == "epoch" {
			options.Epoch, options.HasEpoch = *epoch, true
		}
	})
	var results []*types.IDInfo
	if *everything || compareMode {
		if *everything {
			results = parsers.ParseAllWithOptions(input, options)
		} else {
			results = parsers.ParseTimesWithOptions(input, options)
		}
	} else {
		results = parsers.ParseIDWithOptions(input, *forceFormat, options)
	}

	if compareMode {
		output.ShowComparison(results)
		return
	}
	if len(results) == 0 {
		if *forceFormat != "" {
			fmt.Println("Invalid ID for this format.")
		} else {
			fmt.Println("Unknown ID type.")
		}
		os.Exit(1)
	}

	if *relative {
		for _, info := range results {
			setRelativeTime(info)
		}
	}

	// Handle different output modes
	if *everything {
		output.ShowEverything(results)
		return
	}

	// Show the best match (first result)
	result := results[0]

	switch *outputFormat {
	case "card":
		output.ShowCardColored(result)
	case "short":
		output.ShowShort(result)
	case "json":
		jsonResult := *result
		if !*relative {
			setRelativeTime(&jsonResult)
		}
		jsonOutput, err := json.Marshal(jsonResult)
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

func setRelativeTime(info *types.IDInfo) {
	if info.DateTime == nil {
		return
	}
	diff := time.Since(*info.DateTime)
	value := "in " + relativeDuration(-diff)
	if diff >= 0 {
		value = relativeDuration(diff) + " ago"
	}
	info.Relative = &value
}

func normalizeIDArgs(args []string) []string {
	booleanFlags := map[string]bool{
		"-e": true, "--everything": true, "-c": true, "--compare": true,
		"-r": true, "--relative": true, "--version": true, "--help": true, "-h": true, "-V": true,
	}
	valueFlags := map[string]bool{
		"-f": true, "--force": true, "-o": true, "--output": true, "-g": true, "--generate": true,
		"-a": true, "--alphabet": true, "--salt": true, "--epoch": true,
	}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" || !strings.HasPrefix(arg, "-") {
			if !strings.HasPrefix(arg, "-") {
				continue
			}
			return args
		}
		name := arg
		if equal := strings.IndexByte(name, '='); equal >= 0 {
			name = name[:equal]
		}
		if valueFlags[name] {
			if !strings.ContainsRune(arg, '=') {
				index++
			}
			continue
		}
		if !strings.HasPrefix(arg, "--") && len(arg) > 2 && valueFlags[arg[:2]] {
			result := make([]string, 0, len(args)+1)
			result = append(result, args[:index]...)
			result = append(result, arg[:2], arg[2:])
			result = append(result, args[index+1:]...)
			return result
		}
		if booleanFlags[name] {
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			result := make([]string, 0, len(args)+1)
			result = append(result, args[:index]...)
			result = append(result, "--", arg)
			result = append(result, args[index+1:]...)
			return result
		}
	}
	return args
}

func handleGeneration(format string) {
	// Check if this is a UUID with version specification (e.g., "uuid:v1")
	if strings.HasPrefix(strings.ToLower(format), "uuid:") {
		parts := strings.SplitN(format, ":", 2)
		if len(parts) == 2 {
			version := strings.ToLower(parts[1])
			id, err := generateUUIDWithVersion(version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating UUID %s: %v\n", version, err)
				os.Exit(1)
			}
			fmt.Println(id)
			return
		}
	}

	// Use existing parser for other formats or plain "uuid"
	registry := parsers.NewRegistry()
	var id string
	var err error
	found := false
	for _, parser := range registry.GetAllParsers() {
		if !strings.EqualFold(parser.Name(), format) {
			continue
		}
		found = true
		id, err = parser.Generate()
		if err == nil {
			fmt.Println(id)
			return
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Error: Unsupported format '%s'\n", format)
		fmt.Fprintf(os.Stderr, "Supported formats: ")
		parserNames := registry.GetAvailableParsers()
		for i, name := range parserNames {
			if i > 0 {
				fmt.Fprintf(os.Stderr, ", ")
			}
			fmt.Fprintf(os.Stderr, "%s", name)
		}
		fmt.Fprintf(os.Stderr, "\nFor UUID, you can also specify version: uuid:v1, uuid:v3, uuid:v4, uuid:v5, uuid:v6, uuid:v7\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", format, err)
	os.Exit(1)
}

func generateUUIDWithVersion(version string) (string, error) {
	switch version {
	case "v1":
		// UUID v1: timestamp and MAC address
		u, err := uuid.NewUUID()
		if err != nil {
			return "", fmt.Errorf("failed to generate UUID v1: %w", err)
		}
		return u.String(), nil

	case "v3":
		// UUID v3: namespace name based with MD5
		// Use a default namespace (DNS) and a default name for command-line usage
		u := uuid.NewMD5(uuid.NameSpaceDNS, []byte("idinfo-generated"))
		return u.String(), nil

	case "v4":
		// UUID v4: random (default)
		u := uuid.New()
		return u.String(), nil

	case "v5":
		// UUID v5: namespace name based with SHA-1
		// Use a default namespace (DNS) and a default name for command-line usage
		u := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("idinfo-generated"))
		return u.String(), nil

	case "v6":
		// UUID v6: reordered timestamp and MAC address
		u, err := uuid.NewV6()
		if err != nil {
			return "", fmt.Errorf("failed to generate UUID v6: %w", err)
		}
		return u.String(), nil

	case "v7":
		// UUID v7: sortable timestamp and random
		u, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("failed to generate UUID v7: %w", err)
		}
		return u.String(), nil

	default:
		return "", fmt.Errorf("unsupported UUID version '%s'. Supported versions: v1, v3, v4, v5, v6, v7", version)
	}
}

func showHelp() {
	fmt.Print(`idinfo: ID Information Tool

USAGE:
    idinfo [OPTIONS] <ID>
    idinfo [OPTIONS] -
    idinfo -g <FORMAT>

OPTIONS:
    -f, --force <FORMAT>
                    Force parsing as specific format
                    Available formats: uuid, uuid-b64, uuid25, shortuuid, uuid-int,
                    ulid, julid, upid, sandflake, timeflake, flake, mongodb,
                    ksuid, xid, scru128, scru64, tsid, nuid, typeid, pushid,
                    orderlyid, threads, snowid, nano64, sqid, hashid, youtube,
                    stripe, datadog, breezeid, puid, tid, duns, asin, gdocs,
                    slack, spotify, swhid, iban, commerce, vin, bitcoin,
                    ethereum, ipfs, ipv4, ipv6, mac, imei, isbn, h3, mist, comb,
                    snowflake variants, unix units, hashes, base58, base32.
    -o, --output <OUTPUT>
                    Output format (card, short, json, binary) [default: card]
    -e, --everything
                    Show all possible format interpretations
    -g, --generate <FORMAT>
                    Generate new ID of specified format
                    For UUID, you can specify version: uuid:v1, uuid:v3, uuid:v4, 
                    uuid:v5, uuid:v6, uuid:v7 (default is v4)
    -c, --compare   Compare timestamps from different format interpretations
    -a, --alphabet <ALPHABET>
                    Custom alphabet for Sqids and Nano ID
    -r, --relative  Show relative time if timestamp is available
    --salt <SALT>   Custom salt for Hashids
    --epoch <SEC>   Override epoch (seconds since 1970-01-01 UTC)
    --version       Show version
    --help          Show this help message

EXAMPLES:
    Parse ID:
      idinfo 01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa
      idinfo -f uuid 01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa
      idinfo -o json 01HVZ7JKJJ8M9K9M9M9M9M9M9M
      echo "01941f29-7c00-7aaa-aaaa-aaaaaaaaaaaa" | idinfo -

    Generate ID:
      idinfo -g uuid         # Generate UUID v4 (random)
      idinfo -g uuid:v1      # Generate UUID v1 (timestamp + MAC)
      idinfo -g uuid:v3      # Generate UUID v3 (namespace + name MD5)
      idinfo -g uuid:v4      # Generate UUID v4 (random)
      idinfo -g uuid:v5      # Generate UUID v5 (namespace + name SHA-1)
      idinfo -g uuid:v6      # Generate UUID v6 (reordered timestamp + MAC)
      idinfo -g uuid:v7      # Generate UUID v7 (sortable timestamp + random)
      idinfo -g ulid
      idinfo -g mongodb

SUPPORTED ID FORMATS:
    - UUID (v1-v8), ShortUUID, UUID Base64, UUID25, UUID integer
    - ULID, Julid, UPID, Sandflake, Timeflake, Flake, SCRU128/SCRU64
    - MongoDB ObjectId, KSUID, Xid, TSID, NUID, TypeID, CUID, NanoID
    - Snowflake variants, SnowID, Threads, TID, Firebase PushID, Nano64
    - Sqid/Hashid, YouTube, Stripe, Datadog, BreezeID, PUID, OrderlyID
    - DUNS, ASIN, Google Docs, Slack, Spotify, SWHID, IBAN, barcodes, VIN
    - Bitcoin, Ethereum, IPFS, IPv4/IPv6, MAC, IMEI, ISBN, H3, hashes, COMB
`)
}

func relativeDuration(duration time.Duration) string {
	switch {
	case duration < time.Minute:
		return "less than a minute"
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes", int(duration/time.Minute))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours", int(duration/time.Hour))
	case duration < 365*24*time.Hour:
		return fmt.Sprintf("%d days", int(duration/(24*time.Hour)))
	default:
		return fmt.Sprintf("%d years", int(duration/(365*24*time.Hour)))
	}
}
