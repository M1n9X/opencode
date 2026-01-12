package util

import (
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/term"
)

// CopyOSC52 writes an OSC52 clipboard sequence to stdout.
// Many modern terminals (kitty, wezterm, iTerm2) will copy the payload to the system clipboard.
func CopyOSC52(text string) {
	if text == "" {
		return
	}
	// Avoid emitting control bytes when not a TTY or TERM=dumb
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	termEnv := os.Getenv("TERM")
	if termEnv == "" || termEnv == "dumb" {
		return
	}
	payload := base64.StdEncoding.EncodeToString([]byte(text))
	seq := fmt.Sprintf("\x1b]52;c;%s\x07", payload)
	fmt.Fprint(os.Stdout, seq)
}
