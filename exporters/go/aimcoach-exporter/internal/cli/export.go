package cli

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/example/aimcoach-exporter/internal/config"
    "github.com/example/aimcoach-exporter/internal/kovaaks"
    "github.com/example/aimcoach-exporter/internal/model"
    "github.com/example/aimcoach-exporter/internal/send"
)

func runExport(args []string, g Global) int {
    fs := flag.NewFlagSet("export", flag.ContinueOnError)
    var aimlabsSQLite, aimlabsCSV, kovaaksCSV string
    var dryRun bool
    var idempotencyKey, source string
    var maxRecords int
    var since string
    fs.StringVar(&aimlabsSQLite, "aimlabs-sqlite", getenv("AIMLABS_SQLITE_PATH", ""), "Path to AimLabs SQLite (Klutch.bytes)")
    fs.StringVar(&aimlabsCSV, "aimlabs-csv-dir", getenv("AIMLABS_CSV_DIR", ""), "AimLabs CSV dir (debug only)")
    fs.StringVar(&kovaaksCSV, "kovaaks-csv-dir", getenv("KOVAAKS_CSV_DIR", ""), "Kovaaks Stats CSV dir")
    fs.BoolVar(&dryRun, "dry-run", false, "Print JSON instead of sending")
    fs.StringVar(&idempotencyKey, "idempotency-key", "", "Optional idempotency key for the batch")
    fs.StringVar(&source, "source", getenv("SOURCE", ""), "Source/device identifier")
    fs.IntVar(&maxRecords, "max-records", 0, "Maximum records to read (0=all)")
    fs.StringVar(&since, "since", "", "RFC3339 start time for incremental read")
    if err := fs.Parse(args); err != nil { return 2 }

    cfg := config.Config{
        APIEndpoint:       g.APIEndpoint,
        APIToken:          g.APIToken,
        AimlabsSQLitePath: aimlabsSQLite,
        AimlabsCSVDir:     aimlabsCSV,
        KovaaksCSVDir:     kovaaksCSV,
        DryRun:            dryRun,
    }

    var all []model.NormalizedRecord
    if cfg.KovaaksCSVDir != "" {
        recs, err := kovaaks.ProcessDir(cfg.KovaaksCSVDir)
        if err != nil { log.Fatalf("kovaaks: %v", err) }
        all = append(all, recs...)
    }
    if cfg.AimlabsSQLitePath != "" {
        recs, err := model.NormalizeAimlabsSQLite(cfg.AimlabsSQLitePath)
        if err != nil { log.Printf("aimlabs(sqlite) warning: %v", err) } else { all = append(all, recs...) }
    } else if cfg.AimlabsCSVDir != "" {
        recs, err := model.NormalizeAimlabsCSVDir(cfg.AimlabsCSVDir)
        if err != nil { log.Printf("aimlabs(csv) warning: %v", err) } else { all = append(all, recs...) }
    }

    if maxRecords > 0 && len(all) > maxRecords { all = all[:maxRecords] }

    if len(all) == 0 {
        fmt.Fprintln(os.Stderr, "no input records; specify --aimlabs-sqlite/--aimlabs-csv-dir and/or --kovaaks-csv-dir")
        return 1
    }

    if cfg.DryRun || cfg.APIEndpoint == "" {
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        for _, r := range all { _ = enc.Encode(r) }
        return 0
    }

    client := send.NewClient(cfg.APIEndpoint)
    grouped := map[string][]model.NormalizedRecord{}
    for _, r := range all { grouped[r.Provider] = append(grouped[r.Provider], r) }
    for provider, recs := range grouped {
        path := "/ingest/" + strings.ToLower(provider)
        if err := client.PostJSONAuth(path, recs); err != nil {
            log.Fatalf("send %s: %v", provider, err)
        }
        log.Printf("sent %d records to %s", len(recs), path)
    }
    return 0
}
