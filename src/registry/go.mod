module rebuilder/registry

go 1.22.2

require (
        rebuilder/dockerhub v0.0.0-unpublished
        rebuilder/dockerregistry v0.0.0-unpublished
)

replace rebuilder/dockerhub v0.0.0-unpublished => ../dockerhub
replace rebuilder/dockerregistry v0.0.0-unpublished => ../dockerregistry

