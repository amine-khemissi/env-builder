package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	Install string   `yaml:"install"`
	Remove  string   `yaml:"remove"`
	Check   string   `yaml:"check"` // run as "check <pkg>"; stdout last word = version, non-zero = not installed
	OS      []string `yaml:"os"`    // empty = no OS restriction (user responsibility)
}

func LoadManagers(baseDir string) (map[string]Manager, error) {
	managersPath := filepath.Join(baseDir, "managers.yaml")
	data, err := os.ReadFile(managersPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", managersPath, err)
	}

	var managers map[string]Manager
	if err := yaml.Unmarshal(data, &managers); err != nil {
		return nil, fmt.Errorf("parsing managers.yaml: %w", err)
	}

	return managers, nil
}
