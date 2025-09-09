package send

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"

    auth "github.com/example/aimcoach-exporter/internal/auth"
    "github.com/example/aimcoach-exporter/internal/auth/tokenstore"
)

type Client struct {
    base   string
    client *http.Client
}

func NewClient(base string) *Client {
    return &Client{base: strings.TrimRight(base, "/"), client: &http.Client{Timeout: 30 * time.Second}}
}

// PostJSONAuth sends JSON with Bearer from tokenstore, auto-refreshing when needed.
func (c *Client) PostJSONAuth(path string, v any) error {
    st, err := tokenstore.Open("auto")
    if err != nil { return err }
    tok, err := st.Get("default")
    if err != nil { return fmt.Errorf("no token; run link first: %w", err) }
    // Pre-flight refresh if expiring within 5 days
    if time.Until(tok.AccessExpiry) <= 5*24*time.Hour {
        if err := c.refreshAndSave(st, &tok); err != nil { return err }
    }
    // Send once, refresh-on-401 retry once
    if err := c.postWithToken(path, v, tok.AccessToken); err == nil {
        return nil
    } else if isUnauthorized(err) {
        if err := c.refreshAndSave(st, &tok); err != nil { return err }
        return c.postWithToken(path, v, tok.AccessToken)
    } else {
        return err
    }
}

func (c *Client) refreshAndSave(st tokenstore.Store, t *tokenstore.Token) error {
    a, r, ax, rx, err := auth.RefreshTokens(context.Background(), c.base, t.RefreshToken)
    if err != nil { return err }
    t.AccessToken, t.RefreshToken, t.AccessExpiry, t.RefreshExpiry = a, r, ax, rx
    return st.Set("default", *t)
}

func (c *Client) postWithToken(path string, v any, access string) error {
    b, err := json.Marshal(v)
    if err != nil { return err }
    req, err := http.NewRequest(http.MethodPost, c.base+path, bytes.NewReader(b))
    if err != nil { return err }
    req.Header.Set("Content-Type", "application/json")
    if access != "" { req.Header.Set("Authorization", "Bearer "+access) }
    resp, err := c.client.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusUnauthorized { return fmt.Errorf("unauthorized") }
    if resp.StatusCode/100 != 2 {
        return fmt.Errorf("POST %s: status %d", path, resp.StatusCode)
    }
    return nil
}

func isUnauthorized(err error) bool { return strings.Contains(err.Error(), "unauthorized") }
