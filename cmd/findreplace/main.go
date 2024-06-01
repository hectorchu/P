package main

import (
	"fmt"
	"os"

	"github.com/hectorchu/proc"
)

func main() {
	p := proc.Cat().
		Cmd("find", os.Args[2], "-type", "f").
		Map(func(file string) *proc.Proc {
			p := proc.Cat().Cmd("sed", os.Args[1], file)
			if p.Err() == nil {
				p = p.Cmd("tee", file)
			}
			return p
		})
	if p.Err() != nil {
		fmt.Fprintln(os.Stderr, p.Err())
	}
}
