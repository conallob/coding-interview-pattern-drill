package version

// Version is set at build time via -ldflags "-X github.com/conallob/coding-interview-pattern-drill/version.Version=<tag>".
// Falls back to "dev" for local builds.
var Version = "dev"
