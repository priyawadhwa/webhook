package kubernetes

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func clientSet() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %s", err)
	}
	return kubernetes.NewForConfig(config)
}

func serviceNameFromPR(prNumber int) string {
	return fmt.Sprintf("docs-controller-%d-svc", prNumber)
}

func getService(svc *v1.Service) (*v1.Service, error) {
	clientset, err := clientSet()
	if err != nil {
		return nil, errors.Wrap(err, "getting clientset")
	}
	return clientset.CoreV1().Services(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
}

// Returns the external IP where the server is being hosted
func getExternalIP(s *v1.Service) (string, error) {
	err := wait.PollImmediate(time.Second*5, time.Minute*3, func() (bool, error) {
		svc, err := getService(s)
		if err != nil {
			return false, nil
		}
		return len(svc.Status.LoadBalancer.Ingress) > 0, nil
	})
	svc, err := getService(s)
	if err != nil {
		return "", err
	}
	ip := svc.Status.LoadBalancer.Ingress[0].IP
	return ip, nil
}

// CreateService creates a service for the deployment to bind to
// and returns the external IP of the service
func CreateService(pr *github.PullRequestEvent) (string, error) {
	clientset, err := clientSet()
	if err != nil {
		return "", errors.Wrap(err, "getting clientset")
	}
	labels := generateLabelsFromPullRequestEvent(pr)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   serviceNameFromPR(pr.GetNumber()),
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Port: 1313,
				},
			},
			Selector: serviceSelectorLabel(pr.GetNumber()),
		},
	}
	svc, err = clientset.CoreV1().Services("default").Create(svc)
	if err != nil {
		return "", errors.Wrap(err, "creating service")
	}
	return getExternalIP(svc)
}
