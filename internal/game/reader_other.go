//go:build !windows

package game

import "errors"

type Reader struct{}

func PID() (uint32, error)        { return 0, errors.New("somente Windows") }
func NewReader() (*Reader, error) { return nil, errors.New("somente Windows") }
func (*Reader) Close()            {}
func (*Reader) Sample() RawSample { return RawSample{Status: "somente Windows"} }
func (*Reader) Build() string     { return "" }
func (*Reader) Path() string      { return "" }
