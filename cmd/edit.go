package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const configTemplate = `# config.yaml — your personal tools list for eb
#
# Packages under "system" are installed using the package manager
# that matches your OS (pacman on Arch, apt on Ubuntu, dnf on Fedora...).
# Run 'eb config' to see which manager will be used on your system.

system:
  git:
  curl:
  make:

# post_install runs shell commands after the package manager installs the package.
# post_clean runs shell commands after the package manager removes the package.
#
#  docker:
#    post_install:
#      - sudo systemctl enable --now docker
#      - sudo usermod -aG docker $USER
#    post_clean:
#      - sudo groupdel docker

# AUR packages — Arch / CachyOS only (your responsibility to have paru installed)
# paru:
#   kind:

# Fully custom install/uninstall scripts (run as a single bash -e script)
# custom:
#   go:
#     check: go version | awk '{print $3}'
#     install:
#       - VERSION=$(curl -s "https://go.dev/VERSION?m=text" | head -1)
#       - curl -sL "https://go.dev/dl/${VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
#       - sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
#       - rm /tmp/go.tar.gz
#     uninstall:
#       - sudo rm -rf /usr/local/go
`

func resolveEditor() string {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if e := os.Getenv(env); e != "" {
			return e
		}
	}
	for _, e := range []string{"nvim", "vim", "vi", "nano"} {
		if path, err := exec.LookPath(e); err == nil {
			return path
		}
	}
	return ""
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config.yaml in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		editor := resolveEditor()
		if editor == "" {
			return fmt.Errorf("no editor found — set $EDITOR or install vi, nano, or vim")
		}

		ebDir := filepath.Join(os.Getenv("HOME"), ".eb")
		if err := os.MkdirAll(ebDir, 0755); err != nil {
			return fmt.Errorf("creating ~/.eb: %w", err)
		}

		configPath := filepath.Join(ebDir, "config.yaml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
				return fmt.Errorf("creating config.yaml: %w", err)
			}
			fmt.Printf("Created %s\n", configPath)
		}

		c := exec.Command(editor, configPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
