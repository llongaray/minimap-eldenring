# Compatibilidade — Aray Minimap 0.2 (DLL direta)

| Cenário | Status |
|---|---|
| Go core/testes | Testes unitários executados no runner |
| DLL c-shared Windows x64 | Compilação verificada exclusivamente pelo GitHub Actions |
| Instalador CMD / PowerShell | Simulação de instalação sem jogo real |
| Elden Mod Loader | Estrutura de DLL em Game/mods; carregamento real pendente |
| Edição instalada do Elden Ring 1.16 | Leitor de posição validado isoladamente em versão anterior, sem HUD |
| DLL carregada dentro do eldenring.exe | **Pendente de teste manual** |
| F6, foco, desenho do HUD e controles | **Pendente de teste manual** |
| Marcação e texturas | Interpretador de 1.861 marcadores e importação de 3.125 tiles já foram testados pelo usuário na versão 0.1 |
| Calibração do player para pixels jw-ofs | **Pendente**; não assumir transformação identidade |
| Steam e não-Steam | Deve ser validado por instalação; não garantido |

## Arquitetura

A versão v0.2 não depende de ArayMinimap.exe. O próprio Game/mods/ArayMinimap.dll contém o runtime Go, a interface Win32 e o leitor de memória. O Elden Mod Loader existente permanece intacto, sem injeção adicional.

O estado `calibration.enabled=false` evita representar localização incorreta como real. O HUD pode mostrar texturas/markers, mas a centralização no personagem precisa de aferição real entre coordenadas do jogo e tiles.

A interface é limitada a Win32/GDI e janela sem bordas; telas exclusivas, múltiplas telas e DPI precisam de teste. Logs: `Game/mods/ArayMinimap/data/arayminimap.log`.

## Recuperação

Com jogo fechado, renomeie apenas `Game/mods/ArayMinimap.dll` para `ArayMinimap.dll.disabled` se houver erro de inicialização. Não apague `dinput8.dll`, o save ou `data/`. Os binários anteriores ficam no backup do instalador.

A presença de uma DLL gerada na CI não é evidência de que a DLL foi carregada corretamente em Elden Ring.
