# Local Operator Development

## Prerequisites

* Ensure you have access to a Kubernetes cluster via `kubectl` with cluster admin permissions.
* Install Go version 1.15 or higher.

## Building locally

Run `make build` to compile the operator binary.
The resulting binary will be saved to `bin/operator`.

To build the container image for local use, run `make ko-publish IMAGE_REPO=ko.local`.
This will do the following:

1. Compile the application.
2. Create a container image.
3. Push the container image to your local Docker daemon, with the ref `ko.local/operator:<TAG>`.

The following make options can be set:

* `IMAGE_REPO` - defaults to `quay.io/shipwright`.
  The following special repositories can be used for testing:

  * `ko.local` - this pushes the image to the local Docker daemon.
  * `kind.local` - pushes to a local KinD cluster.

* `TAG` - defaults to `latest`.
* `IMAGE_PUSH` - if false, does not push the image. Defaults to true.

Refer to the [ko documenation](https://github.com/google/ko#local-publishing-options) for more information.
