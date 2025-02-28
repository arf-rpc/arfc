package output

import (
	"fmt"
	"os"
)

func writef(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
}

func Warnf(format string, args ...interface{}) {
	writef(fmt.Sprintf("WARNING: %s", fmt.Sprintf(format, args...)))
}

func Errorf(format string, args ...interface{}) {
	writef(fmt.Sprintf("ERROR: %s", fmt.Sprintf(format, args...)))
	os.Exit(1)
}
