package version

import (
	"os/exec"
	"strings"
)

var (
	// GitVersion stores the git version
	GitVersion = getVersion()
	// GitRevision stores the git revision
	GitRevision = getRevision()
)

func getVersion() string {
	cmd := exec.Command("bash", "-c", "git describe --tags --abbrev=0 || echo 'v0.0.0'")
	versionBytes, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(versionBytes))
}

func getRevision() string {
	branchCmd := exec.Command("bash", "-c", "git rev-parse --abbrev-ref HEAD")
	branchBytes, err := branchCmd.Output()
	if err != nil {
		return "unknown"
	}

	commitCmd := exec.Command("bash", "-c", "git rev-parse --short HEAD")
	commitBytes, err := commitCmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(branchBytes)) + "-" + strings.TrimSpace(string(commitBytes))
}
