package core

import "math"

const MasterPixels = 10496.0

// ZoomAt preserves the map pixel under the pointer when zooming.
// Coordinates are in absolute window pixels; pan is measured in master pixels.
func ZoomAt(cx, cy float64, boxX, boxY, boxW, boxH int, current, next, cursorX, cursorY float64) (float64, float64) {
    if boxW <= 0 || boxH <= 0 || current <= 0 || next <= 0 {
        return cx, cy
    }
    relativeX := cursorX - (float64(boxX) + float64(boxW)*0.5)
    relativeY := cursorY - (float64(boxY) + float64(boxH)*0.5)
    oldScale := float64(boxW) / MasterPixels * current
    newScale := float64(boxW) / MasterPixels * next
    cx += relativeX/oldScale - relativeX/newScale
    cy += relativeY/oldScale - relativeY/newScale
    return ClampCenter(cx, cy, boxW, boxH, next)
}

// ClampCenter prevents the viewport from drifting outside the master map,
// except when the viewport itself exceeds the entire map.
func ClampCenter(cx, cy float64, boxW, boxH int, zoom float64) (float64, float64) {
    if boxW <= 0 || boxH <= 0 || zoom <= 0 { return cx, cy }
    scale := float64(boxW) / MasterPixels * zoom
    halfW := float64(boxW) / (2*scale)
    halfH := float64(boxH) / (2*scale)
    limit := func(v, half float64) float64 {
        if math.IsNaN(v) || math.IsInf(v, 0) { return MasterPixels*0.5 }
        if half >= MasterPixels*0.5 { return MasterPixels*0.5 }
        return math.Max(half, math.Min(MasterPixels-half, v))
    }
    return limit(cx,halfW), limit(cy,halfH)
}
