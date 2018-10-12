FROM golang:1.10
WORKDIR $GOPATH/src/github.com/priyawadhwa/webhook
COPY . .
ENTRYPOINT go run main.go
