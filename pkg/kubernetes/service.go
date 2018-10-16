package kubernetes

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func serviceName(d *appsv1.Deployment) string {
	return fmt.Sprintf("%s-svc", d.Name)
}

func getService(d *appsv1.Deployment) (*v1.Service, error) {
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
	return clientset.CoreV1().Services("default").Get(serviceName(d), metav1.GetOptions{})
}

// Returns the external IP where the server is being hosted
func getExternalIP(d *appsv1.Deployment) (string, error) {
	err := wait.PollImmediate(time.Millisecond*500, time.Minute*3, func() (bool, error) {
		svc, err := getService(d)
		if err != nil {
			return false, nil
		}
		return len(svc.Spec.ExternalIPs) > 0, nil
	})
	svc, err := getService(d)
	if err != nil {
		return "", err
	}
	ip := svc.Spec.ExternalIPs[0]
	log.Printf("Returning ip address %s:1313 for service %s", ip, serviceName(d))
	return fmt.Sprintf("%s:%s", ip, "1313"), nil
}

// CreateServiceFromDeployment creates a service from the deployment
func CreateServiceFromDeployment(d *appsv1.Deployment) (string, error) {
	cmd := exec.Command("kubectl", "expose", "deployment", d.Name, "--name", serviceName(d), "--type=LoadBalancer")
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return getExternalIP(d)
}
