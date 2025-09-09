package cli

import (
    "flag"
    "fmt"
    "time"

    "github.com/example/aimcoach-exporter/internal/config"
    "github.com/example/aimcoach-exporter/internal/auth/tokenstore"
    logx "github.com/example/aimcoach-exporter/internal/log"
)

func runStatus(args []string, g Global) int {
    fs := flag.NewFlagSet("status", flag.ContinueOnError)
    if err := fs.Parse(args); err != nil { return 2 }
    logx.SetLevel(g.LogLevel)
    cfg := config.Resolve(config.Config{ APIEndpoint: g.APIEndpoint })

    fmt.Println("Aim Coach Exporter — status")
    fmt.Printf("api-endpoint: %s\n", cfg.APIEndpoint)
    fmt.Printf("log-level:   %s\n", g.LogLevel)
    fmt.Printf("time:        %s\n", time.Now().Format(time.RFC3339))
    // Token state
    store, _ := tokenstore.Open("auto")
    if store != nil {
        if tok, err := store.Get("default"); err == nil {
            fmt.Printf("token:       access exp=%s refresh exp=%s\n", tok.AccessExpiry.Format(time.RFC3339), tok.RefreshExpiry.Format(time.RFC3339))
        } else {
            fmt.Println("token:       (no token stored)")
        }
    } else {
        fmt.Println("token:       (store unavailable)")
    }
    logx.Infof("status checked", map[string]any{"endpoint": cfg.APIEndpoint})
    return 0
}
