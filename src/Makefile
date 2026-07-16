IMAGE=rebuilder
IMAGE_VERSION=2.2.2
GOVERSION=1.26

.PHONY: dockerhub 

dockerhub:
	buildah build --build-arg GOVERSION=${GOVERSION} --build-arg REBUILDER_VERSION=${IMAGE_VERSION} -t ${IMAGE}:${IMAGE_VERSION} .
	#buildah login docker.io
	buildah push ${IMAGE}:${IMAGE_VERSION} gounix/${IMAGE}:${IMAGE_VERSION}

