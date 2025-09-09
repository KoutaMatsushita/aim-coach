package tokenstore

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
)

type fileStore struct{ path string }

type tokenFile struct {
    Entries map[string]Token `json:"entries"`
}

func openFileStore(path string) (Store, error) {
    if path == "" {
        dir, err := os.UserConfigDir()
        if err != nil || dir == "" {
            // Fallback: $HOME/.config
            if home, herr := os.UserHomeDir(); herr == nil && home != "" {
                dir = filepath.Join(home, ".config")
            } else {
                // Last resort: current working directory
                dir = "."
            }
        }
        path = filepath.Join(dir, "aim-coach", "tokens.json")
    }
    if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
        // Fallback to CWD if permission denied (sandboxed envs)
        cwd, _ := os.Getwd()
        path = filepath.Join(cwd, ".aimcoach", "tokens.json")
        _ = os.MkdirAll(filepath.Dir(path), 0700)
    }
    return &fileStore{path: path}, nil
}

func (s *fileStore) load() (*tokenFile, error) {
    b, err := os.ReadFile(s.path)
    if errors.Is(err, os.ErrNotExist) { return &tokenFile{Entries: map[string]Token{}}, nil }
    if err != nil { return nil, err }
    var tf tokenFile
    if err := json.Unmarshal(b, &tf); err != nil { return nil, err }
    if tf.Entries == nil { tf.Entries = map[string]Token{} }
    return &tf, nil
}

func (s *fileStore) save(tf *tokenFile) error {
    b, err := json.MarshalIndent(tf, "", "  ")
    if err != nil { return err }
    return os.WriteFile(s.path, b, 0600)
}

func (s *fileStore) Get(key string) (Token, error) {
    tf, err := s.load(); if err != nil { return Token{}, err }
    t, ok := tf.Entries[key]
    if !ok { return Token{}, ErrNotFound }
    return t, nil
}

func (s *fileStore) Set(key string, tok Token) error {
    tf, err := s.load(); if err != nil { return err }
    tf.Entries[key] = tok
    return s.save(tf)
}

func (s *fileStore) Delete(key string) error {
    tf, err := s.load(); if err != nil { return err }
    if _, ok := tf.Entries[key]; !ok { return ErrNotFound }
    delete(tf.Entries, key)
    return s.save(tf)
}
