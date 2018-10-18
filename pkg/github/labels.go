package github

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/priyawadhwa/webhook/pkg/constants"
)

func RemoveDocsLabel(client *github.Client, pr *github.PullRequestEvent) error {
	ctx := context.Background()
	_, err := client.Issues.DeleteLabel(ctx, owner, pr.Repo.GetName(), constants.DocsLabel)
	if err != nil {
		return errors.Wrap(err, "deleting label")
	}
	log.Printf("Successfully deleted label from PR %d", pr.GetNumber())
	return nil
}
