package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/llongaray/minimap-eldenring/internal/core"
	"github.com/llongaray/minimap-eldenring/internal/game"
)

func main() {
	dataArg := flag.String("data", "", "Diretório de dados locais")
	check := flag.Bool("check", false, "Verifica configurações e assets sem abrir o jogo")
	diagnose := flag.Bool("diagnose", false, "Lê uma amostra de posição do processo para diagnóstico offline")
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
	if err := os.MkdirAll(dir, 0700); err != nil {
		fatal(err)
	}
	logfile := filepath.Join(dir, "arayminimap.log")
	if f, e := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); e == nil {
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
		log.Printf("Aray Minimap 0.1.0-preview.1 | system=%s | data=%s", runtime.GOOS, dir)
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
		log.Printf("Dados do mapa indisponíveis: %v", err)
		markers = nil
	}
	if *diagnose {
		if runtime.GOOS != "windows" {
			fatal(fmt.Errorf("diagnose requer Windows"))
		}
		reader, e := game.NewReader()
		if e != nil {
			fatal(e)
		}
		defer reader.Close()
		fmt.Printf("BUILD: %s\nAMOSTRA: %+v\n", reader.Build(), reader.Sample())
		return
	}
	if *check {
		fmt.Printf("CONFIG_OK: %s\nMARCADORES: %d\nCAMADA: %s\n", dir, len(markers), store.Get().Layer)
		return
	}
	if err = run(dir, store, markers); err != nil {
		fatal(err)
	}
}
func fatal(err error) { log.Printf("ERRO: %v", err); os.Exit(1) }
