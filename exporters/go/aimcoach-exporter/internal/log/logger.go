package logx

import (
    stdlog "log"
    "os"
    "strings"
    "time"
)

type Level int

const (
    Debug Level = iota
    Info
    Warn
    Error
)

var current = Info

func SetLevel(s string) {
    switch strings.ToLower(s) {
    case "debug": current = Debug
    case "info": current = Info
    case "warn": current = Warn
    case "error": current = Error
    }
}

func log(level string, msg string, fields map[string]any) {
    b := &strings.Builder{}
    b.WriteString("{")
    b.WriteString(`"ts":"`+time.Now().UTC().Format(time.RFC3339Nano)+`"`)
    b.WriteString(`,"level":"`+level+`"`)
    b.WriteString(`,"msg":"`+escape(msg)+`"`)
    for k, v := range fields {
        b.WriteString(`,"`+escape(k)+`":"`+escape(toString(v))+`"`)
    }
    b.WriteString("}")
    stdlog.Println(b.String())
}

func toString(v any) string {
    switch x := v.(type) {
    case string:
        return x
    case error:
        return x.Error()
    default:
        return strings.TrimSpace(strings.ReplaceAll(fmtAny(v), "\n", " "))
    }
}

func fmtAny(v any) string { return stdSprintf("%v", v) }

var stdSprintf = func(format string, a ...any) string { return sprintf(format, a...) }
var sprintf = func(format string, a ...any) string { return Sprintf(format, a...) }

// Sprintf is assigned to avoid import cycle with fmt; replaced at init via closure
var Sprintf = func(format string, a ...any) string {
    // minimal fallback without fmt
    return format
}

func Debugf(msg string, fields map[string]any) { if current <= Debug { log("debug", msg, fields) } }
func Infof(msg string, fields map[string]any)  { if current <= Info  { log("info", msg, fields) } }
func Warnf(msg string, fields map[string]any)  { if current <= Warn  { log("warn", msg, fields) } }
func Errorf(msg string, fields map[string]any) { if current <= Error { log("error", msg, fields) } }

func init() {
    // ensure stdout for logs
    stdlog.SetOutput(os.Stdout)
}

func escape(s string) string {
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, `"`, `\\"`)
    return s
}

