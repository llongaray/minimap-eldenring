//go:build windows && araydll

// SPDX-License-Identifier: GPL-3.0-only
// DLL nativa carregada pelo Elden Mod Loader, sem executavel auxiliar.
package main

/*
#include <stdint.h>
*/
import "C"

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "runtime/debug"
    "strings"
    "time"

    "github.com/llongaray/minimap-eldenring/internal/core"
)

// bootstrapTrace nao depende da configuracao, dos assets ou da pasta do jogo.
// Por isso registra erros que antes ficavam silenciosos no eldenring.exe.
// Log: %TEMP%\ArayMinimap-bootstrap.log
func bootstrapTrace(message string) {
    logPath := filepath.Join(os.TempDir(), "ArayMinimap-bootstrap.log")
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err == nil {
        defer f.Close()
        fmt.Fprintf(f, "%s | PID=%d | %s\n", time.Now().Format(time.RFC3339Nano), os.Getpid(), message)
    }
}

func init() {
    insideGameDLL = true
    bootstrapTrace("init() iniciado - runtime Go operacional")
    // A rotina complexa roda fora da inicializacao da DLL. Nao cria outro processo.
    go nativeDLLStartup()
}

//export ArayMinimapInterfaceVersion
func ArayMinimapInterfaceVersion() C.int { return 3 }

func nativeDLLStartup() {
    bootstrapTrace("nativeDLLStartup() iniciado")
    defer func() {
        if r := recover(); r != nil {
            msg := fmt.Sprintf("PANIC: %v\n%s", r, debug.Stack())
            bootstrapTrace(msg)
            log.Printf("[ARAY DLL] %s", msg)
        }
    }()

    // Dar tempo ao loader de retornar de LoadLibrary, sem depender disso
    // para registrar os eventos de inicializacao.
    time.Sleep(1500 * time.Millisecond)
    exe, err := os.Executable()
    bootstrapTrace(fmt.Sprintf("os.Executable(): %q | erro: %v", exe, err))
    if err != nil {
        bootstrapTrace("ABORTADO: caminho do processo nao foi identificado")
        return
    }
    if !strings.EqualFold(filepath.Base(exe), "eldenring.exe") {
        bootstrapTrace("ABORTADO: processo nao e eldenring.exe")
        return
    }
    dataDir := filepath.Join(filepath.Dir(exe), "mods", "ArayMinimap", "data")
    bootstrapTrace("Diretorio de dados: " + dataDir)
    if err := os.MkdirAll(dataDir, 0700); err != nil {
        bootstrapTrace(fmt.Sprintf("ABORTADO: falha MkdirAll: %v", err))
        return
    }
    f, err := os.OpenFile(filepath.Join(dataDir, "arayminimap.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        bootstrapTrace(fmt.Sprintf("ABORTADO: falha ao abrir log principal: %v", err))
        return
    }
    defer f.Close()
    log.SetOutput(f)
    bootstrapTrace("Log principal aberto com sucesso")
    log.Printf("Aray Minimap DLL direta diagnosticada | jogo=%s | dados=%s", exe, dataDir)

    store, err := core.NewStore(filepath.Join(dataDir, "config.json"))
    if err != nil {
        bootstrapTrace(fmt.Sprintf("ABORTADO: falha configuracao: %v", err))
        log.Printf("[ARAY DLL][ERRO] %v", err)
        return
    }
    markersPath := filepath.Join(dataDir, "markers.json")
    if _, err := os.Stat(markersPath); os.IsNotExist(err) {
        markersPath = filepath.Join(dataDir, "markers.js")
    }
    markers, err := core.LoadMarkers(markersPath)
    if err != nil {
        bootstrapTrace(fmt.Sprintf("Marcadores indisponiveis: %v", err))
        log.Printf("Marcadores indisponiveis: %v", err)
    }
    bootstrapTrace(fmt.Sprintf("Iniciando Win32, marcadores=%d", len(markers)))
    if err := run(dataDir, store, markers); err != nil {
        bootstrapTrace(fmt.Sprintf("ABORTADO: run Win32 retornou erro: %v", err))
        log.Printf("[ARAY DLL][ERRO] %v", err)
        return
    }
    bootstrapTrace("run Win32 terminou")
}
