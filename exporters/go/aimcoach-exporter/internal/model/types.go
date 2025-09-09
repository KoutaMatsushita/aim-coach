package model

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// NormalizedRecord は Workers 側へ送る標準化済みのレコード
type NormalizedRecord struct {
    Provider string                 `json:"provider"`
    Task     string                 `json:"task"`
    Mode     string                 `json:"mode,omitempty"`
    Map      string                 `json:"map,omitempty"`
    TakenAt  string                 `json:"taken_at"`
    Metrics  map[string]any         `json:"metrics"`
    Score    *float64               `json:"score,omitempty"`
    RawJSON  map[string]any         `json:"raw_json,omitempty"`
}

// ---------------- Kovaaks CSV 正規化 ----------------

var fileNameRe = regexp.MustCompile(`^(?P<task>.+?) - (?P<mode>.+?) - (?P<date>\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}) Stats\.csv$`)

// NormalizeKovaaksCSV は単一のKovaaks Stats CSVを正規化する
func NormalizeKovaaksCSV(path string) ([]NormalizedRecord, error) {
    base := filepath.Base(path)
    m := fileNameRe.FindStringSubmatch(base)
    if m == nil {
        return nil, fmt.Errorf("unexpected kovaaks filename: %s", base)
    }
    task := m[1]
    mode := m[2]
    takenAt, err := parseKovaaksTime(m[3])
    if err != nil {
        return nil, err
    }

    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    r := csv.NewReader(f)
    r.FieldsPerRecord = -1

    // 最初のセクション（Kill行）を集約
    var shotsTotal, hitsTotal int
    var damageDone, damagePossible float64
    var lastHeader []string
    lines := [][]string{}
    for {
        rec, err := r.Read()
        if errors.Is(err, io.EOF) { break }
        if err != nil { return nil, err }
        if len(rec) == 0 { continue }
        // セクション切替: "Weapon,Shots,Hits,..." に到達したら終了
        if rec[0] == "Weapon" {
            break
        }
        if rec[0] == "Kill #" { lastHeader = rec; continue }
        // データ行
        if len(lastHeader) > 0 {
            lines = append(lines, rec)
            shotsTotal += atoiSafe(get(rec, headerIndex(lastHeader, "Shots")))
            hitsTotal += atoiSafe(get(rec, headerIndex(lastHeader, "Hits")))
            damageDone += atofSafe(get(rec, headerIndex(lastHeader, "Damage Done")))
            damagePossible += atofSafe(get(rec, headerIndex(lastHeader, "Damage Possible")))
        }
    }

    metrics := map[string]any{
       "shotsTotal": shotsTotal,
       "hitsTotal":  hitsTotal,
       "accuracy":   ratio(hitsTotal, shotsTotal),
    }
    if damagePossible > 0 {
       metrics["damageDone"] = damageDone
       metrics["damagePossible"] = damagePossible
       metrics["efficiency"] = damageDone / damagePossible
    }

    raw := map[string]any{"file": base, "rows": lines}
    rec := NormalizedRecord{
        Provider: "kovaaks",
        Task: task,
        Mode: mode,
        TakenAt: takenAt.Format(time.RFC3339),
        Metrics: metrics,
        RawJSON: raw,
    }
    return []NormalizedRecord{rec}, nil
}

func parseKovaaksTime(s string) (time.Time, error) {
    // "2025.08.04-23.44.33" → YYYY.MM.DD-HH.MM.SS
    t, err := time.Parse("2006.01.02-15.04.05", s)
    if err != nil { return time.Time{}, err }
    return t.UTC(), nil
}

func headerIndex(header []string, name string) int {
    for i, h := range header {
        if strings.EqualFold(strings.TrimSpace(h), name) { return i }
    }
    return -1
}
func get(rec []string, idx int) string {
    if idx < 0 || idx >= len(rec) { return "" }
    return strings.TrimSpace(rec[idx])
}
func atoiSafe(s string) int { v, _ := strconv.Atoi(strings.TrimSpace(s)); return v }
func atofSafe(s string) float64 { v, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(s, "%")), 64); return v }
func ratio(num, den int) float64 { if den == 0 { return 0 }; return float64(num)/float64(den) }

// ---------------- AimLabs 正規化（スタブ＋CSVフォールバック） ----------------

// NormalizeAimlabsSQLite は Klutch.bytes(SQLite) を走査し既知テーブルを正規化する（実装は後続拡張）
func NormalizeAimlabsSQLite(sqlitePath string) ([]NormalizedRecord, error) {
    // ここでは依存解決不要のためスタブ実装。
    // 後続で modernc.org/sqlite を使った SELECT 実装を追加する。
    return nil, fmt.Errorf("NormalizeAimlabsSQLite: not implemented yet (pending driver)")
}

// NormalizeAimlabsCSVDir はCSVディレクトリから一部代表テーブルを正規化する（開発用）
func NormalizeAimlabsCSVDir(dir string) ([]NormalizedRecord, error) {
    var out []NormalizedRecord
    // 代表例: CircleShotData.csv, CircleTrackData.csv, Composite.csv
    // それぞれ存在すれば読み取る（ヘッダの存在/欠損を許容）
    patterns := []string{"CircleShotData.csv", "CircleTrackData.csv", "Composite.csv"}
    for _, p := range patterns {
        path := filepath.Join(dir, p)
        if _, err := os.Stat(path); err == nil {
            recs, err := normalizeAimlabsGenericCSV(path)
            if err != nil { return nil, err }
            out = append(out, recs...)
        }
    }
    return out, nil
}

func normalizeAimlabsGenericCSV(path string) ([]NormalizedRecord, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    r := csv.NewReader(f)
    r.FieldsPerRecord = -1
    header, err := r.Read()
    if err != nil { return nil, err }
    idx := map[string]int{}
    for i, h := range header { idx[strings.TrimSpace(h)] = i }

    var out []NormalizedRecord
    base := filepath.Base(path)
    task := strings.TrimSuffix(strings.TrimSuffix(base, ".csv"), "Data")
    for {
        rec, err := r.Read()
        if errors.Is(err, io.EOF) { break }
        if err != nil { return nil, err }
        metrics := map[string]any{}
        putNum := func(key string) {
            if j, ok := idx[key]; ok { metrics[key] = atofSafe(get(rec, j)) }
        }
        // よく使うキーのみ抽出（存在すれば）
        for _, k := range []string{"accTotal","KPS","SPK","killTotal","shotsTotal","hitsTotal","missesTotal","targetsTotal","OTR","avgTimeOn","avgTimeOff","score"} {
            putNum(k)
        }
        var scorePtr *float64
        if j, ok := idx["score"]; ok {
            s := atofSafe(get(rec, j))
            scorePtr = &s
        }

        ts := time.Now().UTC() // CSVには必須でないため、あれば使う
        if j, ok := idx["timestamp"]; ok {
            if t, err := parseFlexibleTime(get(rec, j)); err == nil { ts = t }
        }
        raw := map[string]any{"file": base, "row": rec}
        out = append(out, NormalizedRecord{
            Provider: "aimlabs",
            Task: task,
            Mode: get(rec, idx["mode"]),
            Map:  get(rec, idx["map"]),
            TakenAt: ts.Format(time.RFC3339),
            Metrics: metrics,
            Score: scorePtr,
            RawJSON: raw,
        })
    }
    return out, nil
}

func parseFlexibleTime(s string) (time.Time, error) {
    layouts := []string{
        time.RFC3339, "2006-01-02 15:04:05", "2006/01/02 15:04:05",
    }
    for _, l := range layouts {
        if t, err := time.ParseInLocation(l, s, time.Local); err == nil { return t.UTC(), nil }
    }
    return time.Time{}, fmt.Errorf("unrecognized time: %s", s)
}

// Pretty はデバッグ用
func Pretty(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }

