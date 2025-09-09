package auth

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type linkCompleteRequest struct {
    Code   string        `json:"code"`
    Device *linkDevice   `json:"device,omitempty"`
}
type linkDevice struct { Name string `json:"name"` }

type linkCompleteResponse struct {
    AccessToken        string `json:"access_token"`
    RefreshToken       string `json:"refresh_token"`
    ExpiresIn          int64  `json:"expires_in"`            // seconds for access
    RefreshExpiresIn   int64  `json:"refresh_expires_in"`    // optional seconds
}

// CompleteLink calls POST /link/complete on the API endpoint and returns tokens with expiries.
func CompleteLink(ctx context.Context, apiEndpoint, code, deviceName string) (access string, refresh string, accessExp time.Time, refreshExp time.Time, err error) {
    if apiEndpoint == "" { return "","", time.Time{}, time.Time{}, fmt.Errorf("api endpoint required") }
    url := strings.TrimRight(apiEndpoint, "/") + "/link/complete"
    reqBody := linkCompleteRequest{ Code: code }
    if deviceName != "" { reqBody.Device = &linkDevice{Name: deviceName} }
    b, _ := json.Marshal(&reqBody)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return "","", time.Time{}, time.Time{}, err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 {
        return "","", time.Time{}, time.Time{}, fmt.Errorf("link failed: status %d", resp.StatusCode)
    }
    var out linkCompleteResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "","", time.Time{}, time.Time{}, err }
    access = out.AccessToken
    refresh = out.RefreshToken
    if out.ExpiresIn <= 0 { out.ExpiresIn = int64((30*24*time.Hour).Seconds()) }
    if out.RefreshExpiresIn <= 0 { out.RefreshExpiresIn = int64((90*24*time.Hour).Seconds()) }
    now := time.Now()
    accessExp = now.Add(time.Duration(out.ExpiresIn) * time.Second)
    refreshExp = now.Add(time.Duration(out.RefreshExpiresIn) * time.Second)
    return
}

