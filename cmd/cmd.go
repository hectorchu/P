package cmd

import (
	"path/filepath"
	"strings"

	"github.com/hectorchu/proc"
)

func FindReplace(pat, dir string) *proc.Proc {
	return proc.Cmd("find", dir, "-type", "f").
		Map(func(file string) *proc.Proc {
			p := proc.Cmd("sed", pat, file)
			if p.Err() == nil {
				p = p.Put(file)
			}
			return p
		})
}

func Pidof(prog string) *proc.Proc {
	return proc.Cmd("ps", "A").Map(func(s string) *proc.Proc {
		if fs := strings.Fields(s); filepath.Base(fs[4]) == prog {
			return proc.Cat(fs[0], "\n")
		}
		return proc.Cat()
	}).Cmd("sort", "-n").Map(func(s string) *proc.Proc {
		return proc.Cat(s, " ")
	}).Cat("\n")
}
