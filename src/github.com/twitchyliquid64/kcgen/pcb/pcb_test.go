package pcb

import (
	"reflect"
	"testing"
)

func TestPCB(t *testing.T) {
	p, err := DecodeFile("testdata/t1.kicad_pcb")
	if err != nil {
		t.Fatalf("DecodeFile() failed: %v", err)
	}

	if got, want := p.FormatVersion, 4; got != want {
		t.Errorf("p.FormatVersion = %v, want %v", got, want)
	}

	if got, want := p.EditorSetup.SegmentWidth, 0.2; got != want {
		t.Errorf("p.EditorSetup.SegmentWidth = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.UViaMinSize, 0.2; got != want {
		t.Errorf("p.EditorSetup.UViaMinSize = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.TextSize, []float64{1.5, 1.5}; !reflect.DeepEqual(got, want) {
		t.Errorf("p.EditorSetup.TextWidth = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.PadSize, []float64{1.524, 1.524}; !reflect.DeepEqual(got, want) {
		t.Errorf("p.EditorSetup.PadSize = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.PadDrill, 0.762; got != want {
		t.Errorf("p.EditorSetup.PadDrill = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.ModTextWidth, 0.15; got != want {
		t.Errorf("p.EditorSetup.ModTextWidth = %v, want %v", got, want)
	}
	if got, want := p.EditorSetup.PlotParams["linewidth"], (PlotParam{
		name:   "linewidth",
		values: []string{"0.100000"},
		order:  4,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("p.EditorSetup.PlotParams['linewidth'] = %v, want %v", got, want)
	}

	if got, want := len(p.LayersByName), 20; got != want {
		t.Errorf("len(p.LayersByName) = %v, want %v", got, want)
		t.Logf("p.LayersByName = %+v", p.LayersByName)
	}
	if got, want := p.LayersByName["F.Mask"].Type, "user"; got != want {
		t.Errorf("p.LayersByName[\"F.Mask\"].Type = %v, want %v", got, want)
		t.Logf("p.LayersByName[\"F.Mask\"] = %+v", p.LayersByName["F.Mask"])
	}

	if got, want := len(p.LayersByName), 20; got != want {
		t.Errorf("len(p.LayersByName) = %v, want %v", got, want)
		t.Logf("p.LayersByName = %+v", p.LayersByName)
	}
	if got, want := p.LayersByName["F.Mask"].Type, "user"; got != want {
		t.Errorf("p.LayersByName[\"F.Mask\"].Type = %v, want %v", got, want)
		t.Logf("p.LayersByName[\"F.Mask\"] = %+v", p.LayersByName["F.Mask"])
	}

	if got, want := len(p.Nets), 7; got != want {
		t.Errorf("len(p.Nets) = %v, want %v", got, want)
		t.Logf("p.Nets = %+v", p.Nets)
	}
	if got, want := p.Nets[1].Name, "GND"; got != want {
		t.Errorf("p.Nets[1].Name = %v, want %v", got, want)
		t.Logf("p.Nets[1] = %+v", p.Nets[1].Name)
	}

	if got, want := len(p.Zones), 1; got != want {
		t.Errorf("len(p.Zones) = %v, want %v", got, want)
		t.Logf("p.Zones = %+v", p.Zones)
	}
	if got, want := p.Zones[0].NetName, "GND"; got != want {
		t.Errorf("p.Zones[0].NetName = %v, want %v", got, want)
	}
	if got, want := p.Zones[0].Layer, "B.Cu"; got != want {
		t.Errorf("p.Zones[0].Layer = %v, want %v", got, want)
	}
	if got, want := p.Zones[0].MinThickness, 0.254; got != want {
		t.Errorf("p.Zones[0].MinThickness = %v, want %v", got, want)
	}
	if got, want := p.Zones[0].ConnectPads.Clearance, 0.508; got != want {
		t.Errorf("p.Zones[0].ConnectPads.Clearance = %v, want %v", got, want)
	}
	if got, want := p.Zones[0].Fill.Enabled, true; got != want {
		t.Errorf("p.Zones[0].Fill.Enabled = %v, want %v", got, want)
	}
	if got, want := p.Zones[0].Fill.Segments, 16; got != want {
		t.Errorf("p.Zones[0].Fill.Segments = %v, want %v", got, want)
	}

	if got, want := len(p.Tracks), 44; got != want {
		t.Errorf("len(p.Tracks) = %v, want %v", got, want)
		t.Logf("p.Tracks = %+v", p.Tracks)
	}
	if got, want := p.Tracks[11].NetIndex, 2; got != want {
		t.Errorf("p.Tracks[11].NetIndex = %v, want %v", got, want)
	}
	if got, want := p.Tracks[11].Start.X, 136.652; got != want {
		t.Errorf("p.Tracks[11].Start.X = %v, want %v", got, want)
	}

	if got, want := len(p.Vias), 1; got != want {
		t.Errorf("len(p.Vias) = %v, want %v", got, want)
		t.Logf("p.Vias = %+v", p.Vias)
	}
	if got, want := p.Vias[0].NetIndex, 1; got != want {
		t.Errorf("p.Vias[0].NetIndex = %v, want %v", got, want)
	}
	if got, want := p.Vias[0].At.X, 88.1; got != want {
		t.Errorf("p.Vias[0].X = %v, want %v", got, want)
	}
	if got, want := p.Vias[0].Drill, 0.4; got != want {
		t.Errorf("p.Vias[0].Drill = %v, want %v", got, want)
	}
	if got, want := p.Vias[0].Layers, []string{"F.Cu", "B.Cu"}; !reflect.DeepEqual(got, want) {
		t.Errorf("p.Vias[0].Layers = %v, want %v", got, want)
	}

	if got, want := len(p.NetClasses), 1; got != want {
		t.Errorf("len(p.NetClasses) = %v, want %v", got, want)
		t.Logf("p.NetClasses = %+v", p.NetClasses)
	}
	if got, want := p.NetClasses[0].Name, "Default"; got != want {
		t.Errorf("p.NetClasses[0].Name = %v, want %v", got, want)
	}
	if got, want := p.NetClasses[0].TraceWidth, 0.25; got != want {
		t.Errorf("p.NetClasses[0].TraceWidth = %v, want %v", got, want)
	}
	if got, want := p.NetClasses[0].Nets[0], "/BUS_A"; got != want {
		t.Errorf("p.NetClasses[0].Nets[0] = %v, want %v", got, want)
	}

	if got, want := len(p.Lines), 4; got != want {
		t.Errorf("len(p.Lines) = %v, want %v", got, want)
		t.Logf("p.Lines = %+v", p.Lines)
	}
	if got, want := p.Lines[0].Width, 0.15; got != want {
		t.Errorf("p.Lines[0].Width = %v, want %v", got, want)
	}
	if got, want := p.Lines[0].Start.X, 173.736; got != want {
		t.Errorf("p.Lines[0].Start.X = %v, want %v", got, want)
	}

	if got, want := len(p.Texts), 1; got != want {
		t.Errorf("len(p.Texts) = %v, want %v", got, want)
		t.Logf("p.Texts = %+v", p.Texts)
	}
	if got, want := p.Texts[0].Text, "Oops"; got != want {
		t.Errorf("p.Texts[0].Text = %v, want %v", got, want)
	}
	if got, want := p.Texts[0].Effects.FontSize.X, 1.5; got != want {
		t.Errorf("p.Texts[0].Effects.FontSize.X = %v, want %v", got, want)
	}

	if got, want := len(p.Dimensions), 2; got != want {
		t.Errorf("len(p.Dimensions) = %v, want %v", got, want)
		t.Logf("p.Dimensions = %+v", p.Dimensions)
	}
	if got, want := p.Dimensions[0].Width, 0.3; got != want {
		t.Errorf("p.Dimensions[0].Width = %v, want %v", got, want)
	}
	if got, want := p.Dimensions[0].Text.Text, "12.446 mm"; got != want {
		t.Errorf("p.Dimensions[0].Text.Text = %v, want %v", got, want)
	}
	if got, want := p.Dimensions[0].Features[1].Feature, "feature2"; got != want {
		t.Errorf("p.Dimensions[0].Features[1].Feature = %v, want %v", got, want)
	}
}
