//go:build windows

package windows

import (
	"errors"
	"fmt"
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

// Clamps the (x, y) coordinates to the client area of the specified HWND.
func clampToClient(hwnd win32.HWND, x, y int32) (int32, int32) {
	var rc win32.RECT
	if ok, winerr := win32.GetClientRect(hwnd, &rc); ok == 0 && winerr == win32.ERROR_SUCCESS {
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		w := rc.Right - rc.Left
		h := rc.Bottom - rc.Top
		if x >= w {
			x = w - 1
		}
		if y >= h {
			y = h - 1
		}
	}
	return x, y
}

// CursorPos returns the current mouse cursor coordinates (x, y) in pixels.
// On Windows, it calls the Win32 GetCursorPos API under the hood.
func Position() POINT {
	cursor := win32.POINT{}
	win32.GetCursorPos(&cursor)
	return POINT{X: int(cursor.X), Y: int(cursor.Y)}
}

// GetScreenDimensions returns the width and height of the primary display in pixels.
func GetScreenDimensions() POINT {
	return POINT{int(win32.GetSystemMetrics(win32.SM_CXSCREEN)), int(win32.GetSystemMetrics(win32.SM_CYSCREEN))}
}

// GetVirtualScreenOffset returns the top‑left origin of the virtual desktop relative
// to the primary monitor. This is often negative if you have put monitors to the left/top.
func GetVirtualScreenOffset() POINT {
	ox := win32.GetSystemMetrics(win32.SM_XVIRTUALSCREEN)
	oy := win32.GetSystemMetrics(win32.SM_YVIRTUALSCREEN)
	return POINT{int(ox), int(oy)}
}

// GetVirtualScreenSize returns the size of the virtual desktop in pixels.
// The virtual desktop is the union of all monitors, including those that are off-screen.s
func GetVirtualScreenSize() POINT {
	cx := win32.GetSystemMetrics(win32.SM_CXVIRTUALSCREEN)
	cy := win32.GetSystemMetrics(win32.SM_CYVIRTUALSCREEN)
	return POINT{int(cx), int(cy)}
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

// createImage creates an image.RGBA with the specified rectangle dimensions.
func createImage(rect image.Rectangle) (img *image.RGBA, e error) {
	img = nil
	e = errors.New("Cannot create image.RGBA")

	// Ensure the rectangle is valid and not too large.
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

// Capture captures a screenshot of the specified area (x, y, width, height).
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

	memDC := win32.CreateCompatibleDC(hdc)
	if memDC == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer win32.DeleteDC(memDC)

	bmp := win32.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bmp == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer win32.DeleteObject(win32.HGDIOBJ(bmp))

	oldObj := win32.SelectObject(memDC, win32.HGDIOBJ(bmp))
	defer win32.SelectObject(memDC, oldObj)

	if ok, _ := win32.BitBlt(memDC, 0, 0, int32(width), int32(height), hdc, int32(x), int32(y), win32.SRCCOPY); ok == 0 {
		code := win32.GetLastError()
		return nil, fmt.Errorf("BitBlt failed, GetLastError=%d", code)
	}

	var bih win32.BITMAPINFOHEADER

	bih = win32.BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(bih)),
		BiPlanes:      1,
		BiBitCount:    32,
		BiWidth:       int32(width),
		BiHeight:      -int32(height),
		BiCompression: win32.BI_RGB,
	}

	byteCount := width * height * 4
	hmem, allocErr := win32.GlobalAlloc(win32.GMEM_MOVEABLE, uintptr(byteCount))
	if allocErr != win32.ERROR_SUCCESS || hmem == 0 {
		return nil, errors.New("GlobalAlloc failed")
	}
	defer win32.GlobalFree(hmem)

	memptr, _ := win32.GlobalLock(hmem)
	defer win32.GlobalUnlock(hmem)

	if win32.GetDIBits(hdc, bmp, 0, uint32(height), memptr, (*win32.BITMAPINFO)(unsafe.Pointer(&bih)), win32.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	buf := unsafe.Slice((*byte)(memptr), byteCount)

	pix := img.Pix
	for dst := 0; dst < byteCount; dst += 4 {
		// RGBA ← buf[B,G,R,_]
		pix[dst+0], pix[dst+1], pix[dst+2], pix[dst+3] =
			buf[dst+2], buf[dst+1], buf[dst+0], 0xFF
	}

	return img, nil
}

// CaptureWindow captures the screenshot of the specified window handle (hwnd).
// It uses the Win32 API to capture the window content, including non-client areas.
func CaptureWindow(hwnd win32.HWND) (*image.RGBA, error) {
	var rc win32.RECT
	ok, winerr := win32.GetWindowRect(hwnd, &rc)
	if ok == 0 || winerr != win32.ERROR_SUCCESS {
		return nil, fmt.Errorf("GetWindowRect failed")
	}

	w, h := int(rc.Right-rc.Left), int(rc.Bottom-rc.Top)

	hdcScreen := win32.GetWindowDC(hwnd)
	if hdcScreen == 0 {
		return nil, fmt.Errorf("GetWindowDC failed")
	}
	defer win32.ReleaseDC(hwnd, hdcScreen)

	hdcMem := win32.CreateCompatibleDC(hdcScreen)
	if hdcMem == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed")
	}
	defer win32.DeleteDC(hdcMem)

	var bmi win32.BITMAPINFO
	bmi.BmiHeader = win32.BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(win32.BITMAPINFOHEADER{})),
		BiWidth:       int32(w),
		BiHeight:      -int32(h), // top-down
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: win32.BI_RGB,
	}

	var bitsPtr unsafe.Pointer
	hBitmap, winerr := win32.CreateDIBSection(
		hdcScreen,
		&bmi,
		win32.DIB_RGB_COLORS,
		unsafe.Pointer(&bitsPtr),
		0,
		0,
	)

	if hBitmap == 0 || bitsPtr == nil || winerr != win32.ERROR_SUCCESS {
		return nil, fmt.Errorf("CreateDIBSection failed")
	}
	defer win32.DeleteObject(win32.HGDIOBJ(hBitmap))

	old := win32.SelectObject(hdcMem, win32.HGDIOBJ(hBitmap))
	defer win32.SelectObject(hdcMem, old)

	if win32.PrintWindow(hwnd, hdcMem, win32.PRINT_WINDOW_FLAGS(2)) == 0 { // win32.PRINT_WINDOW_FLAGS(2) is PW_RENDERFULLCONTENT
		// Flags documented here: https://learn.microsoft.com/en-us/windows/win32/gdi/wm-printclient
		// Without these flags, the window may not render correctly.
		flags := win32.PRF_ERASEBKGND | win32.PRF_CHILDREN | win32.PRF_CLIENT | win32.PRF_NONCLIENT
		win32.SendMessage(hwnd, win32.WM_PRINT, uintptr(hdcMem), uintptr(flags))
	}

	byteCount := w * h * 4 // 4 bytes per pixel (RGBA)

	buf := unsafe.Slice((*byte)(bitsPtr), byteCount)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for dst := 0; dst < byteCount; dst += 4 {
		// RGBA ← buf[B,G,R,_]
		img.Pix[dst+0], img.Pix[dst+1], img.Pix[dst+2], img.Pix[dst+3] =
			buf[dst+2], buf[dst+1], buf[dst+0], 0xFF
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

// CaptureRect captures specified region of desktop.
func CaptureRect(rect image.Rectangle) (*image.RGBA, error) {
	return Capture(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
}

// CaptureDisplay captures the screenshot of the specified display index.
func CaptureDisplay(displayIndex int) (*image.RGBA, error) {
	rect := GetDisplayBounds(displayIndex)
	return CaptureRect(rect)
}

// CapturePrimaryDisplay captures the screenshot of the primary display.
func CapturePrimaryDisplay() (*image.RGBA, error) {
	screen := GetScreenDimensions()
	return Capture(0, 0, screen.X, screen.Y)
}
