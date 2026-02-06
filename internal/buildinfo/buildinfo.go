package buildinfo

// These are populated at build time via -ldflags.
//
// Keep these in a non-main package so they are accessible from commands/tests.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)
