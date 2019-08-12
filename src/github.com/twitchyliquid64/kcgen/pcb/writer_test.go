package pcb

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"

	diff "github.com/sergi/go-diff/diffmatchpatch"
)

func TestPCBWrite(t *testing.T) {
	tcs := []struct {
		name     string
		pcb      PCB
		expected string
	}{
		{
			name: "simple",
			pcb: PCB{
				FormatVersion: 4,
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n \n)\n",
		},
		{
			name: "layers",
			pcb: PCB{
				FormatVersion: 4,
				Layers: []*Layer{
					{Name: "F.Cu", Type: "signal"},
					{Num: 31, Name: "B.Cu", Type: "signal"},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers\n    (0 F.Cu signal)\n    (31 B.Cu signal)\n  )\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n \n)\n",
		},
		{
			name: "nets",
			pcb: PCB{
				FormatVersion: 4,
				Nets: map[int]Net{
					0: {Name: ""},
					1: {Name: "+5C"},
					2: {Name: "GND"},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (net 0 \"\")\n  (net 1 +5C)\n  (net 2 GND)\n\n \n)\n",
		},
		{
			name: "net classes",
			pcb: PCB{
				FormatVersion: 4,
				NetClasses: []NetClass{
					{Name: "Default", Description: "This is the default net class.",
						Clearance: 0.2, TraceWidth: 0.25, Nets: []string{"+5C", "GND"}},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (net_class Default \"This is the default net class.\"\n    (clearance 0.2)\n    (trace_width 0.25)\n    (add_net +5C)\n    (add_net GND)\n  )\n\n \n)\n",
		},
		{
			name: "plot params",
			pcb: PCB{
				EditorSetup: EditorSetup{
					PadDrill: 0.762,
					PlotParams: map[string]PlotParam{
						"usegerberextensions": PlotParam{name: "usegerberextensions", values: []string{"true"}, order: 11},
						"scaleselection":      PlotParam{name: "scaleselection", values: []string{"1"}, order: 10},
						"layerselection":      PlotParam{name: "layerselection", values: []string{"0x010f0_80000001"}},
					},
				},
			},
			expected: "(kicad_pcb (version 0) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n    (pad_drill 0.762)\n    (pcbplotparams\n      (layerselection 0x010f0_80000001)\n      (scaleselection 1)\n      (usegerberextensions true))\n  )\n\n \n)\n",
		},
		{
			name: "vias",
			pcb: PCB{
				FormatVersion: 4,
				Segments: []NetSegment{
					&Via{At: XY{X: 100, Y: 32.5}, Layers: []string{"F.Cu", "B.Cu"}, NetIndex: 2},
					&Via{At: XY{X: 10, Y: 32.5}, Layers: []string{"F.Cu", "B.Cu"}, NetIndex: 2},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (via (at 100 32.5) (size 0) (drill 0) (layers F.Cu B.Cu) (net 2))\n  (via (at 10 32.5) (size 0) (drill 0) (layers F.Cu B.Cu) (net 2))\n)\n",
		},
		{
			name: "tracks",
			pcb: PCB{
				FormatVersion: 4,
				Segments: []NetSegment{
					&Track{Start: XY{X: 100, Y: 32.5}, End: XY{X: 10, Y: 32.5}, Layer: "F.Cu", NetIndex: 2},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (segment (start 100 32.5) (end 10 32.5) (width 0) (layer F.Cu) (net 2))\n)\n",
		},
		{
			name: "lines",
			pcb: PCB{
				FormatVersion: 4,
				Drawings: []Drawing{
					&Line{Start: XY{X: 100, Y: 32.5}, End: XY{X: 10, Y: 32.5}, Layer: "Edge.Cuts", Width: 2},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (gr_line (start 100 32.5) (end 10 32.5) (layer Edge.Cuts) (width 2))\n)\n",
		},
		{
			name: "text",
			pcb: PCB{
				FormatVersion: 4,
				Drawings: []Drawing{
					&Text{At: XYZ{X: 100, Y: 32.5}, Text: "Oops", Layer: "F.SilkS", Effects: TextEffects{
						FontSize:  XY{X: 1.5, Y: 1.5},
						Thickness: 0.3,
					}},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (gr_text Oops (at 100 32.5) (layer F.SilkS)\n    (effects (font (size 1.5 1.5) (thickness 0.3)))\n  )\n)\n",
		},
		{
			name: "zones",
			pcb: PCB{
				FormatVersion: 4,
				Zones: []Zone{
					{NetNum: 42, Tstamp: "0", Layers: []string{"F.Cu"}, NetName: "DBUS", MinThickness: 0.254,
						BasePolys: [][]XY{
							[]XY{{X: 11, Y: 22}, {X: 11.1, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}},
						},
						Polys: [][]XY{
							[]XY{{X: 11, Y: 22}, {X: 11.1, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}, {X: 11, Y: 22}},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (zone (net 42) (net_name DBUS) (layer F.Cu) (tstamp 0) (hatch \"\" 0)\n    (connect_pads (clearance 0))\n    (min_thickness 0.254)\n    (fill (arc_segments 0) (thermal_gap 0) (thermal_bridge_width 0))\n    (polygon\n      (pts\n        (xy 11 22) (xy 11.1 22) (xy 11 22) (xy 11 22) (xy 11 22)\n        (xy 11 22) (xy 11 22)\n      )\n    )\n    (filled_polygon\n      (pts\n        (xy 11 22) (xy 11.1 22) (xy 11 22) (xy 11 22) (xy 11 22)\n        (xy 11 22) (xy 11 22)\n      )\n    )\n  )\n)\n",
		},
		{
			name: "dimensions",
			pcb: PCB{
				FormatVersion: 4,
				Drawings: []Drawing{
					&Dimension{
						CurrentMeasurement: 12.446,
						Width:              0.3,
						Layer:              "F.Fab",
						Text: Text{
							At:    XYZ{X: 125.396, Y: 93.853, Z: 90, ZPresent: true},
							Text:  "12.446 mm",
							Layer: "F.Fab", Effects: TextEffects{
								FontSize:  XY{X: 1.5, Y: 1.5},
								Thickness: 0.3,
							}},
						Features: []DimensionFeature{
							{
								Feature: "feature1",
								Points:  []XY{{X: 173.736, Y: 100.076}, {X: 173.736, Y: 106.586}},
							},
							{
								Feature: "feature2",
								Points:  []XY{{X: 132.08, Y: 100.076}, {X: 132.08, Y: 106.586}},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (dimension 12.446 (width 0.3) (layer F.Fab)\n    (gr_text \"12.446 mm\" (at 125.396 93.853 90) (layer F.Fab)\n      (effects (font (size 1.5 1.5) (thickness 0.3)))\n    )\n    (feature1 (pts (xy 173.736 100.076) (xy 173.736 106.586)))\n    (feature2 (pts (xy 132.08 100.076) (xy 132.08 106.586)))\n  )\n)\n",
		},
		{
			name: "mod simple",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Placement: ModPlacement{
							At: XYZ{X: 159.850666, Y: 90},
						},
						Description: "Through hole straight pin header, 1x04, 2.54mm pitch, single row",
						Tags:        []string{"Through", "hole", "pin", "header", "THT", "1x04", "2.54mm", "single", "row"},
						Path:        "/5ADA7034",
						Attrs:       []string{"smd"},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 159.850666 90)\n    (descr \"Through hole straight pin header, 1x04, 2.54mm pitch, single row\")\n    (tags \"Through hole pin header THT 1x04 2.54mm single row\")\n    (path /5ADA7034)\n    (attr smd)\n  )\n\n \n)\n",
		},
		{
			name: "mod model",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Model: &ModModel{
							Path:   "Resistors_SMD.3dshapes/R_0805_HandSoldering.wrl",
							At:     XYZ{ZPresent: true},
							Scale:  XYZ{X: 1, Y: 1, Z: 1, ZPresent: true},
							Rotate: XYZ{ZPresent: true},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (model Resistors_SMD.3dshapes/R_0805_HandSoldering.wrl\n      (at (xyz 0 0 0))\n      (scale (xyz 1 1 1))\n      (rotate (xyz 0 0 0))\n    )\n  )\n\n \n)\n",
		},
		{
			name: "mod text",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Graphics: []ModGraphic{
							{
								Ident: "fp_text",
								Renderable: &ModText{
									Kind:  RefText,
									Text:  "R9",
									At:    XYZ{X: -1, Y: 0.625},
									Layer: "F.Fab",
									Effects: TextEffects{
										FontSize:  XY{X: 1, Y: 1},
										Thickness: 0.15,
									},
								},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (fp_text reference R9 (at -1 0.625) (layer F.Fab)\n      (effects (font (size 1 1) (thickness 0.15)))\n    )\n  )\n\n \n)\n",
		},
		{
			name: "mod line",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Graphics: []ModGraphic{
							{
								Ident: "fp_line",
								Renderable: &ModLine{
									Start: XY{X: -1, Y: 0.625},
									End:   XY{X: -1, Y: -0.625},
									Layer: "F.Fab",
									Width: 0.1,
								},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (fp_line (start -1 0.625) (end -1 -0.625) (layer F.Fab) (width 0.1))\n  )\n\n \n)\n",
		},
		{
			name: "mod circle",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Graphics: []ModGraphic{
							{
								Ident: "fp_circle",
								Renderable: &ModCircle{
									Center: XY{X: -1, Y: 0.625},
									End:    XY{X: -1, Y: -0.625},
									Layer:  "F.Fab",
									Width:  0.1,
								},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (fp_circle (center -1 0.625) (end -1 -0.625) (layer F.Fab) (width 0.1))\n  )\n\n \n)\n",
		},
		{
			name: "mod arc",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Graphics: []ModGraphic{
							{
								Ident: "fp_arc",
								Renderable: &ModArc{
									Start: XY{X: -1, Y: 0.625},
									End:   XY{X: -1, Y: -0.625},
									Angle: 90,
									Layer: "F.Fab",
									Width: 0.1,
								},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (fp_arc (start -1 0.625) (end -1 -0.625) (angle 90) (layer F.Fab) (width 0.1))\n  )\n\n \n)\n",
		},
		{
			name: "mod polygon",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Graphics: []ModGraphic{
							{
								Ident: "fp_poly",
								Renderable: &ModPolygon{
									Points: []XY{
										{},
										{X: 1},
										{X: 1, Y: 1},
										{Y: 1},
										{},
									},
									Layer: "F.Fab",
									Width: 0.1,
								},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (fp_poly (pts (xy 0 0) (xy 1 0) (xy 1 1) (xy 0 1)\n      (xy 0 0)) (layer F.Fab) (width 0.1))\n  )\n\n \n)\n",
		},
		{
			name: "mod pad",
			pcb: PCB{
				FormatVersion: 4,
				Modules: []Module{
					{
						Name:   "Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm",
						Layer:  "F.Cu",
						Tedit:  "5ADA75A0",
						Tstamp: "5AE3D8AB",
						Pads: []Pad{
							{
								Ident:     "1",
								NetNum:    1,
								NetName:   "GND",
								Layers:    []string{"*.Cu", "*.Mask"},
								Surface:   SurfaceTH,
								Shape:     ShapeRect,
								DrillSize: XY{X: 1, Y: 1},
								Size:      XY{X: 1.7, Y: 1.7},
							},
						},
					},
				},
			},
			expected: "(kicad_pcb (version 4) (host kcgen 0.0.1)\n\n  (general)\n\n  (page A4)\n  (layers)\n\n  (setup\n    (zone_45_only no)\n    (uvias_allowed no)\n  )\n\n  (module Pin_Headers:Pin_Header_Straight_1x04_Pitch2.54mm (layer F.Cu) (tedit 5ADA75A0) (tstamp 5AE3D8AB)\n    (at 0 0)\n    (pad 1 thru_hole rect (at 0 0) (size 1.7 1.7) (drill 1) (layers *.Cu *.Mask)\n      (net 1 GND))\n  )\n\n \n)\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			if err := tc.pcb.Write(&b); err != nil {
				t.Fatalf("pcb.Write() failed: %v", err)
			}
			if tc.expected != b.String() {
				t.Error("output mismatch")
				t.Logf("want = %q", tc.expected)
				t.Logf("got  = %q", b.String())
			}
			// ioutil.WriteFile("test.kicad_pcb", b.Bytes(), 0755)
		})
	}
}

func TestDecodeThenSerializeMatches(t *testing.T) {
	tcs := []struct {
		name  string
		fname string
	}{
		{
			name:  "simple",
			fname: "simple_equality.kicad_pcb",
		},
		{
			name:  "zone",
			fname: "zone_equality.kicad_pcb",
		},
		{
			name:  "dimension",
			fname: "dimension_equality.kicad_pcb",
		},
		{
			name:  "t1",
			fname: "t1.kicad_pcb",
		},
		{
			name:  "fume extractor",
			fname: "anavi-fume-extractor.kicad_pcb",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			pcb, err := DecodeFile(path.Join("testdata", tc.fname))
			if err != nil {
				t.Fatalf("DecodeFile(%q) failed: %v", tc.fname, err)
			}
			var serialized bytes.Buffer
			if err := pcb.Write(&serialized); err != nil {
				t.Fatalf("Write() failed: %v", err)
			}

			d, err := ioutil.ReadFile(path.Join("testdata", tc.fname))
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(d, serialized.Bytes()) {
				t.Error("outputs differ")
				diffs := diff.New()
				dm := diffs.DiffMain(string(d), serialized.String(), false)
				// t.Log(diffs.DiffPrettyText(dm))
				// t.Log(diffs.DiffToDelta(dm))
				t.Log(diffs.PatchToText(diffs.PatchMake(dm)))
				// ioutil.WriteFile("test.kicad_pcb", serialized.Bytes(), 0755)
			}
		})
	}
}
