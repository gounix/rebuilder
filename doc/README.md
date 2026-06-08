# Overview

The rebuilder deployment makes it possible to automatically update your images once a base image is updated. This can be very convenient to mitigate against vulnerabilities. There are a number of images on hub.docker.com that get regular updates without getting a new tag. With the rebuilder deployment you create a manifest that models the dependency and specifies how the rebuild should be done. The rebuilder deployment does this by checking the created date of the base and derived images.

# Installation

Get the helm chart from [github](https://github.com/gounix/rebuilder/tree/main/helm-charts)
There following paramers should be changed in the provided values.yaml:

The deployment runs a cronjob, the schedule defines when the job should be run
```
schedule: "0 17 * * *"
```

The next environment settings apply to the builder container that is spawned to do the actual rebuilding. The BUILDER_NAMESPACE should match the namespace where the rebuilder deployment is installed, for this example we will use rebuilder.
```
env:
  BUILDER_REPO: "docker.io"
  BUILDER_IMAGE: "gounix/builder"
  BUILDER_TAG: "1.2.0"
  BUILDER_NAMESPACE: "rebuilder"
  REBUILDER_NAMESPACE: "rebuilder"
```
The next section specifies the rebuilder image, these settings can be left as is unless you are mirroring the image to a local repo.
```
image:
  repository: "docker.io/gounix/rebuilder"
  # This sets the pull policy for images.
  pullPolicy: Always
  # Overrides the image tag whose default is the chart appVersion.
  tag: "1.4.0"
```
On the first deployment the custom resource definitions will be loaded:
```
$ kubectl api-resources
NAME                                SHORTNAMES                         APIVERSION                          NAMESPACED   KIND
rebuilds                                                               gounix.nl/v1                        true         Rebuild
```

# Extending deployments

When the basic installation is performed, we can continue to model all images. This can be done by adding a rebuild.yaml to an existing deployment. The basic rebuild.yaml looks as follows:
```
apiVersion: gounix.nl/v1
kind: Rebuild
metadata:
  name: kube-sec-board
spec:
  base:
    host: "registry-1.docker.io"
    type: "dockerHub"
    image: "library/python"
    tag: "3.14.2-slim"
    authenticated: true
    secretName: "rebuilder/regcred"
  git:
    host: "git.int.gounix.nl"
    project: "images/kube-sec-board.git"
    user: "git"
    dir: "."
    tag: "1.0"
    secretName: "ssh-key"
    sshKeyName: "id_rsa"
  registry:
    host: "prive.gounix.nl"
    type: "dockerRegistry"
    image: "kube-sec-board"
    tag: "1.0"
    authenticated: false
    secretName: ""
  actions:
    - objecttype: deployment
      name: kube-sec-board
```
In this example spec.base specifies the base image. Our custom image is built on-top of docker.io/library/python:3.14.2-slim. The spec.base.type and spec.registry.type fields can contain "dockerHub", "ghcr" or "dockerRegistry". In which dockerRegistry can be used for any registry that uses the docker registry v2 api (Quay.io for example). If a registry is authenticated set base.authenticated to true and specify the fully qualified path of a valid registry credential.  

The spec.registry section specifies where our derived image is stored, in this case kube-sec-board:1.0 on our private registry that uses no authentication.   

The spec.git section specifies how the derived image can be build. In this case it will git clone git@git.int.gounix.nl:images/kube-sec-board.git and it expects a Makefile in the "." directory.  

The spec.actions section specifies post build actions that should be performed, for example restarting a deployment. The objecttypes that can be specified are deployment, daemonset, statefulset and replicaset.  

# Example Makefile

A simple Makefile to rebuild an image can be as simple as the next example:
```
REGISTRY=registry-tst.int.gounix.nl
IMAGE=kube-sec-board
IMAGE_VERSION=1.0

.PHONY: target

target:
        buildah build -t ${IMAGE}:${IMAGE_VERSION} .
        buildah push ${IMAGE}:${IMAGE_VERSION} docker://${REGISTRY}/${IMAGE}:${IMAGE_VERSION}
```

# Adding secrets

The secret that is used in the spec.git section of the rebuild.yaml can be created in the following way:
```
kubectl create secret -n rebuilder generic ssh-key   --from-file=./id_rsa
```
This secret should be created in the rebuilder namespace.  
The registry credentials are stored in a standard docker-registry secret and can be created in the following way
```
kubectl create secret docker-registry regcred --docker-server=ghcr.io --docker-username=<your-github-username> --docker-password=<your-personal-access-token> --docker-email=<your-email>
```
You can also use your favourite way(like vault for example) to create secrets in a secure way.

# Notes about pullPolicy

The rebuilder deployment does not change version numbers, it just rebuild existing images. To make sure kubernetes will pull the new image the pullPolicy should be set to Always on all deployments that use rebuild.

# Builder image

The builder image contains the following software:
- a ssh client to fetch git repos
- ca-certificates to trust registry certificates
- make to build images from a Makefile
- buildah for the actual building
- wget and curl to fetch other dependencies

If you need other software you can derive a custom image from the builder image and add additional software.

# Migration notes

When migrating to 1.2.0 the new helm chart should be used since it includes a change in the CRD. The builder image needs to be at version 1.2.0.
When migrating to 1.4.0 the new helm chart should be used since a new environment variable is present.

# Change history

* 1.0 5/12/2026 Initial version.
* 1.1.0 5/26/2026 rebuilder checks if an image is newer than a running pod and if so restarts the pod.
* 1.2.0 5/28/2026 The git section of the rebuild.yaml now supports a tag that can be used to checkout a specific version
* 1.3.0 6/4/2026 Merged all registry code into one
* 1.4.0 6/8/2026 Allow authentication on registries
