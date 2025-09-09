package tokenstore

import (
    "errors"
    "runtime"
    "time"
)

var ErrNotFound = errors.New("token not found")

type Token struct {
    AccessToken   string    `json:"access_token"`
    RefreshToken  string    `json:"refresh_token"`
    AccessExpiry  time.Time `json:"access_expiry"`
    RefreshExpiry time.Time `json:"refresh_expiry"`
}

type Store interface {
    Get(key string) (Token, error)
    Set(key string, tok Token) error
    Delete(key string) error
}

// Open selects a backend by name. For now, only "auto" and "file" are supported cross‑platform.
func Open(kind string) (Store, error) {
    switch kind {
    case "", "auto":
        if runtime.GOOS == "windows" {
            if st, err := openDPAPIStore(); err == nil {
                return st, nil
            }
        }
        return openFileStore("")
    case "file":
        return openFileStore("")
    case "dpapi":
        return openDPAPIStore()
    default:
        return nil, errors.New("unsupported token store: " + kind)
    }
}
