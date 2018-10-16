

.PHONY: image
image: 
	docker build -t gcr.io/k8s-skaffold/docs-controller:latest .
	docker push gcr.io/k8s-skaffold/docs-controller:latest

