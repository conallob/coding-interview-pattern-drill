package main

import (
	"os"

	"github.com/conallob/coding-interview-pattern-drill/cli"
	"github.com/conallob/coding-interview-pattern-drill/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		server.Run(os.Args[2:])
		return
	}
	cli.Run(os.Args[1:])
}
