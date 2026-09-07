package parsers

import (
	"testing"

	"github.com/zcyc/idinfo/internal/types"
)

func TestAlignedFormats(t *testing.T) {
	cases := []struct {
		format string
		input  string
		idType string
	}{
		{"uuid", "16689a10-a518-11ef-aa74-4ec6089be97a", "UUID (RFC-4122)"},
		{"uuid-b64", "UHKjBazX_UG8dEAJaikK1g", "Unpadded Base64 of UUID (RFC-4122)"},
		{"uuid25", "dpoadk8izg9y4tte7vy1xt94o", "Uuid25 of UUID (RFC-4122)"},
		{"shortuuid", "32CQvwbvpbnkmkhhguznVH", "ShortUUID of UUID (RFC-4122)"},
		{"shortuuid", "XBCdxzsCR2FEFeSwhnjCo", "ShortUUID of Microsoft GUID"},
		{"uuid-int", "2093703425379131962944436515747969848", "Integer of UUID (RFC-9562)"},
		{"ulid", "01JCXSGZMZQQJ2M93WC0T8KT02", "ULID"},
		{"julid", "01K3ESSGBY0002QCB9YXT6Q6MN", "Julid"},
		{"upid", "abcd_2adnrb7b6jkyos6xusvmaa", "UPID"},
		{"scru128", "03cwivkme1qhj3crprqujv4lu", "SCRU128"},
		{"scru64", "0v20wcjrb21p", "SCRU64"},
		{"xid", "cst4p962941gd9baqg70", "Xid"},
		{"ksuid", "1HCpXwx2EK9oYluWbacgeCnFcLf", "KSUID"},
		{"cuid1", "cm3xemk9o00070cm7ghnl6toe", "CUID"},
		{"cuid2", "byab6ewccgwheoshq1wk9hds", "CUID"},
		{"typeid", "prefix_01h2xcejqtf2nbrexx3vqjhp41", "TypeID"},
		{"pushid", "-OFrJ24CPTXLcIPPjvh3", "PushID (Firebase)"},
		{"nanoid", "XBCdxzsCR2FEFeSwhnjCo", "Nano ID"},
		{"tsid", "0J4AEXRN106Z0", "TSID"},
		{"sqid", "86Rf07", "Sqid"},
		{"youtube", "gocwRvLhDf8", "YouTube Video ID"},
		{"stripe", "cus_lO1DEQWBbQAACfHO", "Stripe ID"},
		{"datadog", "6772800700000000d97a8af26532e259", "Datadog Trace ID"},
		{"breezeid", "9NU6-XQLZ-BDIH-6HKE", "Breeze ID"},
		{"breezeid", "9nu6-xqlz-bdih-6hke", "Breeze ID"},
		{"puid", "he5fps6l2504cd1w3ag8ut8e", "Puid"},
		{"puid", "aeby6ob5sso4zd", "Puid"},
		{"tid", "3lfegaoywdk2w", "TID (AT Protocol, Bluesky)"},
		{"threads", "DEr_fXvuw6D", "Thread ID (Meta Threads)"},
		{"snowid", "HYOYoYloLw", "SnowID"},
		{"duns", "15-048-3782", "DUNS Number"},
		{"asin", "B00DQC2FPM", "ASIN (Amazon)"},
		{"h3", "89283082e73ffff", "H3 Grid System"},
		{"gdocs", "1ZQWherERWu_ZXMGhW0Yw_VxnHFPc3hxLBQ2FjSEalFE", "Google Docs ID"},
		{"slack", "C12345ABCDE", "Slack ID"},
		{"spotify", "4PTG3Z6ehGkBFwjybzWkR8", "Spotify ID"},
		{"nano64", "199C01B6659-5861C", "Nano64"},
		{"orderlyid", "order_00myngy59c0003000dfk59mg3e36j3rr-9xgg", "OrderlyID, type order"},
		{"swhid", "swh:1:dir:65a597ec22d11d3a406784f6f5787a252605561b", "SWHID (Software Hash ID)"},
		{"iban", "NO9386011117947", "IBAN"},
		{"bitcoin", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", "Bitcoin Address (from Satoshi Nakamoto)"},
		{"bitcoin", "bc1p5cyxnuxmeuwuvkwfem96lqzszee2t0p688raqku9m3f8uafvhvqqhkz45z", "Bitcoin Address"},
		{"ethereum", "0xd8da6bf26964af9d7eed9e03e53415d37aa96045", "Ethereum Address"},
		{"ipfs", "QmbWqxBEKC3P8tqsKc98xmWNzrzDtRLMiMPL8wBuTGsMnR", "IPFS"},
		{"ipfs", "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", "IPFS"},
		{"ipv4", "127.0.0.1", "IPv4 Address"},
		{"ipv6", "::1", "IPv6 Address"},
		{"mac", "00:00:00:00:00:00", "MAC Address"},
		{"imei", "355889060149777", "IMEI"},
		{"isbn", "9780553382570", "ISBN-13"},
		{"commerce", "5901234123457", "Commerce Barcode"},
		{"vin", "1HGCM82633A004352", "VIN (Vehicle Identification Number)"},
		{"comb", "019e793c-efb1-4848-9b57-4534777cb348", "COMB"},
		{"comb", "d7a5115f-220b-4928-8c3e-019e793cefb1", "COMB"},
	}

	for _, testCase := range cases {
		t.Run(testCase.format+"/"+testCase.input, func(t *testing.T) {
			results := ParseIDWithOptions(testCase.input, testCase.format, types.ParseOptions{})
			if len(results) != 1 {
				t.Fatalf("expected one result, got %d", len(results))
			}
			if results[0].IDType != testCase.idType {
				t.Fatalf("expected %q, got %q", testCase.idType, results[0].IDType)
			}
		})
	}
}

func TestAlignedMetadataAndOptions(t *testing.T) {
	cases := []struct {
		format, input, version string
	}{
		{"julid", "01K3ESSGBY0002QCB9YXT6Q6MN", "-"},
		{"hashid", "80JTEquWr", "No salt"},
		{"orderlyid", "order_00myngy59c0003000dfk59mg3e36j3rr-9xgg", "Version 1, privacy off, with checksum"},
		{"breezeid", "9nu6-xqlz-bdih-6hke", "Lowercase alphabet"},
		{"ulid", "01933b98-7e9f-bde4-2a24-7c603489e802", "-"},
		{"upid", "01933b9c-a723-e1ea-609d-d6372631d096", "A (default)"},
		{"sandflake", "015c4733-dc14-3e63-2d0a-7fffffd26fb6", "-"},
		{"flake", "00000134-d423-5a10-109a-dd5e0e8f0005", "-"},
		{"timeflake", "016fb420-9023-b444-fd07-590f81b7b0eb", "-"},
		{"scru128", "01936600-0a18-12a4-811f-c295c399b412", "-"},
	}
	for _, testCase := range cases {
		t.Run(testCase.format, func(t *testing.T) {
			results := ParseIDWithOptions(testCase.input, testCase.format, types.ParseOptions{})
			if len(results) != 1 {
				t.Fatalf("expected one result, got %d", len(results))
			}
			version := results[0].Version
			if version == "" {
				version = "-"
			}
			if version != testCase.version {
				t.Fatalf("expected version %q, got %q", testCase.version, version)
			}
		})
	}

	custom := types.ParseOptions{Salt: "this is my salt"}
	results := ParseIDWithOptions("7nnhzEsDkiYa", "hashid", custom)
	if len(results) != 1 || results[0].Version != "Custom salt" || results[0].Node1 == nil {
		t.Fatalf("custom-salt Hashid was not decoded")
	}

	sqid := ParseIDWithOptions("01JCXSGZMZQQJ2M93WC0T8KT02", "sqid", types.ParseOptions{})
	if len(sqid) != 1 || sqid[0].Node1 == nil || *sqid[0].Node1 != "336192318" || sqid[0].HighConfidence {
		t.Fatalf("unexpected checked Sqid decode: %#v", sqid)
	}

	hashid := ParseIDWithOptions("gocwRvLhDf8", "hashid", types.ParseOptions{})
	if len(hashid) != 1 || hashid[0].Node1 == nil || *hashid[0].Node1 != "30, 665257, 29, 31" {
		t.Fatalf("unexpected checked Hashid decode: %#v", hashid)
	}
	if got := ParseIDWithOptions("01JCXSGZMZQQJ2M93WC0T8KT02", "hashid", types.ParseOptions{}); len(got) != 0 {
		t.Fatalf("overflowing Hashid should be rejected: %#v", got)
	}

	flakeID := ParseIDWithOptions("1000000000000000000", "sf-flakeid", types.ParseOptions{})[0]
	if flakeID.Timestamp == nil || *flakeID.Timestamp != "238418579.101" {
		t.Fatalf("unexpected Flake ID timestamp: %#v", flakeID.Timestamp)
	}

	sandflake := ParseIDWithOptions("05E4ECYW2GZ66B8AFZZZZMKFPR", "sandflake", types.ParseOptions{})[0]
	if sandflake.Standard != "05E4ECYW2GZ66B8AFZZZZMKFPR" || sandflake.Timestamp == nil || *sandflake.Timestamp != "1495843200.020" {
		t.Fatalf("unexpected Sandflake metadata: %#v", sandflake)
	}

	zeroResults := ParseAllWithOptions("0", types.ParseOptions{})
	if len(zeroResults) != 2 || zeroResults[0].IDType != "Integer of Nil UUID (all zeros)" {
		t.Fatalf("unexpected all-format result for zero: %#v", zeroResults)
	}

	auto := ParseIDWithOptions("01K3ESSGBY0002QCB9YXT6Q6MN", "", types.ParseOptions{})
	if len(auto) != 1 || auto[0].IDType != "Julid" {
		t.Fatalf("expected auto-detected Julid, got %#v", auto)
	}

	for input, want := range map[string]string{
		"HamVxsto6jDM":           "SCRU64",
		"aeby6ob5sso4":           "SCRU64",
		"EQyuCsA4ysv7ezXReOrk4i": "NUID",
		"JERHwh5PXjL":            "SnowID",
		"DEr_fXvuw6D":            "Thread ID (Meta Threads)",
		"15-048-3782":            "DUNS Number",
	} {
		results := ParseIDWithOptions(input, "", types.ParseOptions{})
		if len(results) != 1 || results[0].IDType != want {
			t.Fatalf("auto-detected %q as %#v, want %q", input, results, want)
		}
	}

	for _, input := range []string{"0-42100-00526-4"} {
		for _, info := range ParseAllWithOptions(input, types.ParseOptions{}) {
			if info.IDType == "MAC Address" {
				t.Fatalf("misclassified barcode as MAC address: %#v", info)
			}
		}
	}

	isbn := ParseIDWithOptions("9780553382570", "isbn", types.ParseOptions{})[0]
	if isbn.Standard != "978-0-553-38257-0" || *isbn.Node2 != "553 (Publisher ID)" || *isbn.Sequence != 38257 {
		t.Fatalf("unexpected ISBN metadata: %#v", isbn)
	}
	for input, want := range map[string]string{
		"9780131103627": "978-0-13-110362-7",
		"9780306406157": "978-0-306-40615-7",
		"9783161484100": "978-3-16-148410-0",
		"9787111213826": "978-7-111-21382-6",
		"9787505715660": "978-7-5057-1566-0",
		"0-9752298-0-X": "0-9752298-0-X",
	} {
		results := ParseIDWithOptions(input, "isbn", types.ParseOptions{})
		if len(results) != 1 || results[0].Standard != want {
			t.Fatalf("unexpected ISBN hyphenation for %q: %#v", input, results)
		}
	}
}

func TestAlignedForceFormats(t *testing.T) {
	cases := []struct {
		format, input, idType, version string
	}{
		{"uuid", "215d3d9f-e980-2cf4-9191-7dd485ba4fee", "UUID (RFC-4122)", "2 (DCE security)"},
		{"uuid", "906b4e7f-84a3-a0ed-1191-2dea8b497113", "NCS UUID", ""},
		{"uuid-int", "2093703425379131962944436515747969848", "Integer of UUID (RFC-9562)", "7 (sortable timestamp and random)"},
		{"ulid", "01933b98-7e9f-bde4-2a24-7c603489e802", "ULID wrapped in UUID", ""},
		{"julid", "01JCXSGZMZQQJ2M93WC0T8KT02", "Julid", ""},
		{"sf-twitter", "1777150623882019211", "Snowflake", "Twitter"},
		{"sf-discord", "1304369705066434662", "Snowflake", "Discord"},
		{"sf-instagram", "1671390786412876801", "Snowflake", "Instagram"},
		{"sf-sony", "540226260526170119", "Snowflake", "Sony"},
		{"sf-spaceflake", "1015189130756840860", "Snowflake", "Spaceflake"},
		{"sf-mastodon", "112277929257317646", "Snowflake", "Mastodon"},
		{"sf-linkedin", "7256902784527069184", "Snowflake", "LinkedIn"},
		{"sf-flakeid", "5828128208445124608", "Snowflake", "Flake ID"},
		{"sf-frostflake", "7423342004626526207", "Snowflake", "Frostflake"},
		{"sf-frostflake", "JERHwh5PXjL", "Snowflake", "Frostflake"},
		{"sf-simpleflake", "3594162604452825250", "Snowflake", "Simpleflake"},
		{"unix-s", "1734971723", "Unix timestamp", "As seconds"},
		{"unix-ms", "1734971723000", "Unix timestamp", "As milliseconds"},
		{"unix-us", "1734971723000000", "Unix timestamp", "As microseconds"},
		{"unix-ns", "1734971723000000000", "Unix timestamp", "As nanoseconds"},
		{"isbn", "0-553-38257-8", "ISBN-10", ""},
		{"hashid", "gocwRvLhDf8", "Hashid", "No salt"},
		{"commerce", "0-42100-00526-4", "Commerce Barcode", "UPC-A (GTIN-12)"},
		{"commerce", "9638-5074", "Commerce Barcode", "EAN-8 (GTIN-8)"},
		{"commerce", "1-06-14141-000415", "Commerce Barcode", "GTIN-14, grouping/packaging level"},
		{"ipfs", "k51qzi5uqu5dlvj2baxnqndepeb86cbk3ng7n3i46uzyxzyqj2xjonzllnv0v8", "IPFS", "CID v1 (IPNS)"},
		{"iban", "NO9386011117947", "IBAN", "NO (Norway)"},
		{"mist", "171671", "Mist", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.format+"/"+testCase.input, func(t *testing.T) {
			results := ParseIDWithOptions(testCase.input, testCase.format, types.ParseOptions{})
			if len(results) != 1 {
				t.Fatalf("expected one result, got %d", len(results))
			}
			if results[0].IDType != testCase.idType || results[0].Version != testCase.version {
				t.Fatalf("got %q / %q, want %q / %q", results[0].IDType, results[0].Version, testCase.idType, testCase.version)
			}
		})
	}
}

func TestGenerationMatchesAlignedParsers(t *testing.T) {
	for _, format := range []string{"scru128", "nanoid"} {
		t.Run(format, func(t *testing.T) {
			parser := NewRegistry().GetParser(format)
			if parser == nil {
				t.Fatalf("missing generator for %s", format)
			}
			generated, err := parser.Generate()
			if err != nil {
				t.Fatalf("failed to generate %s: %v", format, err)
			}
			if results := ParseIDWithOptions(generated, format, types.ParseOptions{}); len(results) != 1 {
				t.Fatalf("generated %s is not accepted by aligned parser: %q", format, generated)
			}
		})
	}
}
