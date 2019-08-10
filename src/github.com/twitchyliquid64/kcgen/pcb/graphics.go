package pcb

import (
	"fmt"

	"github.com/nsf/sexp"
)

// TODO(twitchyliquid64): Refactor graphical features to implement
// some common interface.

// Text represents some text to be rendered.
type Text struct {
	Text     string `json:"value"`
	Layer    string `json:"layer"`
	At       XYZ    `json:"position"`
	Unlocked bool   `json:"unlocked"`

	Effects TextEffects `json:"effects"`
	Hidden  bool        `json:"hidden"`

	order int
}

type TextJustify uint8

func (j TextJustify) String() string {
	switch j {
	case JustifyMirror:
		return "mirror"
	case JustifyLeft:
		return "left"
	case JustifyRight:
		return "right"
	case JustifyTop:
		return "top"
	case JustifyBottom:
		return "bottom"
	}
	return "???????"
}

const (
	JustifyNone TextJustify = iota
	JustifyMirror
	JustifyLeft
	JustifyRight
	JustifyTop
	JustifyBottom
)

// TextEffects describes styling which can be applied to a text drawing.
type TextEffects struct {
	FontSize  XY          `json:"size"`
	Thickness float64     `json:"thickness"`
	Justify   TextJustify `json:"justify"`

	Bold   bool `json:"bold"`
	Italic bool `json:"italic"`
}

// Line represents a graphical line.
type Line struct {
	Start XY      `json:"start"`
	End   XY      `json:"end"`
	Layer string  `json:"layer"`
	Width float64 `json:"width"`

	order int
}

// Dimension represents a measurement graphic.
type Dimension struct {
	CurrentMeasurement float64 `json:"value"`

	Text     Text               `json:"text"`
	Features []DimensionFeature `json:"features"`

	Width float64 `json:"width"`
	Layer string  `json:"layer"`

	order int
}

// DimensionFeature is a graphical element used as part of a
// dimension.
type DimensionFeature struct {
	Feature string `json:"feature"`
	Points  []XY   `json:"points,omitempty"`
}

func parseDimension(n sexp.Helper, ordering int) (Dimension, error) {
	d := Dimension{
		CurrentMeasurement: n.Child(1).MustFloat64(),
		order:              ordering,
	}
	for x := 2; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "width":
			d.Width = c.Child(1).MustFloat64()
		case "layer":
			d.Layer = c.Child(1).MustString()
		case "gr_text":
			t, err := parseGRText(c, x)
			if err != nil {
				return Dimension{}, err
			}
			d.Text = t
		case "feature1", "feature2", "crossbar", "arrow1a", "arrow1b", "arrow2a", "arrow2b":
			f := DimensionFeature{
				Feature: c.Child(0).MustString(),
			}
			for y := 1; y < c.MustNode().NumChildren(); y++ {
				c := c.Child(y)
				switch c.Child(0).MustString() {
				case "pts":
					for z := 1; z < c.MustNode().NumChildren(); z++ {
						c := c.Child(z)
						switch c.Child(0).MustString() {
						case "xy":
							p := XY{X: c.Child(1).MustFloat64(), Y: c.Child(2).MustFloat64()}
							f.Points = append(f.Points, p)
						}
					}
				}
			}
			d.Features = append(d.Features, f)
		}
	}
	return d, nil
}

func parseGRText(n sexp.Helper, ordering int) (Text, error) {
	t := Text{
		Text:  n.Child(1).MustString(),
		order: ordering,
	}
	for x := 2; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "at":
			t.At.X = c.Child(1).MustFloat64()
			t.At.Y = c.Child(2).MustFloat64()
			if c.MustNode().NumChildren() >= 4 {
				if f, err := c.Child(3).Float64(); err == nil {
					t.At.Z = f
					t.At.ZPresent = true
				} else if c.Child(3).MustString() == "unlocked" {
					t.Unlocked = true
				}
			}
		case "hide":
			t.Hidden = true
		case "layer":
			t.Layer = c.Child(1).MustString()
		case "effects":
			effects, err := parseTextEffects(c)
			if err != nil {
				return Text{}, err
			}
			t.Effects = effects
		}
	}
	return t, nil
}

func parseTextEffects(n sexp.Helper) (TextEffects, error) {
	var e TextEffects
	for y := 1; y < n.MustNode().NumChildren(); y++ {
		c := n.Child(y)
		switch c.Child(0).MustString() {
		case "font":
			for z := 1; z < c.MustNode().NumChildren(); z++ {
				c := c.Child(z)

				if c.IsScalar() {
					switch c.MustNode().Value {
					case "italic":
						e.Italic = true
					case "bold":
						e.Bold = true
					default:
						return TextEffects{}, fmt.Errorf("unhandled scalar in text effects: %v", c.MustNode().Value)
					}
				}

				switch c.Child(0).MustString() {
				case "size":
					e.FontSize.X = c.Child(1).MustFloat64()
					e.FontSize.Y = c.Child(2).MustFloat64()
				case "thickness":
					e.Thickness = c.Child(1).MustFloat64()
				}
			}
		case "justify":
			switch c.Child(1).MustString() {
			case "mirror":
				e.Justify = JustifyMirror
			case "top":
				e.Justify = JustifyTop
			case "bottom":
				e.Justify = JustifyBottom
			case "left":
				e.Justify = JustifyLeft
			case "right":
				e.Justify = JustifyRight
			default:
				return TextEffects{}, fmt.Errorf("unknown justify value: %q", c.Child(1).MustString())
			}
		}
	}
	return e, nil
}

func parseGRLine(n sexp.Helper, ordering int) (Line, error) {
	l := Line{order: ordering}
	for x := 1; x < n.MustNode().NumChildren(); x++ {
		c := n.Child(x)
		switch c.Child(0).MustString() {
		case "start":
			l.Start.X = c.Child(1).MustFloat64()
			l.Start.Y = c.Child(2).MustFloat64()
		case "end":
			l.End.X = c.Child(1).MustFloat64()
			l.End.Y = c.Child(2).MustFloat64()
		case "width":
			l.Width = c.Child(1).MustFloat64()
		case "layer":
			l.Layer = c.Child(1).MustString()
		}
	}
	return l, nil
}
