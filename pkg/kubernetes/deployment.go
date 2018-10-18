package kubernetes

import (
	"context"
	"fmt"
	"path"
	"time"

	"log"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/priyawadhwa/webhook/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	typedv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const (
	image        = "gcr.io/priya-wadhwa/hugo@sha256:29a9ef0427218772b3d6a982feac3d855b1949cc78326ad1aeea8af32fa2ba38"
	repo         = "https://github.com/GoogleContainerTools/skaffold.git"
	emptyVol     = "empty-vol"
	emptyVolPath = "/empty"
	port         = 1313
)

func getDeployment(pr *github.PullRequestEvent, ip string) *appsv1.Deployment {
	labels := generateLabelsFromPullRequestEvent(pr)

	userRepo := fmt.Sprintf("https://github.com/%s.git", *pr.PullRequest.Head.Repo.FullName)
	docsPath := path.Join(emptyVolPath, *pr.PullRequest.Head.Repo.Name, "docs")

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("docs-controller-deployment-%d", pr.PullRequest.GetNumber()),
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:       "git-clone",
							Image:      image,
							Args:       []string{"git", "clone", userRepo, "--branch", pr.PullRequest.Head.GetRef()},
							WorkingDir: emptyVolPath,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      emptyVol,
									MountPath: emptyVolPath,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:       "server",
							Image:      image,
							Args:       []string{"hugo", "server", "--bind=0.0.0.0", "-D", "--baseURL", util.GetWebsiteURL(ip)},
							WorkingDir: docsPath,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      emptyVol,
									MountPath: emptyVolPath,
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: port,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: emptyVol,
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	return d
}

func deploymentsClient() (typedv1.DeploymentInterface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "getting clientset")
	}

	deploymentsClient := clientset.AppsV1().Deployments("default")
	return deploymentsClient, nil
}

// WaitForDeploymentToStabilize waits till the Deployment has a matching generation/replica count between spec and status.
// TODO: handle ctx.Done()
func WaitForDeploymentToStabilize(ctx context.Context, deployments typedv1.DeploymentInterface, name string) error {
	return wait.PollImmediateUntil(time.Millisecond*500, func() (bool, error) {
		d, err := deployments.Get(name, metav1.GetOptions{
			IncludeUninitialized: true,
		})
		if err != nil {
			log.Printf("Getting deployment %s", err)
			return false, err
		}
		return d.Status.ReadyReplicas == 1, nil
	}, ctx.Done())
}

func waitForDeployment(d *appsv1.Deployment) error {
	client, err := clientSet()
	if err != nil {
		return errors.Wrap(err, "getting clientset")
	}
	return WaitForDeploymentToStabilize(context.Background(), client.AppsV1().Deployments(d.Namespace), d.Name)
}

func CreateDeployment(pre *github.PullRequestEvent, ip string) error {
	client, err := deploymentsClient()
	if err != nil {
		return errors.Wrap(err, "error getting clientset")
	}
	dep := getDeployment(pre, ip)
	d, err := client.Create(dep)
	if err != nil {
		return errors.Wrap(err, "creating deployment")
	}
	return waitForDeployment(d)
}
