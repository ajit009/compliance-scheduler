IMAGE_REPO=ajlocalau-docker-prod-public.jfrog.io/ops/dalek
TAG=$(shell cut -d'=' -f2- .release)

.DEFAULT_GOAL := build
.PHONY: release check-git-status build container-image pre-build tag-image publish test system-check

help:
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t

release: check-git-status test container-image tag-image publish
	@echo "Successfully releeased version $(TAG)"

# Check git status
check-git-status:
	@echo "Checking git status"
	@if [ -n "$(shell git tag | grep $(TAG))" ] ; then echo 'ERROR: Tag already exists' && exit 1 ; fi
	@if [ -z "$(shell git remote -v)" ] ; then echo 'ERROR: No remote to push tags to' && exit 1 ; fi
	@if [ -z "$(shell git config user.email)" ] ; then echo 'ERROR: Unable to detect git credentials' && exit 1 ; fi

# test
test: system-check
	@echo "Starting unit and integration tests"
	@./hack/runtests.sh

# Build the binary
build: pre-build
	@cd cmd/dalek;GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -o $(PWD)/build/bin/dalek
	@echo "Build completed successfully"

run_sch: build
	@echo "Running dalek"
	@$(PWD)/build/bin/dalek
	
# Build the image
container-image: pre-build
	@echo "Building docker image"
	@docker build --build-arg GOOS_VAL=$(shell go env GOOS) --build-arg GOARCH_VAL=$(shell go env GOARCH) -t $(IMAGE_REPO) -f Dockerfile --no-cache .
	@echo "Docker image build successfully"

# system checks
system-check:
	@echo "Checking system information"
	@if [ -z "$(shell go env GOOS)" ] || [ -z "$(shell go env GOARCH)" ] ; \
	then \
	echo 'ERROR: Could not determine the system architecture.' && exit 1 ; \
	else \
	echo 'GOOS: $(shell go env GOOS)' ; \
	echo 'GOARCH: $(shell go env GOARCH)' ; \
	echo 'System information checks passed.'; \
	fi ;

# Pre-build checks
pre-build: system-check

# Tag images
tag-image: container-image
	@echo 'Tagging image'
	@docker tag $(IMAGE_REPO) $(IMAGE_REPO):$(TAG)

# Docker push image
publish: tag-image
	@echo "Pushing docker image to repository"
	# @docker login
	@docker push $(IMAGE_REPO):$(TAG)
	@docker push $(IMAGE_REPO):latest

