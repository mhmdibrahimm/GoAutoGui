//go:build windows

package windows

import (
	"github.com/zzl/go-win32api/v2/win32"
)

func init() {
	// opt into true‑pixel metrics
	win32.SetProcessDPIAware()
}
