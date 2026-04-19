package clihelp

import (
	"fmt"
	"strings"
)

type Command struct {
	Name        string
	Description string
}

type Document struct {
	Binary         string
	Description    string
	DefaultCommand string
	Commands       []Command
}

func Render(doc Document) string {
	var builder strings.Builder

	binary := strings.TrimSpace(doc.Binary)
	if binary == "" {
		binary = "app"
	}

	builder.WriteString(binary)
	if description := strings.TrimSpace(doc.Description); description != "" {
		builder.WriteString(" - ")
		builder.WriteString(description)
	}
	builder.WriteString("\n\nUsage:\n")
	builder.WriteString(fmt.Sprintf("  %s <command>\n", binary))

	if defaultCommand := strings.TrimSpace(doc.DefaultCommand); defaultCommand != "" {
		builder.WriteString(fmt.Sprintf("\nDefault command:\n  %s\n", defaultCommand))
	}

	if len(doc.Commands) > 0 {
		builder.WriteString("\nCommands:\n")
		for _, command := range doc.Commands {
			builder.WriteString(fmt.Sprintf("  %-22s %s\n", strings.TrimSpace(command.Name), strings.TrimSpace(command.Description)))
		}
	}

	builder.WriteString(fmt.Sprintf("\nRun `%s help` to print this list again.\n", binary))
	return builder.String()
}
