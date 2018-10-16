package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/go-github/github"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	typedv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const (
	image        = "gcr.io/priya-wadhwa/hugo@sha256:29a9ef0427218772b3d6a982feac3d855b1949cc78326ad1aeea8af32fa2ba38"
	repo         = "https://github.com/priyawadhwa/docs.git"
	emptyVol     = "empty-vol"
	emptyVolPath = "/empty"
	docsPath     = "/empty/docs"
	port         = 1313
)

func getDeployment(pr *github.PullRequestEvent) *appsv1.Deployment {
	labels := generateLabelsFromPullRequestEvent(pr)

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
							Args:       []string{"git", "clone", repo},
							WorkingDir: emptyVolPath,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      emptyVol,
									MountPath: emptyVolPath,
								},
							},
						},
						{
							Name:       "curl-patch",
							Image:      image,
							Args:       []string{"curl", "-L", pr.PullRequest.GetPatchURL(), "-o", "patch"},
							WorkingDir: emptyVolPath,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      emptyVol,
									MountPath: emptyVolPath,
								},
							},
						},
						{
							Name:       "git-patch",
							Image:      image,
							Args:       []string{"git", "apply", "../patch"},
							WorkingDir: docsPath,
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
							Args:       []string{"hugo", "server", "--bind=0.0.0.0", "-D"},
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

func CreateDeployment(pre *github.PullRequestEvent) (*appsv1.Deployment, error) {
	client, err := deploymentsClient()
	if err != nil {
		return nil, errors.Wrap(err, "error getting clientset")
	}
	d := getDeployment(pre)
	return client.Create(d)
}
