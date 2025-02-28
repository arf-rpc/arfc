package common

import (
	"fmt"
	"strings"
)

type Writer struct {
	lines       []string
	indentLevel int
	needsIndent bool
}

func (w *Writer) IncreaseIndent() { w.indentLevel++ }
func (w *Writer) DecreaseIndent() { w.indentLevel-- }
func (w *Writer) indent() {
	if !w.needsIndent {
		return
	}
	w.lines = append(w.lines, strings.Repeat("  ", w.indentLevel))
	w.needsIndent = false
}

func (w *Writer) Writef(format string, args ...any) {
	w.indent()
	w.lines = append(w.lines, fmt.Sprintf(format, args...))
	if strings.HasSuffix(format, "\n") {
		w.needsIndent = true
	}
}

func (w *Writer) Writelnf(format string, args ...any) {
	w.indent()
	w.lines = append(w.lines, fmt.Sprintf(format, args...), "\n")
	w.needsIndent = true
}

func (w *Writer) Break() {
	w.lines = append(w.lines, "\n")
	w.needsIndent = true
}

func (w *Writer) String() string {
	return strings.Join(w.lines, "")
}

func (w *Writer) Merge(o *Writer) {
	w.lines = append(o.lines, w.lines...)
}
