package build

import "testing"

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input      string
		wantOK     bool
		major      string
		minor      string
		patch      string
		prerelease string
		build      string
		version    string
		full       string
		majorMinor string
	}{
		// Basic versions
		{input: "1.2.3", wantOK: true, major: "1", minor: "2", patch: "3", version: "1.2.3", full: "1.2.3", majorMinor: "1.2"},
		{input: "0.0.0", wantOK: true, major: "0", minor: "0", patch: "0", version: "0.0.0", full: "0.0.0", majorMinor: "0.0"},
		{input: "10.20.30", wantOK: true, major: "10", minor: "20", patch: "30", version: "10.20.30", full: "10.20.30", majorMinor: "10.20"},

		// With v prefix
		{input: "v1.2.3", wantOK: true, major: "1", minor: "2", patch: "3", version: "1.2.3", full: "1.2.3", majorMinor: "1.2"},
		{input: "v0.1.0", wantOK: true, major: "0", minor: "1", patch: "0", version: "0.1.0", full: "0.1.0", majorMinor: "0.1"},

		// With prerelease
		{input: "1.2.3-rc.1", wantOK: true, major: "1", minor: "2", patch: "3", prerelease: "rc.1", version: "1.2.3", full: "1.2.3-rc.1", majorMinor: "1.2"},
		{input: "v2.0.0-beta.1", wantOK: true, major: "2", minor: "0", patch: "0", prerelease: "beta.1", version: "2.0.0", full: "2.0.0-beta.1", majorMinor: "2.0"},
		{input: "1.0.0-alpha", wantOK: true, major: "1", minor: "0", patch: "0", prerelease: "alpha", version: "1.0.0", full: "1.0.0-alpha", majorMinor: "1.0"},
		{input: "1.0.0-0.3.7", wantOK: true, major: "1", minor: "0", patch: "0", prerelease: "0.3.7", version: "1.0.0", full: "1.0.0-0.3.7", majorMinor: "1.0"},

		// With build metadata
		{input: "1.2.3+build.42", wantOK: true, major: "1", minor: "2", patch: "3", build: "build.42", version: "1.2.3", full: "1.2.3", majorMinor: "1.2"},
		{input: "v1.0.0+20260214", wantOK: true, major: "1", minor: "0", patch: "0", build: "20260214", version: "1.0.0", full: "1.0.0", majorMinor: "1.0"},

		// With prerelease and build metadata
		{input: "v2.0.0-beta.1+build.42", wantOK: true, major: "2", minor: "0", patch: "0", prerelease: "beta.1", build: "build.42", version: "2.0.0", full: "2.0.0-beta.1", majorMinor: "2.0"},
		{input: "1.0.0-rc.1+sha.abc123", wantOK: true, major: "1", minor: "0", patch: "0", prerelease: "rc.1", build: "sha.abc123", version: "1.0.0", full: "1.0.0-rc.1", majorMinor: "1.0"},

		// Whitespace handling
		{input: "  v1.2.3  ", wantOK: true, major: "1", minor: "2", patch: "3", version: "1.2.3", full: "1.2.3", majorMinor: "1.2"},

		// Invalid inputs
		{input: "", wantOK: false},
		{input: "not-a-version", wantOK: false},
		{input: "1.2", wantOK: false},
		{input: "1", wantOK: false},
		{input: "v", wantOK: false},
		{input: "v1.2", wantOK: false},
		{input: "1.2.3.4", wantOK: false},
		{input: "latest", wantOK: false},
		{input: "main", wantOK: false},
		{input: "abc1234", wantOK: false},
		{input: "vv1.2.3", wantOK: false},
		{input: "V1.2.3", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			sv, ok := ParseSemver(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ParseSemver(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				if sv != nil {
					t.Errorf("ParseSemver(%q) returned non-nil Semver on failure", tt.input)
				}
				return
			}

			if sv.Major != tt.major {
				t.Errorf("Major = %q, want %q", sv.Major, tt.major)
			}
			if sv.Minor != tt.minor {
				t.Errorf("Minor = %q, want %q", sv.Minor, tt.minor)
			}
			if sv.Patch != tt.patch {
				t.Errorf("Patch = %q, want %q", sv.Patch, tt.patch)
			}
			if sv.Prerelease != tt.prerelease {
				t.Errorf("Prerelease = %q, want %q", sv.Prerelease, tt.prerelease)
			}
			if sv.Build != tt.build {
				t.Errorf("Build = %q, want %q", sv.Build, tt.build)
			}
			if sv.Version() != tt.version {
				t.Errorf("Version() = %q, want %q", sv.Version(), tt.version)
			}
			if sv.VersionFull() != tt.full {
				t.Errorf("VersionFull() = %q, want %q", sv.VersionFull(), tt.full)
			}
			if sv.MajorMinor() != tt.majorMinor {
				t.Errorf("MajorMinor() = %q, want %q", sv.MajorMinor(), tt.majorMinor)
			}
		})
	}
}

func TestParseSemverRawPreserved(t *testing.T) {
	sv, ok := ParseSemver("v1.2.3-rc.1+build.42")
	if !ok {
		t.Fatal("expected ok")
	}
	if sv.Raw != "v1.2.3-rc.1+build.42" {
		t.Errorf("Raw = %q, want %q", sv.Raw, "v1.2.3-rc.1+build.42")
	}
}
