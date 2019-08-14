PACKAGE=github.com/CermakM/kube-await

CURRENT_DIR=$(shell pwd)
DIST_DIR?=${CURRENT_DIR}/dist

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

EXEC=kube-await

NAMESPACE?=default     # OpenShift Namespace

BUILDCONFIG?=kube-await  # OpenShift Build Config name
BUILDCONFIG_EXISTS := $(shell oc get -n ${NAMESPACE} buildconfigs ${BUILDCONFIG} &> /dev/null && echo 0 || echo 1)

all: build deploy
.PHONY: all

build:
	rm -r ${DIST_DIR}
	GO111MODULE=on go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${EXEC} ./cmd

deploy: build
ifeq ($(BUILDCONFIG_EXISTS), 1)
	$(info  Creating build config ${BUILDCONFIG} )
	oc -n ${NAMESPACE} new-build --strategy docker --binary --docker-image scratch --name ${BUILDCONFIG}
endif
	oc -n ${NAMESPACE} start-build kube-await --from-dir . --follow
