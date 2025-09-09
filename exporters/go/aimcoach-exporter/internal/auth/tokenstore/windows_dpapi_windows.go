//go:build windows

package tokenstore

import (
    "encoding/json"
    "errors"
    "syscall"
    "unsafe"
    "os"
    "path/filepath"
)

type dpapiStore struct{ path string }

func openDPAPIStore() (Store, error) {
    dir, err := os.UserConfigDir()
    if err != nil || dir == "" { dir = "." }
    p := filepath.Join(dir, "aim-coach", "tokens.dpapi")
    if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil { return nil, err }
    return &dpapiStore{path: p}, nil
}

func (s *dpapiStore) Get(key string) (Token, error) {
    m, err := s.load()
    if err != nil { return Token{}, err }
    t, ok := m[key]
    if !ok { return Token{}, ErrNotFound }
    return t, nil
}

func (s *dpapiStore) Set(key string, tok Token) error {
    m, err := s.load()
    if err != nil { return err }
    m[key] = tok
    return s.save(m)
}

func (s *dpapiStore) Delete(key string) error {
    m, err := s.load()
    if err != nil { return err }
    if _, ok := m[key]; !ok { return ErrNotFound }
    delete(m, key)
    return s.save(m)
}

func (s *dpapiStore) load() (map[string]Token, error) {
    b, err := os.ReadFile(s.path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) { return map[string]Token{}, nil }
        return nil, err
    }
    plain, err := dpapiUnprotect(b)
    if err != nil { return nil, err }
    var m map[string]Token
    if err := json.Unmarshal(plain, &m); err != nil { return nil, err }
    if m == nil { m = map[string]Token{} }
    return m, nil
}

func (s *dpapiStore) save(m map[string]Token) error {
    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil { return err }
    enc, err := dpapiProtect(b)
    if err != nil { return err }
    return os.WriteFile(s.path, enc, 0600)
}

// Minimal DPAPI wrappers (user scope)
var (
    modcrypt32              = syscall.NewLazyDLL("Crypt32.dll")
    procCryptProtectData    = modcrypt32.NewProc("CryptProtectData")
    procCryptUnprotectData  = modcrypt32.NewProc("CryptUnprotectData")
)

type dataBlob struct {
    cbData uint32
    pbData *byte
}

func bytesToBlob(b []byte) dataBlob {
    if len(b) == 0 {
        return dataBlob{}
    }
    return dataBlob{cbData: uint32(len(b)), pbData: &b[0]}
}

func blobToBytes(blob dataBlob) []byte {
    if blob.cbData == 0 { return nil }
    return unsafe.Slice(blob.pbData, blob.cbData)
}

func dpapiProtect(plain []byte) ([]byte, error) {
    in := bytesToBlob(plain)
    var out dataBlob
    r, _, err := procCryptProtectData.Call(
        uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&out)),
    )
    if r == 0 { return nil, err }
    defer syscall.LocalFree(syscall.Handle(uintptr(unsafe.Pointer(out.pbData))))
    return append([]byte(nil), blobToBytes(out)...), nil
}

func dpapiUnprotect(enc []byte) ([]byte, error) {
    in := bytesToBlob(enc)
    var out dataBlob
    r, _, err := procCryptUnprotectData.Call(
        uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&out)),
    )
    if r == 0 { return nil, err }
    defer syscall.LocalFree(syscall.Handle(uintptr(unsafe.Pointer(out.pbData))))
    return append([]byte(nil), blobToBytes(out)...), nil
}

