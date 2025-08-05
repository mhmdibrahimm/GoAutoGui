package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	goautogui "github.com/mhmdibrahimm/goautogui/windows"
	"github.com/zzl/go-win32api/v2/win32"
)

// testResult holds the outcome of a single test.
type testResult struct {
	name string
	err  error
}

func main() {
	fmt.Println("Running goautogui full test suite in 3 seconds...")
	time.Sleep(3 * time.Second)

	tests := []func() testResult{
		// Mouse tests
		testMoveAbsolute,
		testMoveRelative,
		testClick,
		testDoubleClick,
		testDrag,
		testScroll,
		testSize,

		// Keyboard tests
		testVKeyDownUp,
		testKeyDownUpRune,
		testPressString,
		testVPressKeys,
		testHoldContext,
		testTypeWrite,
		testWrite,
		testHotKey,
		testWritingLinkUrl,
		testScreenShot,

		// Screen tests
		testPosition,
		testGetScreenDimensions,
		testGetVirtualScreenOffset,
		testGetVirtualScreenSize,
		testOnScreen,
		testGetDisplayBounds,
		testCapture,
		testCaptureWindow,
		testScreenShot,
		testCaptureRect,
		testCaptureDisplay,
		testCapturePrimaryDisplay,
	}

	var passed, failed int
	var failures []testResult
	for _, test := range tests {
		res := test()
		if res.err == nil {
			fmt.Printf("✅ %s\n", res.name)
			passed++
		} else {
			fmt.Printf("❌ %s: %v\n", res.name, res.err)
			failed++
			failures = append(failures, res)
		}
	}

	fmt.Println("\n=== Test Summary ===")
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(tests), passed, failed)
	if failed > 0 {
		fmt.Println("Failures details:")
		for _, f := range failures {
			fmt.Printf(" - %s: %v\n", f.name, f.err)
		}
	}
}

// Mouse Tests

func testMoveAbsolute() testResult {
	orig := goautogui.Position()
	goautogui.SetCursorPosition(100, 100)
	time.Sleep(100 * time.Millisecond)
	p := goautogui.Position()
	if p.X != 100 || p.Y != 100 {
		return testResult{"MoveTo", fmt.Errorf("expected (100,100), got (%d,%d)", p.X, p.Y)}
	}
	goautogui.SetCursorPosition(orig.X, orig.Y)
	return testResult{"MoveTo", nil}
}

func testMoveRelative() testResult {
	orig := goautogui.Position()
	goautogui.Move(10, 15)
	time.Sleep(100 * time.Millisecond)
	p := goautogui.Position()
	if p.X != orig.X+10 || p.Y != orig.Y+15 {
		return testResult{"MoveRel", fmt.Errorf("expected (%d,%d), got (%d,%d)", orig.X+10, orig.Y+15, p.X, p.Y)}
	}
	goautogui.SetCursorPosition(orig.X, orig.Y)
	return testResult{"MoveRel", nil}
}

func testClick() testResult {
	err := goautogui.Click(goautogui.MouseLeftButton, 484, 766)
	if err != nil {
		return testResult{"Click", err}
	}
	return testResult{"Click", nil}
}

func testDoubleClick() testResult {
	err := goautogui.DoubleClick(goautogui.MouseLeftButton, 484, 766)
	if err != nil {
		return testResult{"DoubleClick", err}
	}
	return testResult{"DoubleClick", nil}
}

func testDrag() testResult {
	goautogui.SetCursorPosition(469, 577)
	point := goautogui.POINT{X: 597, Y: 744}
	err := goautogui.DragTo(point.X, point.Y, 0.2, goautogui.MouseLeftButton)
	if err != nil {
		return testResult{"DragTo", err}
	}
	return testResult{"DragTo", nil}
}

func testScroll() testResult {
	goautogui.Scroll(300, 300, 3)
	time.Sleep(1000 * time.Millisecond) // Allow some time for the scroll to take effect
	goautogui.Scroll(300, 300, -3)
	return testResult{"Scroll", nil}
}

func testSize() testResult {
	dim := goautogui.GetScreenDimensions()
	if dim.X <= 0 || dim.Y <= 0 {
		return testResult{"Size", fmt.Errorf("invalid size %dx%d", dim.X, dim.Y)}
	}
	return testResult{"Size", nil}
}

func testVirtualOffset() testResult {
	offset := goautogui.GetVirtualScreenOffset()
	_ = offset
	return testResult{"VirtualOffset", nil}
}

func testLerp() testResult {
	x, y := goautogui.Lerp(0, 0, 10, 10, 0.5)
	if x != 5 || y != 5 {
		return testResult{"Lerp", fmt.Errorf("expected (5,5), got (%.2f,%.2f)", x, y)}
	}
	return testResult{"Lerp", nil}
}

// Keyboard Tests
func testVKeyDownUp() testResult {
	err := goautogui.VKeyDown(goautogui.KEY_LWIN)
	if err != nil {
		return testResult{"VKeyDown", err}
	}
	err = goautogui.VKeyUp(goautogui.KEY_LWIN)
	if err != nil {
		return testResult{"VKeyUp", err}
	}
	return testResult{"VKeyDownUp", nil}
}

func testKeyDownUpRune() testResult {
	err := goautogui.KeyDown('a')
	if err != nil {
		return testResult{"KeyDown('a')", err}
	}
	err = goautogui.KeyUp('a')
	if err != nil {
		return testResult{"KeyUp('a')", err}
	}
	err = goautogui.KeyDown('A')
	if err != nil {
		return testResult{"KeyDown('A')", err}
	}
	err = goautogui.KeyUp('A')
	if err != nil {
		return testResult{"KeyUp('A')", err}
	}
	return testResult{"KeyDownUpRune", nil}
}

func testPressString() testResult {
	err := goautogui.Press("abc", 2, 50*time.Millisecond)
	if err != nil {
		return testResult{"Press('abc')", err}
	}
	return testResult{"PressString", nil}
}

func testVPressKeys() testResult {
	err := goautogui.VPress(1, 50*time.Millisecond, goautogui.KEY_CONTROL, goautogui.KEY_SHIFT)
	if err != nil {
		return testResult{"VPress", err}
	}
	return testResult{"VPressKeys", nil}
}

func testHoldContext() testResult {
	hc, err := goautogui.Hold(goautogui.KEY_CONTROL, goautogui.KEY_SHIFT)
	if err != nil {
		return testResult{"HoldContextDown", err}
	}
	err = hc.Release()
	if err != nil {
		return testResult{"HoldContextRelease", err}
	}
	return testResult{"HoldContext", nil}
}

func testTypeWrite() testResult {
	err := goautogui.TypeWrite("Test", 10)
	if err != nil {
		return testResult{"TypeWrite", err}
	}
	return testResult{"TypeWrite", nil}
}

func testWrite() testResult {
	err := goautogui.Write("Hello", 5)
	if err != nil {
		return testResult{"Write", err}
	}
	return testResult{"Write", nil}
}

func testHotKey() testResult {
	err := goautogui.HotKey(20*time.Millisecond, goautogui.KEY_CONTROL, goautogui.KEY_C)
	if err != nil {
		return testResult{"HotKey", err}
	}
	return testResult{"HotKey", nil}
}

func testWritingLinkUrl() testResult {
	err := goautogui.TypeWrite("https://example.com", 5)
	if err != nil {
		return testResult{"WriteLinkURL", err}
	}
	return testResult{"WriteLinkURL", nil}
}

// Screen Tests

func testPosition() testResult {
	pos := goautogui.Position()
	fmt.Println("Current mouse position:", pos)
	if pos.X < 0 || pos.Y < 0 {
		return testResult{"Position", fmt.Errorf("invalid position (%d,%d)", pos.X, pos.Y)}
	}
	return testResult{"Position", nil}
}

func testGetScreenDimensions() testResult {
	dim := goautogui.GetScreenDimensions()
	if dim.X <= 0 || dim.Y <= 0 {
		return testResult{"GetScreenDimensions", fmt.Errorf("invalid dimensions %dx%d", dim.X, dim.Y)}
	}
	fmt.Printf("Screen dimensions: %dx%d\n", dim.X, dim.Y)
	return testResult{"GetScreenDimensions", nil}
}
func testGetVirtualScreenOffset() testResult {
	offset := goautogui.GetVirtualScreenOffset()
	if offset.X < 0 || offset.Y < 0 {
		return testResult{"GetVirtualScreenOffset", fmt.Errorf("invalid offset (%d,%d)", offset.X, offset.Y)}
	}
	fmt.Printf("Virtual screen offset: (%d,%d)\n", offset.X, offset.Y)
	return testResult{"GetVirtualScreenOffset", nil}
}
func testGetVirtualScreenSize() testResult {
	size := goautogui.GetVirtualScreenSize()
	if size.X <= 0 || size.Y <= 0 {
		return testResult{"GetVirtualScreenSize", fmt.Errorf("invalid size %dx%d", size.X, size.Y)}
	}
	fmt.Printf("Virtual screen size: %dx%d\n", size.X, size.Y)
	return testResult{"GetVirtualScreenSize", nil}
}
func testOnScreen() testResult {
	dim := goautogui.GetScreenDimensions()
	if !goautogui.OnScreen(dim.X/2, dim.Y/2) {
		return testResult{"OnScreen", errors.New("center should be on-screen")}
	}
	if goautogui.OnScreen(-10, -10) {
		return testResult{"OnScreen", errors.New("off-screen not detected")}
	}
	fmt.Println("OnScreen test passed")
	return testResult{"OnScreen", nil}
}
func testGetDisplayBounds() testResult {
	displayIndex := 0 // Test primary display
	bounds := goautogui.GetDisplayBounds(displayIndex)
	if bounds.Empty() {
		return testResult{"GetDisplayBounds", fmt.Errorf("bounds for display %d are empty", displayIndex)}
	}
	fmt.Printf("Display %d bounds: %v\n", displayIndex, bounds)
	return testResult{"GetDisplayBounds", nil}
}

func openImage(img image.Image, test_file_name string) {
	if img == nil {
		fmt.Println("Cannot open nil image")
		return
	}
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, test_file_name+".png")
	f, err := os.Create(tmpFile)
	if err != nil {
		fmt.Println("Failed to create temp file for image:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Println("Failed to encode image:", err)
		return
	}
	f.Close()
	// Open with default viewer
	cmd := exec.Command("rundll32", "shell32.dll,ShellExec_RunDLL", tmpFile)

	_ = cmd.Start()
	go func() {
		time.Sleep(50 * time.Second)
		os.Remove(tmpFile)
	}()
}

func testCapture() testResult {
	img, err := goautogui.Capture(100, 100, 200, 200)
	if err != nil {
		return testResult{"Capture", err}
	}
	if img == nil {
		return testResult{"Capture", errors.New("captured image is nil")}
	}
	openImage(img, "capture_test_image")
	fmt.Println("Capture test passed")
	return testResult{"Capture", nil}
}
func testCaptureWindow() testResult {
	// Assuming we have a window handle to capture
	hwnd := win32.GetForegroundWindow()
	if hwnd == 0 {
		return testResult{"CaptureWindow", errors.New("failed to get foreground window")}
	}

	img, err := goautogui.CaptureWindow(hwnd)
	if err != nil {
		return testResult{"CaptureWindow", err}
	}
	if img == nil {
		return testResult{"CaptureWindow", errors.New("captured image is nil")}
	}
	openImage(img, "capture_window_test_image")
	fmt.Println("CaptureWindow test passed")
	return testResult{"CaptureWindow", nil}
}
func testScreenShot() testResult {
	img, err := goautogui.ScreenShot(100, 100, 200, 200)
	if err != nil {
		return testResult{"ScreenShot", err}
	}
	if img == nil {
		return testResult{"ScreenShot", errors.New("screenshot image is nil")}
	}
	openImage(img, "screenshot_test_image")
	fmt.Println("ScreenShot test passed")
	return testResult{"ScreenShot", nil}
}
func testCaptureRect() testResult {
	rect := goautogui.GetDisplayBounds(0) // Capture primary display
	if rect.Empty() {
		return testResult{"CaptureRect", errors.New("display bounds are empty")}
	}

	img, err := goautogui.Capture(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	if err != nil {
		return testResult{"CaptureRect", err}
	}
	if img == nil {
		return testResult{"CaptureRect", errors.New("captured image is nil")}
	}
	openImage(img, "capture_rect_test_image")
	fmt.Println("CaptureRect test passed")
	return testResult{"CaptureRect", nil}
}
func testCaptureDisplay() testResult {
	displayIndex := 0 // Test primary display
	rect := goautogui.GetDisplayBounds(displayIndex)
	if rect.Empty() {
		return testResult{"CaptureDisplay", fmt.Errorf("bounds for display %d are empty", displayIndex)}
	}

	img, err := goautogui.Capture(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	if err != nil {
		return testResult{"CaptureDisplay", err}
	}
	if img == nil {
		return testResult{"CaptureDisplay", errors.New("captured image is nil")}
	}
	openImage(img, "capture_display_test_image")
	fmt.Println("CaptureDisplay test passed")
	return testResult{"CaptureDisplay", nil}
}
func testCapturePrimaryDisplay() testResult {
	rect := goautogui.GetDisplayBounds(0) // Primary display
	if rect.Empty() {
		return testResult{"CapturePrimaryDisplay", errors.New("primary display bounds are empty")}
	}

	img, err := goautogui.Capture(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	if err != nil {
		return testResult{"CapturePrimaryDisplay", err}
	}
	if img == nil {
		return testResult{"CapturePrimaryDisplay", errors.New("captured image is nil")}
	}
	openImage(img, "capture_primary_display_test_image")
	fmt.Println("CapturePrimaryDisplay test passed")
	return testResult{"CapturePrimaryDisplay", nil}
}
