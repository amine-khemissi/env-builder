package cmd

import (
	"env-builder/internal/config"
	"env-builder/internal/installer"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all packages defined in config.yaml",
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

		if err := installer.Clean(cfg, managers, baseDir); err != nil {
			return opError("%v", err)
		}

		log.Info().Msg("Done.")
		return nil
	},
}
