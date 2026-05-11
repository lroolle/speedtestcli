package version

var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = ""
)

func Full() string {
	if GitCommit == "none" || GitCommit == "" {
		return Version
	}
	short := GitCommit
	if len(short) > 7 {
		short = short[:7]
	}
	return Version + " (" + short + ")"
}
