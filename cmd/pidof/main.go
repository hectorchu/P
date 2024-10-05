package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hectorchu/proc/cmd"
)

func main() {
	if _, err := io.Copy(os.Stdout, cmd.Pidof(os.Args[1])); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
