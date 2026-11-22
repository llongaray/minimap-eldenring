//go:build !windows

package main

import (
	"errors"
	"github.com/llongaray/minimap-eldenring/internal/core"
)

func run(_ string, _ *core.Store, _ []core.Marker) error { return errors.New("Windows obrigatório") }
