// Package version provides build-time version information for the AsyncAPI Go Code Generator.
package version

import (
	"fmt"
	"runtime"
)

// Build-time variables set by ldflags during compilation
var (
	// Version is the semantic version of the application
	Version = "dev"

	// BuildTime is the timestamp when the binary was built
	BuildTime = "unknown"

	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"

	// GitBranch is the git branch of the build
	GitBranch = "unknown"
)

// Info contains version and build information
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	GitCommit string `json:"gitCommit"`
	GitBranch string `json:"gitBranch"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("%s (built %s from %s on %s with %s for %s)",
		i.Version, i.BuildTime, i.GitCommit, i.GitBranch, i.GoVersion, i.Platform)
}

// Short returns a short version string
func (i Info) Short() string {
	return i.Version
}

// GetVersion returns just the version string
func GetVersion() string {
	return Version
}

// GetBuildInfo returns formatted build information
func GetBuildInfo() string {
	info := Get()
	return fmt.Sprintf(`Version:    %s
Build Time: %s
Git Commit: %s
Git Branch: %s
Go Version: %s
Platform:   %s`,
		info.Version,
		info.BuildTime,
		info.GitCommit,
		info.GitBranch,
		info.GoVersion,
		info.Platform)
}
