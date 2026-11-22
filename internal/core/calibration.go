package core

import (
    "errors"
    "math"
)

// Anchor pairs an independently read game-map position with a manually
// confirmed point on the local 10496-pixel map.
type Anchor struct {
    NativeX, NativeY float64
    PixelX, PixelY float64
    Layer Layer
}

func (a Anchor) Valid() bool {
    return a.Layer.Valid() && finite(a.NativeX) && finite(a.NativeY) &&
        finite(a.PixelX) && finite(a.PixelY) &&
        a.NativeX>=0 && a.NativeX<=MasterPixels &&
        a.NativeY>=0 && a.NativeY<=MasterPixels &&
        a.PixelX>=0 && a.PixelX<=MasterPixels &&
        a.PixelY>=0 && a.PixelY<=MasterPixels
}

// One point is only an offset estimate with a 1:1 scale assumption.
// It is intentionally marked provisional until a second separated point exists.
func CalibrateOne(a Anchor) (Calibration,error) {
    if !a.Valid() { return Calibration{},errors.New("ponto de referencia invalido") }
    return Calibration{Enabled:true, Layer:a.Layer, Anchors:1,
        ScaleX:1, ScaleY:1, OffsetX:a.PixelX-a.NativeX, OffsetY:a.PixelY-a.NativeY},nil
}

func CalibrateTwo(a,b Anchor) (Calibration,error) {
    if !a.Valid() || !b.Valid() || a.Layer!=b.Layer {
        return Calibration{},errors.New("referencias precisam estar na mesma camada")
    }
    if math.Abs(b.NativeX-a.NativeX)<80 || math.Abs(b.NativeY-a.NativeY)<80 {
        return Calibration{},errors.New("os pontos devem estar distantes nos eixos X e Y (>80 unidades)")
    }
    sx:=(b.PixelX-a.PixelX)/(b.NativeX-a.NativeX)
    sy:=(b.PixelY-a.PixelY)/(b.NativeY-a.NativeY)
    if !finite(sx)||!finite(sy)||math.Abs(sx)<0.1||math.Abs(sy)<0.1||math.Abs(sx)>10||math.Abs(sy)>10 {
        return Calibration{},errors.New("escala calculada invalida; confira os pontos")
    }
    return Calibration{Enabled:true,Layer:a.Layer,Anchors:2,
        ScaleX:sx,ScaleY:sy,OffsetX:a.PixelX-sx*a.NativeX,
        OffsetY:a.PixelY-sy*a.NativeY},nil
}
