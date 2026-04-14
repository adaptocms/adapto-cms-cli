package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Table renders data as a clean ASCII table.
func Table(headers []string, rows [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		r := make(table.Row, len(row))
		for i, cell := range row {
			r[i] = cell
		}
		t.AppendRow(r)
	}

	t.SetStyle(table.Style{
		Name: "clean",
		Box:  table.StyleBoxDefault,
		Format: table.FormatOptions{
			Header: text.FormatUpper,
		},
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: true,
			SeparateHeader:  true,
			SeparateRows:    false,
		},
	})
	t.Render()
}

// KeyValue renders key-value pairs in a simple format.
func KeyValue(pairs [][2]string) {
	maxLen := 0
	for _, p := range pairs {
		if len(p[0]) > maxLen {
			maxLen = len(p[0])
		}
	}
	for _, p := range pairs {
		fmt.Printf("%-*s  %s\n", maxLen, p[0], p[1])
	}
}

// Truncate shortens a string to maxLen with ellipsis.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// StringOrEmpty returns the string value of an interface or empty string.
func StringOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// JoinStrings joins a slice of strings.
func JoinStrings(items []string, sep string) string {
	return strings.Join(items, sep)
}
