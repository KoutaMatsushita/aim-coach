package kovaaks

import (
    "os"
    "path/filepath"
    "strings"

    "github.com/example/aimcoach-exporter/internal/model"
)

// ProcessDir はディレクトリ内の Stats.csv を走査し正規化する
func ProcessDir(dir string) ([]model.NormalizedRecord, error) {
    entries, err := os.ReadDir(dir)
    if err != nil { return nil, err }
    var out []model.NormalizedRecord
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        if strings.HasSuffix(name, " Stats.csv") {
            path := filepath.Join(dir, name)
            recs, err := model.NormalizeKovaaksCSV(path)
            if err != nil { return nil, err }
            out = append(out, recs...)
        }
    }
    return out, nil
}

