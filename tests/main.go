package main

import (
	"errors"
	"fmt"
	"time"

	goautogui "goautogui/windows"
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
		testPosition,
		testMoveAbsolute,
		testMoveRelative,
		testClick,
		testDoubleClick,
		testDrag,
		testScroll,
		testSize,
		testVirtualOffset,
		testOnScreen,
		testLerp,
		// Keyboard tests
		testVKeyDownUp,
		testKeyDownUpRune,
		testPressString,
		testVPressKeys,
		testHoldContext,
		testTypeWrite,
		testWrite,
		testHotKey,
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
func testPosition() testResult {
	pos := goautogui.Position()
	fmt.Println("Current mouse position:", pos)
	if pos.X < 0 || pos.Y < 0 {
		return testResult{"Position", fmt.Errorf("invalid position (%d,%d)", pos.X, pos.Y)}
	}
	return testResult{"Position", nil}
}

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
	w, h := goautogui.Size()
	if w <= 0 || h <= 0 {
		return testResult{"Size", fmt.Errorf("invalid size %dx%d", w, h)}
	}
	return testResult{"Size", nil}
}

func testVirtualOffset() testResult {
	offset := goautogui.VirtualOffset()
	_ = offset
	return testResult{"VirtualOffset", nil}
}

func testOnScreen() testResult {
	w, h := goautogui.Size()
	if !goautogui.OnScreen(w/2, h/2) {
		return testResult{"OnScreen", errors.New("center should be on-screen")}
	}
	if goautogui.OnScreen(-10, -10) {
		return testResult{"OnScreen", errors.New("off-screen not detected")}
	}
	return testResult{"OnScreen", nil}
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
