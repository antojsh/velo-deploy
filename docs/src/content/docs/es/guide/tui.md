---
title: Dashboard TUI
description: La interfaz de terminal para gestionar apps desde una sola pantalla manejada con teclado.
---

import { Aside } from '@astrojs/starlight/components';

La TUI es una capa opcional por encima de la CLI. Es útil cuando tenés un puñado de apps en un único servidor y querés ver su estado de un vistazo.

## Lanzar la TUI

```bash
velo-deploy
```

La primera pantalla muestra una lista de apps registradas con su tipo, estado y puerto asignado. Usá el teclado para navegar y actuar.

## Atajos de teclado

### Navegación

| Tecla | Acción |
| --- | --- |
| `↑` / `↓` | Moverse entre apps. |
| `Tab` | Cambiar entre la lista y el panel de detalle. |
| `Enter` | Abrir la app seleccionada. |
| `Esc` | Volver a la lista. |
| `Q` | Salir. |

### Acciones

| Tecla | Acción |
| --- | --- |
| `N` | Nuevo deploy desde un URL de Git. |
| `A` | Agregar una app existente desde un path local. |
| `R` | Reiniciar la app seleccionada (solo Node.js). |
| `S` | Detener la app seleccionada (solo Node.js). |
| `L` | Ver logs de la app seleccionada. |
| `D` | Borrar la app seleccionada. |
| `?` | Mostrar el overlay de ayuda. |

## Soporte de mouse

La TUI es completamente mouse-aware en terminales que lo soportan (la mayoría de los emuladores modernos, incluyendo iTerm2, WezTerm, Ghostty y Windows Terminal). Podés hacer click para seleccionar, doble click para abrir, y usar la rueda del mouse para moverte por el visor de logs.

## Cuándo usar la CLI en su lugar

La TUI es ideal para gestión one-off en un único servidor. Para automatización, scripts y CI, preferí la [CLI](/velo-deploy/es/guide/cli/) — cada acción de la TUI mapea a un comando de la CLI por debajo, así que se mantienen en sync.

<Aside type="tip" title="Corré la TUI sobre SSH">
	La TUI usa Bubble Tea y renderiza correctamente sobre cualquier conexión SSH. Para mejores resultados usá un terminal moderno (WezTerm, Ghostty, o la última versión de Windows Terminal).
</Aside>
