package cmd

import (
	"fmt"

	"env-builder/internal/config"
	"env-builder/internal/osdetect"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "List all components managed by eb",
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

		fmt.Printf("OS: %s\n\n", osInfo)

		for _, section := range cfg.Sections {
			label := section.Manager
			if section.Manager == "system" {
				for name, mgr := range managers {
					if len(mgr.OS) > 0 && osInfo.Matches(mgr.OS) {
						label = fmt.Sprintf("system → %s", name)
						break
					}
				}
			}
			fmt.Printf("%s:\n", label)
			for _, pkg := range section.Packages {
				fmt.Printf("  - %s\n", pkg.Name)
			}
			fmt.Println()
		}

		if len(cfg.Custom) > 0 {
			fmt.Println("custom:")
			for _, pkg := range cfg.Custom {
				fmt.Printf("  - %s\n", pkg.Name)
			}
		}

		return nil
	},
}
