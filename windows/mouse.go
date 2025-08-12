//go:build windows

package windows

import (
	"fmt"
	"time"

	"github.com/zzl/go-win32api/v2/win32"
)

// MouseButton represents different mouse buttons
type MouseButton int

const (
	MOUSEEVENTF_LEFTCLICK   = win32.MOUSEEVENTF_LEFTDOWN | win32.MOUSEEVENTF_LEFTUP     // Left click event
	MOUSEEVENTF_RIGHTCLICK  = win32.MOUSEEVENTF_RIGHTDOWN | win32.MOUSEEVENTF_RIGHTUP   // Right click event
	MOUSEEVENTF_MIDDLECLICK = win32.MOUSEEVENTF_MIDDLEDOWN | win32.MOUSEEVENTF_MIDDLEUP // Middle click event
)

const (
	MouseLeftButton      MouseButton = iota // MouseLeftButton is the left mouse button
	MouseMiddleButton                       // MouseMiddleButton is the middle mouse button
	MouseRightButton                        // MouseRightButton is the right mouse button
	MouseX1Button                           // MouseX1Button is the first extra mouse button (usually the back button)
	MouseX2Button                           // MouseX2Button is the second extra mouse button (usually the forward button)
	MousePrimaryButton                      // MousePrimaryButton is the primary mouse button (left by default, right if swapped)
	MouseSecondaryButton                    // MouseSecondaryButton is the secondary mouse button (right by default, left if swapped)
)

// String method for better printing
func (mb MouseButton) String() string {
	switch mb {
	case MouseLeftButton:
		return "Left Button"
	case MouseRightButton:
		return "Right Button"
	case MouseMiddleButton:
		return "Middle Button"
	case MousePrimaryButton:
		return "Primary Button"
	case MouseSecondaryButton:
		return "Secondary Button"
	case MouseX1Button:
		return "X1 Button "
	case MouseX2Button:
		return "X2 Button"
	default:
		return "Unknown Button"
	}
}

// Checks if the mouse buttons are swapped.
// On Windows, it uses GetSystemMetrics with SM_SWAPBUTTON to determine this.
// Returns true if the left and right mouse buttons are swapped, false otherwise.
func IsMouseSwapped() bool {
	return win32.GetSystemMetrics(win32.SM_SWAPBUTTON) != 0
}

// normalizeMouseButton normalizes the mouse button to a valid MouseButton type.
// It converts MousePrimaryButton to MouseLeftButton or MouseRightButton based on the swap state
// This only applies to the primary and secondary buttons.
func normalizeMouseButton(mb MouseButton) (MouseButton, error) {
	switch mb {
	// already physical buttons → no change
	case MouseLeftButton,
		MouseMiddleButton,
		MouseRightButton,
		MouseX1Button,
		MouseX2Button:
		return mb, nil

	// “Primary” is left unless swapped
	case MousePrimaryButton:
		if IsMouseSwapped() {
			return MouseRightButton, nil
		}
		return MouseLeftButton, nil

	// “Secondary” is right unless swapped
	case MouseSecondaryButton:
		if IsMouseSwapped() {
			return MouseLeftButton, nil
		}
		return MouseRightButton, nil

	default:
		return MouseButton(-1), fmt.Errorf("invalid mouse button: %v", mb)
	}
}

// Makes the call to the mouse_event() win32 function.
// dwData: if event has MOUSEEVENTF_WHEEL or MOUSEEVENTF_HWHEEL, then it specifies the amount
// of wheel movement which is usually 120 units per notch (WHEEL_DELTA).
// If event has MOUSEEVENTF_XDOWN or MOUSEEVENTF_XUP, then it specifies the X button number (1 or 2).
// Else, it should be 0.
func sendMouseEvent(event win32.MOUSE_EVENT_FLAGS, x, y, dwData int) {
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y

	convertedX := x * 65535 / (width - 1)
	convertedY := y * 65535 / (height - 1)

	win32.Mouse_event(event, int32(convertedX), int32(convertedY), int32(dwData), 0)
}

func sendMessageTimeout(hwnd win32.HWND, msg uint32, wparam win32.WPARAM, lparam win32.LPARAM) {
	// SMTO_ABORTIFHUNG: return if target thread is not responding
	// 2000 ms timeout is plenty
	win32.SendMessageTimeout(hwnd, msg, wparam, lparam,
		win32.SMTO_ABORTIFHUNG, 2000, nil)
}

// Send the down up event to Windows by calling the mouse_event() win32
// function.
func MouseDown(mb MouseButton, x, y int) (bool, error) {
	mb, err := normalizeMouseButton(mb)
	if err != nil {
		return false, err
	}

	var event win32.MOUSE_EVENT_FLAGS
	dwData := 0
	switch mb {
	case MouseLeftButton:
		event = win32.MOUSEEVENTF_LEFTDOWN
	case MouseRightButton:
		event = win32.MOUSEEVENTF_RIGHTDOWN
	case MouseMiddleButton:
		event = win32.MOUSEEVENTF_MIDDLEDOWN
	case MouseX1Button:
		event = win32.MOUSEEVENTF_XDOWN
		dwData = int(KEY_XBUTTON1)
	case MouseX2Button:
		event = win32.MOUSEEVENTF_XDOWN
		dwData = int(KEY_XBUTTON2)
	}
	sendMouseEvent(event, x, y, dwData)

	return true, nil
}

// Send the mouse up event to Windows by calling the mouse_event() win32
// function.
func MouseUp(mb MouseButton, x, y int) (bool, error) {
	mb, err := normalizeMouseButton(mb)
	if err != nil {
		return false, err
	}

	var event win32.MOUSE_EVENT_FLAGS
	dwData := 0

	switch mb {
	case MouseLeftButton:
		event = win32.MOUSEEVENTF_LEFTUP
	case MouseRightButton:
		event = win32.MOUSEEVENTF_RIGHTUP
	case MouseMiddleButton:
		event = win32.MOUSEEVENTF_MIDDLEUP
	case MouseX1Button:
		event = win32.MOUSEEVENTF_XUP
		dwData = int(KEY_XBUTTON1)
	case MouseX2Button:
		event = win32.MOUSEEVENTF_XUP
		dwData = int(KEY_XBUTTON2)
	}
	sendMouseEvent(event, x, y, dwData)

	return true, nil
}

// Click performs a mouse button click at the specified (x, y) coordinates supporting multiple clicks.
func ClickAt(mb MouseButton, x, y, clicks int) error {
	if mb != MouseLeftButton && mb != MouseRightButton && mb != MouseMiddleButton {
		return fmt.Errorf("mouse button must be one of MouseLeftButton, MouseRightButton, or Middle; received %v", mb)
	}

	var event win32.MOUSE_EVENT_FLAGS
	switch mb {
	case MouseLeftButton:
		event = MOUSEEVENTF_LEFTCLICK
	case MouseRightButton:
		event = MOUSEEVENTF_RIGHTCLICK
	case MouseMiddleButton:
		event = MOUSEEVENTF_MIDDLECLICK
	}
	for i := 0; i < clicks; i++ {
		sendMouseEvent(event|win32.MOUSEEVENTF_ABSOLUTE|win32.MOUSEEVENTF_MOVE, x, y, 0)
	}
	return nil
}

// Click performs a mouse button click at the specified (x, y) coordinates.
func Click(mb MouseButton, x, y int) error {
	return ClickAt(mb, x, y, 1)
}

// LeftClick performs a left mouse button click at the specified (x, y) coordinates.
func LeftClick(x, y int) error {
	return Click(MouseLeftButton, x, y)
}

// RightClick performs a right mouse button click at the specified (x, y) coordinates.
func RightClick(x, y int) error {
	return Click(MouseRightButton, x, y)
}

// MiddleClick performs a middle mouse button click at the specified (x, y) coordinates.
func MiddleClick(x, y int) error {
	return Click(MouseMiddleButton, x, y)
}

// X1Click performs a click with the first extra mouse button (usually the back button) at the specified (x, y) coordinates.
func X1Click(x, y int) error {
	return Click(MouseX1Button, x, y)
}

// X2Click performs a click with the second extra mouse button (usually the forward button) at the specified (x, y) coordinates.
func X2Click(x, y int) error {
	return Click(MouseX2Button, x, y)
}

// PrimaryClick perform a click with the primary mouse button at the specified (x, y) coordinates.
// The primary button is usually the physical left button, only if swapped, it will be the right button.
func PrimaryClick(x, y int) error {
	mb, err := normalizeMouseButton(MousePrimaryButton)
	if err != nil {
		return fmt.Errorf("invalid primary mouse button: %v", err)
	}
	return Click(mb, x, y)
}

// SecondaryClick perform a click with the secondary mouse button at the specified (x, y) coordinates.
// The secondary button is usually the physical right button, only if swapped, it will be the left button.
func SecondaryClick(x, y int) error {
	mb, err := normalizeMouseButton(MouseSecondaryButton)
	if err != nil {
		return fmt.Errorf("invalid secondary mouse button: %v", err)
	}
	return Click(mb, x, y)
}

// DoubleClick performs a double mopuse button click at the specified (x, y) coordinates.
func DoubleClick(mb MouseButton, x, y int) error {
	return ClickAt(mb, x, y, 2)
}

// TripleClick performs a triple mouse button click at the specified (x, y) coordinates.
func TripleClick(mb MouseButton, x, y int) error {
	return ClickAt(mb, x, y, 3)
}

// Click a specific HWND at a SCREEN point (no z-order issues).
func ClickHwnd(hwnd win32.HWND, screenX, screenY int) {
	pt := win32.POINT{X: int32(screenX), Y: int32(screenY)}

	// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-mapwindowpoints
	// The MapWindowPoints function converts (maps) a set of points from a coordinate space relative to one window to a coordinate space relative to another window.
	// from=HWND(0) meaning we are mapping from whole screen coordinates to the client area of the specified window.
	win32.MapWindowPoints(win32.HWND(0), hwnd, &pt, 1)
	// Clamp to client rect just to be safe
	cx, cy := clampToClient(hwnd, pt.X, pt.Y)

	lp := win32.LPARAM(uintptr(cx) | uintptr(cy)<<16)
	// SendMessage is synchronous; use SendMessageTimeout if you fear hangs.
	sendMessageTimeout(hwnd, win32.WM_MOUSEMOVE, 0, lp)
	sendMessageTimeout(hwnd, win32.WM_LBUTTONDOWN, win32.WPARAM(win32.MK_LBUTTON), lp)
	sendMessageTimeout(hwnd, win32.WM_LBUTTONUP, 0, lp)
}

// Scroll performs a mouse scroll at the specified (x, y) coordinates.
// Each scroll notch is typically 120 units in Windows.
// Call ScrollRaw to specify the exact scroll amount.
func Scroll(x, y, notches int) {
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y
	x = max(0, min(x, width-1))
	y = max(0, min(y, height-1))
	dwData := notches * int(win32.WHEEL_DELTA) // 120 is the standard scroll amount in Windows
	sendMouseEvent(win32.MOUSEEVENTF_WHEEL, x, y, dwData)
}

// ScrollRaw performs a mouse scroll at the specified (x, y) coordinates with a custom scroll amount.
// This allows for more precise control over the scroll amount. dwData is the number of scroll notches.
func ScrollRaw(x, y, dwData int) {
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y
	x = max(0, min(x, width-1))
	y = max(0, min(y, height-1))
	sendMouseEvent(win32.MOUSEEVENTF_WHEEL, x, y, dwData)
}

// HorizontalScroll performs a horizontal mouse scroll at the specified (x, y) coordinates.
func HorizontalScroll(x, y, notches int) {
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y
	x = max(0, min(x, width-1))
	y = max(0, min(y, height-1))

	sendMouseEvent(win32.MOUSEEVENTF_HWHEEL, x, y, notches)
}

// VerticalScroll performs a vertical mouse scroll at the specified (x, y) coordinates.
func VerticalScroll(x, y, notches int) {
	Scroll(x, y, notches)
}

// MoveTo moves the mouse cursor to the specified (x, y) coordinates.
func SetCursorPosition(x, y int) {
	// TODO: use sendInput instead of win32.SetCursorPos for better compatibility
	win32.SetCursorPos(int32(x), int32(y))
}

// Moves the mouse cursor by the specified (x, y) offsets from its current position.
func Move(x, y int) {
	currentPos := Position()
	newX := currentPos.X + x
	newY := currentPos.Y + y

	// Ensure new position is within screen bounds
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y
	newX = max(0, min(newX, width-1))
	newY = max(0, min(newY, height-1))

	SetCursorPosition(newX, newY)
}

func DragTo(x, y int, duration float64, mb MouseButton) error {
	// Validate mouse button
	if mb != MouseLeftButton && mb != MouseRightButton && mb != MouseMiddleButton {
		return fmt.Errorf("mouse button must be one of MouseLeftButton, MouseRightButton, or Middle; received %v", mb)
	}

	// Get current mouse position
	startPos := Position()
	startX, startY := float64(startPos.X), float64(startPos.Y)
	endX, endY := float64(x), float64(y)

	// If already at target position, just return
	if startPos.X == x && startPos.Y == y {
		return nil
	}

	// Press mouse button down at start position
	_, err := MouseDown(mb, startPos.X, startPos.Y)
	if err != nil {
		return fmt.Errorf("failed to press mouse button down: %v", err)
	}

	// If duration is very small, move instantly
	const MINIMUM_DURATION = 0.1 // seconds
	if duration <= MINIMUM_DURATION {
		SetCursorPosition(x, y)
		_, err := MouseUp(mb, x, y)
		if err != nil {
			return fmt.Errorf("failed to release mouse button: %v", err)
		}
		return nil
	}

	// Calculate steps for smooth movement
	dim := GetScreenDimensions()
	width, height := dim.X, dim.Y
	numSteps := max(width, height)
	sleepAmount := duration / float64(numSteps)
	const MINIMUM_SLEEP = 0.001 // seconds
	if sleepAmount < MINIMUM_SLEEP {
		numSteps = int(duration / MINIMUM_SLEEP)
		sleepAmount = duration / float64(numSteps)
	}

	// Perform smooth drag movement
	for i := 0; i < numSteps; i++ {
		t := float64(i) / float64(numSteps)
		tweenX, tweenY := Lerp(startX, startY, endX, endY, t)

		SetCursorPosition(int(tweenX+0.5), int(tweenY+0.5)) // Round to nearest int

		// Sleep between steps
		time.Sleep(time.Duration(sleepAmount*1000) * time.Millisecond)
	}

	// Ensure we end at the exact target position
	SetCursorPosition(x, y)

	// Release mouse button
	_, err = MouseUp(mb, x, y)
	if err != nil {
		return fmt.Errorf("failed to release mouse button: %v", err)
	}

	return nil
}
