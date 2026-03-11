package cmd

import (
	"fmt"
	"os"
	"strings"

	"env-builder/internal/config"
	"env-builder/internal/installer"
	"env-builder/internal/osdetect"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)


const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func color(s, code string) string {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return s
	}
	return code + s + colorReset
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show install status of all managed packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := config.ResolveConfigDir()

		cfg, err := config.Load(baseDir)
		if err != nil {
			return cfgError("loading config: %v", err)
		}

		managers, err := config.LoadManagers(baseDir)
		if err != nil {
			return cfgError("loading managers: %v", err)
		}

		osInfo, err := osdetect.Detect()
		if err != nil {
			return cfgError("detecting OS: %v", err)
		}

		sections, customs, err := installer.Status(cfg, managers)
		if err != nil {
			return cfgError("%v", err)
		}

		fmt.Printf("%s: %s\n\n", color("OS", colorBold), osInfo)

		for _, section := range sections {
			printSection(section.Label, section.Packages)
		}

		if len(customs) > 0 {
			var customStatuses []installer.PackageStatus
			for _, c := range customs {
				customStatuses = append(customStatuses, c)
			}
			printSection("custom", customStatuses)
		}

		return nil
	},
}

func printSection(label string, packages []installer.PackageStatus) {
	if len(packages) == 0 {
		return
	}

	fmt.Printf("%s:\n", color(label, colorBold))

	// find max name width for alignment
	maxLen := 0
	for _, p := range packages {
		if len(p.Name) > maxLen {
			maxLen = len(p.Name)
		}
	}

	for _, p := range packages {
		padding := strings.Repeat(" ", maxLen-len(p.Name))

		var icon, versionStr string
		switch {
		case !p.Checkable:
			icon = color("?", colorYellow)
			versionStr = color("no check configured", colorDim)
		case p.Installed:
			icon = color("✔", colorGreen)
			versionStr = color(p.InstalledVersion, colorGreen)
			if p.DesiredVersion != "" && p.InstalledVersion != p.DesiredVersion {
				versionStr += color(fmt.Sprintf("  (want %s)", p.DesiredVersion), colorYellow)
			}
		default:
			icon = color("✗", colorRed)
			if p.DesiredVersion != "" {
				versionStr = color(fmt.Sprintf("not installed  (want %s)", p.DesiredVersion), colorRed)
			} else {
				versionStr = color("not installed", colorRed)
			}
		}

		fmt.Printf("  %s  %s%s  %s\n", icon, p.Name, padding, versionStr)
	}
	fmt.Println()
}

