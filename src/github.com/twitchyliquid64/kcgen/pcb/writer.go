package pcb

import (
	"fmt"
	"io"
	"sort"

	"github.com/twitchyliquid64/kcgen/swriter"
)

// Write produces the file on disk. IMPLEMENTATION IS NOT YET COMPLETE.
func (p *PCB) Write(w io.Writer) error {
	sw, err := swriter.NewSExpWriter(w)
	if err != nil {
		return err
	}
	sw.StartList(false)
	sw.StringScalar("kicad_pcb")

	// Version
	sw.StartList(false)
	sw.StringScalar("version")
	sw.IntScalar(p.FormatVersion)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	// EG: host pcbnew 4.0.7
	sw.StartList(false)
	sw.StringScalar("host")
	if p.CreatedBy.Tool == "" {
		sw.StringScalar("kcgen")
	} else {
		sw.StringScalar(p.CreatedBy.Tool)
	}
	if p.CreatedBy.Version == "" {
		sw.StringScalar("0.0.1")
	} else {
		sw.StringScalar(p.CreatedBy.Version)
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}
	sw.Newlines(2)

	// EG: general (no_connects 0) ...
	sw.StartList(false)
	sw.StringScalar("general")
	if len(p.generalFields) > 0 {
		for _, section := range p.generalFields {
			sw.StartList(true)
			for _, v := range section {
				sw.StringScalar(v)
			}
			if err := sw.CloseList(false); err != nil {
				return err
			}
		}
		if err := sw.CloseList(true); err != nil {
			return err
		}
	} else {
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	sw.Separator()

	// EG: page A4
	sw.StartList(false)
	sw.StringScalar("page")
	sw.StringScalar("A4")
	if err := sw.CloseList(false); err != nil {
		return err
	}
	sw.Newlines(1)

	// Layers
	sw.StartList(false)
	sw.StringScalar("layers")
	if len(p.Layers) > 0 {
		sw.Newlines(1)
		for i, layer := range p.Layers {
			if err := layer.write(sw); err != nil {
				return err
			}
			if i < len(p.Layers)-1 {
				sw.Newlines(1)
			}
		}
		if err := sw.CloseList(len(p.Layers) > 0); err != nil {
			return err
		}
	} else {
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	sw.Separator()

	// Setup
	if err := p.EditorSetup.write(sw); err != nil {
		return err
	}

	// Nets
	if err := p.writeNets(sw); err != nil {
		return err
	}

	// Net classes
	for _, nc := range p.NetClasses {
		if err := nc.write(sw); err != nil {
			return err
		}
	}
	if len(p.NetClasses) > 0 {
		sw.Separator()
	}

	// Modules
	for i, m := range p.Modules {
		if err := m.write(sw, true); err != nil {
			return err
		}
		if i < len(p.Modules)-1 {
			sw.Separator()
		}
	}
	if len(p.Modules) > 0 {
		sw.Separator()
	}

	// Dimensions
	for _, d := range p.Dimensions {
		if err := d.write(sw); err != nil {
			return err
		}
		sw.Newlines(1)
	}
	// (Graphical) Text
	for _, t := range p.Texts {
		if err := t.write(sw); err != nil {
			return err
		}
	}

	// (Graphical) Lines
	if len(p.Lines) > 0 {
		sw.Newlines(1)
	}
	for i, l := range p.Lines {
		if err := l.write(sw); err != nil {
			return err
		}
		if i < len(p.Lines)-1 {
			sw.Newlines(1)
		}
	}
	if len(p.Lines) > 0 || len(p.Vias) > 0 {
		sw.Separator()
	}

	// Vias
	for i, v := range p.Vias {
		if err := v.write(sw); err != nil {
			return err
		}
		if i < len(p.Vias)-1 {
			sw.Newlines(1)
		}
	}
	// Tracks
	if len(p.Tracks) > 0 {
		sw.Newlines(1)
	}
	for i, t := range p.Tracks {
		if err := t.write(sw); err != nil {
			return err
		}
		if i < len(p.Tracks)-1 {
			sw.Newlines(1)
		}
	}
	// Zones
	if len(p.Zones) > 0 {
		sw.Separator()
	}
	for i, z := range p.Zones {
		if err := z.write(sw); err != nil {
			return err
		}
		if i < len(p.Zones)-1 {
			sw.Newlines(1)
		}
	}

	if err := sw.CloseList(true); err != nil {
		return err
	}
	w.Write([]byte("\n"))
	return nil
}

type netPair struct {
	num int
	net Net
}

func (p *PCB) writeNets(sw *swriter.SExpWriter) error {
	var nets []netPair
	for num, net := range p.Nets {
		nets = append(nets, netPair{num: num, net: net})
	}
	sort.Slice(nets, func(i, j int) bool {
		return nets[i].num < nets[j].num
	})

	for i, n := range nets {
		sw.StartList(false)
		sw.StringScalar("net")
		sw.IntScalar(n.num)
		sw.StringScalar(n.net.Name)
		if err := sw.CloseList(false); err != nil {
			return err
		}
		if i < len(nets)-1 {
			sw.Newlines(1)
		}
	}

	if len(nets) > 0 {
		sw.Separator()
	}
	return nil
}

// write generates an s-expression describing the layer.
func (l *Layer) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.IntScalar(l.Num)
	sw.StringScalar(l.Name)
	sw.StringScalar(l.Type)
	return sw.CloseList(false)
}

func f(f float64) string {
	t := fmt.Sprintf("%f", f)
	if t[len(t)-1] != '0' {
		return t
	}

	for i := len(t) - 1; i >= 0; i-- {
		if t[i] != '0' {
			if t[i] == '.' {
				return t[:i]
			}
			return t[:i+1]
		}
	}
	return t
}

// write generates an s-expression describing the point.
func (p *XY) write(prefix string, sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar(prefix)
	sw.StringScalar(f(p.X))
	sw.StringScalar(f(p.Y))
	return sw.CloseList(false)
}

// write generates an s-expression describing the point.
func (p *XYZ) write(prefix string, sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar(prefix)
	sw.StringScalar(f(p.X))
	sw.StringScalar(f(p.Y))
	if p.ZPresent {
		sw.StringScalar(f(p.Z))
	}
	return sw.CloseList(false)
}

// write generates an s-expression describing the line.
func (l *Line) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar("gr_line")
	if err := l.Start.write("start", sw); err != nil {
		return err
	}
	if err := l.End.write("end", sw); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(l.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("width")
	sw.StringScalar(f(l.Width))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	return sw.CloseList(false)
}

// write generates an s-expression describing the text.
func (t *Text) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar("gr_text")
	sw.StringScalar(t.Text)
	if err := t.At.write("at", sw); err != nil {
		return err
	}

	if t.Hidden {
		sw.StringScalar("hide")
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(t.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(true)
	sw.StringScalar("effects")
	sw.StartList(false)
	sw.StringScalar("font")
	if err := t.Effects.FontSize.write("size", sw); err != nil {
		return err
	}
	sw.StartList(false)
	sw.StringScalar("thickness")
	sw.StringScalar(f(t.Effects.Thickness))
	if err := sw.CloseList(false); err != nil {
		return err
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}
	if t.Effects.Justify != JustifyNone {
		sw.StartList(false)
		sw.StringScalar("justify")
		sw.StringScalar(t.Effects.Justify.String())
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}

	if err := sw.CloseList(true); err != nil {
		return err
	}
	return nil
}

// write generates an s-expression describing the dimension.
func (d *Dimension) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar("dimension")
	sw.StringScalar(f(d.CurrentMeasurement))

	sw.StartList(false)
	sw.StringScalar("width")
	sw.StringScalar(f(d.Width))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(d.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}
	sw.Newlines(1)

	if err := d.Text.write(sw); err != nil {
		return err
	}
	sw.Newlines(1)

	for i, f := range d.Features {
		sw.StartList(false)
		sw.StringScalar(f.Feature)

		sw.StartList(false)
		sw.StringScalar("pts")
		for _, pt := range f.Points {
			if err := pt.write("xy", sw); err != nil {
				return err
			}
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
		if i < len(d.Features)-1 {
			sw.Newlines(1)
		}
	}

	if err := sw.CloseList(true); err != nil {
		return err
	}
	return nil
}

// write generates an s-expression describing the layer.
func (l *EditorSetup) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar("setup")

	if l.LastTraceWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("last_trace_width")
		sw.StringScalar(f(l.LastTraceWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	for _, w := range l.UserTraceWidths {
		sw.StartList(true)
		sw.StringScalar("user_trace_width")
		sw.StringScalar(f(w))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.TraceClearance > 0 {
		sw.StartList(true)
		sw.StringScalar("trace_clearance")
		sw.StringScalar(f(l.TraceClearance))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.ZoneClearance > 0 {
		sw.StartList(true)
		sw.StringScalar("zone_clearance")
		sw.StringScalar(f(l.ZoneClearance))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	sw.StartList(true)
	sw.StringScalar("zone_45_only")
	if l.Zone45Only {
		sw.StringScalar("yes")
	} else {
		sw.StringScalar("no")
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}
	if l.TraceMin > 0 {
		sw.StartList(true)
		sw.StringScalar("trace_min")
		sw.StringScalar(f(l.TraceMin))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.SegmentWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("segment_width")
		sw.StringScalar(f(l.SegmentWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.EdgeWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("edge_width")
		sw.StringScalar(f(l.EdgeWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if l.ViaSize > 0 {
		sw.StartList(true)
		sw.StringScalar("via_size")
		sw.StringScalar(f(l.ViaSize))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.ViaDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("via_drill")
		sw.StringScalar(f(l.ViaDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.ViaMinSize > 0 {
		sw.StartList(true)
		sw.StringScalar("via_min_size")
		sw.StringScalar(f(l.ViaMinSize))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.ViaMinDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("via_min_drill")
		sw.StringScalar(f(l.ViaMinDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.UViaSize > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_size")
		sw.StringScalar(f(l.UViaSize))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.UViaDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_drill")
		sw.StringScalar(f(l.UViaDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	sw.StartList(true)
	sw.StringScalar("uvias_allowed")
	if l.AllowUVias {
		sw.StringScalar("yes")
	} else {
		sw.StringScalar("no")
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}
	if l.UViaMinSize > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_min_size")
		sw.StringScalar(f(l.UViaMinSize))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.UViaMinDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_min_drill")
		sw.StringScalar(f(l.UViaMinDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if l.TextWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("pcb_text_width")
		sw.StringScalar(f(l.TextWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if len(l.TextSize) > 0 {
		sw.StartList(true)
		sw.StringScalar("pcb_text_size")
		for _, w := range l.TextSize {
			sw.StringScalar(f(w))
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if l.ModEdgeWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("mod_edge_width")
		sw.StringScalar(f(l.ModEdgeWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if len(l.ModTextSize) > 0 {
		sw.StartList(true)
		sw.StringScalar("mod_text_size")
		for _, w := range l.ModTextSize {
			sw.StringScalar(f(w))
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.ModTextWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("mod_text_width")
		sw.StringScalar(f(l.ModTextWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if len(l.PadSize) > 0 {
		sw.StartList(true)
		sw.StringScalar("pad_size")
		for _, w := range l.PadSize {
			sw.StringScalar(f(w))
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.PadDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("pad_drill")
		sw.StringScalar(f(l.PadDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.PadToMaskClearance > 0 {
		sw.StartList(true)
		sw.StringScalar("pad_to_mask_clearance")
		sw.StringScalar(f(l.PadToMaskClearance))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if len(l.AuxAxisOrigin) > 0 {
		sw.StartList(true)
		sw.StringScalar("aux_axis_origin")
		for _, i := range l.AuxAxisOrigin {
			sw.StringScalar(f(i))
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if l.VisibleElements != "" {
		sw.StartList(true)
		sw.StringScalar("visible_elements")
		sw.StringScalar(l.VisibleElements)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if len(l.PlotParams) > 0 {
		sw.StartList(true)
		sw.StringScalar("pcbplotparams")
		var pps []PlotParam
		for _, pp := range l.PlotParams {
			pps = append(pps, pp)
		}
		sort.Slice(pps, func(i, j int) bool {
			return pps[i].order < pps[j].order
		})

		for _, pp := range pps {
			sw.StartList(true)
			sw.StringScalar(pp.name)
			for _, v := range pp.values {
				sw.StringScalar(v)
			}
			if err := sw.CloseList(false); err != nil {
				return err
			}
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if err := sw.CloseList(true); err != nil {
		return err
	}
	sw.Separator()
	return nil
}

// write generates an s-expression describing the layer.
func (c *NetClass) write(sw *swriter.SExpWriter) error {
	sw.StartList(false)
	sw.StringScalar("net_class")
	sw.StringScalar(c.Name)
	sw.StringScalar(c.Description)

	if c.Clearance > 0 {
		sw.StartList(true)
		sw.StringScalar("clearance")
		sw.StringScalar(f(c.Clearance))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if c.TraceWidth > 0 {
		sw.StartList(true)
		sw.StringScalar("trace_width")
		sw.StringScalar(f(c.TraceWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if c.ViaDiameter > 0 {
		sw.StartList(true)
		sw.StringScalar("via_dia")
		sw.StringScalar(f(c.ViaDiameter))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if c.ViaDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("via_drill")
		sw.StringScalar(f(c.ViaDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if c.UViaDiameter > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_dia")
		sw.StringScalar(f(c.UViaDiameter))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if c.UViaDrill > 0 {
		sw.StartList(true)
		sw.StringScalar("uvia_drill")
		sw.StringScalar(f(c.UViaDrill))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	for _, net := range c.Nets {
		sw.StartList(true)
		sw.StringScalar("add_net")
		sw.StringScalar(net)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if err := sw.CloseList(true); err != nil {
		return err
	}
	return nil
}
