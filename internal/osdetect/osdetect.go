package osdetect

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type OSInfo struct {
	ID     string
	IDLike []string
}

func (o OSInfo) String() string {
	if len(o.IDLike) > 0 {
		return fmt.Sprintf("%s (like: %s)", o.ID, strings.Join(o.IDLike, ", "))
	}
	return o.ID
}

func (o OSInfo) Matches(candidates []string) bool {
	for _, c := range candidates {
		if o.ID == c {
			return true
		}
		for _, like := range o.IDLike {
			if like == c {
				return true
			}
		}
	}
	return false
}

func Detect() (OSInfo, error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return OSInfo{}, fmt.Errorf("reading /etc/os-release: %w", err)
	}
	defer f.Close()

	var info OSInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "ID="):
			info.ID = strings.ToLower(unquote(strings.TrimPrefix(line, "ID=")))
		case strings.HasPrefix(line, "ID_LIKE="):
			val := strings.ToLower(unquote(strings.TrimPrefix(line, "ID_LIKE=")))
			info.IDLike = strings.Fields(val)
		}
	}
	return info, scanner.Err()
}

func unquote(s string) string {
	return strings.Trim(s, `"`)
}
