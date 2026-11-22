//go:build windows && araydll

// SPDX-License-Identifier: GPL-3.0-only
// Aray Minimap: DLL Go carregada diretamente pelo Elden Mod Loader.
// O bootstrap segue o modelo de inicializacao validado no TarnishedAtlas
// (GPLv3), sem executar processos auxiliares.
package main

/*
#include <stdint.h>
*/
import "C"

import (
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "runtime/debug"
    "strings"
    "time"

    "github.com/llongaray/minimap-eldenring/internal/core"
)

func init() {
    insideGameDLL = true
    // Inicia somente depois do carregamento do runtime Go. Nao inicia EXE
    // nem intercepta teclado, DirectX ou modifica memoria do jogo.
    go nativeDLLStartup()
}

//export ArayMinimapInterfaceVersion
func ArayMinimapInterfaceVersion() C.int { return 2 }

func nativeDLLStartup() {
    defer func() {
        if e := recover(); e != nil {
            dllLogError(fmt.Errorf("panic no minimapa: %v\n%s", e, debug.Stack()))
        }
    }()
    // Permite que o Elden Mod Loader conclua o carregamento da DLL.
    time.Sleep(1500 * time.Millisecond)
    exe, err := os.Executable()
    if err != nil { dllLogError(err); return }
    if !strings.EqualFold(filepath.Base(exe), "eldenring.exe") { return }
    dataDir := filepath.Join(filepath.Dir(exe), "mods", "ArayMinimap", "data")
    if err := os.MkdirAll(dataDir, 0700); err != nil { dllLogError(err); return }
    if f, err := os.OpenFile(filepath.Join(dataDir, "arayminimap.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err == nil {
        defer f.Close()
        // Log isolado: nao altera o logger de terceiros que compartilhem processo.
        log.SetOutput(io.MultiWriter(os.Stderr, f))
    }
    log.Printf("Aray Minimap DLL direta | jogo=%s | dados=%s", exe, dataDir)
    store, err := core.NewStore(filepath.Join(dataDir, "config.json"))
    if err != nil { dllLogError(err); return }
    path := filepath.Join(dataDir, "markers.json")
    if _, err := os.Stat(path); os.IsNotExist(err) { path = filepath.Join(dataDir, "markers.js") }
    markers, err := core.LoadMarkers(path)
    if err != nil { log.Printf("Marcadores indisponiveis: %v", err) }
    log.Printf("Inicializando janelas Win32 dentro do jogo | marcadores=%d", len(markers))
    if err := run(dataDir, store, markers); err != nil { dllLogError(err); return }
    log.Printf("Janela Win32 encerrada")
}

func dllLogError(err error) {
    log.Printf("[ARAY DLL][ERRO] %v", err)
}
