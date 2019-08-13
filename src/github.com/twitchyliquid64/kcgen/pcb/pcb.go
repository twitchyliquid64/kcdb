// Package pcb parses & serializes the KiCad PCB format.
package pcb

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/nsf/sexp"
	"github.com/twitchyliquid64/kcgen/swriter"
)

// Layer describes the attributes of a layer.
type Layer struct {
	Num    int    `json:"num"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Hidden bool   `json:'hidden'`

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

	DiffPairWidth float64 `json:"diff_pair_width"`
	DiffPairGap   float64 `json:"diff_pair_gap"`

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

	TitleInfo   *TitleInfo  `json:"title_info"`
	EditorSetup EditorSetup `json:"editor_setup"`

	LayersByName map[string]*Layer `json:"-"`
	Layers       []*Layer          `json:"layers"`

	Segments []NetSegment `json:"segments"`
	Drawings []Drawing    `json:"drawings"`

	Nets       map[int]Net `json:"nets"`
	NetClasses []NetClass  `json:"net_classes"`
	Zones      []Zone      `json:"zones"`
	Modules    []Module    `json:"modules"`

	// TODO(twitchyliquid64): Compute these & expose them.
	generalFields [][]string
}

// Drawing represents a drawable element.
type Drawing interface {
	write(sw *swriter.SExpWriter) error
}

// NetSegment represents copper regions which form part of a net.
type NetSegment interface {
	write(sw *swriter.SExpWriter) error
}

// TitleInfo describes information about the document.
type TitleInfo struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	Revision string `json:"revision"`
	Company  string `json:"company"`

	Comments [4]string `json:"comments"`

	order int
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

	UserVia                [2]float64
	BlindBuriedViasAllowed bool
	GridOrigin             [2]int

	ModEdgeWidth       float64
	ModTextSize        []float64
	ModTextWidth       float64
	PadSize            []float64
	PadDrill           float64
	PadToMaskClearance float64
	SolderMaskMinWidth float64

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

// ZoneConnectMode describes how the zone should connect.
type ZoneConnectMode int8

// Valid ZoneConnectMode values.
const (
	ZoneConnectInherited ZoneConnectMode = iota - 1
	ZoneConnectNone
	ZoneConnectThermal
)

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

			case "title_block":
				t, err := parseTitleBlock(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.TitleInfo = t

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

					if c.MustNode().NumChildren() > 3 && c.Child(3).IsScalar() &&
						c.Child(3).MustString() == "hide" {
						l.Hidden = true
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
				pcb.Segments = append(pcb.Segments, &t)

			case "via":
				v, err := parseVia(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Segments = append(pcb.Segments, &v)

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
				pcb.Drawings = append(pcb.Drawings, &l)

			case "gr_text":
				t, err := parseGRText(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Drawings = append(pcb.Drawings, &t)

			case "gr_arc":
				a, err := parseGRArc(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Drawings = append(pcb.Drawings, &a)

			case "dimension":
				d, err := parseDimension(n, ordering)
				if err != nil {
					return nil, err
				}
				pcb.Drawings = append(pcb.Drawings, &d)

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
		case "diff_pair_width":
			nc.DiffPairWidth = c.Child(1).MustFloat64()
		case "diff_pair_gap":
			nc.DiffPairGap = c.Child(1).MustFloat64()
		}
	}
	return &nc, nil
}

func parseTitleBlock(n sexp.Helper, ordering int) (*TitleInfo, error) {
	t := TitleInfo{order: ordering}
	for x := 1; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "title":
			t.Title = c.Child(1).MustString()
		case "date":
			t.Date = c.Child(1).MustString()
		case "rev":
			t.Revision = c.Child(1).MustString()
		case "company":
			t.Company = c.Child(1).MustString()
		case "comment":
			idx := c.Child(1).MustInt() - 1
			if idx < len(t.Comments) {
				t.Comments[idx] = c.Child(2).MustString()
			}
		}
	}
	return &t, nil
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
		case "solder_mask_min_width":
			e.SolderMaskMinWidth = c.Child(1).MustFloat64()

		case "aux_axis_origin":
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				e.AuxAxisOrigin = append(e.AuxAxisOrigin, c.Child(y).MustFloat64())
			}
		case "visible_elements":
			e.VisibleElements = c.Child(1).MustString()

		case "user_via":
			e.UserVia = [2]float64{c.Child(1).MustFloat64(), c.Child(2).MustFloat64()}
		case "blind_buried_vias_allowed":
			e.BlindBuriedViasAllowed = c.Child(1).MustString() == "yes"
		case "grid_origin":
			e.GridOrigin = [2]int{c.Child(1).MustInt(), c.Child(2).MustInt()}

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
