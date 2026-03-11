# env-builder

`eb` is a portable, OS-aware environment installer. You declare the tools you want once — `eb` figures out which package manager to use based on the system it runs on, and installs everything in one shot.

It supports `pacman`, `apt`, `dnf` out of the box, and is trivially extensible to any other manager by editing a single YAML file.

---

## Install

```bash
curl -sL https://raw.githubusercontent.com/amine-khemissi/env-builder/main/install.sh | bash
eb edit      # create and edit ~/.eb/config.yaml
eb install
```

The install script downloads the `eb` binary to `~/.local/bin`, `managers.yaml` to `~/.eb/`, and generates shell completions for fish, bash, and zsh. No `sudo` required.

---

## Usage

```bash
eb install        # install everything in ~/.eb/config.yaml (idempotent)
eb clean          # remove everything in ~/.eb/config.yaml
eb status         # show install status and version of each package
eb edit           # open ~/.eb/config.yaml in $EDITOR
eb config         # show detected OS and resolved package manager
eb export         # export config.yaml with installed versions pinned
```

`eb status` shows what is installed, what is missing, and flags version mismatches:

```
OS: cachyos (like: arch)

system → pacman:
  ✔  git     2.44.0
  ✔  docker  26.1.3
  ✗  meld    not installed

custom:
  ✔  go      go1.22.4
  ?  goland  no check configured
```

---

## Configuration

### `~/.eb/config.yaml` — what to install

```yaml
system:           # resolved to pacman / apt / dnf based on OS
  curl:
  git:
  vim:
    version: "9.1.0016-1"   # pin to a specific version (optional)
  docker:
    post_install:
      - sudo systemctl enable --now docker
      - sudo usermod -aG docker $USER
    post_clean:
      - sudo groupdel docker

paru:             # AUR helper — Arch/CachyOS only, your responsibility
  kind:
    version: "0.22.0"

custom:           # fully custom install/uninstall scripts
  go:
    check: go version | awk '{print $3}'   # stdout = installed version
    install:
      - VERSION=$(curl -s "https://go.dev/VERSION?m=text" | head -1)
      - curl -sL "https://go.dev/dl/${VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
      - sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
      - rm /tmp/go.tar.gz
    uninstall:
      - sudo rm -rf /usr/local/go
```

Packages are declared as map keys — the YAML parser enforces uniqueness, so duplicates are impossible.

| Field | Description |
|---|---|
| `version` | Pin the package to a specific version. If omitted, the latest available is used. `eb status` flags mismatches between the desired and installed version. |
| `post_install` | Shell commands run after the package manager installs the package |
| `post_clean` | Shell commands run after the package manager removes the package |
| `check` | Command whose stdout is the installed version (exit non-zero = not installed). Used by `eb status` and `eb install` to skip already-installed packages. |
| `install` / `uninstall` | Shell commands for fully custom tools (run as a single `bash -e` script) |
| `comment` | Human-readable note, ignored by `eb` |

### `~/.eb/managers.yaml` — how to install per OS

Downloaded automatically by `install.sh`. Defines install/remove commands per manager, and which OS IDs each applies to.

```yaml
pacman:
  install: sudo pacman -S --noconfirm --needed
  remove: sudo pacman -Rns --noconfirm
  check: pacman -Q
  os: [arch, cachyos, manjaro, artix, endeavouros, garuda]

apt:
  install: sudo apt install -y
  remove: sudo apt remove -y
  check: dpkg-query -W -f=${Version}
  os: [debian, ubuntu, linuxmint, pop]
```

To add a new package manager or OS, add an entry — no code changes needed:

```yaml
zypper:
  install: sudo zypper install -y
  remove: sudo zypper remove -y
  check: rpm -q --queryformat %{VERSION}
  os: [opensuse-leap, opensuse-tumbleweed]
```

---

## Build from source

```bash
git clone <repo-url>
cd env-builder
make install   # builds eb and copies it to ~/.local/bin
```
