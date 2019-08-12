package pcb

import (
	"strings"

	"github.com/twitchyliquid64/kcgen/swriter"
)

func (m *Module) write(sw *swriter.SExpWriter, doPlacement bool) error {
	sw.StartList(false)
	sw.StringScalar("module")
	sw.StringScalar(m.Name)

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(m.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	if m.Tedit != "" {
		sw.StartList(false)
		sw.StringScalar("tedit")
		sw.StringScalar(m.Tedit)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if m.Tstamp != "" {
		sw.StartList(false)
		sw.StringScalar("tstamp")
		sw.StringScalar(m.Tstamp)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	sw.Newlines(1)

	if doPlacement {
		if err := m.Placement.At.write("at", sw); err != nil {
			return err
		}
	}

	if m.Description != "" {
		sw.StartList(true)
		sw.StringScalar("descr")
		sw.StringScalar(m.Description)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if len(m.Tags) > 0 {
		sw.StartList(true)
		sw.StringScalar("tags")
		sw.StringScalar(strings.Join(m.Tags, " "))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if m.Path != "" {
		sw.StartList(true)
		sw.StringScalar("path")
		sw.StringScalar(m.Path)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if len(m.Attrs) > 0 {
		sw.StartList(true)
		sw.StringScalar("attr")
		for _, a := range m.Attrs {
			sw.StringScalar(a)
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	for _, g := range m.Graphics {
		if err := g.Renderable.write(sw); err != nil {
			return err
		}
	}

	for _, p := range m.Pads {
		if err := p.write(sw); err != nil {
			return err
		}
	}

	if m.Model != nil {
		sw.StartList(true)
		sw.StringScalar("model")
		sw.StringScalar(m.Model.Path)

		if m.Model.At.X == 0 && m.Model.At.Y == 0 && m.Model.At.Z == 0 &&
			(m.Model.Offset.X != 0 || m.Model.Offset.Y != 0 || m.Model.Offset.Z != 0) {
			sw.StartList(true)
			sw.StringScalar("offset")
			if err := m.Model.Offset.writeDouble("xyz", sw); err != nil {
				return err
			}
			if err := sw.CloseList(false); err != nil {
				return err
			}
		} else {
			sw.StartList(true)
			sw.StringScalar("at")
			if err := m.Model.At.writeDouble("xyz", sw); err != nil {
				return err
			}
			if err := sw.CloseList(false); err != nil {
				return err
			}
		}

		sw.StartList(true)
		sw.StringScalar("scale")
		if err := m.Model.Scale.writeDouble("xyz", sw); err != nil {
			return err
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}
		sw.StartList(true)
		sw.StringScalar("rotate")
		if err := m.Model.Rotate.writeDouble("xyz", sw); err != nil {
			return err
		}
		if err := sw.CloseList(false); err != nil {
			return err
		}

		if err := sw.CloseList(true); err != nil {
			return err
		}
	}

	if err := sw.CloseList(true); err != nil {
		return err
	}
	return nil
}

func (l *ModLine) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("fp_line")
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

func (a *ModArc) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("fp_arc")
	if err := a.Start.write("start", sw); err != nil {
		return err
	}
	if err := a.End.write("end", sw); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("angle")
	sw.StringScalar(f(a.Angle))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(a.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("width")
	sw.StringScalar(f(a.Width))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	return sw.CloseList(false)
}

func (c *ModCircle) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("fp_circle")
	if err := c.Center.write("center", sw); err != nil {
		return err
	}
	if err := c.End.write("end", sw); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(c.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("width")
	sw.StringScalar(f(c.Width))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	return sw.CloseList(false)
}

func (t *ModText) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("fp_text")
	sw.StringScalar(t.Kind.String())
	sw.StringScalar(t.Text)
	if err := t.At.write("at", sw); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(t.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}
	if t.Hidden {
		sw.StringScalar("hide")
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

func (p *ModPolygon) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("fp_poly")

	sw.StartList(false)
	sw.StringScalar("pts")
	sw.AdjustIndent(-2)
	for i, pts := range p.Points {
		if err := pts.write("xy", sw); err != nil {
			return err
		}
		if (i%4 == 3) && i < len(p.Points)-1 {
			sw.Newlines(1)
		}
	}
	sw.AdjustIndent(2)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("layer")
	sw.StringScalar(p.Layer)
	if err := sw.CloseList(false); err != nil {
		return err
	}

	sw.StartList(false)
	sw.StringScalar("width")
	sw.StringScalar(f(p.Width))
	if err := sw.CloseList(false); err != nil {
		return err
	}

	return sw.CloseList(false)
}

func (p *Pad) write(sw *swriter.SExpWriter) error {
	sw.StartList(true)
	sw.StringScalar("pad")
	sw.StringScalar(p.Ident)
	sw.StringScalar(p.Surface.String())
	sw.StringScalar(p.Shape.String())

	if err := p.At.write("at", sw); err != nil {
		return err
	}
	if err := p.Size.write("size", sw); err != nil {
		return err
	}

	if p.RectDelta.X != 0 || p.RectDelta.Y != 0 {
		if err := p.RectDelta.write("rect_delta", sw); err != nil {
			return err
		}
	}

	if p.DrillSize.X > 0 || p.DrillSize.Y > 0 || p.DrillOffset.X != 0 || p.DrillOffset.Y != 0 {
		sw.StartList(false)
		sw.StringScalar("drill")
		if p.DrillShape == ShapeDrillOblong {
			sw.StringScalar("oval")
		}

		if p.DrillSize.X > 0 {
			sw.StringScalar(f(p.DrillSize.X))
		}
		if p.DrillSize.Y > 0 && p.DrillSize.Y != p.DrillSize.X {
			sw.StringScalar(f(p.DrillSize.Y))
		}
		if p.DrillOffset.X != 0 || p.DrillOffset.Y != 0 {
			if err := p.DrillOffset.write("offset", sw); err != nil {
				return err
			}
		}

		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	sw.StartList(false)
	sw.StringScalar("layers")
	for _, l := range p.Layers {
		sw.StringScalar(l)
	}
	if err := sw.CloseList(false); err != nil {
		return err
	}

	doNewline := p.NetNum != 0 ||
		p.DieLength != 0 ||
		p.SolderMaskMargin != 0 ||
		p.SolderPasteMargin != 0 ||
		p.SolderPasteMarginRatio != 0 ||
		p.Clearance != 0 ||
		p.ThermalWidth != 0 ||
		p.ThermalGap != 0

	if p.Shape == ShapeRoundRect || p.Shape == ShapeChamferedRect {
		sw.StartList(false)
		sw.StringScalar("roundrect_rratio")
		sw.StringScalar(f(p.RoundRectRRatio))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if doNewline {
		sw.Newlines(1)
	}

	// if p.Shape == ShapeChamferedRect {
	//   sw.Newlines(1)
	//   // TODO: Implement
	// }
	if p.NetNum != 0 {
		sw.StartList(false)
		sw.StringScalar("net")
		sw.IntScalar(p.NetNum)
		sw.StringScalar(p.NetName)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}

	if p.DieLength != 0 {
		sw.StartList(false)
		sw.StringScalar("die_length")
		sw.StringScalar(f(p.DieLength))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.SolderMaskMargin != 0 {
		sw.StartList(false)
		sw.StringScalar("solder_mask_margin")
		sw.StringScalar(f(p.SolderMaskMargin))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.SolderPasteMargin != 0 {
		sw.StartList(false)
		sw.StringScalar("solder_paste_margin")
		sw.StringScalar(f(p.SolderPasteMargin))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.SolderPasteMarginRatio != 0 {
		sw.StartList(false)
		sw.StringScalar("solder_paste_margin_ratio")
		sw.StringScalar(f(p.SolderPasteMarginRatio))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.Clearance != 0 {
		sw.StartList(false)
		sw.StringScalar("clearance")
		sw.StringScalar(f(p.Clearance))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.ZoneConnect != 0 {
		sw.StartList(false)
		sw.StringScalar("zone_connect")
		sw.IntScalar(p.ZoneConnect)
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.ThermalWidth != 0 {
		sw.StartList(false)
		sw.StringScalar("thermal_width")
		sw.StringScalar(f(p.ThermalWidth))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	if p.ThermalGap != 0 {
		sw.StartList(false)
		sw.StringScalar("thermal_gap")
		sw.StringScalar(f(p.ThermalGap))
		if err := sw.CloseList(false); err != nil {
			return err
		}
	}
	// TODO: Custom mode.

	return sw.CloseList(false)
}
