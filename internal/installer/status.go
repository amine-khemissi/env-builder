package installer

import (
	"os/exec"
	"strings"

	"env-builder/internal/config"
	"env-builder/internal/osdetect"
)

type PackageStatus struct {
	Name        string
	DesiredVersion string // from tools.yaml, empty = any
	InstalledVersion string // empty = not installed
	Installed   bool
	Checkable   bool // false when no check command is configured
}

type SectionStatus struct {
	Label    string // display name (e.g. "system → pacman")
	Packages []PackageStatus
}

func Status(cfg config.Config, managers map[string]config.Manager) ([]SectionStatus, []PackageStatus, error) {
	osInfo, err := osdetect.Detect()
	if err != nil {
		return nil, nil, err
	}

	var sections []SectionStatus
	for _, section := range cfg.Sections {
		resolvedName, mgr, err := resolveManager(section.Manager, managers, osInfo)
		if err != nil {
			return nil, nil, err
		}

		label := resolvedName
		if section.Manager == "system" {
			label = "system → " + resolvedName
		}

		var statuses []PackageStatus
		for _, pkg := range section.Packages {
			statuses = append(statuses, checkManagedPackage(mgr, pkg))
		}
		sections = append(sections, SectionStatus{Label: label, Packages: statuses})
	}

	var customs []PackageStatus
	for _, pkg := range cfg.Custom {
		customs = append(customs, checkCustomPackage(pkg))
	}

	return sections, customs, nil
}

func checkManagedPackage(mgr config.Manager, pkg config.Package) PackageStatus {
	s := PackageStatus{Name: pkg.Name, DesiredVersion: pkg.Version}
	if mgr.Check == "" {
		return s
	}
	s.Checkable = true

	parts := strings.Fields(mgr.Check)
	out, err := exec.Command(parts[0], append(parts[1:], pkg.Name)...).Output()
	if err != nil {
		return s // not installed
	}

	s.Installed = true
	words := strings.Fields(strings.TrimSpace(string(out)))
	if len(words) > 0 {
		s.InstalledVersion = words[len(words)-1]
	}
	return s
}

func checkCustomPackage(pkg config.CustomPackage) PackageStatus {
	s := PackageStatus{Name: pkg.Name}
	if pkg.Check == "" {
		return s
	}
	s.Checkable = true

	out, err := exec.Command("bash", "-c", pkg.Check).Output()
	if err != nil {
		return s
	}

	s.Installed = true
	s.InstalledVersion = strings.TrimSpace(string(out))
	return s
}
