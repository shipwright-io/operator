# Local Operator Development

## Prerequisites

* Ensure you have access to a Kubernetes cluster via `kubectl` with cluster admin permissions.
* Install [Go](https://go.dev/doc/install) version 1.21 or higher.

## Building locally

Run `make build` to compile the operator binary.
The resulting binary will be saved to `bin/operator`.

To build the container image for local use, run `make container-push IMAGE_REPO=ko.local`.
This will do the following:

1. Compile the application.
2. Create a container image.
3. Push the container image to your local Docker daemon, with the ref `ko.local/operator:<TAG>`.

The following make options can be set:

* `IMAGE_REPO` - defaults to `ghcr.io/shipwright-io/operator`.
  The following special repositories can be used for testing:

  * `ko.local` - this pushes the image to the local Docker daemon.
  * `kind.local` - pushes to a local KinD cluster.

* `VERSION` - defaults to the current version of Shipwright.
  This must be a valid [semantic version](https://semver.org/), and will be used as the tag for the resulting image.

Refer to the [ko documentation](https://ko.build/) for more information.

## Deploy to Kubernetes

To test the operator on a Kubernetes cluster, you first must have the following:

* Access to a Kubernetes cluster v1.20 or higher, with cluster admin permissions.
* Install [Tekton operator](https://github.com/tektoncd/operator) v0.50 or higher on the cluster.

```bash
$ export KUBECONFIG=/path/to/kubeconfig
$ kubectl apply -f https://github.com/tektoncd/pipeline/releases/download/v0.21.0/release.notags.yaml
```

If pushing to an external image registry, you may need to provide credentials to ko:

```bash
$ make ko
$ ko login <IMAGE_REGISTRY> -u <USERNAME> -p <PASSWORD>
```

Next, use the `make deploy` command with appropriate `IMAGE_REPO` and `VERSION` arguments to deploy the operator to the cluster.

```bash
$ make deploy IMAGE_REPO="<IMAGE_REGISTRY>/<USERNAME>" VERSION="<VERSION>"
```

_Note:_

Scripts in `hack` folder may require `sed` (GNU), therefore in platforms other than Linux you may have it with a different name. For instance, on macOS it's usually named `gsed`, in this case provide the `SED_BIN` make variable with the alternative name.

```bash
$ make build SED_BIN=gsed ...
```
