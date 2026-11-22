package core

import (
    "math"
    "testing"
)

func TestZoomAtPreservesPointerWorldPixel(t *testing.T) {
    x0, y0 := 5000.0, 4800.0
    boxX, boxY, w, h := 270, 82, 1630, 850
    px, py := 1270.0, 420.0
    oldScale := float64(w)/MasterPixels
    oldX := x0 + (px-float64(boxX)-float64(w)*0.5)/oldScale
    oldY := y0 + (py-float64(boxY)-float64(h)*0.5)/oldScale
    x1,y1 := ZoomAt(x0,y0,boxX,boxY,w,h,1,1.25,px,py)
    nextScale := oldScale*1.25
    newX := x1 + (px-float64(boxX)-float64(w)*0.5)/nextScale
    newY := y1 + (py-float64(boxY)-float64(h)*0.5)/nextScale
    if math.Abs(newX-oldX) > .0001 || math.Abs(newY-oldY) > .0001 {
        t.Fatalf("anchoring lost: before=(%f,%f) after=(%f,%f)",oldX,oldY,newX,newY)
    }
}
func TestClampCenterAtMapEdges(t *testing.T) {
    x,y:=ClampCenter(-100,200000,1400,850,2)
    if x <= 0 || x>= 10496 || y <= 0 || y>=10496 {
        t.Fatalf("clamp failure: %f,%f",x,y)
    }
    x,y=ClampCenter(2000,1234,1400,850,.5)
    if x != 5248 || y != 5248 { t.Fatalf("full-sized canvas should be centered: %f,%f",x,y) }
}
