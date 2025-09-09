package auth

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestCompleteLink_Basic(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost || r.URL.Path != "/link/complete" {
            w.WriteHeader(http.StatusNotFound); return
        }
        _ = json.NewEncoder(w).Encode(map[string]any{
            "access_token": "a",
            "refresh_token": "r",
            "expires_in": 60,
            "refresh_expires_in": 120,
        })
    }))
    defer srv.Close()

    a, r, ax, rx, err := CompleteLink(context.Background(), srv.URL, "123456", "dev")
    if err != nil { t.Fatal(err) }
    if a != "a" || r != "r" { t.Fatalf("unexpected tokens: %s %s", a, r) }
    if ax.Sub(time.Now()) <= 0 || rx.Sub(ax) <= 0 { t.Fatalf("unexpected expiries: %v %v", ax, rx) }
}

