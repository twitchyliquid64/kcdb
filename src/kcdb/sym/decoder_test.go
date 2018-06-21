package sym

import (
	"bytes"
	"testing"
)

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
#
#End Library`)

	parts, err := DecodeSymbolLibrary(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(parts) != 1 {
		t.Fatalf("Got %d parts, expected 1", len(parts))
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
}
