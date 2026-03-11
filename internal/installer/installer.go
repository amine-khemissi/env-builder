package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"env-builder/internal/config"
	"env-builder/internal/osdetect"
	"env-builder/internal/runner"

	"github.com/rs/zerolog/log"
)

func Install(cfg config.Config, managers map[string]config.Manager, scriptDir string) error {
	osInfo, err := osdetect.Detect()
	if err != nil {
		return fmt.Errorf("detecting OS: %w", err)
	}
	log.Debug().Msgf("detected OS: %s", osInfo)

	for _, section := range cfg.Sections {
		resolvedName, mgr, err := resolveManager(section.Manager, managers, osInfo)
		if err != nil {
			return err
		}
		if err := installSection(resolvedName, mgr, section.Packages, scriptDir); err != nil {
			return err
		}
	}
	return installCustomSection(cfg.Custom, scriptDir)
}

func Clean(cfg config.Config, managers map[string]config.Manager, scriptDir string) error {
	osInfo, err := osdetect.Detect()
	if err != nil {
		return fmt.Errorf("detecting OS: %w", err)
	}
	log.Debug().Msgf("detected OS: %s", osInfo)

	if err := cleanCustomSection(cfg.Custom, scriptDir); err != nil {
		return err
	}
	for i := len(cfg.Sections) - 1; i >= 0; i-- {
		section := cfg.Sections[i]
		resolvedName, mgr, err := resolveManager(section.Manager, managers, osInfo)
		if err != nil {
			return err
		}
		if err := cleanSection(resolvedName, mgr, section.Packages, scriptDir); err != nil {
			return err
		}
	}
	return nil
}

func resolveManager(sectionName string, managers map[string]config.Manager, osInfo osdetect.OSInfo) (string, config.Manager, error) {
	if sectionName != "system" {
		mgr, ok := managers[sectionName]
		if !ok {
			return "", config.Manager{}, fmt.Errorf("unknown manager %q — add it to managers.yaml", sectionName)
		}
		return sectionName, mgr, nil
	}

	for name, mgr := range managers {
		if len(mgr.OS) > 0 && osInfo.Matches(mgr.OS) {
			log.Debug().Msgf("system resolved to %q for OS %q", name, osInfo.ID)
			return name, mgr, nil
		}
	}
	return "", config.Manager{}, fmt.Errorf("no package manager found for OS %q — add a matching os field in managers.yaml", osInfo.ID)
}

func installSection(name string, mgr config.Manager, packages []config.Package, scriptDir string) error {
	if len(packages) == 0 {
		return nil
	}

	// Converge: skip packages already installed at the desired version.
	var toInstall []config.Package
	if mgr.Check != "" {
		for _, pkg := range packages {
			s := checkManagedPackage(mgr, pkg)
			if s.Installed && (pkg.Version == "" || s.InstalledVersion == pkg.Version) {
				log.Debug().Msgf("[%s] %s already installed (%s), skipping", name, pkg.Name, s.InstalledVersion)
				continue
			}
			toInstall = append(toInstall, pkg)
		}
	} else {
		toInstall = packages
	}

	if len(toInstall) == 0 {
		log.Info().Msgf("[%s] all packages already installed", name)
		return nil
	}

	pkgArgs := make([]string, len(toInstall))
	pkgNames := make([]string, len(toInstall))
	for i, p := range toInstall {
		pkgNames[i] = p.Name
		if p.Version != "" {
			pkgArgs[i] = p.Name + "=" + p.Version
		} else {
			pkgArgs[i] = p.Name
		}
	}

	log.Info().Msgf("[%s] installing: %s", name, strings.Join(pkgNames, " "))

	parts := strings.Fields(mgr.Install)
	cmd := exec.Command(parts[0], append(parts[1:], pkgArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[%s] one or more packages could not be installed", name)
	}

	return runPostInstall(toInstall, scriptDir)
}

func runPostInstall(packages []config.Package, scriptDir string) error {
	for _, pkg := range packages {
		if len(pkg.PostInstall) == 0 {
			continue
		}
		log.Debug().Msgf("[%s] running post-install", pkg.Name)
		if err := runner.RunSteps(pkg.Name, scriptDir, pkg.PostInstall); err != nil {
			return err
		}
	}
	return nil
}

func installCustomSection(customs []config.CustomPackage, scriptDir string) error {
	for _, pkg := range customs {
		if len(pkg.Install) == 0 {
			continue
		}
		if pkg.Check != "" {
			s := checkCustomPackage(pkg)
			if s.Installed {
				log.Debug().Msgf("[custom] %s already installed (%s), skipping", pkg.Name, s.InstalledVersion)
				continue
			}
		}
		log.Info().Msgf("[custom] installing: %s", pkg.Name)
		if err := runner.RunSteps(pkg.Name, scriptDir, pkg.Install); err != nil {
			return err
		}
	}
	return nil
}

func cleanSection(name string, mgr config.Manager, packages []config.Package, scriptDir string) error {
	if len(packages) == 0 {
		return nil
	}

	pkgNames := make([]string, len(packages))
	for i, p := range packages {
		pkgNames[i] = p.Name
	}

	log.Info().Msgf("[%s] removing: %s", name, strings.Join(pkgNames, " "))
	parts := strings.Fields(mgr.Remove)
	cmd := exec.Command(parts[0], append(parts[1:], pkgNames...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run() // intentionally ignore error: some packages may not be installed

	return runPostClean(packages, scriptDir)
}

func runPostClean(packages []config.Package, scriptDir string) error {
	for _, pkg := range packages {
		if len(pkg.PostClean) == 0 {
			continue
		}
		log.Debug().Msgf("[%s] running post-clean", pkg.Name)
		if err := runner.RunSteps(pkg.Name, scriptDir, pkg.PostClean); err != nil {
			return err
		}
	}
	return nil
}

func cleanCustomSection(customs []config.CustomPackage, scriptDir string) error {
	for _, pkg := range customs {
		if len(pkg.Uninstall) == 0 {
			continue
		}
		log.Info().Msgf("[custom] removing: %s", pkg.Name)
		if err := runner.RunSteps(pkg.Name, scriptDir, pkg.Uninstall); err != nil {
			return err
		}
	}
	return nil
}
