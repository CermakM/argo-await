PACKAGE 		= github.com/CermakM/argo-await

CURRENT_DIR     = $(shell pwd)
DIST_DIR       ?= ${CURRENT_DIR}/dist

VERSION         = $(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE      = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT      = $(shell git rev-parse HEAD)
GIT_TAG         = $(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE  = $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

EXEC                = argo-await
NAMESPACE          ?= $(shell oc whoami --show-context | cut -d'/' -f 1)     # OpenShift Namespace

BUILDCONFIG        ?= argo-await  # OpenShift Build Config name
BUILDCONFIG_EXISTS := $(shell oc get -n ${NAMESPACE} buildconfigs ${BUILDCONFIG} &> /dev/null && echo 0 || echo 1)

.PHONY: all
all: build image

.PHONY: all-linux
all-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make all

.PHONY: build
build:
	- rm -r ${DIST_DIR}
	GO111MODULE=on go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${EXEC} ./cmd

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

docker: build
	docker build --rm -t ${EXEC}:latest ${CURRENT_DIR}

docker-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make docker

image: build
ifeq ($(BUILDCONFIG_EXISTS), 1)
	$(info  Creating build config ${BUILDCONFIG} )
	oc -n ${NAMESPACE} new-build --strategy docker --binary --docker-image scratch --name ${BUILDCONFIG}
endif
	oc -n ${NAMESPACE} start-build argo-await --from-dir . --follow

image-linux: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make image
