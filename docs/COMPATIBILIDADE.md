# Matriz de compatibilidade e riscos

| Cenário | Compila | Validado no jogo | Situação |
|---|---:|---:|---|
| Go core em Linux | Sim | N/A | Testes unitários passaram |
| Go auxiliar Windows x64 (cross-compilação) | Sim | Não | PE válido; não executado em Windows |
| Bridge DLL x64 com imports kernel32 | Sim | Não | Imports PE inspecionados; não carregada em EML |
| Elden Ring 1.16 instalado pelo usuário | Não aplicável | Não | Assinaturas e offset podem divergir |
| Elden Ring oficial/Steam | Não aplicável | Não | Inicialização EAC desativado requerida |
| Instalador em E:\Games\ELDEN RING | Fonte gerada | Não | Sem execução em Windows |
| Integridade dos controles durante uso | Código sem hook de remapeamento | Não | Necessita teste manual |
| Tiles e marcadores jw-ofs | Importador implementado | Não | Requer cópia local autorizada |

## Falhas fechadas

A ausência de dados do mapa desativa texturas/marcadores, sem criar posições inventadas. Assinaturas inválidas desativam a localização. O valor padrão `calibration.enabled=false` impede mostrar posição incorreta por suposição de alinhamento. Isso exige calibração e teste reais antes de considerar a função de rastreamento pronta.

## Limitações adicionais

O HUD nativo utiliza Win32 GDI e não suporta ainda rotação da câmera. O marcador direcional é renderizado quando a calibração estiver validada. A interface não inclui todos os pop-ups, os 30 fluxos de missão do site de referência nem banco interno de itens em português. Múltiplas telas, DPI não padrão e tela cheia exclusiva precisam ser testados em Windows.
