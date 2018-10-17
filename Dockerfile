FROM golang:1.10

ENV KUBECTL_VERSION v1.12.0
RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl  && \
    chmod +x /usr/local/bin/kubectl

WORKDIR $GOPATH/src/github.com/priyawadhwa/webhook
COPY . .
ENTRYPOINT go run main.go
