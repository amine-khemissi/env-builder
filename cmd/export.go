package cmd

import (
	"fmt"
	"os"

	"env-builder/internal/config"
	"env-builder/internal/installer"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export config.yaml with currently installed versions",
	Long:  "Outputs a config.yaml capturing installed versions of all managed packages.\nRedirect to a file to save: eb export > config.yaml",
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

		sections, customs, err := installer.Status(cfg, managers)
		if err != nil {
			return cfgError("%v", err)
		}

		// Build a lookup: package name → installed version
		installedVersion := map[string]string{}
		for _, section := range sections {
			for _, p := range section.Packages {
				if p.Installed {
					installedVersion[p.Name] = p.InstalledVersion
				}
			}
		}

		w := os.Stdout
		fmt.Fprintln(w, "# config.yaml — exported by eb")

		for _, section := range cfg.Sections {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%s:\n", section.Manager)
			for _, pkg := range section.Packages {
				version, installed := installedVersion[pkg.Name]
				if !installed {
					log.Warn().Msgf("%s: not installed, skipping", pkg.Name)
					continue
				}
				fmt.Fprintf(w, "  %s:\n", pkg.Name)
				fmt.Fprintf(w, "    version: %s\n", version)
				writeListField(w, "post_install", pkg.PostInstall)
				writeListField(w, "post_clean", pkg.PostClean)
			}
		}

		if len(customs) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "custom:")
			for i, p := range customs {
				fmt.Fprintf(w, "  %s:\n", p.Name)
				src := cfg.Custom[i]
				if src.Check != "" {
					fmt.Fprintf(w, "    check: %s\n", src.Check)
				}
				writeListField(w, "install", src.Install)
				writeListField(w, "uninstall", src.Uninstall)
			}
		}

		return nil
	},
}

func writeListField(w *os.File, key string, steps []string) {
	if len(steps) == 0 {
		return
	}
	fmt.Fprintf(w, "    %s:\n", key)
	for _, step := range steps {
		fmt.Fprintf(w, "      - %s\n", step)
	}
}
