package cli

import (
    "flag"
    "fmt"
)

func runUnlink(args []string, g Global) int {
    fs := flag.NewFlagSet("unlink", flag.ContinueOnError)
    if err := fs.Parse(args); err != nil { return 2 }
    // TODO: delete stored tokens (T11)
    fmt.Println("(stub) unlinked tokens (local only)")
    return 0
}

