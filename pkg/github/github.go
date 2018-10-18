package github

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/priyawadhwa/webhook/pkg/util"
)

const (
	owner = "GoogleContainerTools"
)

// CommentOnPr comments on the PR to visit IP to see changes to docs
func CommentOnPr(client *github.Client, pr *github.PullRequestEvent, ip string) error {
	log.Printf("trying to comment on PR %d now", pr.PullRequest.GetNumber())
	ctx := context.Background()
	url := util.GetWebsiteURL(ip)
	comment := &github.IssueComment{
		Body: &[]string{fmt.Sprintf("Please visit [%s](%s) to view changes to the docs.", url, url)}[0],
	}

	log.Printf("Creating comment on %s %s %d", owner, *pr.Repo.Name, pr.PullRequest.GetNumber())
	_, _, err := client.Issues.CreateComment(ctx, owner, *pr.Repo.Name, pr.PullRequest.GetNumber(), comment)
	if err != nil {
		return errors.Wrap(err, "creating github comment")
	}
	log.Printf("Succesfully commented on PR.")
	return nil
}
