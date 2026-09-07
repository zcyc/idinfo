package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
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
		fmt.Println("idinfo 0.7.5")
		return
	}

	// Handle ID generation
	if *generate != "" {
		handleGeneration(*generate)
		return
	}
	validateOptions(*forceFormat, *outputFormat)

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please provide an ID to parse\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <ID>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Try '%s --help' for more information.\n", os.Args[0])
		os.Exit(2)
	}
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "error: unexpected argument %q found\n", args[1])
		os.Exit(2)
	}

	compareMode := *compare || *compareShort
	input := args[0]
	if args[0] == "-" && !compareMode {
		// Read from stdin
		var err error
		input, err = readStdin(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please ensure valid input is provided via pipe.\n")
			os.Exit(1)
		}
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
		if *everything {
			fmt.Println("Unknown ID type.")
			return
		}
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
	if info.Timestamp == nil {
		return
	}
	timestamp, err := strconv.ParseFloat(*info.Timestamp, 64)
	if err != nil {
		return
	}
	diff := int64(timestamp) - time.Now().UTC().Unix()
	value := relativeDuration(diff)
	info.Relative = &value
}

func readStdin(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		return scanner.Text(), scanner.Err()
	}
	return "", scanner.Err()
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
	var flags, positional []string
	seenPositional := false
	changed := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			return args
		}
		name := arg
		if equal := strings.IndexByte(name, '='); equal >= 0 {
			name = name[:equal]
		}
		if valueFlags[name] {
			if !seenPositional && strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.ContainsRune(arg, '=') {
				positional = append(positional, arg)
				seenPositional = true
				changed = true
				continue
			}
			flags = append(flags, arg)
			if !strings.ContainsRune(arg, '=') && index+1 < len(args) {
				flags = append(flags, args[index+1])
				index++
			}
			if seenPositional {
				changed = true
			}
			continue
		}
		if seenPositional && !strings.HasPrefix(arg, "--") && len(arg) > 2 && valueFlags[arg[:2]] {
			flags = append(flags, arg[:2], arg[2:])
			changed = true
			continue
		}
		if !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			shortFlags := arg[1:]
			cluster := make([]string, 0, len(shortFlags))
			parsed := true
			for shortIndex := 0; shortIndex < len(shortFlags); shortIndex++ {
				shortName := "-" + shortFlags[shortIndex:shortIndex+1]
				if valueFlags[shortName] {
					parsed = false
					break
				}
				if !booleanFlags[shortName] {
					parsed = false
					break
				}
				cluster = append(cluster, shortName)
			}
			if parsed {
				flags = append(flags, cluster...)
				changed = true
				continue
			}
		}
		if booleanFlags[name] {
			flags = append(flags, arg)
			if seenPositional {
				changed = true
			}
			continue
		}
		if strings.HasPrefix(arg, "-") && seenPositional {
			flags = append(flags, arg)
			changed = true
			continue
		}
		positional = append(positional, arg)
		if !seenPositional {
			seenPositional = true
			if strings.HasPrefix(arg, "-") {
				changed = true
			}
		}
	}
	if !changed {
		return args
	}
	result := make([]string, 0, len(args)+1)
	result = append(result, flags...)
	if len(positional) > 0 && strings.HasPrefix(positional[0], "-") {
		result = append(result, "--")
	}
	return append(result, positional...)
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
                    Available formats: uuid, shortuuid, uuid-int, uuid-b64, uuid25,
                    ulid, sandflake, julid, upid, comb, timeflake, flake,
                    scru128, scru64, mongodb, ksuid, xid, cuid1, cuid2, nanoid,
                    tsid, sqid, hashid, youtube, stripe, datadog, nuid, typeid,
                    breezeid, puid, pushid, tid, threads, duns, asin, snowid,
                    gdocs, slack, spotify, nano64, orderlyid, swhid, iban,
                    commerce, vin, bitcoin, ethereum, sf-twitter, sf-mastodon,
                    sf-discord, sf-instagram, sf-linkedin, sf-sony, sf-spaceflake,
                    sf-frostflake, sf-flakeid, sf-simpleflake, mist, unix, unix-s,
                    unix-ms, unix-us, unix-ns, hash, ipfs, ipv4, ipv6, mac, imei,
                    isbn, h3.
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
    -V, --version   Show version
    -h, --help      Show this help message

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

func relativeDuration(seconds int64) string {
	beforeCurrent := seconds < 0
	var duration uint64
	if beforeCurrent {
		duration = uint64(-(seconds + 1)) + 1
	} else {
		duration = uint64(seconds)
	}

	const (
		minute = 60
		hour   = 60 * minute
		day    = 24 * hour
		month  = 30 * day
	)
	ceil := func(value, unit uint64) uint64 { return (value + unit - 1) / unit }

	var value string
	switch {
	case duration <= 44:
		value = "a few seconds"
	case duration <= 89:
		value = "a minute"
	case duration <= 44*minute:
		value = fmt.Sprintf("%d minutes", ceil(duration, minute))
	case duration <= 89*minute:
		value = "an hour"
	case duration <= 21*hour:
		value = fmt.Sprintf("%d hours", ceil(duration, hour))
	case duration <= 35*hour:
		value = "a day"
	case duration <= 25*day:
		value = fmt.Sprintf("%d days", ceil(duration, day))
	case duration <= 45*day:
		value = "a month"
	case duration <= 10*month:
		value = fmt.Sprintf("%.0f months", float32(duration)/float32(month))
	case duration <= 17*month:
		value = "a year"
	default:
		value = fmt.Sprintf("%.0f years", float32(duration)/float32(12*month))
	}
	if beforeCurrent {
		return value + " ago"
	}
	return "in " + value
}

func validateOptions(force, output string) {
	if force != "" && !parsers.IsCanonicalForceFormat(force) {
		fmt.Fprintf(os.Stderr, "error: invalid value %q for '--force <FORCE>'\n", force)
		fmt.Fprintf(os.Stderr, "  [possible values: %s]\n", strings.Join(parsers.CanonicalForceFormats(), ", "))
		os.Exit(2)
	}
	if output != "card" && output != "short" && output != "json" && output != "binary" {
		fmt.Fprintf(os.Stderr, "error: invalid value %q for '--output <OUTPUT>'\n", output)
		fmt.Fprintln(os.Stderr, "  [possible values: card, short, json, binary]")
		os.Exit(2)
	}
}
