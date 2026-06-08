IMAGE=rebuilder
IMAGE_VERSION=1.4.0
GOVERSION=1.26.4

.PHONY: dockerhub  ghcr

dockerhub:
	buildah build --build-arg GOVERSION=${GOVERSION} --build-arg REBUILDER_VERSION=${IMAGE_VERSION} -t ${IMAGE}:${IMAGE_VERSION} .
	#buildah login docker.io
	buildah push ${IMAGE}:${IMAGE_VERSION} gounix/${IMAGE}:${IMAGE_VERSION}

ghcr:
	buildah build --build-arg GOVERSION=${GOVERSION} --build-arg REBUILDER_VERSION=${IMAGE_VERSION} -t ${IMAGE}:${IMAGE_VERSION} .
	#buildah login ghcr.io
	buildah push ${IMAGE}:${IMAGE_VERSION} ghcr.io/gounix/${IMAGE}:${IMAGE_VERSION}

