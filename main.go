package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_POPUP            = 0x80000000
	WS_EX_TOPMOST       = 0x00000008
	WS_EX_TOOLWINDOW    = 0x00000080
	BS_PUSHBUTTON       = 0
	WM_DESTROY          = 2
	WM_PAINT            = 15
	WM_ERASEBKGND       = 20
	WM_SETICON          = 0x0080
	ICON_BIG            = 1
	ICON_SMALL          = 0
	IMAGE_ICON          = 1
	LR_LOADFROMFILE     = 0x0010
	CF_DIB              = 8
	CF_HDROP            = 15
	GMEM_MOVEABLE       = 0x0002
	VK_C                = 0x43
	VK_S                = 0x53
	WM_COMMAND          = 273
	WM_KEYDOWN          = 256
	WM_LBUTTONDOWN      = 513
	WM_MOUSEMOVE        = 512
	WM_MOUSEWHEEL       = 0x020A
	WM_SIZE             = 5
	VK_CONTROL          = 0x11
	VK_SHIFT            = 0x10
	WM_LBUTTONUP        = 514
	WM_RBUTTONDOWN      = 0x0204
	VK_DELETE           = 0x2E
	VK_Z                = 0x5A
	WM_APP              = 0x8000
	SW_SHOW             = 5
	SW_MINIMIZE         = 6
	SW_RESTORE          = 9
	VK_ESCAPE           = 0x1B
	VK_RETURN           = 0x0D
	WM_SETFONT          = 0x0030
	SRCCOPY             = 0x00CC0020
	DIB_RGB_COLORS      = 0
	BI_RGB              = 0
	IDC_ARROW           = 32512
	IDC_CROSS           = 32515
	COLOR_WINDOW        = 5
	TRANSPARENT         = 1
	SM_CXSCREEN         = 0
	SM_CYSCREEN         = 1
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	gdi32                   = syscall.NewLazyDLL("gdi32.dll")
	pRegisterClass          = user32.NewProc("RegisterClassW")
	pCreateWindow           = user32.NewProc("CreateWindowExW")
	pDefWindow              = user32.NewProc("DefWindowProcW")
	pShowWindow             = user32.NewProc("ShowWindow")
	pUpdateWindow           = user32.NewProc("UpdateWindow")
	pGetMessage             = user32.NewProc("GetMessageW")
	pTranslate              = user32.NewProc("TranslateMessage")
	pDispatch               = user32.NewProc("DispatchMessageW")
	pPostQuit               = user32.NewProc("PostQuitMessage")
	pGetCursorPos           = user32.NewProc("GetCursorPos")
	pSetWindowText          = user32.NewProc("SetWindowTextW")
	pPostMessage            = user32.NewProc("PostMessageW")
	pInvalidateRect         = user32.NewProc("InvalidateRect")
	pBeginPaint             = user32.NewProc("BeginPaint")
	pEndPaint               = user32.NewProc("EndPaint")
	pGetClientRect          = user32.NewProc("GetClientRect")
	pMessageBox             = user32.NewProc("MessageBoxW")
	pLoadCursor             = user32.NewProc("LoadCursorW")
	pSetCapture             = user32.NewProc("SetCapture")
	pReleaseCapture         = user32.NewProc("ReleaseCapture")
	pGetDC                  = user32.NewProc("GetDC")
	pReleaseDC              = user32.NewProc("ReleaseDC")
	pDestroyWindow          = user32.NewProc("DestroyWindow")
	pGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
	pGetKeyState            = user32.NewProc("GetKeyState")
	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pGetDIBits              = gdi32.NewProc("GetDIBits")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pStretchDIBits          = gdi32.NewProc("StretchDIBits")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pTextOut                = gdi32.NewProc("TextOutW")
	pFillRect               = user32.NewProc("FillRect")
	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pSetTextColor           = gdi32.NewProc("SetTextColor")
	pCreateFont             = gdi32.NewProc("CreateFontW")
	pSetFocus               = user32.NewProc("SetFocus")
	pLoadImage              = user32.NewProc("LoadImageW")
	pSendMessage            = user32.NewProc("SendMessageW")
	pOpenClipboard          = user32.NewProc("OpenClipboard")
	pEmptyClipboard         = user32.NewProc("EmptyClipboard")
	pSetClipboardData       = user32.NewProc("SetClipboardData")
	pCloseClipboard         = user32.NewProc("CloseClipboard")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	pGlobalAlloc            = kernel32.NewProc("GlobalAlloc")
	pGlobalLock             = kernel32.NewProc("GlobalLock")
	pGlobalUnlock           = kernel32.NewProc("GlobalUnlock")
	pGlobalFree             = kernel32.NewProc("GlobalFree")
	pRtlMoveMemory          = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlMoveMemory")

	mainWnd                      uintptr
	rawPaintPath                 string
	appIcon                      uintptr
	languageIndex                int // English, Russian, Uzbek new Latin
	mainButtons                  = map[int]uintptr{}
	editorWnd                    uintptr
	statusWnd                    uintptr
	selectWnd                    uintptr
	stopWnd                      uintptr
	mu                           sync.Mutex
	selected                     image.Rectangle
	capturing                    bool
	frames                       []*image.RGBA
	offsets                      []image.Point
	result                       *image.RGBA
	masks                        []image.Rectangle
	selectedMask                 = -1
	editorMode                   = 0 // 0 mask, 1 select, 2 crop
	cropSelection                image.Rectangle
	cropSnapshot                 *image.RGBA
	cropMasksSnapshot            []image.Rectangle
	moveMask                     = false
	moveMaskOrigin               image.Rectangle
	moveStart                    image.Point
	editorFont                   uintptr
	dragging                     bool
	dragStart                    image.Point
	dragEnd                      image.Point
	editorZoom                   float64
	editorScrollX, editorScrollY int
	overlayImage                 *image.RGBA
	selDragging                  bool
	selStart                     image.Point
	selEnd                       image.Point
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type wndclass struct {
	Style                                                   uint32
	WndProc                                                 uintptr
	ClsExtra, WndExtra                                      int32
	Instance, Icon, Cursor, Background, MenuName, ClassName uintptr
}
type msg struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	Private        uint32
}
type paint struct {
	HDC                uintptr
	Erase              int32
	RcPaint            rect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type bitmapinfoheader struct {
	Size                         uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, SizeImage       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ClrUsed, ClrImportant        uint32
}
type bitmapinfo struct {
	Header bitmapinfoheader
	Colors [1]uint32
}

func str(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func message(s string) {
	pMessageBox.Call(mainWnd, uintptr(unsafe.Pointer(str(s))), uintptr(unsafe.Pointer(str("ScrollScreen 1.2"))), 0)
}
func setStatus(s string) { pSetWindowText.Call(statusWnd, uintptr(unsafe.Pointer(str(s)))) }
func low(w uintptr) int  { return int(uint16(w)) }
func xy(v uintptr) image.Point {
	return image.Pt(int(int16(uint16(v))), int(int16(uint16(v>>16))))
}
func windowEx(exStyle uint32, class, title string, style uint32, x, y, w, h int, parent uintptr) uintptr {
	r, _, _ := pCreateWindow.Call(uintptr(exStyle), uintptr(unsafe.Pointer(str(class))), uintptr(unsafe.Pointer(str(title))), uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, 0, 0, 0)
	return r
}
func window(class, title string, x, y, w, h int, parent uintptr) uintptr {
	return windowEx(0, class, title, WS_OVERLAPPEDWINDOW|WS_VISIBLE, x, y, w, h, parent)
}
func button(parent uintptr, label string, id, x, y, w, h int) {
	handle, _, _ := pCreateWindow.Call(0, uintptr(unsafe.Pointer(str("BUTTON"))), uintptr(unsafe.Pointer(str(label))), WS_VISIBLE|WS_CHILD|BS_PUSHBUTTON, uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), 0, 0)
	if parent == mainWnd && handle != 0 {
		mainButtons[id] = handle
	}
}
func getMouse() image.Point {
	var p point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return image.Pt(int(p.X), int(p.Y))
}
func screenShot(r image.Rectangle) *image.RGBA {
	w, h := r.Dx(), r.Dy()
	if w < 1 || h < 1 {
		return nil
	}
	dc, _, _ := pGetDC.Call(0)
	defer pReleaseDC.Call(0, dc)
	mem, _, _ := pCreateCompatibleDC.Call(dc)
	defer pDeleteDC.Call(mem)
	bmp, _, _ := pCreateCompatibleBitmap.Call(dc, uintptr(w), uintptr(h))
	if bmp == 0 {
		return nil
	}
	defer pDeleteObject.Call(bmp)
	old, _, _ := pSelectObject.Call(mem, bmp)
	defer pSelectObject.Call(mem, old)
	pBitBlt.Call(mem, 0, 0, uintptr(w), uintptr(h), dc, uintptr(int32(r.Min.X)), uintptr(int32(r.Min.Y)), SRCCOPY)
	info := bitmapinfo{Header: bitmapinfoheader{Size: 40, Width: int32(w), Height: -int32(h), Planes: 1, BitCount: 32, Compression: BI_RGB}}
	buf := make([]byte, w*h*4)
	pSelectObject.Call(mem, old)
	rv, _, _ := pGetDIBits.Call(mem, bmp, 0, uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS)
	if rv == 0 {
		return nil
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(buf); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = buf[i+2], buf[i+1], buf[i], 255
	}
	return img
}
func drawRGBA(hdc uintptr, img *image.RGBA, dst image.Rectangle) {
	if img == nil || dst.Empty() {
		return
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w < 1 || h < 1 {
		return
	}
	buf := make([]byte, w*h*4)
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = img.Pix[i+2], img.Pix[i+1], img.Pix[i], 0
	}
	info := bitmapinfo{Header: bitmapinfoheader{Size: 40, Width: int32(w), Height: -int32(h), Planes: 1, BitCount: 32, Compression: BI_RGB}}
	pStretchDIBits.Call(hdc, uintptr(dst.Min.X), uintptr(dst.Min.Y), uintptr(dst.Dx()), uintptr(dst.Dy()), 0, 0, uintptr(w), uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS, SRCCOPY)
}
func grayAt(im *image.RGBA, x, y int) int {
	i := im.PixOffset(x, y)
	return (int(im.Pix[i])*77 + int(im.Pix[i+1])*150 + int(im.Pix[i+2])*29) >> 8
}

// V4 vertical-only matching: robust horizontal strip profiles and a short capture interval.
// Spatial profiles heavily penalize mismatched text lines, while constant backgrounds
// contribute little. Use overlapping rows only; UI-fixed bars are trimmed off.
func profile(im *image.RGBA) [][]int {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	bands := 20
	out := make([][]int, h)
	left := w / 12
	right := w - w/12
	for y := 0; y < h; y++ {
		out[y] = make([]int, bands)
		for k := 0; k < bands; k++ {
			xx := left + (right-left)*(k*2+1)/(2*bands)
			xx = min(w-2, max(1, xx))
			out[y][k] = grayAt(im, xx, y)
		}
	}
	return out
}
func stripeScore(a, b [][]int, shift, step int) (float64, int) {
	h := len(a)
	// remove static header/footer which do not scroll with the content
	pad := max(15, h/16)
	start := max(pad, shift+pad)
	end := h - pad
	if end-start < max(48, h/8) {
		return 1e9, 0
	}
	sum := 0.0
	n := 0
	for y := start; y < end; y += step {
		for k := 0; k < len(a[y]); k++ {
			av := a[y][k]
			bv := b[y-shift][k]
			// both pixels and local contrast help distinguish repeated Splunk boxes
			d := abs(av - bv)
			if y+2 < end {
				ad := abs(av - a[y+2][k])
				bd := abs(bv - b[y-shift+2][k])
				d += abs(ad-bd) / 2
			}
			sum += float64(d)
			n++
		}
	}
	if n == 0 {
		return 1e9, 0
	}
	return sum / float64(n), n
}
func verticalAlign(a, b *image.RGBA) (int, float64, bool, string) {
	if a.Bounds().Size() != b.Bounds().Size() {
		return 0, 1e9, false, "size_mismatch"
	}
	h := a.Bounds().Dy()
	maxShift := h * 85 / 100
	pa, pb := profile(a), profile(b)
	bestScore := 1e9
	bestShift := 0
	// Each candidate uses only its true overlapping content. A shorter overlap
	// must have enough structure to be accepted, never blindly append a frame.
	for shift := 0; shift <= maxShift; shift += 4 {
		sc, _ := stripeScore(pa, pb, shift, 4)
		if sc < bestScore {
			bestScore, bestShift = sc, shift
		}
	}
	for shift := max(0, bestShift-7); shift <= min(maxShift, bestShift+7); shift++ {
		sc, _ := stripeScore(pa, pb, shift, 2)
		if sc < bestScore {
			bestScore, bestShift = sc, shift
		}
	}
	if bestShift < 5 {
		return bestShift, bestScore, false, "duplicate_or_no_scroll"
	}
	score, _ := stripeScore(pa, pb, bestShift, 1)
	// stricter than a loose visual match; threshold depends on the strip score,
	// and is intended to protect from repeating interface rows.
	if score > 31 {
		return bestShift, score, false, "low_confidence_or_gap"
	}
	// Require the best alignment to be distinguishable from unrelated offsets.
	second := 1e9
	for shift := 0; shift <= maxShift; shift += 8 {
		if abs(shift-bestShift) <= 18 {
			continue
		}
		sc, _ := stripeScore(pa, pb, shift, 4)
		if sc < second {
			second = sc
		}
	}
	if second < score+1.5 {
		return bestShift, score, false, "ambiguous_overlap"
	}
	return bestShift, score, true, "accepted"
}

var capturesSeen, capturesRejected int

func addFrame() {
	mu.Lock()
	r := selected
	mu.Unlock()
	im := screenShot(r)
	if im == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if !capturing {
		return
	}
	capturesSeen++
	// Keep every captured frame. Offline V8 alignment compares adjacent frames,
	// so rejecting a frame now could destroy the motion chain.
	frames = append(frames, im)
	pPostMessage.Call(mainWnd, WM_APP+1, 0, 0)
}
func stitchV8Delta(a, b *image.RGBA, top, bottom int) (int, float64) {
	w := a.Bounds().Dx()
	h := a.Bounds().Dy()
	center := h - top - bottom
	if center < 80 || w < 50 {
		return 0, 1e9
	}
	maxD := min(center-35, 470)
	// Match sequential screen frames only in the moving middle region.
	best := math.MaxFloat64
	bestD := 0
	for dy := 0; dy < maxD; dy++ {
		var sum float64
		var n int
		for y := top + dy; y < h-bottom; y += 3 {
			for x := 18; x < w-18; x += 3 {
				va := grayAt(a, x, y)
				vb := grayAt(b, x, y-dy)
				sum += math.Abs(float64(va - vb))
				n++
			}
		}
		if n == 0 {
			continue
		}
		score := sum/float64(n) + 0.004*float64(dy)
		if score < best {
			best = score
			bestD = dy
		}
	}
	return bestD, best
}
func beginSelectArea() {
	sw, _, _ := pGetSystemMetrics.Call(SM_CXSCREEN)
	sh, _, _ := pGetSystemMetrics.Call(SM_CYSCREEN)
	overlayImage = screenShot(image.Rect(0, 0, int(sw), int(sh)))
	if overlayImage == nil {
		message("Could not capture the screen for region selection.")
		return
	}
	selDragging = false
	selStart, selEnd = image.Point{}, image.Point{}
	pShowWindow.Call(mainWnd, SW_MINIMIZE)
	selectWnd = windowEx(WS_EX_TOPMOST|WS_EX_TOOLWINDOW, "PanoramaSelector", "", WS_POPUP|WS_VISIBLE, 0, 0, int(sw), int(sh), 0)
	pShowWindow.Call(selectWnd, SW_SHOW)
	pUpdateWindow.Call(selectWnd)
}
func start() {
	mu.Lock()
	if selected.Empty() {
		mu.Unlock()
		message("Click 'Select area' first, then drag a rectangle like in Snagit.")
		return
	}
	frames = nil
	offsets = nil
	capturesSeen, capturesRejected = 0, 0
	// Privacy: do not write diagnostic logs or unredacted capture frames to disk.
	capturing = true
	mu.Unlock()
	setStatus("CAPTURING: scroll manually. Use the small 'Stop && assemble' window when done.")
	pShowWindow.Call(mainWnd, SW_MINIMIZE)
	if stopWnd == 0 {
		stopWnd = windowEx(WS_EX_TOPMOST|WS_EX_TOOLWINDOW, "PanoramaStop", "ScrollScreen", WS_OVERLAPPEDWINDOW|WS_VISIBLE, 30, 30, 280, 120, 0)
		button(stopWnd, "Stop && assemble", 301, 30, 20, 200, 35)
	}
	pShowWindow.Call(stopWnd, SW_SHOW)
	pUpdateWindow.Call(stopWnd)
	go func() {
		runtime.LockOSThread()
		time.Sleep(350 * time.Millisecond)
		addFrame()
		tick := time.NewTicker(300 * time.Millisecond)
		defer tick.Stop()
		for range tick.C {
			mu.Lock()
			on := capturing
			mu.Unlock()
			if !on {
				return
			}
			addFrame()
		}
	}()
}
func assemble() {
	mu.Lock()
	capturing = false
	fs := append([]*image.RGBA(nil), frames...)
	mu.Unlock()
	if stopWnd != 0 {
		pDestroyWindow.Call(stopWnd)
		stopWnd = 0
	}
	pShowWindow.Call(mainWnd, SW_RESTORE)
	if len(fs) == 0 {
		message("No frames captured.")
		return
	}
	w, h := fs[0].Bounds().Dx(), fs[0].Bounds().Dy()
	// Calibrated Splunk profile from 50 real captures: 100 px top, 186 px bottom
	// for a 887 px selection. Proportionally adjusted for different selections.
	top := 100 * h / 887
	bottom := 186 * h / 887
	if h-top-bottom < 100 {
		top = 0
		bottom = 0
	}
	center := h - top - bottom
	shifts := make([]int, len(fs))
	scores := make([]float64, len(fs))
	total := 0
	for i := 1; i < len(fs); i++ {
		d, score := stitchV8Delta(fs[i-1], fs[i], top, bottom)
		if d < 0 {
			d = 0
		}
		shifts[i] = d
		scores[i] = score
		total += d
	}
	height := h + total
	if height > 16000 || int64(height)*int64(w) > 150000000 {
		message("Panorama too large.")
		return
	}
	out := image.NewRGBA(image.Rect(0, 0, w, height))
	draw.Draw(out, image.Rect(0, 0, w, top), fs[0], image.Pt(0, 0), draw.Src)
	draw.Draw(out, image.Rect(0, top, w, top+center), fs[0], image.Pt(0, top), draw.Src)
	write := top + center
	for i := 1; i < len(fs); i++ {
		d := shifts[i]
		if d == 0 {
			continue
		}
		src := image.Rect(0, h-bottom-d, w, h-bottom)
		draw.Draw(out, image.Rect(0, write, w, write+d), fs[i], src.Min, draw.Src)
		write += d
	}
	if bottom > 0 {
		draw.Draw(out, image.Rect(0, write, w, write+bottom), fs[len(fs)-1], image.Pt(0, h-bottom), draw.Src)
	}
	result = out
	masks = nil
	selectedMask = -1
	setStatus(fmt.Sprintf("Scroll screenshot assembled: %d frames, %dx%d. Editor opened locally; original remains in memory.", len(fs), w, height))
	openInternalEditor()
}

// Mask rectangles remain separate until a new redacted PNG is explicitly saved.
// No unredacted panorama is written to disk during ordinary editing.
func openInternalEditor() {
	if result == nil {
		message("Capture a scroll screenshot first.")
		return
	}
	if editorWnd != 0 {
		pShowWindow.Call(editorWnd, SW_RESTORE)
		pInvalidateRect.Call(editorWnd, 0, 0)
		pSetFocus.Call(editorWnd)
		return
	}
	masks = nil
	selectedMask = -1
	editorMode = 0
	cropSelection = image.Rectangle{}
	cropSnapshot = nil
	cropMasksSnapshot = nil
	dragging = false
	moveMask = false
	editorWnd = window("PanoramaEditor", "ScrollScreen 1.2 | "+tr("Open editor"), 30, 45, 1560, 880, 0)
	editorZoom = 1
	editorScrollX = 0
	editorScrollY = 0
	// Open at approximately 100% so text is legible rather than a tiny full-page preview.
	fit := editorScale()
	if fit > 0 {
		editorZoom = 1 / fit
	}
	clampEditorScroll()
	pShowWindow.Call(editorWnd, SW_SHOW)
	pUpdateWindow.Call(editorWnd)
	pSetFocus.Call(editorWnd)
}

// The editor uses a viewport into the original PNG. Zoom never changes saved pixels.
func editorViewport() image.Rectangle {
	var rc rect
	pGetClientRect.Call(editorWnd, uintptr(unsafe.Pointer(&rc)))
	return image.Rect(0, 84, int(rc.Right), max(85, int(rc.Bottom)-38))
}
func editorScale() float64 {
	if result == nil {
		return 1
	}
	v := editorViewport()
	if v.Empty() {
		return 1
	}
	fit := math.Min(float64(v.Dx())/float64(result.Bounds().Dx()), float64(v.Dy())/float64(result.Bounds().Dy()))
	if fit <= 0 {
		fit = 1
	}
	return fit * editorZoom
}
func clampEditorScroll() {
	if result == nil {
		return
	}
	v := editorViewport()
	z := editorScale()
	maxX := max(0, int(math.Ceil(float64(result.Bounds().Dx())*z))-v.Dx())
	maxY := max(0, int(math.Ceil(float64(result.Bounds().Dy())*z))-v.Dy())
	editorScrollX = max(0, min(editorScrollX, maxX))
	editorScrollY = max(0, min(editorScrollY, maxY))
}
func editorImagePoint(p image.Point) image.Point {
	if result == nil {
		return image.Point{}
	}
	z := editorScale()
	if z <= 0 {
		return image.Point{}
	}
	v := editorViewport()
	x := int(math.Floor(float64(p.X+editorScrollX-v.Min.X) / z))
	y := int(math.Floor(float64(p.Y+editorScrollY-v.Min.Y) / z))
	return image.Pt(max(0, min(result.Bounds().Dx(), x)), max(0, min(result.Bounds().Dy(), y)))
}
func setEditorZoom(zoom float64) {
	if result == nil {
		return
	}
	old := editorScale()
	v := editorViewport()
	cx, cy := float64(v.Dx())/2, float64(v.Min.Y+v.Dy()/2)
	ix := (cx + float64(editorScrollX)) / old
	iy := (cy - float64(v.Min.Y) + float64(editorScrollY)) / old
	editorZoom = math.Max(0.1, math.Min(100, zoom))
	now := editorScale()
	editorScrollX = int(ix*now - cx)
	editorScrollY = int(iy*now - (cy - float64(v.Min.Y)))
	clampEditorScroll()
	pInvalidateRect.Call(editorWnd, 0, 1)
}

// The rendered image is the ONLY version sent to the clipboard or written to PNG.
var translations = map[string][3]string{
	"Instructions":                    {"Instructions", "Инструкция", "Yöriqnoma"},
	"Select area":                     {"Select area", "Выбрать область", "Hududni tanlaş"},
	"Start capture":                   {"Start capture", "Начать захват", "Tasvirga oliş"},
	"Open editor":                     {"Open editor", "Открыть редактор", "Tahrirchini oçiş"},
	"About / О программе":             {"About", "О программе", "Dastur haqida"},
	"Crop":                            {"Crop", "Обрезка", "Kesish"},
	"Apply":                           {"Apply", "Применить", "Qoʻllash"},
	"Cancel":                          {"Cancel", "Отмена", "Bekor qilish"},
	"Mask":                            {"Mask", "Маска", "Niqob"},
	"Select":                          {"Select", "Выбор", "Tanlaş"},
	"Undo":                            {"Undo", "Отменить", "Bekor qiliş"},
	"Delete":                          {"Delete", "Удалить", "Öçiriş"},
	"Fit":                             {"Fit", "Вписать", "Siğdiriş"},
	"Copy":                            {"Copy", "Копировать", "Nusxalaş"},
	"Copy ZIP":                        {"Copy ZIP", "Копировать ZIP", "ZIP nusxalaş"},
	"Save PNG":                        {"Save PNG", "Сохранить PNG", "PNG saqlaş"},
	"Red":                             {"Red", "Красный", "Qizil"},
	"Mask: drag to cover":             {"Mask: drag to cover", "Маска: выделите область", "Niqob: hududni belgilaş"},
	"Crop: drag, then Enter to apply": {"Crop: drag, then Enter to apply", "Обрезка: выделите, затем Enter", "Kesish: tanlang, soʻng Enter"},
	"Select: move or delete masks":    {"Select: move or delete masks", "Выбор: переместить или удалить", "Tanlaş: suriş yoki öçiriş"},
}

func tr(k string) string {
	if v, ok := translations[k]; ok {
		return v[languageIndex]
	}
	return k
}
func languageName() string {
	return []string{"English", "Русский", "Özbek (Ğ Ö Ş Ç)"}[languageIndex]
}
func refreshLanguage() {
	for id, k := range map[int]string{101: "Instructions", 102: "Start capture", 103: "Select area", 105: "Open editor", 106: "About / О программе"} {
		if h := mainButtons[id]; h != 0 {
			glyph := map[int]string{101: "☰", 102: "●", 103: "▣", 105: "✎", 106: "ⓘ"}[id]
			pSetWindowText.Call(h, uintptr(unsafe.Pointer(str(glyph+"  "+tr(k)))))
		}
	}
	if h := mainButtons[107]; h != 0 {
		pSetWindowText.Call(h, uintptr(unsafe.Pointer(str("🌐  "+languageName()))))
	}
	if mainWnd != 0 {
		pSetWindowText.Call(mainWnd, uintptr(unsafe.Pointer(str("ScrollScreen 1.2"))))
	}
	if editorWnd != 0 {
		pSetWindowText.Call(editorWnd, uintptr(unsafe.Pointer(str("ScrollScreen 1.2 | "+tr("Open editor")))))
		pInvalidateRect.Call(editorWnd, 0, 0)
	}
}

// Windows owns HGLOBAL after SetClipboardData succeeds. Copy into the locked
// Windows allocation through RtlMoveMemory rather than constructing a Go
// pointer from the uintptr returned by GlobalLock.
func writeGlobalMemory(mem uintptr, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	ptr, _, _ := pGlobalLock.Call(mem)
	if ptr == 0 {
		return false
	}
	pRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	runtime.KeepAlive(data)
	pGlobalUnlock.Call(mem)
	return true
}

// Windows CF_HDROP: 20-byte DROPFILES header, UTF-16 path, double terminator.
func dropfilesBytes(path string) []byte {
	utf := syscall.StringToUTF16(path)
	b := make([]byte, 20+(len(utf)+1)*2)
	binary.LittleEndian.PutUint32(b[0:4], 20)
	binary.LittleEndian.PutUint32(b[16:20], 1) // fWide
	for i, v := range utf {
		binary.LittleEndian.PutUint16(b[20+i*2:22+i*2], v)
	}
	return b
}

// Windows CF_DIB: BITMAPINFOHEADER followed by bottom-up BGRA pixels.
// BI_RGB 32-bit DIB expects reserved alpha bytes (set to zero).
func dibBytes(out *image.RGBA) []byte {
	w, h := out.Bounds().Dx(), out.Bounds().Dy()
	b := make([]byte, 40+w*h*4)
	binary.LittleEndian.PutUint32(b[0:4], 40)
	binary.LittleEndian.PutUint32(b[4:8], uint32(w))
	binary.LittleEndian.PutUint32(b[8:12], uint32(h))
	binary.LittleEndian.PutUint16(b[12:14], 1)
	binary.LittleEndian.PutUint16(b[14:16], 32)
	binary.LittleEndian.PutUint32(b[20:24], uint32(w*h*4))
	for y := 0; y < h; y++ {
		sourceY := out.Bounds().Min.Y + h - 1 - y
		for x := 0; x < w; x++ {
			i := out.PixOffset(out.Bounds().Min.X+x, sourceY)
			j := 40 + (y*w+x)*4
			b[j], b[j+1], b[j+2], b[j+3] = out.Pix[i+2], out.Pix[i+1], out.Pix[i], 0
		}
	}
	return b
}

func copyZIP() {
	img := redactedImage()
	if img == nil {
		message("No screenshot to copy.")
		return
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		message("PNG encoding failed: " + err.Error())
		return
	}
	stamp := time.Now().Format("20060102_150405")
	folder := filepath.Join(os.TempDir(), "ScrollScreen_1_1")
	if err := os.MkdirAll(folder, 0700); err != nil {
		message("Cannot create temp folder: " + err.Error())
		return
	}
	zipPath := filepath.Join(folder, "ScrollScreen_Redacted_"+stamp+".zip")
	f, err := os.Create(zipPath)
	if err != nil {
		message("Cannot create ZIP: " + err.Error())
		return
	}
	zw := zip.NewWriter(f)
	entry, err := zw.Create("ScrollScreen_Redacted_" + stamp + ".png")
	if err == nil {
		_, err = entry.Write(pngBuf.Bytes())
	}
	ze := zw.Close()
	ce := f.Close()
	if err == nil {
		err = ze
	}
	if err == nil {
		err = ce
	}
	if err != nil {
		os.Remove(zipPath)
		message("ZIP error: " + err.Error())
		return
	}
	// CF_HDROP contains a DROPFILES header followed by UTF-16 absolute paths and a double terminator.
	buf := dropfilesBytes(zipPath)
	mem, _, _ := pGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(len(buf)))
	if mem == 0 {
		message("ZIP saved at:\n" + zipPath + "\n\nClipboard allocation failed.")
		return
	}
	if !writeGlobalMemory(mem, buf) {
		pGlobalFree.Call(mem)
		message("ZIP saved at:\n" + zipPath + "\nClipboard lock failed.")
		return
	}
	ok, _, _ := pOpenClipboard.Call(editorWnd)
	if ok == 0 {
		pGlobalFree.Call(mem)
		message("ZIP saved at:\n" + zipPath + "\nClipboard is busy.")
		return
	}
	emptied, _, _ := pEmptyClipboard.Call()
	if emptied == 0 {
		pCloseClipboard.Call()
		pGlobalFree.Call(mem)
		message("ZIP saved at:\n" + zipPath + "\nClipboard cannot be cleared.")
		return
	}
	copied, _, _ := pSetClipboardData.Call(CF_HDROP, mem)
	pCloseClipboard.Call()
	if copied == 0 {
		pGlobalFree.Call(mem)
		message("ZIP saved at:\n" + zipPath + "\nClipboard file copy failed.")
		return
	}
	message("Only the REDACTED PNG was added to ZIP.\n\nZIP file copied to Windows clipboard. Try Ctrl+V in ChatGPT. Some browsers do not support pasting ZIP files.\n\nIf Ctrl+V fails, attach this file manually:\n" + zipPath + "\n\nNOTE: the ZIP remains in your TEMP folder until you delete it.")
}

func redactedImage() *image.RGBA {
	if result == nil {
		return nil
	}
	out := image.NewRGBA(result.Bounds())
	draw.Draw(out, out.Bounds(), result, image.Point{}, draw.Src)
	for _, r := range masks {
		draw.Draw(out, r.Intersect(out.Bounds()), &image.Uniform{C: color.RGBA{R: 0xE5, G: 0x1B, B: 0x23, A: 255}}, image.Point{}, draw.Src)
	}
	return out
}

// Put CF_DIB (a bottom-up 32-bit Windows bitmap) on the clipboard.
// Clipboard takes ownership of the HGLOBAL on successful SetClipboardData.
func copyRedacted() {
	out := redactedImage()
	if out == nil {
		message("No screenshot to copy.")
		return
	}
	w, h := out.Bounds().Dx(), out.Bounds().Dy()
	if w <= 0 || h <= 0 || int64(w)*int64(h) > 150000000 {
		message("Image is too large.")
		return
	}
	buf := dibBytes(out)
	mem, _, _ := pGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(len(buf)))
	if mem == 0 {
		message("Clipboard allocation failed.")
		return
	}
	if !writeGlobalMemory(mem, buf) {
		pGlobalFree.Call(mem)
		message("Clipboard memory lock failed.")
		return
	}
	ok, _, _ := pOpenClipboard.Call(editorWnd)
	if ok == 0 {
		pGlobalFree.Call(mem)
		message("Cannot open clipboard. Close other clipboard applications and try again.")
		return
	}
	empty, _, _ := pEmptyClipboard.Call()
	if empty == 0 {
		pCloseClipboard.Call()
		pGlobalFree.Call(mem)
		message("Cannot clear clipboard.")
		return
	}
	copied, _, _ := pSetClipboardData.Call(CF_DIB, mem)
	pCloseClipboard.Call()
	if copied == 0 {
		pGlobalFree.Call(mem)
		message("Clipboard copy failed.")
		return
	}
	// The unredacted buffer is never placed in the clipboard.
	message("Redacted screenshot copied. Switch to ChatGPT and press Ctrl+V.\n\nReview all masks first: copying does not erase previous clipboard history.")
}
func save() {
	out := redactedImage()
	if out == nil {
		return
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Desktop")
	if st, e := os.Stat(dir); e != nil || !st.IsDir() {
		dir = home
	}
	path := filepath.Join(dir, "ScrollScreen_Redacted_"+time.Now().Format("20060102_150405")+".png")
	f, e := os.Create(path)
	if e != nil {
		message(e.Error())
		return
	}
	e = png.Encode(f, out)
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		message(e.Error())
		return
	}
	message("Redacted PNG saved:\n" + path + "\n\nVerify the entire image before sharing.")
}
func about() {
	switch languageIndex {
	case 1:
		message("ScrollScreen 1.2\n\nДлинные снимки экрана с ручной прокруткой и маскированием данных.\n\nОткрытый проект для создания длинных снимков экрана.\n\nЛокальная обработка. Без автоматической отправки в интернет.")
	case 2:
		message("ScrollScreen 1.2\n\nQöl bilan aylantirib uzun ekran tasvirini oliş va maxfiy ma'lumotlarni niqoblaş.\n\nUzun ekran tasvirlari uchun ochiq manbali loyiha.\n\nMa'lumotlar qurilmada qayta işlanadi. Internetga avtomatik yuborilmaydi.")
	default:
		message("ScrollScreen 1.2\n\nManual scrolling screenshots and privacy redaction.\n\nOpen-source project for scrolling screenshots.\n\nLocal processing. No automatic upload.")
	}
}
func applyIcon(hwnd uintptr) {
	if appIcon != 0 && hwnd != 0 {
		pSendMessage.Call(hwnd, WM_SETICON, ICON_BIG, appIcon)
		pSendMessage.Call(hwnd, WM_SETICON, ICON_SMALL, appIcon)
	}
}
func initIcon() {
	executable, err := os.Executable()
	if err != nil {
		return
	}
	path := filepath.Join(filepath.Dir(executable), "ScrollScreen.ico")
	if _, err := os.Stat(path); err != nil {
		return
	}
	handle, _, _ := pLoadImage.Call(0, uintptr(unsafe.Pointer(str(path))), IMAGE_ICON, 0, 0, LR_LOADFROMFILE)
	appIcon = handle
}
func drawBorder(hdc uintptr, r image.Rectangle, rgb uint32) {
	if r.Empty() {
		return
	}
	if r.Dx() < 2 || r.Dy() < 2 {
		return
	}
	brush, _, _ := pCreateSolidBrush.Call(uintptr(rgb))
	defer pDeleteObject.Call(brush)
	rs := []rect{
		{int32(r.Min.X), int32(r.Min.Y), int32(r.Max.X), int32(r.Min.Y + 2)},
		{int32(r.Min.X), int32(r.Max.Y - 2), int32(r.Max.X), int32(r.Max.Y)},
		{int32(r.Min.X), int32(r.Min.Y), int32(r.Min.X + 2), int32(r.Max.Y)},
		{int32(r.Max.X - 2), int32(r.Min.Y), int32(r.Max.X), int32(r.Max.Y)},
	}
	for _, rr := range rs {
		rr2 := rr
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rr2)), brush)
	}
}
func showSelection(hdc uintptr) {
	if overlayImage != nil {
		drawRGBA(hdc, overlayImage, overlayImage.Bounds())
	}
	pSetBkMode.Call(hdc, TRANSPARENT)
	pSetTextColor.Call(hdc, 0x00FFFFFF)
	txt := "Drag with the mouse to select the capture area. Release to confirm. Esc = cancel"
	t := syscall.StringToUTF16(txt)
	pTextOut.Call(hdc, 12, 12, uintptr(unsafe.Pointer(&t[0])), uintptr(len(t)-1))
	if selDragging {
		r := image.Rect(min(selStart.X, selEnd.X), min(selStart.Y, selEnd.Y), max(selStart.X, selEnd.X), max(selStart.Y, selEnd.Y))
		drawBorder(hdc, r, 0x00FFFF00)
	}
}

// The editor canvas is drawn to an offscreen bitmap, then blitted once to avoid flicker.
func solidRect(hdc uintptr, r image.Rectangle, c uint32) {
	if r.Empty() {
		return
	}
	brush, _, _ := pCreateSolidBrush.Call(uintptr(c))
	defer pDeleteObject.Call(brush)
	rr := rect{int32(r.Min.X), int32(r.Min.Y), int32(r.Max.X), int32(r.Max.Y)}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rr)), brush)
}
func uiText(hdc uintptr, x, y int, label string, c uint32) {
	pSetBkMode.Call(hdc, TRANSPARENT)
	pSetTextColor.Call(hdc, uintptr(c))
	t := syscall.StringToUTF16(label)
	if len(t) > 1 {
		pTextOut.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&t[0])), uintptr(len(t)-1))
	}
}

type toolAction struct {
	name string
	area image.Rectangle
}

func toolbarActions() []toolAction {
	return []toolAction{
		{"Mask", image.Rect(16, 16, 109, 58)},
		{"Select", image.Rect(116, 16, 211, 58)},
		{"Crop", image.Rect(218, 16, 310, 58)},
		{"Undo", image.Rect(317, 16, 402, 58)},
		{"Delete", image.Rect(409, 16, 503, 58)},
		{"Zoom -", image.Rect(518, 16, 582, 58)},
		{"100%", image.Rect(589, 16, 649, 58)},
		{"Zoom +", image.Rect(656, 16, 722, 58)},
		{"Fit", image.Rect(729, 16, 802, 58)},
	}
}
func undoEditor() {
	if editorMode == 2 && !cropSelection.Empty() {
		cropSelection = image.Rectangle{}
	} else if len(masks) > 0 {
		masks = masks[:len(masks)-1]
		selectedMask = -1
	} else if cropSnapshot != nil {
		result = cropSnapshot
		masks = cropMasksSnapshot
		cropSnapshot = nil
		cropMasksSnapshot = nil
		cropSelection = image.Rectangle{}
		editorZoom = 1
		editorScrollX, editorScrollY = 0, 0
	}
}
func applyCrop() {
	if result == nil || cropSelection.Empty() {
		return
	}
	r := cropSelection.Intersect(result.Bounds())
	if r.Dx() < 5 || r.Dy() < 5 {
		return
	}
	// Undo restores the preceding image and its mask geometry.
	cropSnapshot = result
	cropMasksSnapshot = append([]image.Rectangle(nil), masks...)
	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(out, out.Bounds(), result, r.Min, draw.Src)
	newMasks := make([]image.Rectangle, 0, len(masks))
	for _, mask := range masks {
		clipped := mask.Intersect(r)
		if !clipped.Empty() {
			newMasks = append(newMasks, clipped.Sub(r.Min))
		}
	}
	result = out
	masks = newMasks
	selectedMask = -1
	cropSelection = image.Rectangle{}
	editorMode = 0
	editorZoom = 1
	editorScrollX, editorScrollY = 0, 0
	clampEditorScroll()
	pInvalidateRect.Call(editorWnd, 0, 0)
}
func toolbarClick(p image.Point) bool {
	if p.Y > 74 {
		return false
	}
	if editorMode == 2 {
		if p.X >= 822 && p.X < 920 {
			applyCrop()
			return true
		}
		if p.X >= 926 && p.X < 1021 {
			cropSelection = image.Rectangle{}
			editorMode = 0
			pInvalidateRect.Call(editorWnd, 0, 0)
			return true
		}
	}
	for _, t := range toolbarActions() {
		if !p.In(t.area) {
			continue
		}
		switch t.name {
		case "Mask":
			editorMode = 0
			selectedMask = -1
			cropSelection = image.Rectangle{}
		case "Select":
			editorMode = 1
			cropSelection = image.Rectangle{}
		case "Crop":
			editorMode = 2
			selectedMask = -1
		case "Undo":
			undoEditor()
		case "Delete":
			removeSelectedMask()
		case "Zoom -":
			setEditorZoom(editorZoom / 1.25)
		case "Zoom +":
			setEditorZoom(editorZoom * 1.25)
		case "100%":
			fit := editorScale() / editorZoom
			if fit > 0 {
				setEditorZoom(1 / fit)
			}
		case "Fit":
			editorZoom = 1
			editorScrollX, editorScrollY = 0, 0
			clampEditorScroll()
		}
		pInvalidateRect.Call(editorWnd, 0, 0)
		return true
	}
	var rc rect
	pGetClientRect.Call(editorWnd, uintptr(unsafe.Pointer(&rc)))
	width := int(rc.Right)
	if p.X >= width-393 && p.X < width-263 && p.Y >= 15 && p.Y < 59 {
		copyZIP()
		return true
	}
	if p.X >= width-257 && p.X < width-138 && p.Y >= 15 && p.Y < 59 {
		copyRedacted()
		return true
	}
	if p.X >= width-132 && p.X < width-12 && p.Y >= 15 && p.Y < 59 {
		save()
		return true
	}
	return true
}
func removeSelectedMask() {
	if selectedMask < 0 || selectedMask >= len(masks) {
		return
	}
	masks = append(masks[:selectedMask], masks[selectedMask+1:]...)
	selectedMask = -1
}
func findMask(pt image.Point) int {
	for i := len(masks) - 1; i >= 0; i-- {
		if pt.In(masks[i]) {
			return i
		}
	}
	return -1
}
func showImage(hdc uintptr) {
	if result == nil {
		return
	}
	clampEditorScroll()
	v := editorViewport()
	if v.Empty() {
		return
	}
	solidRect(hdc, v, 0x00383037)
	z := editorScale()
	w, h := result.Bounds().Dx(), result.Bounds().Dy()
	sx := max(0, int(float64(editorScrollX)/z))
	sy := max(0, int(float64(editorScrollY)/z))
	ex := min(w, int(math.Ceil(float64(editorScrollX+v.Dx())/z))+1)
	ey := min(h, int(math.Ceil(float64(editorScrollY+v.Dy())/z))+1)
	if sx >= ex || sy >= ey {
		return
	}
	cw, ch := ex-sx, ey-sy
	if int64(cw)*int64(ch) > 25000000 {
		return
	}
	buf := make([]byte, cw*ch*4)
	active := image.Rectangle{}
	if dragging && (editorMode == 0 || editorMode == 2) {
		active = image.Rect(min(dragStart.X, dragEnd.X), min(dragStart.Y, dragEnd.Y), max(dragStart.X, dragEnd.X), max(dragStart.Y, dragEnd.Y))
	}
	// Copy original pixels and apply temporary mask objects only to the preview.
	for yy := 0; yy < ch; yy++ {
		src := result.PixOffset(sx, sy+yy)
		dst := yy * cw * 4
		for xx := 0; xx < cw; xx++ {
			buf[dst+xx*4], buf[dst+xx*4+1], buf[dst+xx*4+2] = result.Pix[src+xx*4+2], result.Pix[src+xx*4+1], result.Pix[src+xx*4]
		}
	}
	all := append([]image.Rectangle(nil), masks...)
	if !active.Empty() && editorMode == 0 {
		all = append(all, active)
	}
	for _, mask := range all {
		clip := mask.Intersect(image.Rect(sx, sy, ex, ey))
		for yy := clip.Min.Y; yy < clip.Max.Y; yy++ {
			row := (yy - sy) * cw * 4
			for xx := clip.Min.X; xx < clip.Max.X; xx++ {
				i := row + (xx-sx)*4
				buf[i], buf[i+1], buf[i+2] = 0x23, 0x1B, 0xE5
			}
		}
	}
	info := bitmapinfo{Header: bitmapinfoheader{Size: 40, Width: int32(cw), Height: -int32(ch), Planes: 1, BitCount: 32, Compression: BI_RGB}}
	px := v.Min.X + int(math.Round(float64(sx)*z)) - editorScrollX
	py := v.Min.Y + int(math.Round(float64(sy)*z)) - editorScrollY
	dw, dh := int(math.Ceil(float64(cw)*z)), int(math.Ceil(float64(ch)*z))
	pStretchDIBits.Call(hdc, uintptr(px), uintptr(py), uintptr(dw), uintptr(dh), 0, 0, uintptr(cw), uintptr(ch), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS, SRCCOPY)
	if editorMode == 2 {
		r := cropSelection
		if dragging {
			r = image.Rect(min(dragStart.X, dragEnd.X), min(dragStart.Y, dragEnd.Y), max(dragStart.X, dragEnd.X), max(dragStart.Y, dragEnd.Y))
		}
		if !r.Empty() {
			x0 := v.Min.X + int(float64(r.Min.X)*z) - editorScrollX
			y0 := v.Min.Y + int(float64(r.Min.Y)*z) - editorScrollY
			x1 := v.Min.X + int(float64(r.Max.X)*z) - editorScrollX
			y1 := v.Min.Y + int(float64(r.Max.Y)*z) - editorScrollY
			drawBorder(hdc, image.Rect(x0, y0, x1, y1), 0x00F2B462)
		}
	}
	if selectedMask >= 0 && selectedMask < len(masks) {
		r := masks[selectedMask]
		x0 := v.Min.X + int(float64(r.Min.X)*z) - editorScrollX
		y0 := v.Min.Y + int(float64(r.Min.Y)*z) - editorScrollY
		x1 := v.Min.X + int(float64(r.Max.X)*z) - editorScrollX
		y1 := v.Min.Y + int(float64(r.Max.Y)*z) - editorScrollY
		drawBorder(hdc, image.Rect(x0, y0, x1, y1), 0x00F0AD19)
	}
}
func toolGlyph(name string) string {
	switch name {
	case "Mask":
		return "▣"
	case "Select":
		return "⌖"
	case "Crop":
		return "✂"
	case "Undo":
		return "↶"
	case "Delete":
		return "×"
	case "Zoom -":
		return "−"
	case "Zoom +":
		return "+"
	case "Fit":
		return "⛶"
	case "Copy ZIP":
		return "▤"
	case "Copy":
		return "⧉"
	case "Save PNG":
		return "▣"
	case "Apply":
		return "✓"
	case "Cancel":
		return "×"
	default:
		return ""
	}
}
func toolButton(hdc uintptr, area image.Rectangle, name string, selected, primary bool) {
	bg := uint32(0x003A3440) // dark graphite (BGR)
	fg := uint32(0x00F5F4F6)
	if selected {
		bg = 0x006D4932
	}
	if primary {
		bg = 0x00BD7640
	}
	solidRect(hdc, area, bg)
	if selected {
		drawBorder(hdc, area, 0x00F6B76C)
	}
	icon := toolGlyph(name)
	if icon != "" {
		uiText(hdc, area.Min.X+9, area.Min.Y+11, icon, fg)
	}
	label := tr(name)
	if name == "Zoom -" || name == "Zoom +" {
		label = ""
	}
	if label != "" {
		uiText(hdc, area.Min.X+28, area.Min.Y+12, label, fg)
	}
}
func showEditor(hdc uintptr) {
	var rc rect
	pGetClientRect.Call(editorWnd, uintptr(unsafe.Pointer(&rc)))
	width, height := int(rc.Right), int(rc.Bottom)
	if width < 1 || height < 1 {
		return
	}
	// Dark toolbar and workbench; original pixels are never changed by theme.
	solidRect(hdc, image.Rect(0, 0, width, height), 0x00241E23)
	solidRect(hdc, image.Rect(0, 0, width, 74), 0x002C2831)
	solidRect(hdc, image.Rect(0, 73, width, 74), 0x00514851)
	oldFont := uintptr(0)
	if editorFont != 0 {
		oldFont, _, _ = pSelectObject.Call(hdc, editorFont)
		defer pSelectObject.Call(hdc, oldFont)
	}
	for _, t := range toolbarActions() {
		if t.area.Max.X > width-405 && width < 1250 {
			continue
		}
		chosen := (t.name == "Mask" && editorMode == 0) || (t.name == "Select" && editorMode == 1) || (t.name == "Crop" && editorMode == 2)
		toolButton(hdc, t.area, t.name, chosen, false)
	}
	if editorMode == 2 && width > 1450 {
		toolButton(hdc, image.Rect(822, 16, 920, 58), "Apply", false, true)
		toolButton(hdc, image.Rect(926, 16, 1021, 58), "Cancel", false, false)
	}
	toolButton(hdc, image.Rect(width-393, 16, width-263, 58), "Copy ZIP", false, false)
	toolButton(hdc, image.Rect(width-257, 16, width-138, 58), "Copy", false, false)
	toolButton(hdc, image.Rect(width-132, 16, width-12, 58), "Save PNG", false, true)
	showImage(hdc)
	solidRect(hdc, image.Rect(0, height-37, width, height), 0x002D2932)
	solidRect(hdc, image.Rect(0, height-38, width, height-37), 0x00514851)
	zoom := int(math.Round(100 * editorScale()))
	mode := "Mask: drag to cover"
	if editorMode == 1 {
		mode = "Select: move or delete masks"
	}
	if editorMode == 2 {
		mode = "Crop: drag, then Enter to apply"
	}
	uiText(hdc, 17, height-27, fmt.Sprintf("%s   |   %d%%   |   %d × %d   |   Masks: %d", tr(mode), zoom, result.Bounds().Dx(), result.Bounds().Dy(), len(masks)), 0x00F3E8E5)
	uiText(hdc, max(20, width-350), height-27, "Ctrl+Wheel: zoom  |  Shift+Wheel: pan", 0x00BDB1B1)
}
func wndProc(hwnd uintptr, umsg uint32, wparam, lparam uintptr) uintptr {
	switch umsg {
	case WM_COMMAND:
		switch hwnd {
		case mainWnd:
			switch low(wparam) {
			case 101:
				message("1. Select area with mouse, then Start capture.\n2. Scroll DOWN manually, then Stop & assemble.\n3. Editor: Mask = draw opaque red rectangles; Select = move or delete masks.\n4. Wheel = vertical scroll; Ctrl+wheel = zoom; Shift+wheel = horizontal scroll.\n5. Copy / Ctrl+C copies REDACTED image to clipboard; Ctrl+V in ChatGPT.\n6. Save PNG / Ctrl+S creates a NEW redacted file on your Desktop.\n\nThe raw panorama stays in memory; no automatic upload.")
			case 102:
				start()
			case 103:
				beginSelectArea()
			case 105:
				openInternalEditor()
			case 106:
				about()
			case 107:
				languageIndex = (languageIndex + 1) % 3
				refreshLanguage()

			}
		case stopWnd:
			if low(wparam) == 301 {
				assemble()
			}
		case editorWnd:
			// Toolbar is drawn directly and handled in mouse events.
		}
	case WM_APP + 1:
		mu.Lock()
		n := len(frames)
		mu.Unlock()
		setStatus(fmt.Sprintf("Capturing... %d accepted frames. Strict vertical mode. Stop && assemble when finished.", n))
	case WM_MOUSEWHEEL:
		if hwnd == editorWnd && result != nil {
			delta := int(int16(uint16(wparam >> 16)))
			ctrl, _, _ := pGetKeyState.Call(VK_CONTROL)
			shift, _, _ := pGetKeyState.Call(VK_SHIFT)
			if ctrl&0x8000 != 0 {
				if delta > 0 {
					setEditorZoom(editorZoom * 1.2)
				} else {
					setEditorZoom(editorZoom / 1.2)
				}
			} else if shift&0x8000 != 0 {
				editorScrollX -= delta / 120 * 100
				clampEditorScroll()
				pInvalidateRect.Call(hwnd, 0, 1)
			} else {
				editorScrollY -= delta / 120 * 120
				clampEditorScroll()
				pInvalidateRect.Call(hwnd, 0, 1)
			}
			return 0
		}
	case WM_SIZE:
		if hwnd == editorWnd {
			clampEditorScroll()
			pInvalidateRect.Call(hwnd, 0, 0)
			return 0
		}
	case WM_KEYDOWN:
		if hwnd == editorWnd {
			if wparam == VK_RETURN && editorMode == 2 {
				applyCrop()
				return 0
			}
			if wparam == VK_ESCAPE && editorMode == 2 {
				cropSelection = image.Rectangle{}
				editorMode = 0
				pInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
			if wparam == VK_DELETE {
				removeSelectedMask()
				pInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
			ctrl, _, _ := pGetKeyState.Call(VK_CONTROL)
			if wparam == VK_C && ctrl&0x8000 != 0 {
				copyRedacted()
				return 0
			}
			if wparam == VK_S && ctrl&0x8000 != 0 {
				save()
				return 0
			}
			if wparam == VK_Z && ctrl&0x8000 != 0 {
				undoEditor()
				pInvalidateRect.Call(hwnd, 0, 0)
				return 0
			}
		}
		if hwnd == selectWnd && wparam == VK_ESCAPE {
			pDestroyWindow.Call(selectWnd)
			selectWnd = 0
			pShowWindow.Call(mainWnd, SW_RESTORE)
			setStatus("Selection canceled.")
			return 0
		}
	case WM_LBUTTONDOWN:
		if hwnd == editorWnd && result != nil {
			p := xy(lparam)
			if p.Y < 74 {
				toolbarClick(p)
				return 0
			}
			if !p.In(editorViewport()) {
				return 0
			}
			ip := editorImagePoint(p)
			if editorMode == 1 {
				selectedMask = findMask(ip)
				if selectedMask >= 0 {
					moveMask = true
					moveStart = ip
					moveMaskOrigin = masks[selectedMask]
					pSetCapture.Call(hwnd)
				}
			} else {
				selectedMask = -1
				dragging = true
				dragStart = ip
				dragEnd = ip
				pSetCapture.Call(hwnd)
			}
			pInvalidateRect.Call(hwnd, 0, 0)
			return 0
		}
		if hwnd == selectWnd {
			selDragging = true
			selStart = xy(lparam)
			selEnd = selStart
			pSetCapture.Call(hwnd)
			pInvalidateRect.Call(hwnd, 0, 1)
			return 0
		}
	case WM_MOUSEMOVE:
		if hwnd == editorWnd && result != nil {
			if dragging {
				dragEnd = editorImagePoint(xy(lparam))
				pInvalidateRect.Call(hwnd, 0, 0)
			}
			if moveMask && selectedMask >= 0 && selectedMask < len(masks) {
				pt := editorImagePoint(xy(lparam))
				delta := pt.Sub(moveStart)
				rr := moveMaskOrigin.Add(delta)
				w, h := result.Bounds().Dx(), result.Bounds().Dy()
				rr = rr.Add(image.Pt(max(0, -rr.Min.X)-max(0, rr.Max.X-w), max(0, -rr.Min.Y)-max(0, rr.Max.Y-h)))
				masks[selectedMask] = rr
				pInvalidateRect.Call(hwnd, 0, 0)
			}
		}
		if hwnd == selectWnd && selDragging {
			selEnd = xy(lparam)
			pInvalidateRect.Call(hwnd, 0, 1)
			return 0
		}
	case WM_LBUTTONUP:
		if hwnd == editorWnd && (dragging || moveMask) {
			if dragging {
				dragEnd = editorImagePoint(xy(lparam))
				r := image.Rect(min(dragStart.X, dragEnd.X), min(dragStart.Y, dragEnd.Y), max(dragStart.X, dragEnd.X), max(dragStart.Y, dragEnd.Y))
				if r.Dx() > 2 && r.Dy() > 2 && editorMode == 2 {
					cropSelection = r
				} else if r.Dx() > 2 && r.Dy() > 2 {
					masks = append(masks, r)
					selectedMask = len(masks) - 1
				}
			}
			dragging = false
			moveMask = false
			pReleaseCapture.Call()
			pInvalidateRect.Call(hwnd, 0, 0)
			return 0
		}
		if hwnd == selectWnd && selDragging {
			selDragging = false
			pReleaseCapture.Call()
			selEnd = xy(lparam)
			selected = image.Rect(min(selStart.X, selEnd.X), min(selStart.Y, selEnd.Y), max(selStart.X, selEnd.X), max(selStart.Y, selEnd.Y))
			if selected.Dx() < 120 || selected.Dy() < 120 {
				selected = image.Rectangle{}
				message("Region too small. Please try again.")
				pInvalidateRect.Call(hwnd, 0, 1)
				return 0
			}
			pDestroyWindow.Call(selectWnd)
			selectWnd = 0
			pShowWindow.Call(mainWnd, SW_RESTORE)
			setStatus(fmt.Sprintf("Region selected: %dx%d. Click Start capture.", selected.Dx(), selected.Dy()))
			return 0
		}
	case WM_ERASEBKGND:
		if hwnd == selectWnd || hwnd == editorWnd {
			return 1
		}
	case WM_RBUTTONDOWN:
		if hwnd == editorWnd && result != nil && editorMode == 1 {
			selectedMask = findMask(editorImagePoint(xy(lparam)))
			removeSelectedMask()
			pInvalidateRect.Call(hwnd, 0, 0)
			return 0
		}
	case WM_PAINT:
		if hwnd == editorWnd {
			var ps paint
			hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			var rc rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
			w, h := int(rc.Right), int(rc.Bottom)
			if w > 0 && h > 0 {
				mem, _, _ := pCreateCompatibleDC.Call(hdc)
				bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
				if mem != 0 && bmp != 0 {
					old, _, _ := pSelectObject.Call(mem, bmp)
					showEditor(mem)
					pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), mem, 0, 0, SRCCOPY)
					pSelectObject.Call(mem, old)
				} else {
					showEditor(hdc)
				}
				if bmp != 0 {
					pDeleteObject.Call(bmp)
				}
				if mem != 0 {
					pDeleteDC.Call(mem)
				}
			}
			pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			return 0
		}
		if hwnd == selectWnd {
			var ps paint
			hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			var rc rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
			w, h := int(rc.Right-rc.Left), int(rc.Bottom-rc.Top)
			if w > 0 && h > 0 {
				mem, _, _ := pCreateCompatibleDC.Call(hdc)
				if mem != 0 {
					bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
					if bmp != 0 {
						old, _, _ := pSelectObject.Call(mem, bmp)
						showSelection(mem)
						pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), mem, 0, 0, SRCCOPY)
						pSelectObject.Call(mem, old)
						pDeleteObject.Call(bmp)
					} else {
						showSelection(hdc)
					}
					pDeleteDC.Call(mem)
				} else {
					showSelection(hdc)
				}
			}
			pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			return 0
		}
	case WM_DESTROY:
		switch hwnd {
		case mainWnd:
			pPostQuit.Call(0)
			return 0
		case editorWnd:
			editorWnd = 0
			return 0
		case stopWnd:
			stopWnd = 0
			return 0
		case selectWnd:
			selectWnd = 0
			return 0
		}
	}
	r, _, _ := pDefWindow.Call(hwnd, uintptr(umsg), wparam, lparam)
	return r
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func main() {
	runtime.LockOSThread()
	user32.NewProc("SetProcessDPIAware").Call()
	initIcon()
	cb := syscall.NewCallback(wndProc)
	cur, _, _ := pLoadCursor.Call(0, IDC_ARROW)
	cls := wndclass{WndProc: cb, Cursor: cur, Background: COLOR_WINDOW + 1, ClassName: uintptr(unsafe.Pointer(str("ScrollScreenNative")))}
	r, _, regErr := pRegisterClass.Call(uintptr(unsafe.Pointer(&cls)))
	if r == 0 {
		message("Window registration failed: " + regErr.Error())
		return
	}
	curCross, _, _ := pLoadCursor.Call(0, IDC_CROSS)
	cls2 := cls
	cls2.ClassName = uintptr(unsafe.Pointer(str("PanoramaEditor")))
	pRegisterClass.Call(uintptr(unsafe.Pointer(&cls2)))
	cls3 := cls
	cls3.ClassName = uintptr(unsafe.Pointer(str("PanoramaSelector")))
	cls3.Cursor = curCross
	cls3.Background = 0 // the overlay paints the complete buffer, no white background erase
	pRegisterClass.Call(uintptr(unsafe.Pointer(&cls3)))
	cls4 := cls
	cls4.ClassName = uintptr(unsafe.Pointer(str("PanoramaStop")))
	pRegisterClass.Call(uintptr(unsafe.Pointer(&cls4)))

	editorFont, _, _ = pCreateFont.Call(uintptr(uint32(0xffffFFEF)), 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 0, 0, uintptr(unsafe.Pointer(str("Segoe UI"))))
	mainWnd = window("ScrollScreenNative", "ScrollScreen 1.2", 180, 150, 760, 315, 0)
	applyIcon(mainWnd)
	if mainWnd == 0 {
		message("Main window could not be created.")
		return
	}
	button(mainWnd, "☰  Instructions", 101, 18, 22, 155, 35)
	button(mainWnd, "▣  Select area", 103, 190, 22, 155, 35)
	button(mainWnd, "●  Start capture", 102, 362, 22, 155, 35)
	button(mainWnd, "✎  Open editor", 105, 18, 70, 155, 35)
	button(mainWnd, "ⓘ  About", 106, 190, 70, 200, 35)
	button(mainWnd, "Language: English", 107, 405, 70, 270, 35)
	statusWnd, _, _ = pCreateWindow.Call(0, uintptr(unsafe.Pointer(str("STATIC"))), uintptr(unsafe.Pointer(str("Select area -> Start capture -> Stop & assemble -> redaction editor. The unredacted panorama stays in memory; save a redacted PNG only."))), WS_VISIBLE|WS_CHILD, 18, 125, 710, 110, mainWnd, 0, 0, 0)
	refreshLanguage()

	var m msg
	for {
		v, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(v) <= 0 {
			break
		}
		pTranslate.Call(uintptr(unsafe.Pointer(&m)))
		pDispatch.Call(uintptr(unsafe.Pointer(&m)))
	}
}
