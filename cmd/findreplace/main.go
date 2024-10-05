package main

import (
	"fmt"
	"os"

	"github.com/hectorchu/proc/cmd"
)

func main() {
	if err := cmd.FindReplace(os.Args[1], os.Args[2]).Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
