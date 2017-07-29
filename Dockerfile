FROM golang:1.8.3-alpine3.6

ENV GOPATH /go
ENV GOROOT /usr/local/go

COPY kube-monkey /kube-monkey
