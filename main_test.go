package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

func TestGrayAt(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})
	img.SetRGBA(1, 0, color.RGBA{255, 255, 255, 255})
	if got := grayAt(img, 0, 0); got != 0 {
		t.Errorf("Black pixel: got %d, want 0", got)
	}
	if got := grayAt(img, 1, 0); got != 255 {
		t.Errorf("White pixel: got %d, want 255", got)
	}
}
func TestProfile(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 120))
	p := profile(img)
	if len(p) != 120 {
		t.Fatalf("Profile height: got %d, want 120", len(p))
	}
	for y, row := range p {
		if len(row) != 20 {
			t.Fatalf("Row %d: got %d bands, want 20", y, len(row))
		}
	}
}
func TestStripeScoreIdentical(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 200; x++ {
			v := uint8((x + y*3) % 256)
			img.SetRGBA(x, y, color.RGBA{v, v, v, 255})
		}
	}
	p := profile(img)
	score, count := stripeScore(p, p, 0, 1)
	if count == 0 {
		t.Fatal("No pixels compared")
	}
	if score != 0 {
		t.Errorf("Identical profiles: score=%f, want 0", score)
	}
}
func TestMathHelpers(t *testing.T) {
	if min(3, 8) != 3 {
		t.Error("min failed")
	}
	if max(3, 8) != 8 {
		t.Error("max failed")
	}
	if abs(-12) != 12 {
		t.Error("abs failed")
	}
}

// Test the CF_HDROP representation without calling the Windows clipboard.
func TestDropfilesBytes(t *testing.T) {
	b := dropfilesBytes(`C:\\Temp\\sample.zip`)
	if got := binary.LittleEndian.Uint32(b[0:4]); got != 20 {
		t.Fatalf("pFiles: %d", got)
	}
	if got := binary.LittleEndian.Uint32(b[16:20]); got != 1 {
		t.Fatalf("fWide: %d", got)
	}
	if b[len(b)-1] != 0 || b[len(b)-2] != 0 || b[len(b)-3] != 0 || b[len(b)-4] != 0 {
		t.Fatal("UTF-16 double terminator missing")
	}
}

// Verify the bottom-up 32-bit DIB channel layout used for copying to the clipboard.
func TestDIBBytes(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	img.SetRGBA(0, 1, color.RGBA{R: 40, G: 50, B: 60, A: 255})
	b := dibBytes(img)
	if len(b) != 48 {
		t.Fatalf("DIB size: %d", len(b))
	}
	if binary.LittleEndian.Uint32(b[0:4]) != 40 || binary.LittleEndian.Uint32(b[4:8]) != 1 || binary.LittleEndian.Uint32(b[8:12]) != 2 {
		t.Fatal("invalid DIB header")
	}
	if b[12] != 1 || b[14] != 32 {
		t.Fatal("invalid planes/bit count")
	}
	if !bytes.Equal(b[40:48], []byte{60, 50, 40, 0, 30, 20, 10, 0}) {
		t.Fatalf("bad DIB pixels: %v", b[40:48])
	}
}
