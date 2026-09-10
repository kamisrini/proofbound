package main

import (
	"context"
	"os"

	"github.com/kamisrini/proofbound/kernel/internal/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), "proofbound", os.Args[1:], os.Stdout, os.Stderr))
}
