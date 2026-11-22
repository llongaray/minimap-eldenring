package game

import (
	"encoding/binary"
	"math"
	"testing"
)

func sample(id uint32, layer byte, x float32) []byte {
	b := make([]byte, 36)
	binary.LittleEndian.PutUint32(b[:4], 0x0A000000)
	binary.LittleEndian.PutUint32(b[16:20], id)
	binary.LittleEndian.PutUint32(b[20:24], math.Float32bits(x))
	binary.LittleEndian.PutUint32(b[24:28], math.Float32bits(3200))
	b[28] = layer
	binary.LittleEndian.PutUint32(b[32:36], math.Float32bits(90))
	return b
}
func TestParseValidSamples(t *testing.T) {
	for _, tt := range []struct {
		a    []byte
		want string
	}{{sample(0, 0, 200), "M00"}, {sample(1, 1, 200), "M01"}, {sample(10, 0, 200), "M10"}} {
		p, ok := ParseView(tt.a)
		if !ok || p.Layer != tt.want || p.Heading != 90 {
			t.Fatal(p, ok, tt.want)
		}
	}
}
func TestRejectBadSamples(t *testing.T) {
	if _, ok := ParseView([]byte{0}); ok {
		t.Fatal("short accepted")
	}
	if _, ok := ParseView(sample(0, 0, float32(math.NaN()))); ok {
		t.Fatal("NaN accepted")
	}
	if _, ok := ParseView(sample(0, 0, 15000)); ok {
		t.Fatal("out of range accepted")
	}
	if _, ok := ParseView(sample(17, 0, 230)); ok {
		t.Fatal("mapID accepted")
	}
}
