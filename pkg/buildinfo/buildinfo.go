package buildinfo

import (
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildTime = ""
)

type Info struct {
	Binary      string `json:"binary"`
	Version     string `json:"version"`
	Commit      string `json:"commit,omitempty"`
	BuildTime   string `json:"build_time,omitempty"`
	GoVersion   string `json:"go_version"`
	Platform    string `json:"platform"`
	VCSModified bool   `json:"vcs_modified"`
}

func Current(binary string) Info {
	info := Info{
		Binary:    strings.TrimSpace(binary),
		Version:   strings.TrimSpace(Version),
		Commit:    strings.TrimSpace(Commit),
		BuildTime: strings.TrimSpace(BuildTime),
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if info.Version == "" || info.Version == "dev" {
			if version := strings.TrimSpace(buildInfo.Main.Version); version != "" && version != "(devel)" {
				info.Version = version
			}
		}
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if info.Commit == "" {
					info.Commit = shortCommit(setting.Value)
				}
			case "vcs.time":
				if info.BuildTime == "" {
					info.BuildTime = strings.TrimSpace(setting.Value)
				}
			case "vcs.modified":
				info.VCSModified = strings.EqualFold(strings.TrimSpace(setting.Value), "true")
			}
		}
	}

	if info.Version == "" {
		info.Version = "dev"
	}
	return info
}

func shortCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}
