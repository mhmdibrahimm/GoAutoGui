//go:build windows

package windows

import (
	"errors"
	"fmt"
	"strings"
	"time"

	win32 "github.com/zzl/go-win32api/v2/win32"
)

// Helper function to send keyboard events using keybd_event
func sendKeyboardEvent(vk win32.VIRTUAL_KEY, scanCode uint16, flags win32.KEYBD_EVENT_FLAGS) error {
	win32.Keybd_event(byte(vk), byte(scanCode), flags, 0)
	return nil
}

// Returns True if the “character“ is a keyboard key that would require the shift key to be held down, such as
// uppercase letters or the symbols on the keyboard's number row.
func isShiftCharacter(character string) bool {
	shiftChars := "~!@#$%^&*()_+{}|:\"<>?"
	if len(character) == 1 && strings.Contains(shiftChars, character) {
		return true
	}
	if len(character) == 1 && character == strings.ToUpper(character) && character != strings.ToLower(character) {
		return true
	}
	return false
}

func VKeyDown(key KeyboardKeys) error {
	// Presses the specified key down. If the key is not valid, it returns an error.
	// Press the actual key down
	err := sendKeyboardEvent(win32.VIRTUAL_KEY(key), 0, win32.KEYBD_EVENT_FLAGS(0)) // KEYEVENTF_KEYDOWN = 0
	if err != nil {
		return fmt.Errorf("failed to press key: %v", err)
	}
	return nil

}

// Presses the specified key up. If the key is not valid, it returns an error.
func VKeyUp(key KeyboardKeys) error {
	// Press the actual key down
	err := sendKeyboardEvent(win32.VIRTUAL_KEY(key), 0, win32.KEYEVENTF_KEYUP)
	if err != nil {
		return fmt.Errorf("failed to press key: %v", err)
	}
	return nil

}

func KeyDown(key rune) error {
	needsShift := isShiftCharacter(string(key))
	vkCode := win32.VkKeyScanW(uint16(key))
	if vkCode == -1 {
		return fmt.Errorf("there is no vk code for key \"%s\"", string(key))
	}
	if vkCode > 0x100 { // the vk code will be > 0x100 if it needs shift
		vkCode -= 0x100
		needsShift = true
	}
	if needsShift {
		// KEYEVENTF_KEYDOWN = 0 (Technically this constant doesn't exist in the MS documentation. It's the lack of KEYEVENTF_KEYUP that means pressing the key down.)
		err := sendKeyboardEvent(win32.VK_SHIFT, 0, win32.KEYBD_EVENT_FLAGS(0))
		if err != nil {
			return fmt.Errorf("failed to press shift key: %v", err)
		}
	}
	// Press the actual key down
	err := sendKeyboardEvent(win32.VIRTUAL_KEY(vkCode), 0, win32.KEYBD_EVENT_FLAGS(0))
	if err != nil {
		return fmt.Errorf("failed to press key: %v", err)
	}
	return nil
}

// Presses the specified key up. If the key is not valid, it returns an error.
func KeyUp(key rune) error {
	needsShift := isShiftCharacter(string(key))
	vkCode := win32.VkKeyScanW(uint16(key))
	if vkCode == -1 {
		return fmt.Errorf("there is no vk code for key \"%s\"", string(key))
	}
	if vkCode > 0x100 { // the vk code will be > 0x100 if it needs shift
		vkCode -= 0x100
		needsShift = true
	}
	if needsShift {
		err := sendKeyboardEvent(win32.VK_SHIFT, 0, win32.KEYEVENTF_KEYUP) // KEYEVENTF_KEYUP = 2
		if err != nil {
			return fmt.Errorf("failed to press shift key: %v", err)
		}
	}
	// Release the actual key down
	err := sendKeyboardEvent(win32.VIRTUAL_KEY(vkCode), 0, win32.KEYEVENTF_KEYUP)
	if err != nil {
		return fmt.Errorf("failed to press key: %v", err)
	}
	return nil
}

func Press(keys string, presses int, interval time.Duration) error {
	for i := 0; i < presses; i++ {
		for _, k := range keys {

			err := KeyDown(k)
			if err != nil {
				return fmt.Errorf("failed to press key down '%s': %v", string(k), err)
			}

			err = KeyUp(k)
			if err != nil {
				return fmt.Errorf("failed to release key '%s': %v", string(k), err)
			}
		}

		if i < presses-1 { // Don't sleep after the last press
			time.Sleep(interval)
		}
	}

	return nil
}

func VPress(presses int, interval time.Duration, keys ...KeyboardKeys) error {
	for i := 0; i < presses; i++ {
		for _, k := range keys {
			err := VKeyDown(k)
			if err != nil {
				return fmt.Errorf("failed to press key down '%d': %v", k, err)
			}

			err = VKeyUp(k)
			if err != nil {
				return fmt.Errorf("failed to release key '%d': %v", k, err)
			}
		}

		if i < presses-1 { // Don't sleep after the last press
			time.Sleep(interval)
		}
	}

	return nil
}

type HoldContext struct {
	keys []KeyboardKeys
}

// Hold presses the specified key(s) down and returns a cleanup function to release them.
// This simulates Python's context manager behavior.
func Hold(keys ...KeyboardKeys) (*HoldContext, error) {
	var pressed []KeyboardKeys

	// Try to press each key down, tracking successes.
	for _, k := range keys {
		if err := VKeyDown(k); err != nil {
			// Cleanup: release only the keys that succeeded
			for _, pk := range pressed {
				_ = VKeyUp(pk) // ignore cleanup errors
			}
			return nil, fmt.Errorf("failed to press key down '%d': %w", k, err)
		}
		pressed = append(pressed, k)
	}

	// Only the successfully pressed keys get stored in the context
	return &HoldContext{keys: pressed}, nil
}

func (hc *HoldContext) Release() error {
	// Release all keys in reverse order
	for i := len(hc.keys) - 1; i >= 0; i-- {
		err := VKeyUp(hc.keys[i])
		if err != nil {
			return fmt.Errorf("failed to release key '%d': %v", hc.keys[i], err)
		}
	}
	hc.keys = nil // Clear the keys to prevent double release
	return nil
}

// Typewrite simulates typing a message character by character with an optional interval between each character.
func TypeWrite(message string, interval time.Duration) error {
	for _, char := range message {
		charStr := string(char)

		err := Press(charStr, 1, 0) // Press once with no interval between key down/up
		if err != nil {
			return fmt.Errorf("failed to type character '%s': %v", charStr, err)
		}

		if interval > 0 {
			time.Sleep(interval * time.Millisecond)
		}
	}

	return nil
}

// Write simulates typing a message character by character with an optional interval between each character.
func Write(message string, interval time.Duration) error {
	return TypeWrite(message, interval)
}

// Performs key down presses on the arguments passed in order, then performs key releases in reverse order.
func HotKey(interval time.Duration, keys ...KeyboardKeys) error {
	if len(keys) == 0 {
		return errors.New("no keys provided for HotKey")
	}
	fmt.Println("Pressing keys:", keys)
	for _, key := range keys {
		err := VKeyDown(key)
		if err != nil {
			return fmt.Errorf("failed to release key '%s': %v", fmt.Sprint(key), err)
		}
		time.Sleep(interval)
	}
	for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
		keys[i], keys[j] = keys[j], keys[i]
	}
	fmt.Println(keys)
	for _, key := range keys {
		err := VKeyUp(key)
		if err != nil {
			return fmt.Errorf("failed to release key '%s': %v", fmt.Sprint(key), err)
		}
		time.Sleep(interval)
	}
	return nil
}
