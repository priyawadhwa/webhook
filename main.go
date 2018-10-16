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
		log.Printf("found pull request %d", prEvent.GetNumber())
		if err := handlePREvent(prEvent); err != nil {
			log.Printf("error handling pr event: %v", err)
		}

	} else if eventType == constants.LabelEvent {
		labelEvent := new(github.LabelEvent)
		if err := json.NewDecoder(r.Body).Decode(labelEvent); err != nil {
			log.Printf("error decoding label event: %v", err)
		}
		log.Print(labelEvent)
	} else {
		log.Printf("Event %s not supported yet.\n", eventType)
	}
}

func handlePREvent(event *github.PullRequestEvent) error {
	if !docsLabelExists(event.PullRequest.Labels) {
		log.Printf("label %s not found on PR %d", constants.DocsLabel, event.GetNumber())
		return nil
	}
	d, err := kubernetes.CreateDeployment(event)
	if err != nil {
		return errors.Wrap(err, "creating deployment")
	}
	ip, err := kubernetes.CreateServiceFromDeployment(d)
	if err != nil {
		return errors.Wrap(err, "creating service from deployment")
	}
	return pkggithub.CommentOnPr(client, event, ip)
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
