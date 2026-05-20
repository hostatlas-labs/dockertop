# dockertop

[![Release](https://img.shields.io/github/v/release/hostatlas-labs/dockertop?style=flat-square)](https://github.com/hostatlas-labs/dockertop/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/hostatlas-labs/dockertop?style=flat-square)](https://goreportcard.com/report/github.com/hostatlas-labs/dockertop)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

```
  _   _           _      _   _   _
 | | | | ___  ___| |_   / \ | |_| | __ _ ___
 | |_| |/ _ \/ __| __| / _ \| __| |/ _` / __|
 |  _  | (_) \__ \ |_ / ___ \ |_| | (_| \__ \
 |_| |_|\___/|___/\__/_/   \_\__|_|\__,_|___/
```

A prettier `docker stats` for the terminal. CPU sparklines, color-coded
memory progress, Docker-Compose-aware grouping, keyboard-driven sort and
filter. Single binary by [HostAtlas](https://hostatlas.app), MIT-licensed.

## Features

- **Real-time stats** — CPU, memory, network, block I/O per container
- **CPU sparklines** — short history per container, so you see trend not just snapshot
- **Compose grouping** — containers grouped by `com.docker.compose.project` label
- **Sort modes** — by CPU, memory, name, or network — switch with one keystroke
- **Live filter** — type `/` and filter on container name, image, or label
- **Color-coded thresholds** — green / amber / red on CPU + memory pressure
- **Auto-refresh** — configurable interval (1-60 s), defaults to 2 s
- **Alt-screen TUI** — full screen while running, terminal restored on exit
- **Single binary** — no daemons, no config files, no Python venv
- **Self-update** — `dockertop --update` pulls the latest GitHub release

## Quick Start

### Install

**One-liner:**

```bash
curl -fsSL https://tools.hostatlas.app/install.sh | sh -s dockertop
```

Add `--user` for a `$HOME/.local/bin` install without sudo, or
`--version=1.0.0` to pin a specific release.

**With Go:**

```bash
go install github.com/hostatlas-labs/dockertop/cmd/dockertop@latest
```

**Homebrew (macOS, Linuxbrew):**

```bash
brew install hostatlas-labs/tap/dockertop
```

**Manual:** Browse all releases at [github.com/hostatlas-labs/dockertop/releases](https://github.com/hostatlas-labs/dockertop/releases) or [tools.hostatlas.app](https://tools.hostatlas.app/tools.json).

### Run

```bash
dockertop                       # Launch TUI, default sort: cpu, grouped by compose project
dockertop --sort mem            # Sort by memory
dockertop --no-group            # Flat list, no compose grouping
dockertop --interval 5          # 5-second refresh instead of 2
```

## Usage

```
dockertop                       Launch the TUI
dockertop --sort cpu            Sort by CPU (default)
dockertop --sort mem            Sort by memory
dockertop --sort name           Sort alphabetically
dockertop --sort net            Sort by network I/O
dockertop --group               Group containers by Compose project (default)
dockertop --no-group            Flat list, no grouping
dockertop --interval 2          Refresh every N seconds (1-60, default 2)
dockertop --update              Self-update from GitHub releases
dockertop --version             Show version
dockertop --help                Show all flags
```

## Keyboard

| Key | Action |
|---|---|
| `s` | Cycle sort mode (cpu → mem → name → net → cpu) |
| `g` | Toggle Compose grouping on/off |
| `/` | Enter filter mode — type to filter, Esc to clear |
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `PgUp` / `PgDn` | Page up / down |
| `Home` / `End` | Jump to top / bottom |
| `q` / `Ctrl-C` | Quit |

## How It Looks

```
 dockertop · web-prod-04 · 12 containers · sort:cpu · group:compose
 ─────────────────────────────────────────────────────────────────────
 ▼ compose: hostatlas-app
   nginx       CPU  3.2%  ▁▂▂▃▅▆▆▄    MEM  48 MiB / 256 MiB  ────
   php-fpm     CPU 41.2%  ▃▄▅▇█▇▆▅    MEM 512 MiB / 1.5 GiB  ▓▓░░
   redis       CPU  1.1%  ▁▁▁▂▁▁▁▁    MEM  62 MiB / 256 MiB  ▒───
   postgres    CPU  8.4%  ▁▁▂▂▃▃▃▂    MEM 2.1 GiB / 4 GiB    ▓▓▓░
 ▼ compose: monitoring
   prometheus  CPU  2.1%  ▁▂▁▁▁▁▁▁    MEM 412 MiB / 1 GiB    ▓▒──
   grafana     CPU  0.8%  ▁▁▁▁▁▁▁▁    MEM 180 MiB / 512 MiB  ▓───
 ─────────────────────────────────────────────────────────────────────
 s sort  g group  / filter  ↑↓ scroll  q quit
```

## Configuration

dockertop has no config file. Every option is a CLI flag. To stick with
your preferred defaults, wrap it in a shell alias:

```bash
# ~/.bashrc / ~/.zshrc
alias dt='dockertop --sort mem --interval 5'
```

## Docker Connectivity

dockertop talks to the Docker daemon over the standard socket at
`/var/run/docker.sock` (Linux) or `~/.docker/run/docker.sock` (Docker
Desktop on macOS). The user running dockertop must be a member of the
`docker` group (Linux) or have Docker Desktop running (macOS).

For remote Docker daemons, point `DOCKER_HOST` at the remote socket:

```bash
DOCKER_HOST=ssh://user@remote-host dockertop
DOCKER_HOST=tcp://192.168.1.10:2376 dockertop
```

## Podman

dockertop auto-detects Podman when the Docker socket isn't available but
the Podman socket is. Same UI, same keys, same sparklines. Tested with
rootless Podman 4.x+.

```bash
# Enable Podman's user socket
systemctl --user enable --now podman.socket

# Then just run
dockertop
```

## Examples

### Watch a deploy go live

```bash
# In another tab while you `helm upgrade` / `docker compose up`
dockertop --sort cpu --interval 1
```

CPU sparklines make it obvious which container is doing the work after a
rollout.

### Find the memory hog quickly

```bash
dockertop --sort mem
# Top of the list is the answer.
```

### Filter to one Compose project

```bash
dockertop
# Then press / and type: myapp
```

## How It Compares

| | dockertop | `docker stats` | ctop | lazydocker |
|---|---|---|---|---|
| Sparklines | ✓ | — | — | partial |
| Compose grouping | ✓ | — | — | ✓ |
| Sort + live filter | ✓ | — | partial | ✓ |
| Alt-screen, restores terminal | ✓ | — | ✓ | ✓ |
| Single binary, no deps | ✓ | ✓ | ✓ | ✓ |
| Podman auto-detect | ✓ | — | partial | partial |

`docker stats` is the baseline; dockertop is what you reach for when you
want to actually look at the numbers. ctop is a great alternative — we
just wanted something with real sparklines and Compose grouping.

## FAQ

**Does it modify or restart my containers?**
No. dockertop is read-only — it streams stats and renders them. It never
sends `restart`, `stop`, `kill`, or `exec`. (Action keys may land in a
future major version; they're not in v0.7.)

**Why not call it `htop` for Docker?**
htop is for processes on a host. dockertop is for containers across one or
more Docker / Podman daemons. Same vibe, different scope.

**Does it work with Docker Swarm?**
It shows containers that the local daemon sees. For a multi-node Swarm
view, point `DOCKER_HOST` at each node and run an instance per node, or
use `docker stack ps` for the per-service view.

**Multi-host?**
Currently single-daemon per invocation. Multi-host is roadmap. For now run
an instance per host (one terminal pane each, tmux works well).

**License?**
MIT. Fork, modify, sell, anything. See [LICENSE](LICENSE).

## Troubleshooting

### "Could not connect to Docker daemon"

The user running dockertop isn't in the `docker` group, or the daemon
isn't running, or you're on macOS without Docker Desktop running. Fix:

```bash
# Linux
sudo usermod -aG docker $USER
newgrp docker

# macOS
open -a Docker
```

### Sparklines look broken

Your terminal doesn't render the Unicode block characters used for
sparklines (`▁ ▂ ▃ ▄ ▅ ▆ ▇ █`). Use a modern terminal (iTerm2, Alacritty,
WezTerm, Kitty, GNOME Terminal, Windows Terminal). PuTTY and basic xterm
will show boxes instead.

### Refresh interval feels jittery

Drop to `--interval 1` (max sample rate) or up to `--interval 5` (less
chatty). Docker stats sampling is rate-limited by the daemon itself, so
sub-second intervals don't actually deliver more data.

### `terminfo: term not found`

Run `export TERM=xterm-256color` and retry. Bubbletea (the underlying TUI
library) requires a recognisable TERM value.

## Building from Source

```bash
git clone https://github.com/hostatlas-labs/dockertop.git
cd dockertop
go build -o dockertop ./cmd/dockertop
```

Go 1.22+ required.

## Self-Update

```bash
dockertop --update
```

Pulls the latest GitHub release, verifies the signature, replaces the
binary in place.

## Support

DockerTop is primarily maintained by the source control team at HostAtlas Technologies LLC.
Please submit a [GitHub Issue](https://github.com/hostatlas-labs/dockertop/issues) to report any trouble.

## License

MIT — see [LICENSE](LICENSE).

---

Built by [HostAtlas](https://hostatlas.app) — HostAtlas Technologies LLC, an [Akyros Labs](https://akyroslabs.com) brand.  
[www.hostatlas.app](https://hostatlas.app) · [hello@hostatlas.app](mailto:hello@hostatlas.app)
