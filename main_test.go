package main

import (
 "image"
 "image/color"
 "testing"
)

func TestGrayAt(t *testing.T) {
 img := image.NewRGBA(image.Rect(0,0,10,10))
 img.SetRGBA(0,0,color.RGBA{0,0,0,255})
 img.SetRGBA(1,0,color.RGBA{255,255,255,255})
 if got := grayAt(img,0,0); got != 0 { t.Errorf("Black pixel: got %d, want 0", got) }
 if got := grayAt(img,1,0); got != 255 { t.Errorf("White pixel: got %d, want 255", got) }
}
func TestProfile(t *testing.T) {
 img := image.NewRGBA(image.Rect(0,0,200,120))
 p := profile(img)
 if len(p) != 120 { t.Fatalf("Profile height: got %d, want 120",len(p)) }
 for y,row := range p { if len(row) != 20 { t.Fatalf("Row %d: got %d bands, want 20",y,len(row)) } }
}
func TestStripeScoreIdentical(t *testing.T) {
 img := image.NewRGBA(image.Rect(0,0,200,120))
 for y:=0;y<120;y++ { for x:=0;x<200;x++ { v:=uint8((x+y*3)%256); img.SetRGBA(x,y,color.RGBA{v,v,v,255}) } }
 p:=profile(img)
 score,count:=stripeScore(p,p,0,1)
 if count==0 { t.Fatal("No pixels compared") }
 if score!=0 { t.Errorf("Identical profiles: score=%f, want 0",score) }
}
func TestMathHelpers(t *testing.T) {
 if min(3,8)!=3 {t.Error("min failed")}
 if max(3,8)!=8 {t.Error("max failed")}
 if abs(-12)!=12 {t.Error("abs failed")}
}
