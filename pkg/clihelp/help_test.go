package clihelp

import (
	"strings"
	"testing"
)

func TestRenderIncludesDefaultCommandAndCommands(t *testing.T) {
	rendered := Render(Document{
		Binary:         "content-bot.exe",
		Description:    "main runtime",
		DefaultCommand: "run",
		Commands: []Command{
			{Name: "run", Description: "run API and worker"},
			{Name: "help", Description: "show this help"},
		},
	})

	for _, snippet := range []string{
		"content-bot.exe - main runtime",
		"Default command:\n  run",
		"run",
		"help",
		"Run `content-bot.exe help` to print this list again.",
	} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered help to contain %q, got:\n%s", snippet, rendered)
		}
	}
}
