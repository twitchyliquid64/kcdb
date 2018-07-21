package sym

import (
	"bytes"
	"testing"

	"github.com/andreyvit/diff"
)

func TestDecoder2(t *testing.T) {
	f := bytes.NewBufferString(`EESchema-LIBRARY Version 2.3  Date: Mon 27 Oct 2014 05:36:45 PM CET
#encoding utf-8
#
# MSP430G2553-20
#
DEF MSP430G2553-20 U 0 40 Y Y 1 F N
F0 "U" 0 600 60 H V C CNN
F1 "MSP430G2553-20" 0 -600 60 H V C CNN
F2 "~" 0 0 60 H V C CNN
F3 "~" 0 0 60 H V C CNN
$FPLIST
 tssop-20
 DIP-20_300
$ENDFPLIST
DRAW
S -400 550 400 -550 0 1 0 N
X DVCC 1 -500 450 100 R 50 50 1 1 I
X P1.0 2 -500 350 100 R 50 50 1 1 B
X P1.1 3 -500 250 100 R 50 50 1 1 B
X P1.2 4 -500 150 100 R 50 50 1 1 B
X P1.3 5 -500 50 100 R 50 50 1 1 B
X P1.4 6 -500 -50 100 R 50 50 1 1 B
X P1.5 7 -500 -150 100 R 50 50 1 1 B
X P2.0 8 -500 -250 100 R 50 50 1 1 W
X P2.1 9 -500 -350 100 R 50 50 1 1 I
X P2.2 10 -500 -450 100 R 50 50 1 1 O
X DVSS 20 500 450 100 L 50 50 1 1 B
X P2.3 11 500 -450 100 L 50 50 1 1 W
X P2.4 12 500 -350 100 L 50 50 1 1 W
X P2.5 13 500 -250 100 L 50 50 1 1 B
X P1.6 14 500 -150 100 L 50 50 1 1 B
X P1.7 15 500 -50 100 L 50 50 1 1 B
X ~RST~ 16 500 50 100 L 50 50 1 1 B
X TEST 17 500 150 100 L 50 50 1 1 B
X P2.7 18 500 250 100 L 50 50 1 1 B
X P2.6 19 500 350 100 L 50 50 1 1 B
ENDDRAW
ENDDEF`)

	parts, err := DecodeSymbolLibrary(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(parts) != 1 {
		t.Fatalf("Got %d parts, expected 1", len(parts))
	}

	if parts[0].Name != "MSP430G2553-20" {
		t.Error("Name incorrect")
	}
	if parts[0].Reference != "U" {
		t.Error("Reference incorrect")
	}
	if parts[0].ReferenceYOffsetMils != 40 {
		t.Error("ReferenceYOffsetMils incorrect")
	}
	if !parts[0].ShowNames {
		t.Error("ShowNames incorrect")
	}
	if !parts[0].ShowPins {
		t.Error("ShowPins incorrect")
	}

	if len(parts[0].Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(parts[0].Fields))
	}
	if parts[0].Fields[1].Kind != 1 {
		t.Errorf("Expected field kind to be 1, got %d", parts[0].Fields[1].Kind)
	}
	if parts[0].Fields[1].Value != "MSP430G2553-20" {
		t.Errorf("Expected field kind to be WS2812B, got %s", parts[0].Fields[1].Value)
	}

	expectedRawData := `DEF MSP430G2553-20 U 0 40 Y Y 1 F N
F0 "U" 0 600 60 H V C CNN
F1 "MSP430G2553-20" 0 -600 60 H V C CNN
F2 "~" 0 0 60 H V C CNN
F3 "~" 0 0 60 H V C CNN
DRAW
S -400 550 400 -550 0 1 0 N
X DVCC 1 -500 450 100 R 50 50 1 1 I
X P1.0 2 -500 350 100 R 50 50 1 1 B
X P1.1 3 -500 250 100 R 50 50 1 1 B
X P1.2 4 -500 150 100 R 50 50 1 1 B
X P1.3 5 -500 50 100 R 50 50 1 1 B
X P1.4 6 -500 -50 100 R 50 50 1 1 B
X P1.5 7 -500 -150 100 R 50 50 1 1 B
X P2.0 8 -500 -250 100 R 50 50 1 1 W
X P2.1 9 -500 -350 100 R 50 50 1 1 I
X P2.2 10 -500 -450 100 R 50 50 1 1 O
X DVSS 20 500 450 100 L 50 50 1 1 B
X P2.3 11 500 -450 100 L 50 50 1 1 W
X P2.4 12 500 -350 100 L 50 50 1 1 W
X P2.5 13 500 -250 100 L 50 50 1 1 B
X P1.6 14 500 -150 100 L 50 50 1 1 B
X P1.7 15 500 -50 100 L 50 50 1 1 B
X ~RST~ 16 500 50 100 L 50 50 1 1 B
X TEST 17 500 150 100 L 50 50 1 1 B
X P2.7 18 500 250 100 L 50 50 1 1 B
X P2.6 19 500 350 100 L 50 50 1 1 B
ENDDRAW
ENDDEF`

	if parts[0].RawData != expectedRawData {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(parts[0].RawData, expectedRawData))
	}
}

func TestDecoder(t *testing.T) {
	f := bytes.NewBufferString(`EESchema-LIBRARY Version 2.3  Date: Mon 27 Oct 2014 05:36:45 PM CET
#encoding utf-8
#
# WS2812B
#
DEF WS2812B U 0 40 Y Y 1 F N
F0 "U" 0 100 60 H V C CNN
F1 "WS2812B" 0 0 60 H V C CNN
F2 "~" 0 0 60 H V C CNN
F3 "~" 0 0 60 H V C CNN
DRAW
S -250 -100 250 -300 0 1 0 N
S 100 -350 100 -350 0 1 0 N
X VDD 1 -550 -150 300 R 50 50 1 1 W
X DOUT 2 550 -250 300 L 50 50 1 1 O
X GND 3 550 -150 300 L 50 50 1 1 W
X DIN 4 -550 -250 300 R 50 50 1 1 I
ENDDRAW
ENDDEF
DEF kek U 0 40 Y Y 1 F N
ENDDEF
#
#End Library`)

	parts, err := DecodeSymbolLibrary(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(parts) != 2 {
		t.Fatalf("Got %d parts, expected 2", len(parts))
	}

	if parts[0].Name != "WS2812B" {
		t.Error("Name incorrect")
	}
	if parts[0].Reference != "U" {
		t.Error("Reference incorrect")
	}
	if parts[0].ReferenceYOffsetMils != 40 {
		t.Error("ReferenceYOffsetMils incorrect")
	}
	if !parts[0].ShowNames {
		t.Error("ShowNames incorrect")
	}
	if !parts[0].ShowPins {
		t.Error("ShowPins incorrect")
	}

	if len(parts[0].Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(parts[0].Fields))
	}
	if parts[0].Fields[1].Kind != 1 {
		t.Errorf("Expected field kind to be 1, got %d", parts[0].Fields[1].Kind)
	}
	if parts[0].Fields[1].Value != "WS2812B" {
		t.Errorf("Expected field kind to be WS2812B, got %s", parts[0].Fields[1].Value)
	}

	expectedRawData := `DEF WS2812B U 0 40 Y Y 1 F N
F0 "U" 0 100 60 H V C CNN
F1 "WS2812B" 0 0 60 H V C CNN
F2 "~" 0 0 60 H V C CNN
F3 "~" 0 0 60 H V C CNN
DRAW
S -250 -100 250 -300 0 1 0 N
S 100 -350 100 -350 0 1 0 N
X VDD 1 -550 -150 300 R 50 50 1 1 W
X DOUT 2 550 -250 300 L 50 50 1 1 O
X GND 3 550 -150 300 L 50 50 1 1 W
X DIN 4 -550 -250 300 R 50 50 1 1 I
ENDDRAW
ENDDEF`

	if parts[0].RawData != expectedRawData {
		t.Errorf("Expected RawData=%q, got %q.", expectedRawData, parts[0].RawData)
	}
}
