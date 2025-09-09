package cli

import (
    "flag"
    "fmt"
)

func runStatus(args []string, g Global) int {
    fs := flag.NewFlagSet("status", flag.ContinueOnError)
    if err := fs.Parse(args); err != nil { return 2 }
    // TODO: load token store and print expiry (T3)
    fmt.Println("Aim Coach Exporter — status")
    fmt.Printf("api-endpoint: %s\n", g.APIEndpoint)
    fmt.Printf("log-level:   %s\n", g.LogLevel)
    fmt.Println("token:       (not implemented yet)")
    return 0
}

