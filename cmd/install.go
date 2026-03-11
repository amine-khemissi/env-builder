package cmd

import (
	"env-builder/internal/config"
	"env-builder/internal/installer"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all packages from config.yaml",
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

		if err := installer.Install(cfg, managers, baseDir); err != nil {
			return opError("%v", err)
		}

		log.Info().Msg("Done.")
		return nil
	},
}
