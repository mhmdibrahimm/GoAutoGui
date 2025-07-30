package windows

import (
	win32 "github.com/zzl/go-win32api/v2/win32"
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

// VirtualOffset returns the top‑left origin of the virtual desktop relative
// to the primary monitor. This is often negative if you’ve put monitors to the left/top.
func VirtualOffset() POINT {
	ox := win32.GetSystemMetrics(win32.SM_XVIRTUALSCREEN)
	oy := win32.GetSystemMetrics(win32.SM_YVIRTUALSCREEN)
	return POINT{int(ox), int(oy)}
}

// Size returns the width and height of the primary display in pixels.
func Size() (int, int) {
	return int(win32.GetSystemMetrics(win32.SM_CXSCREEN)), int(win32.GetSystemMetrics(win32.SM_CYSCREEN))
}

// OnScreen reports whether the point (x, y) lies within the bounds of the
// primary display. OnScreen returns true if the point
// is on‑screen at its current resolution; otherwise it returns false.
func OnScreen(x, y int) bool {
	width, height := Size()
	return 0 <= x && x < width &&
		0 <= y && y < height
}

func init() {
	// opt into true‑pixel metrics
	win32.SetProcessDPIAware()
}
