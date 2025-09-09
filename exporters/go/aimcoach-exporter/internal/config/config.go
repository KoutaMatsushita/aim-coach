package config

import "os"

type Config struct {
    APIEndpoint      string
    APIToken         string
    AimlabsSQLitePath string
    AimlabsCSVDir    string
    KovaaksCSVDir    string
    DryRun           bool
}

// Resolve fills zero-values in cfg with environment defaults and returns a copy.
// CLI > ENV の優先順位を担保するため、CLI側で渡された値はそのまま。
func Resolve(in Config) Config {
    out := in
    if out.APIEndpoint == "" { out.APIEndpoint = getenv("API_ENDPOINT", "") }
    if out.APIToken == "" { out.APIToken = getenv("API_TOKEN", "") }
    if out.AimlabsSQLitePath == "" { out.AimlabsSQLitePath = getenv("AIMLABS_SQLITE_PATH", "") }
    if out.AimlabsCSVDir == "" { out.AimlabsCSVDir = getenv("AIMLABS_CSV_DIR", "") }
    if out.KovaaksCSVDir == "" { out.KovaaksCSVDir = getenv("KOVAAKS_CSV_DIR", "") }
    return out
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}
