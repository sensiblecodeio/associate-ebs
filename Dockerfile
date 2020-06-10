FROM golang:1.14.4-alpine

# Turn off cgo for a "more static" binary.
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/sensiblecodeio/associate-ebs

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install -v
