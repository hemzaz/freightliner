package banner

import (
	"fmt"
	"runtime"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Logo is the ASCII art freightliner truck
const Logo = `
    _______________________________________________
   |  ___________________________________________  |
   | |                                           | |
   | |     FREIGHTLINER                          | |
   | |     Container Registry Replication        | |
   | |___________________________________________| |
   |_______________________________________________|
    __||__||__||__||__||__||__||__||__||__||__||__
   |______________________________________________|
   /        ___/      \___      ___/      \___    \
  /_________[_]________[_]____[_]________[_]______\
           (o)        (o)    (o)        (o)
`

// SmallLogo is a compact version
const SmallLogo = `
   _________________
  |  FREIGHTLINER  |
  |________________|
     (o)      (o)
`

// Print displays the full banner with version info
func Print() {
	fmt.Print(Logo)
	fmt.Printf("  Version: %s | Commit: %s | Built: %s\n", Version, GitCommit, BuildTime)
	fmt.Printf("  Runtime: Go %s %s/%s\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// PrintSmall displays the compact banner
func PrintSmall() {
	fmt.Print(SmallLogo)
	fmt.Printf("  v%s\n\n", Version)
}

// PrintVersion displays version information only
func PrintVersion() {
	fmt.Printf("Freightliner v%s\n", Version)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Built: %s\n", BuildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
