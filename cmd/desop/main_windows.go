//go:build windows

package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "desop:", err)
		os.Exit(1)
	}
}
