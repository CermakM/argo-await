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

OBSERVER           ?= argo-await-observer
NAMESPACE          ?= $(shell oc whoami --show-context | cut -d'/' -f 1)

BUILDCONFIG        ?= ${OBSERVER}-buildconfig
BUILDCONFIG_EXISTS := $(shell oc get -n ${NAMESPACE} buildconfigs ${BUILDCONFIG} &> /dev/null && echo 0 || echo 1)

IMAGESTREAM        ?= ${OBSERVER}

.PHONY: all
all: build image

.PHONY: all-linux
all-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make all


.PHONY: observer
observer:
	GO111MODULE=on go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${OBSERVER} ./observer/cmd

.PHONY: observer-linux
observer-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make observer

observer-image: observer-linux
	docker build --rm -t ${OBSERVER}:latest -f ${CURRENT_DIR}/observer/Dockerfile .

observer-image-openshift: archive = dist.tar.gz
observer-image-openshift: observer-linux
ifeq ($(BUILDCONFIG_EXISTS), 1)
	oc -n ${NAMESPACE} new-build --strategy docker --binary --name ${BUILDCONFIG} --to ${IMAGESTREAM}
endif
	tar -czf ${archive} $(shell basename ${DIST_DIR}) \
	    -C observer/ --add-file Dockerfile
	oc -n ${NAMESPACE} start-build ${BUILDCONFIG} --from-archive ${archive} --follow