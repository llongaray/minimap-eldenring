package game

import (
	"encoding/binary"
	"math"
)

type RawSample struct {
	X       float64 `json:"native_x"`
	Y       float64 `json:"native_y"`
	Heading float64 `json:"heading"`
	Layer   string  `json:"layer"`
	Ready   bool    `json:"ready"`
	Status  string  `json:"status"`
}

// ParseView is a conservative decoder for an untrusted 36-byte world map
// view. It checks plausibility but DOES NOT claim the numbers are correctly
// calibrated to the third-party map's pixel coordinates.
func ParseView(raw []byte) (RawSample, bool) {
	if len(raw) != 36 {
		return RawSample{}, false
	}
	mapCode := binary.LittleEndian.Uint32(raw)
	id := binary.LittleEndian.Uint32(raw[16:])
	x := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[20:])))
	y := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[24:])))
	heading := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[32:])))
	area := mapCode >> 24
	if area < 10 || area > 65 || (id != 0 && id != 1 && id != 10) || math.IsNaN(x) || math.IsNaN(y) || x < 0 || y < 0 || x > 10496 || y > 10496 || (x == 0 && y == 0) || math.IsNaN(heading) || math.IsInf(heading, 0) || heading < -720 || heading > 720 || raw[28] > 1 {
		return RawSample{}, false
	}
	layer := "M00"
	if id == 1 || raw[28] == 1 {
		layer = "M01"
	}
	if id == 10 || area == 61 {
		layer = "M10"
	}
	return RawSample{X: x, Y: y, Heading: heading, Layer: layer, Ready: true, Status: "Leitura nativa válida; calibração pendente"}, true
}
