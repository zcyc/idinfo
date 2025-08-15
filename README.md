# idinfo: ID Information Tool

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#)

A Go-based command-line tool for debugging and analyzing unique identifiers. This tool can parse, analyze, and generate various types of unique identifiers with detailed information extraction.

## Quick Start

```bash
# Install
go install github.com/zcyc/idinfo@latest

# Generate a UUID
idinfo -g uuid

# Analyze the generated UUID
idinfo $(idinfo -g uuid)

# Try different formats
idinfo -g snowflake
idinfo -g nanoid
idinfo -g ulid
```

## Features

- **Multiple ID Format Support**: UUID (v1-v8), ULID, MongoDB ObjectId, KSUID, Xid, NanoID, NUID, Snowflake variants, Unix timestamps, hex-encoded hashes, Base58, Firebase PushID, Base32, ShortUUID, Sqids, and TypeID
- **Auto-Detection**: Automatically detects ID format using heuristics
- **Force Format**: Override auto-detection to parse as specific format
- **Multiple Output Formats**: Card-style (default), short, JSON, and binary output
- **Comprehensive Analysis**: Extracts timestamps, entropy, node information, sequences, and format-specific details
- **Pipeline Support**: Read from stdin for integration with other tools
- **Comparison Mode**: Compare timestamps from different format interpretations
- **Everything Mode**: Show all possible format interpretations

## Installation

### Using go install

```bash
go install github.com/zcyc/idinfo@latest
```

### Build from Source

```bash
git clone <repository-url>
cd idinfo
go build -o idinfo .
```

## Usage

### Basic Usage

```bash
# Analyze a UUID (using installed binary)
idinfo 550e8400-e29b-41d4-a716-446655440000

# Analyze a MongoDB ObjectId
idinfo 507f1f77bcf86cd799439011

# Analyze a ULID
idinfo 01ARZ3NDEKTSV4RRFFQ69G5FAV

# If you built from source, use ./idinfo instead:
./idinfo 550e8400-e29b-41d4-a716-446655440000
```

### Force Specific Format

```bash
# Parse a number as Twitter Snowflake
idinfo -f snowflake 1777150623882019211

# Parse as UUID even if it could be something else
idinfo -f uuid 550e8400-e29b-41d4-a716-446655440000
```

### Different Output Formats

```bash
# Short format
idinfo -o short 507f1f77bcf86cd799439011

# JSON format
idinfo -o json 550e8400-e29b-41d4-a716-446655440000

# Binary output (useful for further processing)
idinfo -o binary 550e8400-e29b-41d4-a716-446655440000 | xxd
```

### Advanced Features

```bash
# Show all possible format interpretations
idinfo -e 1777150623882019211

# Compare timestamps from different interpretations
idinfo --compare 1777150623882019211

# Generate new IDs
idinfo -g uuid
idinfo -g snowflake
idinfo -g nanoid

# Pipeline usage
echo "550e8400-e29b-41d4-a716-446655440000" | idinfo -
```

## Supported ID Formats

### Core Formats
- **UUID (RFC-9562)**: All versions (1-8), including Nil and Max UUIDs
- **ULID**: Universally Unique Lexicographically Sortable Identifier
- **MongoDB ObjectId**: 96-bit ObjectId with timestamp, machine, process, and counter
- **KSUID**: K-Sortable Unique Identifier with timestamp and payload
- **Xid**: Globally unique sortable id with timestamp, machine, process, and counter

### Additional Formats
- **NanoID**: URL-safe unique ID generator
- **NUID**: NATS Unique Identifier - high-performance 22-character base62 IDs
- **Snowflake Variants**: Twitter, Discord, Instagram formats
- **Unix Timestamps**: Seconds, milliseconds, microseconds, nanoseconds
- **Hex-encoded Hashes**: MD5, SHA-1, SHA-256, SHA-384, SHA-512

### Extended Formats
- **Base58**: Bitcoin-style base58 encoded IDs (no confusing characters)
- **Firebase PushID**: Firebase real-time database push IDs
- **Base32**: RFC 4648 base32 encoded identifiers
- **Hashids**: Reversible obfuscated numeric IDs

### Platform & Service IDs
- **ShortUUID**: 22-character UUID representations

- **Sqids**: Modern Hashids successor with anti-profanity
- **TypeID**: Type-prefixed ULID format (type_id)

## Output Examples

### Card Format (Default)
```
┏━━━━━━━━━━━┯━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ID Type   │ UUID (RFC-9562)                             ┃
┃ Version   │ 4 (random)                                  ┃
┠───────────┼─────────────────────────────────────────────┨
┃ String    │ 550e8400-e29b-41d4-a716-446655440000        ┃
┃ Integer   │ 113059749145936325402354257176981405696     ┃
┃ Base64    │ VQ6EAOKbQdSnFkRmVUQAAA==                    ┃
┠───────────┼─────────────────────────────────────────────┨
┃ Size      │ 128 bits                                    ┃
┃ Entropy   │ 122 bits                                    ┃
┃ Node 1    │ -                                           ┃
┃ Node 2    │ -                                           ┃
┃ Sequence  │ -                                           ┃
┠───────────┼─────────────────────────────────────────────┨
┃ 550e 8400 │ 0101 0101 0000 1110 1000 0100 0000 0000     ┃
┃ e29b 41d4 │ 1110 0010 1001 1011 0100 0001 1101 0100     ┃
┃ a716 4466 │ 1010 0111 0001 0110 0100 0100 0110 0110     ┃
┃ 5544 0000 │ 0101 0101 0100 0100 0000 0000 0000 0000     ┃
┗━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### JSON Format
```json
{
  "id_type": "UUID (RFC-9562)",
  "version": "4 (random)",
  "standard": "550e8400-e29b-41d4-a716-446655440000",
  "integer": "113059749145936325402354257176981405696",
  "base64": "VQ6EAOKbQdSnFkRmVUQAAA==",
  "size": 128,
  "entropy": 122,
  "hex": "550e8400e29b41d4a716446655440000",
  "extra": {
    "variant": "RFC 4122"
  }
}
```

### Short Format
```
ID Type: UUID (RFC-9562), version: 4 (random).
```

## Command Line Options

### Parsing Options
- `-f <FORMAT>`: Force parsing as specific format
- `-o <OUTPUT>`: Output format (card, short, json, binary)
- `-e`: Show all possible format interpretations
- `--compare`: Compare timestamps from different format interpretations
- `--color`: Enable colored output (default: true)

### Generation Options
- `-g <FORMAT>`: Generate new ID of specified format

### General Options
- `--help`: Show help message

### Available Force Formats
- `uuid`, `guid`
- `ulid`
- `objectid`, `mongodb`, `bson`
- `ksuid`
- `xid`
- `nanoid`, `nano-id`
- `snowflake`, `sf`, `sf-twitter`, `sf-discord`, `twitter`, `discord`
- `unixtime`, `unix`, `timestamp`
- `hashhex`, `hash`, `hex`

## Architecture

The tool follows a modular architecture:

```
├── main.go                 # CLI entry point
├── internal/
│   ├── types/             # Common types and interfaces
│   ├── parsers/           # ID parsers for each format
│   │   ├── registry.go    # Parser registration and management
│   │   ├── uuid.go        # UUID parser
│   │   ├── ulid.go        # ULID parser
│   │   └── ...            # Other format parsers
│   └── output/            # Output formatters
│       └── output.go      # Card, short, JSON, binary formatters
```

Each parser implements the `IDParser` interface:
```go
type IDParser interface {
    Name() string
    CanParse(input string) bool
    Parse(input string) (*IDInfo, error)
    Generate() (string, error)  // New: ID generation capability
}
```




## License

This project is licensed under the MIT License.

## Acknowledgments

- Inspired by [uuinfo](https://github.com/Racum/uuinfo)
- Thanks to the Go community for the excellent ID generation libraries