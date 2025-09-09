package send

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type Client struct {
    base   string
    token  string
    client *http.Client
}

func NewClient(base, token string) *Client {
    return &Client{base: strings.TrimRight(base, "/"), token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) PostJSON(path string, v any) error {
    b, err := json.Marshal(v)
    if err != nil { return err }
    req, err := http.NewRequest(http.MethodPost, c.base+path, bytes.NewReader(b))
    if err != nil { return err }
    req.Header.Set("Content-Type", "application/json")
    if c.token != "" { req.Header.Set("Authorization", "Bearer "+c.token) }
    resp, err := c.client.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 {
        return fmt.Errorf("POST %s: status %d", path, resp.StatusCode)
    }
    return nil
}

