package kubernetes

import (
	"os/exec"

	"github.com/google/go-github/github"
)

func CleanupDeployment(pr *github.PullRequestEvent) error {
	cmd := exec.Command("kubectl", "delete", "all", "--selector", getPRLabel(pr.GetNumber()))
	return cmd.Run()
}
