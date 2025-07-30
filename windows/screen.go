//go:build windows

package windows

import (
	"github.com/zzl/go-win32api/v2/win32"
)

// Point represents an x, y coordinate pair
type POINT struct {
	X, Y int
}

// BOX defines a rectangle by left, top position and its width and height.
type BOX struct {
	Left, Top, Width, Height int
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
