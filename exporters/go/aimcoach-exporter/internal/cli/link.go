package cli

import (
    "context"
    "flag"
    "fmt"
    "time"
    auth "github.com/example/aimcoach-exporter/internal/auth"
    "github.com/example/aimcoach-exporter/internal/auth/tokenstore"
)

func runLink(args []string, g Global) int {
    fs := flag.NewFlagSet("link", flag.ContinueOnError)
    var code, device string
    fs.StringVar(&code, "code", "", "6-digit pairing code from /link")
    fs.StringVar(&device, "device", "", "Optional device name")
    if err := fs.Parse(args); err != nil { return 2 }
    if code == "" { fmt.Println("--code is required"); fs.Usage(); return 2 }
    if g.APIEndpoint == "" { fmt.Println("--api-endpoint is required (or API_ENDPOINT)"); return 2 }
    access, refresh, aexp, rexp, err := auth.CompleteLink(context.Background(), g.APIEndpoint, code, device)
    if err != nil { fmt.Printf("link failed: %v\n", err); return 1 }
    st, err := tokenstore.Open("auto")
    if err != nil { fmt.Printf("store error: %v\n", err); return 1 }
    tok := tokenstore.Token{ AccessToken: access, RefreshToken: refresh, AccessExpiry: aexp, RefreshExpiry: rexp }
    if err := st.Set("default", tok); err != nil { fmt.Printf("store error: %v\n", err); return 1 }
    fmt.Printf("linked: device=%s access-exp=%s refresh-exp=%s\n", device, aexp.Format(time.RFC3339), rexp.Format(time.RFC3339))
    return 0
}
