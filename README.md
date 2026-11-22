# Aray Minimap — Elden Ring

![Licença GPLv3](https://img.shields.io/badge/license-GPL--3.0-blue)
![Build Windows](https://github.com/llongaray/minimap-eldenring/actions/workflows/build.yml/badge.svg)

**Estado: v0.2.0-preview.2, experimental. A integração no Elden Ring 1.16 ainda exige teste manual.**

Minimapa e mapa detalhado F6 em **português brasileiro**. Projeto próprio, sem execução de ArayMinimap.exe, sem Map for Goblins e sem modificar o TarnishedAtlas.

## Arquitetura — DLL única

```text
Elden Ring (eldenring.exe)
  └─ Elden Mod Loader (Game/dinput8.dll)
       └─ Game/mods/ArayMinimap.dll [Go c-shared + Win32]
            ├─ Leitura somente da posição na memória
            ├─ HUD de minimapa
            ├─ Mapa detalhado F6
            ├─ Filtros, pesquisa, pinos e progresso
            └─ Game/mods/ArayMinimap/data/
                 ├─ config.json (criado ao salvar)
                 ├─ markers.js (importação local de referência)
                 └─ ctiles/ (texturas locais PNG)
```

A DLL nova implementa a interface e o leitor **dentro do processo Elden Ring**. Não chama `CreateProcess`, não instala segundo executável, não altera o loader `dinput8.dll`, o save nem `regulation.bin`. O modelo de inicialização segue a técnica do TarnishedAtlas (GPLv3), usando uma implementação independente.

## Capacidades

| Recurso | Estado |
|---|---|
| DLL Windows x64 em processo, sem EXE auxiliar | Implementada; CI compila |
| HUD nativo e painel F6 PT-BR | Implementados; comportamento real pendente |
| 1.861 marcadores do site jw-ofs | Interpretador disponível; dados importados pelo usuário |
| 3.125 tiles M00/M01/M10 | Importador local disponível; não incluídos no ZIP |
| Filtros, busca, marcadores fixados e concluídos | Implementados; UI pendente de teste real |
| Leitura de memória sem escrita | Funcional no teste anterior do usuário; integração DLL nova não validada |
| Posicionamento correto em pixels jw-ofs | **Calibração pendente; não inventa coordenadas** |
| Teclas originais preservadas | Projetado sem remapeamentos; requer teste manual |
| Suporte a versões modificadas do jogo | Não garantido; validação por assinatura de executável |

**Aviso:** o estado de CI não substitui teste do jogo. A renderização usa Win32, cuja sobreposição em fullscreen exclusivo pode não funcionar; utilize janela sem bordas. Recomendado apenas offline com anticheat desativado pelo método permitido na instalação do usuário.

## Download pelo CMD

Em uma janela CMD do Windows, com Git e GitHub CLI autenticados:

```cmd
cd /d "E:\Games\ELDEN RING"
git clone https://github.com/llongaray/minimap-eldenring.git
cd /d "E:\Games\ELDEN RING\minimap-eldenring"
gh run list -R llongaray/minimap-eldenring --limit 5
```

Identifique a última execução de CI bem-sucedida após o commit da DLL nativa (v0.2). Substitua **RUN_ID** pelo número real:

```cmd
gh run download RUN_ID -R llongaray/minimap-eldenring -n ArayMinimap-native-dll-preview -D "_ci"
mkdir "_ci\extraido"
tar -xf "_ci\ArayMinimap-v0.2.0-preview.2-win64.zip" -C "_ci\extraido"
if not exist "dist" mkdir "dist"
for /r "_ci\extraido" %F in (ArayMinimap.dll) do @copy /Y "%F" "dist\"
install.cmd "E:\Games\ELDEN RING"
```

Não execute `install.cmd` sem baixar a **nova DLL**. Se já clonou o repo, use `git pull --ff-only` em vez de clonar de novo. `install.cmd` pergunta se deve desativar Map for Goblins, faz backup das DLLs antigas e preserva os dados importados.

## Instalação pelo PowerShell

```powershell
.\install.ps1 -GamePath 'E:\Games\ELDEN RING' -DisableMapForGoblins -NoPrompt
```

Requer `dist/ArayMinimap.dll` da build v0.2 nativa. Alternativamente, os instaladores podem ser usados em qualquer pasta; o caminho de exemplo não é fixo.

### Migração da versão v0.1

A instalação passa a manter somente `Game/mods/ArayMinimap.dll`, enquanto os executáveis antigos `ArayMinimap.exe`, `ArayMinimap-Check.exe` e a DLL anterior são movidos para um diretório de backup. Os marcadores, tiles e progresso em `Game/mods/ArayMinimap/data` permanecem intactos. O jogo deve estar fechado durante a migração. Se a DLL não carregar no jogo, renomeie **somente** `ArayMinimap.dll` para `ArayMinimap.dll.disabled` e reinicie offline.

## Dados opcionais do mapa

O projeto do mapa de referência **não declara autorização explícita para republicação integral dos assets**. Baixe apenas para uso local, de acordo com seus direitos:

```cmd
cd /d "E:\Games\ELDEN RING"
git clone --depth 1 https://github.com/jw-ofs/elden-ring-map.git "_Aray-Referencia-Mapa"
copy /Y "_Aray-Referencia-Mapa\markers.js" "Game\mods\ArayMinimap\data\markers.js"
py -m pip install Pillow
py "minimap-eldenring\scripts\import_tiles.py" --source "_Aray-Referencia-Mapa\ctiles" --destination "Game\mods\ArayMinimap\data\ctiles"
```

O importador converte WebP a PNG, mantendo camadas M00 (superfície), M01 (subterrâneo) e M10 (DLC). O site de origem permanece separado e seus assets não são publicados neste projeto.

## Controles (apenas quando mapa F6 estiver aberto)

- **F6:** abrir ou fechar o mapa.
- **Esc:** fechar.
- **1 / 2 / 3:** camadas.
- **+ / -:** zoom.
- **F:** pesquisa por nome.
- **L:** centralizar no jogador quando calibrado.
- **R:** centralizar mapa.
- **Clique esquerdo:** fixar/desafixar próximo marcador.
- **Clique direito:** marcar concluído/não concluído.

A HUD não captura entrada de mouse. A tecla F6 é lida pelo Windows somente durante o jogo, sem remapeamento de teclas. O desempenho de foco e desenho ainda requer validação manual.

## Diagnóstico

Os registros da DLL são gravados em:

```text
Game/mods/ArayMinimap/data/arayminimap.log
```

O log registra início da DLL, criação de janelas, estados de foco e tecla F6. A ausência de log pode indicar que o Elden Mod Loader não carregou a DLL.

A calibração fica desativada por padrão. É preciso aferir a correspondência da posição real à imagem do mapa (tiles jw-ofs). O leitor não inventa posição quando não há leitura válida. Nunca habilite valores arbitrários sem aferir.

## Compilação

Requer **Go 1.23+ e MinGW-w64 GCC para Windows x64**, com `CGO_ENABLED=1`. Em um terminal Windows com GCC no PATH:

```cmd
go test ./...
go vet ./...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -trimpath -tags araydll -buildmode=c-shared -ldflags="-s -w" -o dist\ArayMinimap.dll .\cmd\aray-minimap
```

O GitHub Actions usa MSYS2 UCRT64 no Windows e publica somente `ArayMinimap-native-dll-preview`, sem release estável até validação no jogo. O antigo bridge em C e o executável auxiliar não fazem mais parte do processo de build.

## Desinstalação

CMD:
```cmd
uninstall.cmd "E:\Games\ELDEN RING"
```
PowerShell:
```powershell
.\uninstall.ps1 -GamePath 'E:\Games\ELDEN RING'
```

Os comandos preservam `dinput8.dll`, o save e os dados em `data/`. Há uma opção PowerShell `-RemoveUserData` que move os dados para backup fora do mod. Restaurar Map for Goblins exige recolocar os arquivos manualmente do backup da migração.

## Verificações necessárias antes da release estável

- [ ] DLL carrega no `eldenring.exe` e não impede iniciar.
- [ ] HUD aparece na janela sem bordas, sem processo auxiliar.
- [ ] F6 abre e fecha com controles normais preservados.
- [ ] Posição real é lida e calibrada em dois pontos.
- [ ] Marcadores, camadas, zoom e tiles renderizam corretamente.
- [ ] Estado dos pinos persiste após reiniciar o jogo.
- [ ] CPU, GPU e memória não apresentam sobrecarga relevante.
- [ ] Desinstalação não modifica outros mods.

## Licenciamento e créditos

**GPL-3.0** para o código Aray Minimap. Referências técnicas: [TarnishedAtlas](https://github.com/) (GPLv3, código recebido do usuário), [Elden Mod Loader](https://github.com/techiew/EldenRingModLoader), [jw-ofs/elden-ring-map](https://github.com/jw-ofs/elden-ring-map). Os assets externos não estão licenciados como parte do Aray Minimap e não são distribuídos.
