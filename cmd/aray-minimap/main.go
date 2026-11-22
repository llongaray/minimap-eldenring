package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/llongaray/minimap-eldenring/internal/core"
)

func main() {
	dataArg := flag.String("data", "", "Diretório de dados locais")
	check := flag.Bool("check", false, "Verifica configurações e assets sem abrir o jogo")
	flag.Parse()
	if runtime.GOOS != "windows" && !*check {
		fmt.Fprintln(os.Stderr, "A interface do jogo requer Windows")
		os.Exit(2)
	}
	dir := *dataArg
	if dir == "" {
		// Standard install: DLL alongside an ArayMinimap/ folder containing
		// ArayMinimap.exe and data/. If absent, use user config directory.
		if exe, e := os.Executable(); e == nil {
			candidate := filepath.Join(filepath.Dir(exe), "data")
			if st, se := os.Stat(candidate); se == nil && st.IsDir() {
				dir = candidate
			}
		}
	}
	if dir == "" {
		var e error
		dir, e = os.UserConfigDir()
		if e != nil {
			fatal(e)
		}
		dir = filepath.Join(dir, "ArayMinimap")
	}
	store, err := core.NewStore(filepath.Join(dir, "config.json"))
	if err != nil {
		fatal(err)
	}
	markersPath := filepath.Join(dir, "markers.json")
	if _, err = os.Stat(markersPath); os.IsNotExist(err) {
		markersPath = filepath.Join(dir, "markers.js")
	}
	markers, err := core.LoadMarkers(markersPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Dados do mapa indisponíveis: %v\n", err)
		markers = nil
	}
	if *check {
		fmt.Printf("CONFIG_OK: %s\nMARCADORES: %d\nCAMADA: %s\n", dir, len(markers), store.Get().Layer)
		return
	}
	if err = run(dir, store, markers); err != nil {
		fatal(err)
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "ERRO:", err); os.Exit(1) }
