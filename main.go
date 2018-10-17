package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/priyawadhwa/webhook/pkg/constants"
	pkggithub "github.com/priyawadhwa/webhook/pkg/github"
	"github.com/priyawadhwa/webhook/pkg/kubernetes"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	githubToken = os.Getenv("GITHUB_ACCESS_TOKEN")

	client *github.Client
)

func main() {
	// Setup the token for github authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken}, //pfortin-urbn token
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Get a client instance from github
	client = github.NewClient(tc)
	// Github is now ready to receive information from us!!!

	//Setup the serve route to receive guthub events
	http.HandleFunc("/receive", handleGithubEvent)

	// Start the server
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//This is the event handler for github events.
func handleGithubEvent(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")

	if eventType == constants.PullRequestEvent {
		prEvent := new(github.PullRequestEvent)
		if err := json.NewDecoder(r.Body).Decode(prEvent); err != nil {
			log.Printf("error decoding pr event: %v", err)
		}
		if err := handlePREvent(prEvent); err != nil {
			log.Printf("error handling pr event: %v", err)
		}

		// } else if eventType == constants.LabelEvent {
		// 	labelEvent := new(github.LabelEvent)
		// 	if err := json.NewDecoder(r.Body).Decode(labelEvent); err != nil {
		// 		log.Printf("error decoding label event: %v", err)
		// 	}
		// 	log.Println("label event")
		// 	if err := handleLabelEvent(labelEvent); err != nil {
		// 		log.Printf("error handling label event: %v", err)
		// 	}

	}
}

func handleLabelEvent(event *github.LabelEvent) error {
	if event.Label.GetName() != constants.DocsLabel {
		return nil
	}
	if event.Action != nil && *event.Action != "created" {
		return nil
	}
	return nil
}

func handlePREvent(event *github.PullRequestEvent) error {
	if event.GetAction() == "closed" {
		return cleanup(event)
	}

	if event.GetAction() != "labeled" {
		return nil
	}

	if event.GetPullRequest() == nil {
		return nil
	}

	// Make sure pull request is open
	if event.GetPullRequest().GetMerged() {
		return nil
	}

	log.Printf("Received event for pull request %d", prEvent.GetNumber())

	if !docsLabelExists(event.PullRequest.Labels) {
		log.Printf("Label %s not found on PR %d", constants.DocsLabel, event.GetNumber())
		return nil
	}

	if err := cleanup(event); err != nil {
		return errors.Wrap(err, "cleaning up")
	}

	log.Printf("Found %s label on pull request %d, creating deployment", constants.DocsLabel, event.GetNumber())

	d, err := kubernetes.CreateDeployment(event)
	if err != nil {
		return errors.Wrap(err, "creating deployment")
	}
	log.Printf("Creating service from deployment %s", d.Name)
	ip, err := kubernetes.CreateServiceFromDeployment(d)
	if err != nil {
		return errors.Wrap(err, "creating service from deployment")
	}

	log.Printf("Commenting on PR %d that IP %s is ready", event.GetNumber(), ip)
	if err := pkggithub.CommentOnPr(client, event, ip); err != nil {
		return errors.Wrap(err, "commenting on pr")
	}

	log.Printf("Removing %s label from pull request %d", constants.DocsLabel, event.GetNumber())
	return pkggithub.RemoveDocsLabel(client, event)
}

func cleanup(event *github.PullRequestEvent) error {
	log.Printf("Cleaning up deployment for pull request %d", event.GetNumber())
	return kubernetes.CleanupDeployment(event)
}

func docsLabelExists(labels []*github.Label) bool {
	for _, l := range labels {
		if l != nil {
			if *l.Name == constants.DocsLabel {
				return true
			}
		}
	}
	return false
}
