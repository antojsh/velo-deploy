---
title: TUI dashboard
description: The terminal user interface for managing apps from a single keyboard-driven screen.
---

import { Aside } from '@astrojs/starlight/components';

The TUI is an optional layer on top of the CLI. It is useful when you have a handful of apps on a single server and want to see their status at a glance.

## Launching the TUI

```bash
velo-deploy
```

The first screen shows a list of registered apps with their type, status, and assigned port. Use the keyboard to navigate and act.

## Keyboard shortcuts

### Navigation

| Key | Action |
| --- | --- |
| `↑` / `↓` | Move between apps. |
| `Tab` | Switch between the app list and the detail panel. |
| `Enter` | Open the selected app. |
| `Esc` | Back to the list. |
| `Q` | Quit. |

### Actions

| Key | Action |
| --- | --- |
| `N` | New deploy from a Git URL. |
| `A` | Add an existing app from a local path. |
| `R` | Restart the selected app (Node.js only). |
| `S` | Stop the selected app (Node.js only). |
| `L` | View logs for the selected app. |
| `D` | Delete the selected app. |
| `?` | Show the help overlay. |

## Mouse support

The TUI is fully mouse-aware on terminals that support it (most modern terminal emulators, including iTerm2, WezTerm, Ghostty, and Windows Terminal). You can click to select, double-click to open, and use the scroll wheel to move through the log viewer.

## When to use the CLI instead

The TUI is great for one-off management on a single server. For automation, scripting, and CI, prefer the [CLI](/velo-deploy/guide/cli/) — every TUI action maps to a CLI command under the hood, so they stay in sync.

<Aside type="tip" title="Run the TUI over SSH">
	The TUI uses Bubble Tea and renders correctly over any SSH connection. For best results use a modern terminal (e.g. WezTerm, Ghostty, or the latest Windows Terminal).
</Aside>
