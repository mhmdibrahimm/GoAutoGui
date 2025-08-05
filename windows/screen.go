//go:build windows

package windows

import (
	"errors"
	"image"
	"syscall"
	"unsafe"

	"github.com/zzl/go-win32api/v2/win32"
)

// Point represents an x, y coordinate pair
type POINT struct {
	X, Y int
}

// Lerp returns the point at fraction t along the line from (x1, y1) to (x2, y2).
// t specifies how far along the line the returned point is (t=0 yields (x1, y1),
// t=1 yields (x2, y2)).
func Lerp(x1, y1, x2, y2, t float64) (x, y float64) {
	x = x1 + ((x2 - x1) * t)
	y = y1 + ((y2 - y1) * t)
	return
}

// CursorPos returns the current mouse cursor coordinates (x, y) in pixels.
// On Windows, it calls the Win32 GetCursorPos API under the hood.
func Position() POINT {
	cursor := win32.POINT{}
	win32.GetCursorPos(&cursor)
	return POINT{X: int(cursor.X), Y: int(cursor.Y)}
}

// GetVirtualScreenOffset returns the top‑left origin of the virtual desktop relative
// to the primary monitor. This is often negative if you’ve put monitors to the left/top.
func GetVirtualScreenOffset() POINT {
	ox := win32.GetSystemMetrics(win32.SM_XVIRTUALSCREEN)
	oy := win32.GetSystemMetrics(win32.SM_YVIRTUALSCREEN)
	return POINT{int(ox), int(oy)}
}

func GetVirtualScreenSize() POINT {
	cx := win32.GetSystemMetrics(win32.SM_CXVIRTUALSCREEN)
	cy := win32.GetSystemMetrics(win32.SM_CYVIRTUALSCREEN)
	return POINT{int(cx), int(cy)}
}

// GetScreenDimensions returns the width and height of the primary display in pixels.
func GetScreenDimensions() POINT {
	return POINT{int(win32.GetSystemMetrics(win32.SM_CXSCREEN)), int(win32.GetSystemMetrics(win32.SM_CYSCREEN))}
}

// OnScreen reports whether the point (x, y) lies within the bounds of the
// primary display. OnScreen returns true if the point
// is on‑screen at its current resolution; otherwise it returns false.
func OnScreen(x, y int) bool {
	p := GetScreenDimensions()
	width, height := p.X, p.Y
	return 0 <= x && x < width &&
		0 <= y && y < height
}

// GetDisplayBounds returns the bounds of the display at the specified index.
// The index starts at 0 for the primary display.
func GetDisplayBounds(displayIndex int) image.Rectangle {
	var rects []win32.RECT

	monitorEnumProc := syscall.NewCallback(func(hMonitor win32.HMONITOR, hdcMonitor win32.HDC, lprcMonitor *win32.RECT, dwData uintptr) uintptr {
		rects = append(rects, *lprcMonitor)
		return 1 // continue enumeration
	})

	win32.EnumDisplayMonitors(
		win32.HDC(0),
		nil,
		monitorEnumProc,
		0,
	)

	if displayIndex < 0 || displayIndex >= len(rects) {
		return image.Rectangle{}
	}
	r := rects[displayIndex]
	return image.Rect(
		int(r.Left), int(r.Top),
		int(r.Right), int(r.Bottom))
}

// CaptureDisplay captures whole region of displayIndex'th display, starts at 0 for primary display.
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	rect := GetDisplayBounds(displayIndex)
	return CaptureRect(rect)
}

// CaptureRect captures specified region of desktop.
func CaptureRect(rect image.Rectangle) (*image.RGBA, error) {
	return Capture(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
}

func createImage(rect image.Rectangle) (img *image.RGBA, e error) {
	img = nil
	e = errors.New("Cannot create image.RGBA")

	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)

	return img, e
}

// Capture captures a screenshot of the specified area (x, y, width, height) and returns it as an image.RGBA.
func Capture(x, y, width, height int) (*image.RGBA, error) {
	rect := image.Rect(0, 0, width, height)
	img, err := createImage(rect)
	if err != nil {
		return nil, err
	}

	hwnd := win32.GetDesktopWindow()
	hdc := win32.GetDC(hwnd)
	if hdc == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer win32.ReleaseDC(hwnd, hdc)

	memory_device := win32.CreateCompatibleDC(hdc)
	if memory_device == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer win32.DeleteDC(memory_device)

	bitmap := win32.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bitmap == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer win32.DeleteObject(win32.HGDIOBJ(bitmap))

	var header win32.BITMAPINFOHEADER
	header.BiSize = uint32(unsafe.Sizeof(header))
	header.BiPlanes = 1
	header.BiBitCount = 32
	header.BiWidth = int32(width)
	header.BiHeight = int32(-height)
	header.BiCompression = win32.BI_RGB
	header.BiSizeImage = 0

	bitmapDataSize := uintptr(((int64(width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(height))
	hmem, errAlloc := win32.GlobalAlloc(win32.GMEM_MOVEABLE, bitmapDataSize)
	if errAlloc != win32.ERROR_SUCCESS || hmem == 0 {
		return nil, errors.New("GlobalAlloc failed")
	}
	defer win32.GlobalFree(hmem)
	memptr, _ := win32.GlobalLock(hmem)
	defer win32.GlobalUnlock(hmem)

	old := win32.SelectObject(memory_device, win32.HGDIOBJ(bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win32.SelectObject(memory_device, old)

	bitBltOk, _ := win32.BitBlt(memory_device, 0, 0, int32(width), int32(height), hdc, int32(x), int32(y), win32.SRCCOPY)
	if bitBltOk == 0 {
		return nil, errors.New("BitBlt failed")
	}

	if win32.GetDIBits(hdc, bitmap, 0, uint32(height), memptr, (*win32.BITMAPINFO)(unsafe.Pointer(&header)), win32.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	i := 0
	src := uintptr(memptr)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + uintptr(1)))
			v2 := *(*uint8)(unsafe.Pointer(src + uintptr(2)))

			// BGRA => RGBA, and set A to 255
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += uintptr(4)
		}
	}

	return img, nil
}

// ScreenShot captures a screenshot of the specified area (x, y, width, height).
func ScreenShot(x, y, width, height int) (*image.RGBA, error) {
	if !OnScreen(x, y) || !OnScreen(x+width-1, y+height-1) {
		return nil, errors.New("coordinates are off-screen")
	}
	return Capture(x, y, width, height)
}

// DisplayScreenShot captures a screenshot of the specified display index.
func DisplayScreenShot(displayIndex int) (*image.RGBA, error) {
	if displayIndex < 0 {
		return nil, errors.New("display index cannot be negative")
	}
	rect := GetDisplayBounds(displayIndex)
	if rect.Empty() {
		return nil, errors.New("invalid display index")
	}
	return CaptureRect(rect)
}

// FullScreenShot captures a screenshot of the entire primary display.
func FullScreenShot() (*image.RGBA, error) {
	screen := GetScreenDimensions()
	return Capture(0, 0, screen.X, screen.Y)
}
