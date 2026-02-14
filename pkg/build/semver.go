package build

import (
	"regexp"
	"strings"
)

// Semver holds parsed semantic version components.
// Follows the Semantic Versioning 2.0.0 spec (https://semver.org).
type Semver struct {
	// Raw is the original input string (e.g., "v1.2.3-rc.1+build.42").
	Raw string

	// Major is the major version component (e.g., "1").
	Major string

	// Minor is the minor version component (e.g., "2").
	Minor string

	// Patch is the patch version component (e.g., "3").
	Patch string

	// Prerelease is the prerelease identifier (e.g., "rc.1"), empty if none.
	Prerelease string

	// Build is the build metadata after '+' (e.g., "build.42"), empty if none.
	Build string
}

// semverRegex matches semantic version strings with optional v prefix.
// Groups: (1) major, (2) minor, (3) patch, (4) prerelease, (5) build metadata
var semverRegex = regexp.MustCompile(
	`^v?(\d+)\.(\d+)\.(\d+)` +
		`(?:-([\dA-Za-z\-]+(?:\.[\dA-Za-z\-]+)*))?` +
		`(?:\+([\dA-Za-z\-]+(?:\.[\dA-Za-z\-]+)*))?$`,
)

// ParseSemver parses a semver string like "v1.2.3", "1.2.3-rc.1", or
// "v2.0.0-beta.1+build.42". Returns nil, false if the string is not a
// valid semver.
func ParseSemver(s string) (*Semver, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	matches := semverRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, false
	}

	return &Semver{
		Raw:        s,
		Major:      matches[1],
		Minor:      matches[2],
		Patch:      matches[3],
		Prerelease: matches[4], // empty string if no prerelease
		Build:      matches[5], // empty string if no build metadata
	}, true
}

// Version returns the clean MAJOR.MINOR.PATCH string without v prefix,
// prerelease, or build metadata. E.g., "1.2.3".
func (sv *Semver) Version() string {
	return sv.Major + "." + sv.Minor + "." + sv.Patch
}

// VersionFull returns MAJOR.MINOR.PATCH with prerelease if present.
// E.g., "1.2.3-rc.1". Build metadata is excluded per semver spec
// (build metadata SHOULD be ignored for version precedence).
func (sv *Semver) VersionFull() string {
	v := sv.Version()
	if sv.Prerelease != "" {
		v += "-" + sv.Prerelease
	}
	return v
}

// MajorMinor returns "MAJOR.MINOR", e.g., "1.2".
func (sv *Semver) MajorMinor() string {
	return sv.Major + "." + sv.Minor
}
