package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// exitError carries an exit code alongside the error message.
// code 1 = operational (package not found, install failed)
// code 2 = configuration (invalid YAML, missing file, unknown manager)
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func opError(format string, args ...any) error {
	return &exitError{1, fmt.Sprintf(format, args...)}
}

func cfgError(format string, args ...any) error {
	return &exitError{2, fmt.Sprintf(format, args...)}
}

var logLevel string

var rootCmd = &cobra.Command{
	Use:           "eb",
	Short:         "env builder — portable environment installer",
	Long:          "Installs or removes packages defined in ~/.eb/config.yaml. Detects your OS at runtime and routes to the right package manager automatically.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			return cfgError("invalid log level %q — use debug, info, warn, or error", logLevel)
		}
		zerolog.SetGlobalLevel(level)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var e *exitError
		if errors.As(err, &e) {
			if e.code == 1 {
				log.Warn().Msg(e.msg)
			} else {
				log.Error().Msg(e.msg)
			}
			os.Exit(e.code)
		}
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		PartsOrder: []string{zerolog.LevelFieldName, zerolog.MessageFieldName},
	}
	log.Logger = zerolog.New(output)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(exportCmd)
}
