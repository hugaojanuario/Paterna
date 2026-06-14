package version

// preenchidos em build time via ldflags pelo goreleaser
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
