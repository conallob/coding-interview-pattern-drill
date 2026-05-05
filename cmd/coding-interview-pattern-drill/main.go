package main

import (
	"os"

	"github.com/conallob/coding-interview-pop-quiz/internal/cli"
	"github.com/conallob/coding-interview-pop-quiz/internal/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		server.Run(os.Args[2:])
		return
	}
	cli.Run(os.Args[1:])
}
