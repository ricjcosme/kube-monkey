all: build

ENVVAR = GOOS=linux GOARCH=amd64 CGO_ENABLED=0
TAG = latest

build: clean
	$(ENVVAR) go build -o kube-monkey

container: build
	docker build -t ricjcosme/kube-monkey:$(TAG) .

clean:
	rm -f kube-monkey

.PHONY: all build container clean
