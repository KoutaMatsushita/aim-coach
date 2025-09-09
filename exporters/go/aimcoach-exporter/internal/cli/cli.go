package cli

import (
    "flag"
    "fmt"
    "os"
)

type Global struct {
    APIEndpoint string
    APIToken    string
    LogLevel    string
}

func Run(argv []string) int {
    if len(argv) < 2 {
        printRootHelp()
        return 2
    }

    // Global flags (before subcommand)
    globals := Global{}
    root := flag.NewFlagSet("aimcoach-exporter", flag.ContinueOnError)
    root.StringVar(&globals.APIEndpoint, "api-endpoint", getenv("API_ENDPOINT", ""), "API endpoint base URL")
    root.StringVar(&globals.APIToken, "api-token", getenv("API_TOKEN", ""), "API token (Bearer)")
    root.StringVar(&globals.LogLevel, "log-level", getenv("LOG_LEVEL", "info"), "log level: debug|info|warn|error")

    // Parse globals until subcommand
    // e.g. aimcoach-exporter --api-endpoint X export --dry-run
    i := 1
    for ; i < len(argv); i++ {
        if !isFlag(argv[i]) { break }
    }
    if err := root.Parse(argv[1:i]); err != nil {
        return 2
    }
    if i >= len(argv) {
        printRootHelp()
        return 2
    }

    cmd := argv[i]
    args := argv[i+1:]
    switch cmd {
    case "help", "-h", "--help":
        printRootHelp()
        return 0
    case "export":
        return runExport(args, globals)
    case "link":
        return runLink(args, globals)
    case "unlink":
        return runUnlink(args, globals)
    case "status":
        return runStatus(args, globals)
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
        printRootHelp()
        return 2
    }
}

func isFlag(s string) bool {
    return len(s) > 0 && s[0] == '-'
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}

func printRootHelp() {
    fmt.Println("Aim Coach Exporter")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  aimcoach-exporter [--api-endpoint URL] [--api-token TOKEN] [--log-level L] <command> [flags]")
    fmt.Println()
    fmt.Println("Commands:")
    fmt.Println("  export    Read AimLabs/Kovaaks and send or print JSON")
    fmt.Println("  link      Link this device with a 6-digit code")
    fmt.Println("  unlink    Unlink and remove stored tokens")
    fmt.Println("  status    Show effective configuration and token status")
    fmt.Println()
    fmt.Println("Use 'aimcoach-exporter <command> -h' for more information about a command.")
}

