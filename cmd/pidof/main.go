package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hectorchu/proc"
)

func main() {
	p := proc.Cmd("ps", "A").Map(func(s string) *proc.Proc {
		if fs := strings.Fields(s); filepath.Base(fs[4]) == os.Args[1] {
			return proc.Cat(fs[0], "\n")
		}
		return proc.Cat()
	}).Cmd("sort", "-n").Map(func(s string) *proc.Proc {
		return proc.Cat(s, " ")
	}).Cat("\n")
	if _, err := io.Copy(os.Stdout, p); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
