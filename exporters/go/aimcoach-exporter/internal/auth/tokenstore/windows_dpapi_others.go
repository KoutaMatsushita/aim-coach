//go:build !windows

package tokenstore

import "errors"

func openDPAPIStore() (Store, error) { return nil, errors.New("dpapi unsupported on this OS") }

