package version

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

var spDumpVersion = "0.0.1"
var prerelease = ""
var SpDumpVersion *semver.Version
var VersionString string

func init() {
	VersionString = spDumpVersion
	if prerelease != "" {
		VersionString = fmt.Sprintf("%s-%s", spDumpVersion, prerelease)
	}
	SpDumpVersion = semver.MustParse(VersionString)
}
