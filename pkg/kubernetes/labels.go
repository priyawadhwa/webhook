package kubernetes

import (
	"fmt"

	"github.com/google/go-github/github"
)

func generateLabelsFromPullRequestEvent(pr *github.PullRequestEvent) map[string]string {
	labels := map[string]string{
		"deployment": fmt.Sprintf("docs-controller-deployment-%d", pr.PullRequest.GetNumber()),
	}
	labels["docs-controller-deployment"] = "true"
	return labels
}

func getPRLabel(prNumber int) string {
	return fmt.Sprintf("deployment=docs-controller-deployment-%d", prNumber)
}
