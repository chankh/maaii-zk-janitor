package version

import (
	"fmt"
	"runtime"
)

// VersionMajor holds the release major number
const VersionMajor = 1

// VersionMinor holds the release minor number
const VersionMinor = 0

// VersionPatch holds the release patch number
const VersionPatch = 0

// Version holds the combination of major, minor and patch as a string
// of format Major.Minor.Patch
var Version = fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)

// BuildHash is filled with the Git revision being used to build the
// program at linking time
var BuildHash = ""

// BuildNumber is filled with the Git revision being used to build the
// program at linking time
var BuildNumber = ""

// BuildDate in ISO8601 format, is filled when building the program
// at linking time
var BuildDate = "1970-01-01T00:00:00Z"

// BuildPlatform returns the OS platform that this build was compiled for
var BuildPlatform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

// GoVersion returns the version of the Go compiler used during build
var GoVersion = runtime.Version()
