package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Package struct {
	Name        string   // populated from map key
	Version     string   `yaml:"version"`
	Comment     string   `yaml:"comment"`
	PostInstall []string `yaml:"post_install"`
	PostClean   []string `yaml:"post_clean"`
}

type CustomPackage struct {
	Name      string   // populated from map key
	Check     string   `yaml:"check"` // command whose stdout is the version, non-zero exit = not installed
	Install   []string `yaml:"install"`
	Uninstall []string `yaml:"uninstall"`
}

type Section struct {
	Manager  string
	Packages []Package
}

type Config struct {
	Sections []Section
	Custom   []CustomPackage
}

func ResolveConfigDir() string {
	// Primary: ~/.eb/
	home, _ := os.UserHomeDir()
	ebDir := filepath.Join(home, ".eb")
	if _, err := os.Stat(filepath.Join(ebDir, "config.yaml")); err == nil {
		return ebDir
	}
	// Fallback: current directory (development / cloned repo)
	wd, _ := os.Getwd()
	return wd
}

func Load(baseDir string) (Config, error) {
	configPath := filepath.Join(baseDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	var cfg Config
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return cfg, nil
	}

	mapping := root.Content[0]
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		key := mapping.Content[i].Value
		val := mapping.Content[i+1]

		if key == "custom" {
			customs, err := parseCustomMap(val)
			if err != nil {
				return Config{}, fmt.Errorf("parsing custom section: %w", err)
			}
			cfg.Custom = customs
			continue
		}

		packages, err := parsePackageMap(val)
		if err != nil {
			return Config{}, fmt.Errorf("parsing section %q: %w", key, err)
		}
		cfg.Sections = append(cfg.Sections, Section{Manager: key, Packages: packages})
	}

	return cfg, nil
}

func parsePackageMap(node *yaml.Node) ([]Package, error) {
	var packages []Package
	for i := 0; i < len(node.Content)-1; i += 2 {
		name := node.Content[i].Value
		pkg := Package{Name: name}
		if err := node.Content[i+1].Decode(&pkg); err != nil {
			return nil, fmt.Errorf("parsing package %q: %w", name, err)
		}
		pkg.Name = name // restore: Decode may zero it out if value is null
		packages = append(packages, pkg)
	}
	return packages, nil
}

func parseCustomMap(node *yaml.Node) ([]CustomPackage, error) {
	var customs []CustomPackage
	for i := 0; i < len(node.Content)-1; i += 2 {
		name := node.Content[i].Value
		pkg := CustomPackage{Name: name}
		if err := node.Content[i+1].Decode(&pkg); err != nil {
			return nil, fmt.Errorf("parsing custom package %q: %w", name, err)
		}
		pkg.Name = name
		customs = append(customs, pkg)
	}
	return customs, nil
}
