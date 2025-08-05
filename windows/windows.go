//go:build windows

package windows

import (
	"github.com/zzl/go-win32api/v2/win32"
)

func init() {
	// Set the DPI awareness context to Per Monitor V2 for true pixel metrics
	win32.SetProcessDpiAwarenessContext(win32.DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)
}
