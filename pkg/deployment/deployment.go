package deployment

import (
	"fmt"

	"github.com/google/go-github/github"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getDeployment(pr github.PullRequestEvent) *appsv1.Deployment {
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("docs-deployment-%s", pr.PullRequest.Number),
		},
	}
	return d
}

func repo() string {
https: //github.com/Zetten/kaniko.git enhance-is-dest-dir
}
