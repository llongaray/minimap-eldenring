package core

import (
    "math"
    "testing"
)

func TestCalibrationAnchors(t *testing.T){
    a:=Anchor{NativeX:2000,NativeY:3000,PixelX:3100,PixelY:4100,Layer:Surface}
    b:=Anchor{NativeX:3000,NativeY:4200,PixelX:4100,PixelY:5300,Layer:Surface}
    one,err:=CalibrateOne(a)
    if err!=nil || one.Anchors!=1 { t.Fatalf("one: %+v / %v",one,err) }
    two,err:=CalibrateTwo(a,b)
    if err!=nil || two.Anchors!=2 { t.Fatalf("two: %+v / %v",two,err) }
    p,err:=two.Apply(2500,3600,Surface,90)
    if err!=nil || math.Abs(p.X-3600)>0.001 || math.Abs(p.Y-4700)>0.001 { t.Fatalf("apply: %+v / %v",p,err) }
    if _,err=two.Apply(2500,3600,Underground,90); err==nil {t.Fatal("expected per-layer calibration guard")}
    b.Layer=Shadow
    if _,err=CalibrateTwo(a,b);err==nil {t.Fatal("expected layer mismatch error")}
    b.Layer=Surface; b.NativeX=2010
    if _,err=CalibrateTwo(a,b);err==nil {t.Fatal("expected insufficient separation error")}
}
