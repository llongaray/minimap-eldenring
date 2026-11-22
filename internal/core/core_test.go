package core

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestPersistenceAndIsolation(t *testing.T) {
	s, e := NewStore(filepath.Join(t.TempDir(), "config.json"))
	if e != nil {
		t.Fatal(e)
	}
	v := s.Get()
	TogglePin(&v, "m1")
	ToggleFound(&v, "m1")
	if e = s.Put(v); e != nil {
		t.Fatal(e)
	}
	other := s.Get()
	if !other.Markers["m1"].Pinned || !other.Markers["m1"].Found {
		t.Fatal(other)
	}
	other.Markers["m1"] = Flags{}
	if !s.Get().Markers["m1"].Pinned {
		t.Fatal("shared map mutation")
	}
	reopened, e := NewStore(filepath.Join(filepath.Dir(s.file), "config.json"))
	if e != nil || !reopened.Get().Markers["m1"].Found {
		t.Fatalf("reload: %v", e)
	}
}
func TestMarkerParser(t *testing.T) {
	raw := []byte(`const MARKERS = [{ id:"g1", cat:"grace", name:"Portão do Castelo", master:"M00", px:3099.5, py:6462.7, desc:"teste" }, {id:"boss",cat:"boss",name:"Chefão",master:"M10",px:3300,py:7000}];`)
	got, e := FromReferenceJS(raw)
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 2 || got[0].X != 3099.5 || got[1].Master != Shadow {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestCalibrationFailClosed(t *testing.T) {
	c := Calibration{}
	if _, e := c.Apply(500, 500, Surface, 0); e == nil {
		t.Fatal("calibration enabled without validation")
	}
	c.Enabled = true
	c.ScaleX = 1
	c.ScaleY = 1
	p, e := c.Apply(200, 300, Surface, math.Pi)
	if e != nil || p.X != 200 {
		t.Fatal(p, e)
	}
	if _, e = c.Apply(math.NaN(), 100, Surface, 0); e == nil {
		t.Fatal("NaN not caught")
	}
}
func TestRejectCorruptConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.json")
	os.WriteFile(p, []byte(`{"schema":998}`), 0600)
	if _, e := NewStore(p); e == nil {
		t.Fatal("old schema accepted")
	}
}
