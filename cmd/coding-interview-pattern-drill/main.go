package main

import (
	"fmt"
	"os"

	"github.com/conallob/coding-interview-pattern-drill/cli"
	"github.com/conallob/coding-interview-pattern-drill/server"
	"github.com/conallob/coding-interview-pattern-drill/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			server.RunLogin(os.Args[2:])
		case "--version", "-version", "version":
			fmt.Println(version.Version)
			return
		case "serve":
			server.Run(os.Args[2:])
			return
		}
	}
	cli.Run(os.Args[1:])
}
