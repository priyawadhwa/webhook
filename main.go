package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	githubToken = ""

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
	http.HandleFunc("/receive", EventHandler)

	// Start the server
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//This is the event handler for github events.
func EventHandler(w http.ResponseWriter, r *http.Request) {
	event_type := r.Header.Get("X-GitHub-Event")

	if event_type == "pull_request" {
		pr_event := new(github.PullRequestEvent)

		json.NewDecoder(r.Body).Decode(pr_event)
		if pr_event.PullRequest.State != nil {

			log.Printf("Event Type: %s, Created by: %s\n", event_type, pr_event.PullRequest.Base.User.Login)

			log.Println("Handler exiting...")
		} else {
			log.Println("PR state not open or reopen")
		}
	} else {
		log.Printf("Event %s not supported yet.\n", event_type)
	}
}
