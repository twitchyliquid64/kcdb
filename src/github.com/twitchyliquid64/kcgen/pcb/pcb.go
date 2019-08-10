// Package pcb parses & serializes the KiCad PCB format.
package pcb

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/nsf/sexp"
)

// Layer describes the attributes of a layer.
type Layer struct {
	Num  int    `json:"num"`
	Name string `json:"name"`
	Type string `json:"type"`

	order int
}

// Net represents a netlist.
type Net struct {
	Name string `json:"name"`

	order int
}

// NetClass represents a net class.
type NetClass struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Clearance    float64 `json:"clearance"`
	TraceWidth   float64 `json:"trace_width"`
	ViaDiameter  float64 `json:"via_dia"`
	ViaDrill     float64 `json:"via_drill"`
	UViaDiameter float64 `json:"uvia_dia"`
	UViaDrill    float64 `json:"uvia_drill"`

	// Nets contains the names of nets which are part of this class.
	Nets []string `json:"connect_pads"`

	order int
}

// PCB represents the parsed contents of a kicad_pcb file.
type PCB struct {
	FormatVersion int `json:"format_version"`
	CreatedBy     struct {
		Tool    string `json:"tool"`
		Version string `json:"version"`
	} `json:"created_by"`

	EditorSetup EditorSetup `json:"editor_setup"`

	LayersByName map[string]*Layer `json:"-"`
	Layers       []*Layer          `json:"layers"`

	Tracks     []Track     `json:"tracks"`
	Vias       []Via       `json:"vias"`
	Lines      []Line      `json:"lines"`
	Texts      []Text      `json:"texts"`
	Dimensions []Dimension `json:"dimensions"`

	Nets       map[int]Net `json:"nets"`
	NetClasses []NetClass  `json:"net_classes"`
	Zones      []Zone      `json:"zones"`
	Modules    []Module    `json:"modules"`

	// TODO(twitchyliquid64): Compute these & expose them.
	generalFields [][]string
}

// EditorSetup describes how the editor should be configured when
// editing this PCB.
type EditorSetup struct {
	LastTraceWidth  float64
	UserTraceWidths []float64
	TraceClearance  float64
	ZoneClearance   float64
	Zone45Only      bool
	TraceMin        float64

	TextWidth    float64
	TextSize     []float64
	SegmentWidth float64
	EdgeWidth    float64

	ViaSize      float64
	ViaMinSize   float64
	ViaDrill     float64
	ViaMinDrill  float64
	UViaSize     float64
	UViaMinSize  float64
	UViaDrill    float64
	UViaMinDrill float64
	AllowUVias   bool

	ModEdgeWidth       float64
	ModTextSize        []float64
	ModTextWidth       float64
	PadSize            []float64
	PadDrill           float64
	PadToMaskClearance float64

	AuxAxisOrigin   []float64
	VisibleElements string

	PlotParams map[string]PlotParam

	Unrecognised map[string]sexp.Helper
	order        int
}

// PlotParam describes a setting for rendering the PCB to another format.
type PlotParam struct {
	name   string
	values []string

	order int
}

// DecodeFile reads a .kicad_pcb file at fpath, returning a parsed representation.
func DecodeFile(fpath string) (*PCB, error) {
	f, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	ast, err := sexp.Parse(strings.NewReader(string(f)), nil)
	if err != nil {
		return nil, err
	}

	if !ast.IsList() {
		return nil, errors.New("invalid format: expected s-expression list at top level")
	}
	if ast.NumChildren() != 1 {
		return nil, errors.New("invalid format: top level list of size 1")
	}
	mainAST, _ := ast.Nth(0)
	if !mainAST.IsList() {
		return nil, errors.New("invalid format: expected s-expression list at 1st level")
	}

	if mainAST.NumChildren() < 5 {
		return nil, errors.New("invalid format: expected at least 5 nodes in main expression")
	}
	if mainAST.Children.Value != "kicad_pcb" {
		return nil, errors.New("invalid format: missing leading element kicad_pcb")
	}

	pcb := &PCB{LayersByName: map[string]*Layer{}, Nets: map[int]Net{}}
	var ordering int

	for i := 1; i < mainAST.NumChildren(); i++ {
		n := sexp.Help(mainAST).Child(i)
		if n.IsList() && n.Child(1).IsValid() {
			switch n.Child(0).MustString() {
			case "version":
				pcb.FormatVersion, err = n.Child(1).Int()
				if err != nil {
					return nil, errors.New("invalid format: version value must be an int")
				}
			case "host":
				pcb.CreatedBy.Tool, err = n.Child(1).String()
				if err != nil {
					return nil, errors.New("invalid format: host value[1] must be a string")
				}
				pcb.CreatedBy.Version, err = n.Child(2).String()
				if err != nil {
					return nil, errors.New("invalid format: host value[2] must be a string")
				}
			case "setup":
				s, err := parseSetup(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.EditorSetup = *s

			case "general":
				for y := 1; y < n.MustNode().NumChildren(); y++ {
					c := n.Child(y)
					var params []string
					for z := 0; z < c.MustNode().NumChildren(); z++ {
						params = append(params, c.Child(z).MustString())
					}
					pcb.generalFields = append(pcb.generalFields, params)
				}

			case "layers":
				for x := 1; x < n.MustNode().NumChildren(); x++ {
					c := n.Child(x)
					num, err2 := c.Child(0).Int()
					if err2 != nil {
						return nil, err
					}
					l := &Layer{
						Num:   num,
						Name:  c.Child(1).MustString(),
						Type:  c.Child(2).MustString(),
						order: ordering,
					}
					pcb.Layers = append(pcb.Layers, l)
					pcb.LayersByName[c.Child(1).MustString()] = l
					ordering++
				}
			case "net":
				num, err2 := n.Child(1).Int()
				if err2 != nil {
					return nil, err
				}
				pcb.Nets[num] = Net{Name: n.Child(2).MustString(), order: ordering}

			case "segment":
				t, err := parseSegment(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Tracks = append(pcb.Tracks, t)

			case "via":
				v, err := parseVia(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Vias = append(pcb.Vias, v)

			case "zone":
				z, err := parseZone(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Zones = append(pcb.Zones, *z)

			case "gr_line":
				l, err := parseGRLine(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Lines = append(pcb.Lines, l)

			case "gr_text":
				t, err := parseGRText(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Texts = append(pcb.Texts, t)

			case "dimension":
				d, err := parseDimension(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Dimensions = append(pcb.Dimensions, d)

			case "net_class":
				c, err := parseNetClass(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.NetClasses = append(pcb.NetClasses, *c)

			case "module":
				m, err := parseModule(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Modules = append(pcb.Modules, *m)
			}
		}
		ordering++
	}

	return pcb, nil
}

func parseNetClass(n sexp.Helper, ordering int) (*NetClass, error) {
	nc := NetClass{
		order:       ordering,
		Name:        n.Child(1).MustString(),
		Description: n.Child(2).MustString(),
	}
	for x := 3; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "clearance":
			nc.Clearance = c.Child(1).MustFloat64()
		case "trace_width":
			nc.TraceWidth = c.Child(1).MustFloat64()
		case "via_dia":
			nc.ViaDiameter = c.Child(1).MustFloat64()
		case "via_drill":
			nc.ViaDrill = c.Child(1).MustFloat64()
		case "uvia_dia":
			nc.UViaDiameter = c.Child(1).MustFloat64()
		case "uvia_drill":
			nc.UViaDrill = c.Child(1).MustFloat64()
		case "add_net":
			nc.Nets = append(nc.Nets, c.Child(1).MustString())
		}
	}
	return &nc, nil
}

func parseSetup(n sexp.Helper, ordering int) (*EditorSetup, error) {
	e := EditorSetup{
		order:        ordering,
		Unrecognised: map[string]sexp.Helper{},
	}
	for x := 1; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "last_trace_width":
			e.LastTraceWidth = c.Child(1).MustFloat64()
		case "user_trace_width":
			e.UserTraceWidths = append(e.UserTraceWidths, c.Child(1).MustFloat64())
		case "trace_clearance":
			e.TraceClearance = c.Child(1).MustFloat64()
		case "zone_clearance":
			e.ZoneClearance = c.Child(1).MustFloat64()
		case "zone_45_only":
			e.Zone45Only = c.Child(1).MustString() == "yes"
		case "trace_min":
			e.TraceMin = c.Child(1).MustFloat64()
		case "segment_width":
			e.SegmentWidth = c.Child(1).MustFloat64()
		case "edge_width":
			e.EdgeWidth = c.Child(1).MustFloat64()

		case "via_size":
			e.ViaSize = c.Child(1).MustFloat64()
		case "via_min_size":
			e.ViaMinSize = c.Child(1).MustFloat64()
		case "via_min_drill":
			e.ViaMinDrill = c.Child(1).MustFloat64()
		case "via_drill":
			e.ViaDrill = c.Child(1).MustFloat64()
		case "uvia_size":
			e.UViaSize = c.Child(1).MustFloat64()
		case "uvia_min_size":
			e.UViaMinSize = c.Child(1).MustFloat64()
		case "uvia_min_drill":
			e.UViaMinDrill = c.Child(1).MustFloat64()
		case "uvia_drill":
			e.UViaDrill = c.Child(1).MustFloat64()
		case "uvias_allowed":
			e.AllowUVias = c.Child(1).MustString() == "yes"

		case "pcb_text_width":
			e.TextWidth = c.Child(1).MustFloat64()
		case "pcb_text_size":
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				e.TextSize = append(e.TextSize, c.Child(y).MustFloat64())
			}

		case "mod_edge_width":
			e.ModEdgeWidth = c.Child(1).MustFloat64()
		case "mod_text_size":
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				e.ModTextSize = append(e.ModTextSize, c.Child(y).MustFloat64())
			}
		case "mod_text_width":
			e.ModTextWidth = c.Child(1).MustFloat64()

		case "pad_size":
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				e.PadSize = append(e.PadSize, c.Child(y).MustFloat64())
			}
		case "pad_drill":
			e.PadDrill = c.Child(1).MustFloat64()
		case "pad_to_mask_clearance":
			e.PadToMaskClearance = c.Child(1).MustFloat64()

		case "aux_axis_origin":
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				e.AuxAxisOrigin = append(e.AuxAxisOrigin, c.Child(y).MustFloat64())
			}
		case "visible_elements":
			e.VisibleElements = c.Child(1).MustString()

		case "pcbplotparams":
			e.PlotParams = map[string]PlotParam{}
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				c := c.Child(y)
				param := PlotParam{
					name:  c.Child(0).MustString(),
					order: y,
				}
				for z := 1; z < c.MustNode().NumChildren(); z++ {
					param.values = append(param.values, c.Child(z).MustString())
				}
				e.PlotParams[param.name] = param
			}

		default:
			e.Unrecognised[c.Child(0).MustString()] = c
		}
	}
	return &e, nil
}
