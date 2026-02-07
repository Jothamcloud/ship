package version

// Version is set at build time via ldflags
var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

// FullVersion returns the full version string
func FullVersion() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildDate + ")"
}
