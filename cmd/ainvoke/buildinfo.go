package main

import (
	"fmt"
	"runtime/debug"
	"sync"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
	loadOnce  sync.Once
)

func ensureBuildInfo() {
	loadOnce.Do(func() {
		if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
			version = info.Main.Version
		}
	})
}

func getVersion() string {
	ensureBuildInfo()

	return version
}

func getGitCommit() string {
	ensureBuildInfo()

	return gitCommit
}

func getBuildDate() string {
	ensureBuildInfo()

	return buildDate
}

func versionString() string {
	return fmt.Sprintf(
		"ainvoke version %s (git: %s, built: %s)",
		getVersion(),
		getGitCommit(),
		getBuildDate(),
	)
}
