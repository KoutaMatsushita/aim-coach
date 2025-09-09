package cli

import (
    "flag"
    "fmt"
)

func runLink(args []string, g Global) int {
    fs := flag.NewFlagSet("link", flag.ContinueOnError)
    var code, device string
    fs.StringVar(&code, "code", "", "6-digit pairing code from /link")
    fs.StringVar(&device, "device", "", "Optional device name")
    if err := fs.Parse(args); err != nil { return 2 }
    if code == "" { fmt.Println("--code is required"); fs.Usage(); return 2 }
    // TODO: call /link/complete and store tokens (T4)
    fmt.Printf("(stub) linking with code=%s device=%s\n", code, device)
    return 0
}

