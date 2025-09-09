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

type refreshResponse struct {
    AccessToken      string `json:"access_token"`
    RefreshToken     string `json:"refresh_token,omitempty"`
    ExpiresIn        int64  `json:"expires_in"`
    RefreshExpiresIn int64  `json:"refresh_expires_in,omitempty"`
}

// RefreshTokens calls POST /token/refresh using the refresh token for auth.
func RefreshTokens(ctx context.Context, apiEndpoint, refreshToken string) (access string, newRefresh string, accessExp time.Time, refreshExp time.Time, err error) {
    if apiEndpoint == "" || refreshToken == "" { return "","", time.Time{}, time.Time{}, fmt.Errorf("endpoint and refresh token required") }
    url := strings.TrimRight(apiEndpoint, "/") + "/token/refresh"
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
    req.Header.Set("Authorization", "Bearer "+refreshToken)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return "","", time.Time{}, time.Time{}, err }
    defer resp.Body.Close()
    if resp.StatusCode/100 != 2 {
        return "","", time.Time{}, time.Time{}, fmt.Errorf("refresh failed: status %d", resp.StatusCode)
    }
    var out refreshResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "","", time.Time{}, time.Time{}, err }
    access = out.AccessToken
    if out.RefreshToken != "" { newRefresh = out.RefreshToken } else { newRefresh = refreshToken }
    if out.ExpiresIn <= 0 { out.ExpiresIn = int64((30*24*time.Hour).Seconds()) }
    if out.RefreshExpiresIn <= 0 { out.RefreshExpiresIn = int64((90*24*time.Hour).Seconds()) }
    now := time.Now()
    accessExp = now.Add(time.Duration(out.ExpiresIn) * time.Second)
    refreshExp = now.Add(time.Duration(out.RefreshExpiresIn) * time.Second)
    return
}

