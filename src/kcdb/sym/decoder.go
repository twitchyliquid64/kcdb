package sym

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Symbol represents a schematic symbol.
type Symbol struct {
	Name                 string `json:"name"`
	Reference            string `json:"reference"`
	ReferenceYOffsetMils int

	ShowPins  bool `json:"show_pins"`
	ShowNames bool `json:"show_names"`

	Fields []SymbolFieldLine `json:"fields"`
	Pins   []Pin `json:"pins"`

	RawData string
}

// SymbolFieldLine represents a data field on a symbol.
type SymbolFieldLine struct {
	Kind         int
	Value        string
	X            int
	Y            int
	Size         int
	IsHorizontal bool
	IsHidden     bool
}

// Pin represents a pin draw line.
type Pin struct {
	Name        string `json:"name"`
	Number      string `json:"num"`
	X           int
	Y           int
	Orientation string `json:"orientation"`
}

// DecodeSymbolLibrary decodes an encoded representation of symbols.
func DecodeSymbolLibrary(r io.Reader) ([]*Symbol, error) {
	b := bufio.NewReader(r)
	var header string
	var err error

	for err == nil && strings.TrimSpace(header) == "" {
		header, err = b.ReadString('\n')
	}
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(header, "EESchema-LIBRARY Version 2.") {
		return decodeV2Library(b)
	}

	return nil, nil
}

const (
	parseStateNone = 0
	parseStateDEF  = 1
	parseStateDRAW = 2
)

func decodeV2Library(r *bufio.Reader) ([]*Symbol, error) {
	var parts []*Symbol
	var parseState int
	var err error
	var line string

	for {
		line, err = r.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Replace(line, "\n", "", -1)

		if strings.HasPrefix(line, "DEF ") && parseState == parseStateNone {
			spl, err := spaceSplit(line)
			if err != nil {
				return nil, err
			}
			if len(spl) < 7 {
				return nil, errors.New("missing tokens on DEF line")
			}
			var p Symbol
			p.Name = spl[1]
			p.Reference = spl[2]
			p.ReferenceYOffsetMils, err = strconv.Atoi(spl[4])
			if err != nil {
				return nil, err
			}
			p.ShowPins = spl[5] == "Y"
			p.ShowNames = spl[6] == "Y"
			parts = append(parts, &p)
			parseState = parseStateDEF
			p.RawData = line + "\n"
		} else if strings.HasPrefix(line, "F") && parseState == parseStateDEF {
			spl, err := spaceSplit(line)
			if err != nil {
				return nil, err
			}
			if len(spl) < 9 {
				return nil, errors.New("missing tokens on field line")
			}
			var d SymbolFieldLine
			d.Kind, err = strconv.Atoi(spl[0][1:])
			if err != nil {
				return nil, err
			}
			d.Value = spl[1]
			d.X, err = strconv.Atoi(spl[2])
			if err != nil {
				return nil, err
			}
			d.Y, err = strconv.Atoi(spl[3])
			if err != nil {
				return nil, err
			}
			d.Size, err = strconv.Atoi(spl[4])
			if err != nil {
				return nil, err
			}
			d.IsHorizontal = spl[5] == "H"
			d.IsHidden = spl[6] == "I"
			parts[len(parts)-1].Fields = append(parts[len(parts)-1].Fields, d)
			parts[len(parts)-1].RawData += line + "\n"
		} else if strings.HasPrefix(line, "DRAW") && parseState == parseStateDEF {
			parseState = parseStateDRAW
			parts[len(parts)-1].RawData += line + "\n"
		} else if strings.HasPrefix(line, "X ") && parseState == parseStateDRAW {
			spl, err := spaceSplit(line)
			if err != nil {
				return nil, err
			}
			if len(spl) < 10 {
				return nil, errors.New("missing tokens on pin line")
			}
			var p Pin
			p.Name = spl[1]
			p.Number = spl[2]
			p.X, err = strconv.Atoi(spl[3])
			if err != nil {
				return nil, err
			}
			p.Y, err = strconv.Atoi(spl[4])
			if err != nil {
				return nil, err
			}
			p.Orientation = spl[6]
			parts[len(parts)-1].Pins = append(parts[len(parts)-1].Pins, p)
			parts[len(parts)-1].RawData += line + "\n"
		} else if strings.HasPrefix(line, "ENDDRAW") && parseState == parseStateDRAW {
			parseState = parseStateDEF
			parts[len(parts)-1].RawData += line + "\n"
		} else if strings.HasPrefix(line, "ENDDEF") && parseState == parseStateDEF {
			parseState = parseStateNone
			parts[len(parts)-1].RawData += line
		}
	}
	if err == io.EOF {
		return parts, nil
	}
	return nil, err
}

func spaceSplit(line string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ' ' // space
	return r.Read()
}
