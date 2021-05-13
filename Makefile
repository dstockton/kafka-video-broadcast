VERSION := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date +%F_%T)
VCS_URL := $(shell basename `git rev-parse --show-toplevel`)
VCS_REF := $(CI_COMMIT_SHORT_SHA)
NAME := $(shell basename `git rev-parse --show-toplevel`)
VENDOR := dstockton
DOCKER_IMAGE_NAME_TAG := $(shell echo "$${CI_REGISTRY_IMAGE:-$(NAME)}:$(VCS_REF)")
COVERAGE_REQUIRED := 80

.PHONY: debug
debug:
	@LOG_LEVEL=debug go run main.go

.PHONY: run
run:
	@go run main.go

.PHONY: build
build:
	@go build

.PHONY: fmt
fmt:
	@echo "Formatting..."
	@go fmt `go list ./...`

.PHONY: vet
vet:
	@echo "Vetting..."
	@go vet `go list ./...`

.PHONY: test
test: fmt vet test-only

.PHONY: test-only
test-only:
	@echo "Testing..."
	@go test -race -coverprofile cp.out `go list ./...`
	@export COVERAGE_LEVEL=`go tool cover -func=cp.out | grep "total:" | awk '{print $$3}' | cut -d'.' -f1`; \
	echo "\n\nOverall coverage level achieved: $${COVERAGE_LEVEL}% (require ${COVERAGE_REQUIRED}%)"; \
	if [ "$${COVERAGE_LEVEL}" -lt "${COVERAGE_REQUIRED}" ]; then \
		echo "FAILURE - below required test code coverage level!"; \
		exit 1; \
	fi

.PHONY: print
print:
	@echo VERSION=${VERSION} 
	@echo BUILD_DATE=${BUILD_DATE}
	@echo VCS_URL=${VCS_URL}
	@echo VCS_REF=${VCS_REF}
	@echo NAME=${NAME}
	@echo VENDOR=${VENDOR}
	@echo DOCKER_IMAGE_NAME_TAG=${DOCKER_IMAGE_NAME_TAG}

.PHONY: build-image
build-image:
	@docker build --pull -t "${DOCKER_IMAGE_NAME_TAG}" --build-arg VERSION="${VERSION}" \
	--build-arg BUILD_DATE="${BUILD_DATE}" \
	--build-arg VCS_URL="${VCS_URL}" \
	--build-arg VCS_REF="${VCS_REF}" \
	--build-arg NAME="${NAME}" \
	--build-arg VENDOR="${VENDOR}" \
	--build-arg 'GOFLAGS=${GOFLAGS}' \
	.

.PHONY: push-image
push-image: build-image
	@docker push "${DOCKER_IMAGE_NAME_TAG}"
