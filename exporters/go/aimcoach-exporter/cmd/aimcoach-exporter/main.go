package main

import (
    "os"

    "github.com/example/aimcoach-exporter/internal/cli"
)

func main() {
    os.Exit(cli.Run(os.Args))
}
