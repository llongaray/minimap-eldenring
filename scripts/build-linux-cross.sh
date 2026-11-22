#!/usr/bin/env bash
# Development-only cross build: builds the CRT-free bridge DLL using a generated
# kernel32 import library for linker resolution; only the REAL Windows kernel32
# is loaded at runtime. No substitute kernel32 DLL is distributed.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p build dist
for cmd in clang-cl lld-link go; do command -v "$cmd" >/dev/null || { echo "Missing $cmd" >&2; exit 2; }; done
cat > build/kernel32.def <<'DEF'
LIBRARY KERNEL32.dll
EXPORTS
 GetModuleFileNameW
 CreateThread
 CreateProcessW
 CloseHandle
 DisableThreadLibraryCalls
DEF
cat > build/kernel32-stubs.c <<'C'
void GetModuleFileNameW(void){}
void CreateThread(void){}
void CreateProcessW(void){}
void CloseHandle(void){}
void DisableThreadLibraryCalls(void){}
C
clang-cl /nologo /c /GS- /Fo:build/kernel32-stubs.obj build/kernel32-stubs.c
lld-link /nologo /nodefaultlib /dll /machine:x64 /subsystem:windows /noentry /def:build/kernel32.def /implib:build/kernel32.lib /out:build/kernel32-fake.dll build/kernel32-stubs.obj
clang-cl /nologo /c /GS- /O2 /Fo:build/ArayMinimap.obj src/launcher/ArayMinimap.c
lld-link /nologo /nodefaultlib /dll /machine:x64 /subsystem:windows /entry:DllMain /out:dist/ArayMinimap.dll build/ArayMinimap.obj build/kernel32.lib
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H=windowsgui' -o dist/ArayMinimap.exe ./cmd/aray-minimap
go test ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...
file dist/ArayMinimap.dll dist/ArayMinimap.exe
