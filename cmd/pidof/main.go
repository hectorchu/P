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
		fs := strings.Fields(s)
		if filepath.Base(fs[4]) != os.Args[1] {
			return proc.Cat()
		}
		return proc.Cmd("echo", fs[0])
	}).Cmd("sort", "-n").Map(func(s string) *proc.Proc {
		return proc.Cmd("echo", "-n", s, "")
	})
	p = proc.Cat(p, proc.Cmd("echo"))
	if _, err := io.Copy(os.Stdout, p); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
