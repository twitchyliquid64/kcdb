package mod

import (
	"errors"
	"io"

	"github.com/nsf/sexp"
)

// Point2D represents a point in 2-dimensional space.
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// FpLine represents a graphical line.
type FpLine struct {
	Start Point2D `json:"start"`
	End   Point2D `json:"end"`
	Layer string  `json:"layer"`
	Width float64 `json:"width"`
}

// FpText represents graphical text.
type FpText struct {
	Pos    Point2D `json:"position"`
	Kind   string  `json:"kind"`
	Value  string  `json:"value"`
	Layer  string  `json:"layer"`
	Hidden bool    `json:"hidden"`

	Size      Point2D `json:"size"`
	Thickness float64 `json:"thickness"`
}

// Pad represents a pad in a component footprint.
type Pad struct {
	Pin   int    `json:"pin"`
	Kind  string `json:"kind"`
	Shape string `json:"shape"`
	Drill Drill  `json:"drill"`

	Pos    Point2D  `json:"position"`
	Size   Point2D  `json:"size"`
	Layers []string `json:"layers"`
}

// Drill represents pad drill parameters.
type Drill struct {
	Kind    string  `json:"kind"`
	Scalar  float64 `json:"scalar"`
	Ellipse Point2D `json:"ellipse"`
	Offset  Point2D `json:"offset"`
}

// Module represents a Kicad module.
type Module struct {
	Name        string `json:"name"`
	Tedit       string `json:"tedit"`
	Description string `json:"description"`
	Layer       string `json:"layer"`
	Model       string `json:"model"`

	Tags  []string `json:"tags"`
	Attrs []string `json:"attrs"`
	Lines []FpLine `json:"lines"`
	Texts []FpText `json:"texts"`
	Pads  []Pad    `json:"pads"`
}

// DecodeModule reads a .kicad_mod file from a reader.
func DecodeModule(r io.RuneReader) (*Module, error) {
	out := &Module{}
	ast, err := sexp.Parse(r, nil)

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

	if mainAST.NumChildren() < 3 {
		return nil, errors.New("invalid format: missing minimum elements")
	}
	if s, err2 := sexp.Help(mainAST).Child(0).String(); err2 != nil || s != "module" {
		return nil, errors.New("invalid format: missing module prefix")
	}

	out.Name, err = sexp.Help(mainAST).Child(1).String()
	if err != nil {
		return nil, errors.New("invalid format: expected string value for module name")
	}

	for i := 2; i < mainAST.NumChildren(); i++ {
		n := sexp.Help(mainAST).Child(i)
		if n.IsList() && n.Child(1).IsValid() {
			switch n.Child(0).MustString() {
			case "layer":
				out.Layer, err = n.Child(1).String()
				if err != nil {
					return nil, errors.New("invalid format: layer value must be a string")
				}
			case "tedit":
				out.Tedit, err = n.Child(1).String()
				if err != nil {
					return nil, errors.New("invalid format: tedit value must be a string")
				}
			case "descr":
				out.Description, err = n.Child(1).String()
				if err != nil {
					return nil, errors.New("invalid format: tedit value must be a string")
				}
			case "tags":
				for x := 1; x < n.MustNode().NumChildren(); x++ {
					var t string
					t, err = n.Child(1).String()
					if err != nil {
						return nil, errors.New("invalid format: tag value must be a string")
					}
					out.Tags = append(out.Tags, t)
				}
			case "attr":
				for x := 1; x < n.MustNode().NumChildren(); x++ {
					var t string
					t, err = n.Child(1).String()
					if err != nil {
						return nil, errors.New("invalid format: tag value must be a string")
					}
					out.Attrs = append(out.Attrs, t)
				}
			case "model":
				out.Model, err = n.Child(1).String()
				if err != nil {
					return nil, errors.New("invalid format: model value must be a string")
				}
			case "fp_line":
				line, err := unmarshalFpLine(n)
				if err != nil {
					return nil, err
				}
				out.Lines = append(out.Lines, line)
			case "fp_text":
				txt, err := unmarshalFpText(n)
				if err != nil {
					return nil, err
				}
				out.Texts = append(out.Texts, txt)
			case "pad":
				pad, err := unmarshalPad(n)
				if err != nil {
					return nil, err
				}
				out.Pads = append(out.Pads, pad)
			default:
				return nil, errors.New("cannot handle expression: " + n.Child(0).MustString())
			}
		}

	}

	return out, nil
}

func unmarshalFpLine(n sexp.Helper) (FpLine, error) {
	line := FpLine{}
	for x := 1; x < n.MustNode().NumChildren(); x++ {
		switch n.Child(x).Child(0).MustString() {
		case "start":
			line.Start = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
		case "end":
			line.End = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
		case "layer":
			line.Layer = n.Child(x).Child(1).MustString()
		case "width":
			line.Width = n.Child(x).Child(1).MustFloat64()
		}
	}
	return line, nil
}

func unmarshalFpText(n sexp.Helper) (FpText, error) {
	txt := FpText{
		Kind:  n.Child(1).MustString(),
		Value: n.Child(2).MustString(),
	}

	for x := 3; x < n.MustNode().NumChildren(); x++ {
		if n.Child(x).IsScalar() {
			if s, err := n.Child(x).String(); err == nil {
				switch s {
				case "hide":
					txt.Hidden = true
				}
				continue
			}
		}

		switch n.Child(x).Child(0).MustString() {
		case "at":
			txt.Pos = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
		case "layer":
			txt.Layer = n.Child(x).Child(1).MustString()
		case "effects":
			s := n.Child(x).Child(1)
			for i := 1; i < s.MustNode().NumChildren(); i++ {
				switch s.Child(i).Child(0).MustString() {
				case "size":
					txt.Size = Point2D{X: s.Child(i).Child(1).MustFloat64(), Y: s.Child(i).Child(2).MustFloat64()}
				case "thickness":
					txt.Thickness = s.Child(i).Child(1).MustFloat64()
				}
			}

		}
	}
	return txt, nil
}

func decodeDrill(n sexp.Helper) (Drill, error) {
	d := Drill{}

	for x := 1; x < n.MustNode().NumChildren(); x++ {
		if n.Child(x).IsList() {
			switch n.Child(x).Child(0).MustString() {
			case "offset":
				d.Offset = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
			}
		} else {
			if _, err := n.Child(x).Float64(); err != nil { // kind + 2d parameters
				d.Kind = n.Child(x).MustString()
				d.Ellipse = Point2D{X: n.Child(x + 1).MustFloat64(), Y: n.Child(x + 2).MustFloat64()}
				x += 2
				return d, nil
			} else {
				// just a scalar (radius)
				d.Scalar = n.Child(1).MustFloat64()
			}
		}
	}

	return d, nil
}

func unmarshalPad(n sexp.Helper) (Pad, error) {
	var err error
	pad := Pad{
		Kind:  n.Child(2).MustString(),
		Shape: n.Child(3).MustString(),
	}
	if v, err := n.Child(1).Int(); err == nil { // Pads without an int pin are just disconnected ones
		pad.Pin = v
	}

	for x := 4; x < n.MustNode().NumChildren(); x++ {
		switch n.Child(x).Child(0).MustString() {
		case "at":
			pad.Pos = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
		case "size":
			pad.Size = Point2D{X: n.Child(x).Child(1).MustFloat64(), Y: n.Child(x).Child(2).MustFloat64()}
		case "drill":
			pad.Drill, err = decodeDrill(n.Child(x))
			if err != nil {
				return pad, err
			}

		case "layers":
			s := n.Child(x)
			for i := 1; i < s.MustNode().NumChildren(); i++ {
				pad.Layers = append(pad.Layers, s.Child(i).MustString())
			}

		}
	}
	return pad, nil
}
