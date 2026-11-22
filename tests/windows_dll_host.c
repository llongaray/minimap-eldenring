// SPDX-License-Identifier: GPL-3.0-only
// Test harness for a DLL-loaded host, not an Elden Ring emulator.
// This is only used in CI and is NOT distributed as a game component.
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <stdio.h>

int main(void) {
    HMODULE mod = LoadLibraryW(L"mods\\ArayMinimap.dll");
    if (!mod) {
        fprintf(stderr, "ERR: LoadLibrary failed, Win32 error=%lu\n", GetLastError());
        return 10;
    }
    FARPROC version = GetProcAddress(mod, "ArayMinimapInterfaceVersion");
    if (!version) {
        fprintf(stderr, "ERR: export ArayMinimapInterfaceVersion missing. Win32 error=%lu\n", GetLastError());
        return 11;
    }
    fprintf(stdout, "OK: LoadLibrary and export confirmed. Waiting for Go init...\n");
    Sleep(6000);
    // Avoid FreeLibrary: Go c-shared runtimes are not designed for unloading
    // from a process still carrying active Go threads.
    return 0;
}
