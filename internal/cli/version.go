package cli

import "fmt"

var (
	// BuildVersion contains the major.minor version of the CLI client
	BuildVersion = "*not set*"

	// BuildTime contains timestamp when the CLI client has been built
	BuildTime = "*not set*"

	// BuildBranch contains Git branch used to build this application
	BuildBranch = "*not set*"

	// BuildCommit contains Git commit used to build this application
	BuildCommit = "*not set*"
)

func printInfo(msg string, val string) {
	fmt.Printf("%s\t%s\n", msg, val)
}

func PrintVersionInfo() {
	printInfo("Version:", BuildVersion)
	printInfo("Build time:", BuildTime)
	printInfo("Branch:", BuildBranch)
	printInfo("Commit:", BuildCommit)
}
