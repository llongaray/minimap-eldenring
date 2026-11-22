# Aray Minimap — Elden Ring

**Status: preview técnico experimental (`0.1.0-preview.1`); ainda não validado dentro do Elden Ring 1.16.**

Aray Minimap é um mod independente para **Elden Ring no Windows x64**, distribuído como uma DLL de integração para **Elden Mod Loader** e um executável auxiliar Win32. A interface é em português brasileiro e contém um minimapa HUD e uma tela ampliada controlada por **F6**. As posições do jogo são lidas **somente em modo leitura** e não são publicadas no minimapa sem uma calibração validada contra os tiles externos.

> **Atenção:** não interprete o estado de compilação ou o sucesso dos testes unitários como garantia de funcionamento no Elden Ring 1.16. A DLL compila, mas ainda requer testes de carregamento, alinhamento de coordenadas, foco de janela e controles no jogo real. Use exclusivamente offline, sem Easy Anti-Cheat.

## Visão geral

- **DLL independente:** `ArayMinimap.dll` carrega via Elden Mod Loader (`Game/mods/`) e lança o auxiliar nativo `ArayMinimap.exe` em thread separada, sem alterar `dinput8.dll`.
- **HUD de minimapa:** janela Win32 sem captura de teclas/mouse, com zoom, posições condicionadas à validação, marcadores e tiles locais quando presentes.
- **F6 — tela ampliada:** alternância de janela com navegação, filtros por categoria, pesquisa textual, marcadores fixados e encontrados, zoom e alternância de camadas.
- **Dados persistentes:** JSON versionado, gravação atômica e backup anterior em `data/config.json`.
- **Dados externos desacoplados:** importação local de `markers.js` e tiles de projetos de terceiros, sem redistribuí-los neste repositório.
- **Leitura defensiva:** `ReadProcessMemory` + validação de assinaturas + rejeição de posições inválidas; nenhuma escrita na memória do jogo.

## Estado dos requisitos

| Recurso | Estado | Observação |
|---|---|---|
| Core, configurações e persistência | Implementado; testes locais passaram | Pode ser testado sem o jogo |
| Aplicativo Win32 x64 | Compilado por cross-build | Execução real no Windows pendente |
| DLL para Elden Mod Loader | Compilada PE32+ | Teste de carregamento real pendente |
| Minimap HUD / janela F6 | Implementação experimental | Necessita testes com jogo, foco e tela |
| Filtros, pesquisa e marcações | Implementados no protótipo nativo | Persistência testada; comportamento Win32 pendente |
| Superfície / subterrâneo / DLC | Seletor implementado | Imagens externas não distribuídas |
| Posição real do personagem | Scanner experimental | **Desativada até calibração** e verificação de compatibilidade |
| Mapas e dados completos jw-ofs | **Não incluídos** | Sem licença de redistribuição explícita do projeto fonte |
| Missões detalhadas e pop-ups de equipamentos | Não implementados | Requer módulo de dados autorizado |
| Zoom, tamanho, opacidade e posição do HUD | Opções no esquema | Zoom/opacidade parciais; posição e rotação pendentes |
| Build/testes em Windows CI | Workflow incluído | Execução no runner será a primeira validação de CI |
| Teste end-to-end no Elden Ring 1.16 | **Pendente** | Necessário para release estável |

## Requisitos

- Windows 10 ou Windows 11 x64.
- Elden Ring no Windows. A versão pretendida é **1.16**; não foi validada em jogo real.
- [Elden Mod Loader](https://www.nexusmods.com/eldenring/mods/117) instalado (DLL `dinput8.dll` e `mod_loader_config.ini`).
- Para compilar: Go 1.23+ e Visual Studio 2022 Build Tools (MSVC + Windows SDK), ou Clang/LLD para o bridge DLL.
- Para importar tiles WebP localmente: Python 3 e Pillow (`python -m pip install Pillow`).

## Download

Obtenha os fontes em <https://github.com/llongaray/minimap-eldenring>. Uma build experimental binária pode ser disponibilizada como artefato de CI/arquivo ZIP de teste. **Não existe release estável validada para download neste momento.** Não confunda código-fonte GitHub com binários prontos.

## Download via Git

```powershell
git clone https://github.com/llongaray/minimap-eldenring.git
cd minimap-eldenring
```

Se `git` não estiver instalado, obtenha-o em <https://git-scm.com/downloads/win>. Os comandos abaixo pressupõem que a compilação produziu `dist/ArayMinimap.dll` e `dist/ArayMinimap.exe`, ou que você extraiu ambos de um pacote experimental verificado.

## Compilação Windows

```powershell
# PowerShell 5.1 ou superior, em terminal de desenvolvedor VS 2022:
go test ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o dist/ArayMinimap.exe ./cmd/aray-minimap
cl /nologo /c /O2 /GS- /Fodist\ArayMinimap.obj src/launcher/ArayMinimap.c
link /nologo /DLL /NODEFAULTLIB /ENTRY:DllMain /OUT:dist\ArayMinimap.dll dist\ArayMinimap.obj kernel32.lib
go build -trimpath -ldflags "-s -w" -o dist/ArayMinimap-Check.exe ./cmd/aray-minimap
```

O comando `cl` depende do ambiente MSVC ativado e pode exigir ajustes conforme a versão do Visual Studio/SDK. A CI usa Visual Studio Build Tools e registra o log da compilação.

## Instalar no PowerShell

```powershell
.\install.ps1
# Para executar sem prompts:
.\install.ps1 -GamePath 'E:\Games\ELDEN RING' -DisableMapForGoblins -NoPrompt
```

Aceita tanto a pasta raiz (`E:\Games\ELDEN RING`) quanto a pasta `Game`. **Não é obrigatório usar esse caminho**; ele é apenas um exemplo de uma instalação personalizada. O instalador detecta o loader, verifica o jogo fechado, faz backup de atualizações e pode desativar `MapForGoblins.dll` com backup. Não sobrescreve o loader nem modifica o save. Também aceita `-MapSource 'C:\meus-dados\jw-map'` para copiar `markers.js` de uma cópia local do site de referência.

## Instalar no CMD

```cmd
install.cmd
rem sem prompt de caminho:
install.cmd "E:\Games\ELDEN RING"
```

O CMD possui validação própria e não chama `install.ps1`. O modo CMD não importa `markers.js` automaticamente. Se desejar, copie manualmente `markers.js` para `Game\mods\ArayMinimap\data\`.

## Estrutura instalada

```text
Game/
├── eldenring.exe
├── dinput8.dll                  (existente; preservado)
├── mod_loader_config.ini       (existente; preservado)
└── mods/
    ├── ArayMinimap.dll          (bridge nativo)
    └── ArayMinimap/
        ├── ArayMinimap.exe      (auxiliar Win32)
        ├── ArayMinimap-Check.exe (diagnóstico CLI)
        └── data/
            ├── config.json     (persistência própria, criada após uso)
            ├── markers.js      (opcional, obtido pelo usuário)
            └── ctiles/         (opcional, obtido e convertido pelo usuário)
```

## Dados do mapa (fontes externas)

Os arquivos gráficos originais e `markers.js` **não são republicados por Aray Minimap**. O projeto [jw-ofs/elden-ring-map](https://github.com/jw-ofs/elden-ring-map) disponibiliza o código de sua interface e oferece instruções de execução local, mas não declara licença que autorize redistribuir todos os seus recursos. O usuário que tiver uma cópia legítima local pode importar, para uso pessoal:

```powershell
# Se você já possui a cópia local da referência:
Copy-Item 'C:\Referencias\elden-ring-map\markers.js' 'E:\Games\ELDEN RING\Game\mods\ArayMinimap\data\markers.js'

# Converter os tiles WebP da cópia do usuário para PNG reconhecido pelo app:
python -m pip install Pillow
python scripts/import_tiles.py --source 'C:\Referencias\elden-ring-map\ctiles' --destination 'E:\Games\ELDEN RING\Game\mods\ArayMinimap\data\ctiles'
```

Os tiles são endereçados como `ctiles/M00/base/{zoom}/{y}/{x}.png`, `M01` e `M10`. A interface pode iniciar sem tiles e sem marcadores, mas **não mostrará um mapa ilustrado completo** nesse caso. O script de conversão não baixa nem distribui arquivos de terceiros.

## Uso e teclas

| Comando | Ação |
|---|---|
| F6 | Abrir/fechar a janela de mapa |
| Esc | Fechar a janela de mapa |
| 1 / 2 / 3 | Selecionar superfície / subterrâneo / DLC |
| `+` / `-` | Aumentar ou reduzir o zoom do mapa ampliado |
| F | Iniciar pesquisa textual |
| R | Retornar à vista central padrão |
| L | Centralizar no jogador **somente quando calibrado** |
| Clique esquerdo sobre ponto | Fixar/remover marcador |
| Clique direito sobre ponto | Concluir/desconcluir marcador |
| Clique sobre categoria | Exibir/ocultar categoria |

O HUD fica visível apenas quando a janela do jogo está ativa; a interface F6 não usa hooks globais de teclado. A tecla é consultada por `GetAsyncKeyState`, sem injetar comandos nem remapear controles. **O comportamento real de foco deve ser validado manualmente.**

## Leitura de posição: calibração e integridade

O leitor utiliza `OpenProcess(PROCESS_VM_READ)` e `ReadProcessMemory`, procura assinaturas com verificação de unicidade e rejeita campos inválidos. Ele não usa endereços absolutos do processo como se fossem estáveis. O retorno em pixels **não é projetado automaticamente para o mapa externo**, porque o alinhamento com os tiles de terceiros e o executável 1.16 ainda precisam ser verificados.

Por padrão, `config.json` tem `calibration.enabled = false`. É intencional: o HUD apresenta **"Sem posição validada"** até que existam pontos-âncora e testes reais que permitam calibrar `scale_x`, `scale_y`, `offset_x`, `offset_y`. Habilitar sem aferição pode posicionar o jogador incorretamente; não recomendamos fazê-lo.

Não faz gravação em memória, save, `regulation.bin`, hooks de DirectX ou bypass de anticheat.

## Configuração e persistência

Os campos incluem tamanho, opacidade, zoom, posição, camada, tecla, categorias ocultas e estado dos marcadores. As preferências residem em `Game/mods/ArayMinimap/data/config.json` quando o arquivo `data/` acompanha a instalação. A atualização preserva esses arquivos. Para resolver problemas de JSON, restaure a cópia `.bak` após fechar o jogo.

## Atualização

Faça backup da pasta `data/` e execute o mesmo instalador sobre a instalação atual. O instalador preserva os dados e cria cópia dos binários substituídos. Não remova manualmente o `dinput8.dll`.

## Desinstalação

PowerShell:

```powershell
.\uninstall.ps1 -GamePath 'E:\Games\ELDEN RING'
# Opcional: mover os dados pessoais para backup fora do mod:
.\uninstall.ps1 -GamePath 'E:\Games\ELDEN RING' -RemoveUserData
```

CMD:

```cmd
uninstall.cmd "E:\Games\ELDEN RING"
```

Os scripts removem apenas os binários do Aray Minimap. Os demais mods, o loader e os saves são preservados. O `-RemoveUserData` do PowerShell **faz backup**, não apaga definitivamente o histórico do mapa.

## Solução de problemas

- **Mapa sem textura:** tiles locais PNG/JPEG ausentes. Verifique a importação de `ctiles` e a estrutura de pastas.
- **Sem marcadores:** `markers.js`/`markers.json` ausente ou inválido na pasta `data/`.
- **Diagnóstico da posição:** abra o jogo offline e execute `ArayMinimap-Check.exe --diagnose` no PowerShell para registrar uma amostra nativa. Consulte também `data/diagnostico-posicao.json`.
- **Sem posição:** assinaturas não reconhecidas, leitura protegida, jogo em tela de carregamento ou calibração pendente. Isso não equivale a um bug do mapa.
- **F6 não abre:** confirme se a DLL e o EXE foram copiados, se o EML carregou a DLL e se o executável auxiliar iniciou. Consulte `mod_loader_log.txt`.
- **Jogo trava / perda de foco:** desligue o mod removendo exclusivamente `Game/mods/ArayMinimap.dll`, restaure o backup e informe build do jogo. Não faça depuração online.
- **Bloqueio de permissão:** feche o jogo e execute um shell com permissões sobre a pasta instalada. Os instaladores não exigem elevação quando já existe permissão de gravação.

## Arquitetura técnica

```text
Elden Mod Loader / eldenring.exe
    -> ArayMinimap.dll (bridge mínimo em C, sem dependência do CRT)
        -> ArayMinimap.exe (Go + Win32)
            -> Leitor read-only de memória com verificação de layout
            -> Conversão por calibração explícita
            -> HUD GDI e janela do mapa F6
            -> Marcadores e camadas (dados locais do usuário)
            -> JSON local (configuração, filtros, progresso)
```

A separação permite que erros no processo auxiliar não bloqueiem a inicialização do Elden Ring e evita carregar uma máquina virtual de linguagem dentro do jogo. O processo auxiliar tenta encerrar após o jogo ser fechado.

## Testes

No ambiente de desenvolvimento:

```powershell
go test ./...
go vet ./...
```

Os testes unitários abrangem validação de configurações, gravação e leitura, prevenção de mutação compartilhada, parse seguro de marcadores, calibração defensiva e decodificação de 36 bytes de posições. Testes de UI e integração real com o jogo **não foram executados**.

### Checklist manual exigido para uma release

- [ ] Confirmar que `ArayMinimap.dll` carrega no Elden Mod Loader no Elden Ring 1.16.
- [ ] Confirmar que o EXE auxiliar inicia e encerra com o jogo.
- [ ] Confirmar todos os controles de jogo: WASD, Shift, Ctrl, Alt, Espaço, mouse, ataques, câmera, menus, interações.
- [ ] Confirmar que F6 não executa ação indesejada no jogo e que o foco retorna corretamente.
- [ ] Validar duas ou mais coordenadas de ancoragem na superfície e a calibração; repetir em subterrâneo/DLC.
- [ ] Confirmar renderização de tiles e seleção de camada no HUD e F6.
- [ ] Confirmar fixar/remover marcador e persistência após reinicialização.
- [ ] Medir CPU, memória, FPS e verificar travamentos em diversas resoluções.
- [ ] Testar instalar, reinstalar e desinstalar sem prejudicar outros mods.
- [ ] Realizar testes também em instalação Steam oficial e em uma instalação não-Steam permitida, sem mecanismos de bypass.

## GitHub Actions, binários e releases

O workflow `.github/workflows/build.yml` é destinado a compilar em Windows, rodar os testes e publicar artefatos temporários de CI (`ArayMinimap-win64-preview.zip`) **sem criar release automaticamente**. Releases versionadas devem ser publicadas somente após testes reais no jogo 1.16.

## Créditos, licenças e atribuição

- **Aray Minimap:** código original, licenciado sob [GPL-3.0](LICENSE).
- **TarnishedAtlas:** referência de análise técnica e das regras de validação de memória; ZIP fornecido pelo usuário sob GPL-3.0. Não é dependência binária.
- **[jw-ofs/elden-ring-map](https://github.com/jw-ofs/elden-ring-map):** referência de formato de `markers.js`, projeção e interface; sem licença geral identificável. Não redistribuímos dados nem tiles.
- **[EthanShoeDev/elden-ring-compass](https://github.com/EthanShoeDev/elden-ring-compass):** referência de dados do mapa; não redistribuímos assets.
- **[Leaflet](https://leafletjs.com/):** biblioteca usada pelo site de referência; não é incluída neste mod nativo.
- **[Elden Mod Loader](https://github.com/techiew/EldenRingModLoader):** loader externo, preservado pelo instalador; não é redistribuído.

Elden Ring, o jogo, suas marcas e recursos gráficos pertencem aos seus respectivos titulares. Este é um projeto comunitário não oficial e independente.

## Changelog

- `0.1.0-preview.1` — primeiro código do app auxiliar, DLL bridge, estado persistente, leitor experimental, filtros e instaladores. **Sem afirmação de compatibilidade comprovada em jogo.**
