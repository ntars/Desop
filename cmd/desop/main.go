//go:build !windows

// Command desop is the Windows desktop beautification suite.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Desop must be built for Windows. See docs/BUILD.md.")
	os.Exit(1)
}
