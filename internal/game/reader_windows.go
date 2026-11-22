//go:build windows

// Package game provides read-only process introspection. It never calls WriteProcessMemory.
// The guarded scan strategy is based on the GPL-3.0 TarnishedAtlas reference;
// no binaries or data from that project are part of Aray Minimap.
package game

import (
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"

	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel         = syscall.NewLazyDLL("kernel32.dll")
	createSnapshot = kernel.NewProc("CreateToolhelp32Snapshot")
	procFirst      = kernel.NewProc("Process32FirstW")
	procNext       = kernel.NewProc("Process32NextW")
	moduleFirst    = kernel.NewProc("Module32FirstW")
	openProcess    = kernel.NewProc("OpenProcess")
	readMemory     = kernel.NewProc("ReadProcessMemory")
	closeHandle    = kernel.NewProc("CloseHandle")
)

const invalidHandle = ^uintptr(0)
const (
	processSnapshot = 0x2
	moduleSnapshot  = 0x8 | 0x10
	processRead     = 0x0010 | 0x0400
)

type processEntry struct {
	Size      uint32
	Usage     uint32
	PID       uint32
	_         uintptr
	ModuleID  uint32
	Threads   uint32
	ParentPID uint32
	Priority  int32
	Flags     uint32
	Name      [260]uint16
}
type moduleEntry struct {
	Size         uint32
	ModuleID     uint32
	PID          uint32
	GlobalUsage  uint32
	ProcessUsage uint32
	Base         uintptr
	ModuleSize   uint32
	ModuleHandle uintptr
	Name         [256]uint16
	Path         [260]uint16
}

type Reader struct {
	process uintptr
	menu    uintptr
	gdm     uintptr
	layout  int
	exe     string
	build   string
}

func PID() (uint32, error) {
	h, _, _ := createSnapshot.Call(processSnapshot, 0)
	if h == invalidHandle {
		return 0, errors.New("snapshot de processos indisponível")
	}
	defer closeHandle.Call(h)
	var e processEntry
	e.Size = uint32(unsafe.Sizeof(e))
	ok, _, _ := procFirst.Call(h, uintptr(unsafe.Pointer(&e)))
	for ok != 0 {
		if strings.EqualFold(syscall.UTF16ToString(e.Name[:]), "eldenring.exe") {
			return e.PID, nil
		}
		ok, _, _ = procNext.Call(h, uintptr(unsafe.Pointer(&e)))
	}
	return 0, errors.New("eldenring.exe não está em execução")
}
func NewReader() (r *Reader, err error) {
	pid, err := PID()
	if err != nil {
		return nil, err
	}
	h, _, _ := openProcess.Call(processRead, 0, uintptr(pid))
	if h == 0 {
		return nil, errors.New("sem acesso de leitura ao jogo")
	}
	r = &Reader{process: h, layout: -1}
	defer func() {
		if err != nil {
			closeHandle.Call(h)
		}
	}()
	snap, _, _ := createSnapshot.Call(moduleSnapshot, uintptr(pid))
	if snap == invalidHandle {
		return nil, errors.New("snapshot de módulos indisponível")
	}
	defer closeHandle.Call(snap)
	var m moduleEntry
	m.Size = uint32(unsafe.Sizeof(m))
	ok, _, _ := moduleFirst.Call(snap, uintptr(unsafe.Pointer(&m)))
	if ok == 0 || !strings.EqualFold(syscall.UTF16ToString(m.Name[:]), "eldenring.exe") {
		return nil, errors.New("módulo do Elden Ring não encontrado")
	}
	r.exe = syscall.UTF16ToString(m.Path[:])
	pf, err := pe.Open(r.exe)
	if err != nil {
		return nil, fmt.Errorf("ler PE: %w", err)
	}
	defer pf.Close()
	textSec, dataSec := pf.Section(".text"), pf.Section(".data")
	if textSec == nil || dataSec == nil || textSec.VirtualSize == 0 || textSec.VirtualSize > 130<<20 {
		return nil, errors.New("layout PE não suportado")
	}
	code := make([]byte, textSec.VirtualSize)
	if !r.read(m.Base+uintptr(textSec.VirtualAddress), code) {
		return nil, errors.New("falha ao ler código do jogo")
	}
	r.build = fmt.Sprintf("%s (%d bytes)", r.exe, m.ModuleSize)
	menu := findGlobal(code, []byte{0x48, 0x8b, 0x0d, 0, 0, 0, 0, 0x48, 0x8b, 0x49, 0x08, 0xe8, 0, 0, 0, 0, 0x48, 0x8b, 0xd0, 0x48, 0x8b, 0xce, 0xe8}, []int{7, 8, 9, 10, 11, 16, 17, 18, 19, 20, 21, 22}, textSec.VirtualAddress, dataSec)
	gdm := findGlobal(code, []byte{0x48, 0x8b, 0x05, 0, 0, 0, 0, 0x48, 0x85, 0xc0, 0x74, 0x05, 0x48, 0x8b, 0x40, 0x58, 0xc3, 0xc3}, []int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, textSec.VirtualAddress, dataSec)
	if menu == 0 || gdm == 0 {
		return nil, errors.New("assinaturas não encontradas ou ambíguas nesta build")
	}
	r.menu = m.Base + uintptr(menu)
	r.gdm = m.Base + uintptr(gdm)
	return r, nil
}
func (r *Reader) Close() {
	if r != nil && r.process != 0 {
		closeHandle.Call(r.process)
		r.process = 0
	}
}
func (r *Reader) Build() string { return r.build }
func (r *Reader) read(addr uintptr, out []byte) bool {
	if addr < 0x10000 || len(out) == 0 {
		return false
	}
	var n uintptr
	v, _, _ := readMemory.Call(r.process, addr, uintptr(unsafe.Pointer(&out[0])), uintptr(len(out)), uintptr(unsafe.Pointer(&n)))
	return v != 0 && n == uintptr(len(out))
}
func (r *Reader) ptr(addr uintptr) uintptr {
	var b [8]byte
	if !r.read(addr, b[:]) {
		return 0
	}
	p := uintptr(binary.LittleEndian.Uint64(b[:]))
	if p < 0x10000 || p > 0x7FFFFFFFFFFF {
		return 0
	}
	return p
}
func findGlobal(code, pattern []byte, required []int, rva uint32, data *pe.Section) uint32 {
	var result uint32
	for i := 0; i+len(pattern) <= len(code); i++ {
		if code[i] != pattern[0] || code[i+1] != pattern[1] || code[i+2] != pattern[2] {
			continue
		}
		same := true
		for _, j := range required {
			if code[i+j] != pattern[j] {
				same = false
				break
			}
		}
		if !same {
			continue
		}
		relative := int32(binary.LittleEndian.Uint32(code[i+3 : i+7]))
		target := int64(rva) + int64(i) + 7 + int64(relative)
		if target < int64(data.VirtualAddress) || target+8 > int64(data.VirtualAddress)+int64(data.VirtualSize) {
			continue
		}
		if result != 0 && result != uint32(target) {
			return 0
		}
		result = uint32(target)
	}
	return result
}

var layouts = []struct{ owner, view, loading uintptr }{{0x80, 0x250, 0x730}, {0x80, 0x248, 0x728}, {0x78, 0x248, 0x718}}

func (r *Reader) Sample() RawSample {
	waiting := RawSample{Status: "Aguardando posição válida"}
	menu, gdm := r.ptr(r.menu), r.ptr(r.gdm)
	if menu == 0 || gdm == 0 || r.ptr(gdm+0x58) == 0 {
		return waiting
	}
	for idx, layout := range layouts {
		if r.layout >= 0 && idx != r.layout {
			continue
		}
		var loading [2]byte
		if !r.read(menu+layout.loading, loading[:]) || binary.LittleEndian.Uint16(loading[:]) != 0 {
			continue
		}
		owner := r.ptr(menu + layout.owner)
		if owner == 0 {
			continue
		}
		view := r.ptr(owner + layout.view)
		if view == 0 {
			continue
		}
		var raw [36]byte
		if !r.read(view+0x14, raw[:]) || menu != r.ptr(r.menu) || gdm != r.ptr(r.gdm) || owner != r.ptr(menu+layout.owner) {
			continue
		}
		if result, ok := decode(raw[:]); ok {
			r.layout = idx
			return result
		}
	}
	return waiting
}
func decode(raw []byte) (RawSample, bool) { return ParseView(raw) }
func (r *Reader) Path() string            { return r.exe }
func Exists(path string) bool             { _, e := os.Stat(path); return e == nil }
