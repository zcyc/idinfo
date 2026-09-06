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

	auto := ParseIDWithOptions("01K3ESSGBY0002QCB9YXT6Q6MN", "", types.ParseOptions{})
	if len(auto) != 1 || auto[0].IDType != "Julid" {
		t.Fatalf("expected auto-detected Julid, got %#v", auto)
	}
}
