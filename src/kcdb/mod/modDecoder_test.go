package mod

import (
	"strings"
	"testing"
)

func TestDecodeMod(t *testing.T) {
	out, err := DecodeModule(strings.NewReader(`
    (module CR2032-white-smd (layer F.Cu) (tedit 566FEBDB)
      (fp_text reference REF** (at 0 -3.3) (layer F.SilkS)
        (effects (font (size 1 1) (thickness 0.15)))
      )
      (fp_text value CR2032-white-smd (at 0 3.35) (layer F.Fab)
        (effects (font (size 1 1) (thickness 0.15)))
      )
      (fp_line (start -1.8 7.45) (end 1.8 7.45) (layer F.SilkS) (width 0.15))
      (fp_line (start -1.8 -7.9) (end 1.8 -7.9) (layer F.SilkS) (width 0.15))
      (fp_text user BAT (at 7 0) (layer F.SilkS)
        (effects (font (size 1 1) (thickness 0.25)))
      )
      (fp_line (start -5.5 0) (end -2.5 2.5) (layer F.SilkS) (width 0.15))
      (fp_line (start 2.5 -2.5) (end 5.5 0) (layer F.SilkS) (width 0.15))
      (fp_text user Adhesive (at 0 0) (layer F.SilkS)
        (effects (font (size 1 1) (thickness 0.15)))
      )
      (fp_line (start -5.5 -2.5) (end -5.5 2.5) (layer F.SilkS) (width 0.15))
      (fp_line (start 5.5 -2.5) (end 5.5 2.5) (layer F.SilkS) (width 0.15))
      (fp_line (start -5.5 -2.5) (end 5.5 -2.5) (layer F.SilkS) (width 0.15))
      (fp_line (start -5.5 2.5) (end 5.5 2.5) (layer F.SilkS) (width 0.15))
      (pad 1 smd rect (at 14.65 0) (size 2.6 3.6) (layers F.Cu F.Paste F.Mask))
      (pad 2 smd rect (at -14.65 0) (size 2.6 3.6) (layers F.Cu F.Paste F.Mask))
    )

    `))

	if err != nil || out == nil {
		t.Errorf("Expected value and no error, got err = %v, out = %+v", err, out)
	}
}

func TestTextHideSet(t *testing.T) {
	tc := `
	(module CE-Logo_11.2x8mm_SilkScreen (layer F.Cu) (tedit 0)
	  (descr "CE marking")
	  (tags "Logo CE certification")
	  (attr virtual)
	  (fp_text reference REF** (at 0 0) (layer F.SilkS) hide
	    (effects (font (size 1 1) (thickness 0.15)))
	  )
	  (fp_text value CE-Logo_11.2x8mm_SilkScreen (at 0.75 0) (layer F.Fab) hide
	    (effects (font (size 1 1) (thickness 0.15)))
	  )
	)`
	out, err := DecodeModule(strings.NewReader(tc))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Texts) != 2 {
		t.Fatalf("Expected 2 texts, got %d", len(out.Texts))
	}
	if !out.Texts[0].Hidden {
		t.Fatal("Expected texts[0] to be hidden")
	}
}

func TestOvalDrill(t *testing.T) {
	tc := `
	(module Valve_ECC-83-2 (layer F.Cu) (tedit 5A030096)
	  (descr "Valve ECC-83-2 flat pins")
	  (tags "Valve ECC-83-2 flat pins")
	  (pad 1 thru_hole oval (at 0 0 306) (size 2.03 3.05) (drill oval 1.02 2.03) (layers *.Cu *.Mask))
	  (pad 2 thru_hole circle (at -3.45 -4.75) (size 4.5 4.5) (drill 3.1) (layers *.Cu *.Mask))
	)`
	out, err := DecodeModule(strings.NewReader(tc))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Pads) != 2 {
		t.Fatalf("Expected 2 pads, got %d", len(out.Pads))
	}
	if out.Pads[0].Drill.Kind != "oval" || out.Pads[0].Drill.Ellipse.X != 1.02 {
		t.Fatal("Incorrect drill[0] value")
	}
	if out.Pads[1].Drill.Scalar != 3.1 {
		t.Fatal("Incorrect drill[1] value")
	}
}

func TestDecodeModAt(t *testing.T) {
	out, err := DecodeModule(strings.NewReader(`
		(module Gauge_50mm_Type2_SilkScreenTop (layer F.Cu)
		  (at 0 0)
		  (descr "Gauge, Massstab, 50mm, SilkScreenTop, Type 2,")
		  (tags "Gauge Massstab 50mm SilkScreenTop Type 2")
		  (attr virtual)
		  (fp_text reference REF** (at 20.50034 9.99998) (layer F.SilkS)
		    (effects (font (size 1 1) (thickness 0.15)))
		  )
		  (fp_line (start 9.99998 0) (end 9.99998 1.99898) (layer F.SilkS) (width 0.15))
		)
    `))

	if err != nil || out == nil {
		t.Errorf("Expected value and no error, got err = %v, out = %+v", err, out)
	}
}

func TestDecodeModPoly(t *testing.T) {
	out, err := DecodeModule(strings.NewReader(`
		(module WEEE-Logo_8.4x12mm_SilkScreen (layer F.Cu) (tedit 0)
		  (descr "Waste Electrical and Electronic Equipment Directive")
		  (tags "Logo WEEE")
		  (attr virtual)
		  (fp_text reference REF** (at 0 0) (layer F.SilkS) hide
		    (effects (font (size 1 1) (thickness 0.15)))
		  )
		  (fp_text value WEEE-Logo_8.4x12mm_SilkScreen (at 0.75 0) (layer F.Fab) hide
		    (effects (font (size 1 1) (thickness 0.15)))
		  )
		  (fp_poly (pts (xy 3.461372 5.976471) (xy -3.511177 5.976471) (xy -3.511177 4.258235) (xy 3.461372 4.258235)
		    (xy 3.461372 5.976471)) (layer F.SilkS) (width 0.01))
		)    `))

	if err != nil || out == nil {
		t.Errorf("Expected value and no error, got err = %v, out = %+v", err, out)
	}
}
