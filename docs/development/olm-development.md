# OLM Development

The Shipwright operator is meant to be deployed on a cluster using the
[Operator Lifecycle Manager](https://olm.operatorframework.io/) (OLM).
OLM provides mechanisms to support over the air upgrades and automatically deploy related operators
that are packaged in operator [catalogs](https://olm.operatorframework.io/).
Additional steps need to be taken to ensure the operator can be deployed with OLM.

## Prerequisites

* Ensure you have access to a Kubernetes cluster via `kubectl` with cluster admin permissions.
* Install [Go](https://go.dev/doc/install) version 1.22 or higher.
* Install OLM on your cluster. This can be done using the `make install-olm` command.
* Ability to push to a container registry that is accessible inside your Kubernetes cluster.

## Step 1: Build and push the operator and bundle

Run `make bundle-push IMAGE_REPO=<your-registry>`, pushing to a container registry that is
accessible inside your Kubernetes cluster.
This make command will push the operator and [operator bundle](https://olm.operatorframework.io/docs/tasks/creating-operator-bundle/)
to the container registry.
An operator bundle is an OCI artifact that tells OLM how to deploy your operator.

Using `ko.local` or `kind.local` for `IMAGE_REPO` is not recommended, as this will not push the
resulting images to an OCI-compliant container registry.
If you are using [KinD](https://kind.sigs.k8s.io/), follow the instructions on how to configure a
[local registry](https://kind.sigs.k8s.io/docs/user/local-registry/).
This will let you use `localhost:<port>` as your container registry.

## Step 2: Build and push an operator catalog

Next, run `make catalog-push IMAGE_REPO=<your-registry>`.
This will build and push an [operator catalog](https://olm.operatorframework.io/docs/tasks/creating-a-catalog/),
which packages your test operator bundle with the other operators available on [operatorhub.io](https://operatorhub.io).
The built catalog uses the new [file-based catalog](https://olm.operatorframework.io/docs/reference/file-based-catalogs/)
architecture, and packages the upstream Tekton operator along with the current released Shipwright
operator.
The catalog is rendered from JSON manifests in the `test/catalog` directory, and full contents of
the built catalog can be inspected in the `_output/catalog` directory.

As in step 1, be sure to use the same container registry for the `IMAGE_REPO` argument.

## Step 3: Deploy the operator using the catalog image

Finally, deploy the operator using `make catalog-run IMAGE_REPO=<your-registry>`, using the same
value for `IMAGE_REPO` as in the previous steps.
This will run a script that does the following:

1. Creates a custom [CatalogSource](https://olm.operatorframework.io/docs/tasks/make-catalog-available-on-cluster/)
   and [OperatorGroup](https://olm.operatorframework.io/docs/advanced-tasks/operator-scoping-with-operatorgroups/),
   which allows the operators in step 3's catalog to be installed anywhere on the cluster.
2. Creates a [Subscription](https://olm.operatorframework.io/docs/tasks/install-operator-with-olm/),
   which instructs OLM to install the operator and any dependent operators.
3. Checks that the operator has successfully been installed and rolled out.

Once the script completes, the Shipwright and Tekton operators will be installed on the cluster.

## Testing a Release

To test a release that has not been published to the Kubernetes Operators
[OperatorHub](https://operatorhub.io/), do the following:

### Step 1: Build and Push an Operator Catalog

Like Step 2 above, you will need to publish a catalog containing the candidate operator release.
Run the following command to set the correct `BUNDLE_IMG` for the catalog:

```sh
$ version=<version to test, no leading v> # example: 0.13.0-rc0
$ make catalog-push IMAGE_REPO=<your-registry> VERSION="$version" BUNDLE_IMG="ghcr.io/shipwright-io/operator/operator-bundle:v$version"
```

### Step 2: Deploy the operator

Similar to Step 3 above, deploy the operator with the catalog image:

```sh
$ version=<version to test, no leading `v`> # example: 0.13.0-rc0
$ make catalog-run IMAGE_REPO=<your-registry> VERSION="$version"
```

## Troubleshooting

### `sed` Command Not Found

Scripts in `hack` folder may require `sed` (GNU) and assume they are running on Linux.
On platforms other than Linux, use the `SED_BIN` make variable to use a different command for `sed`.
For instance, on MacOS sed functions are provided by `gsed`:

```bash
$ make catalog-run SED_BIN=gsed ...
```

### Catalog Source Fails - Cannot Access Registry Over grpc

OLM uses `grpc` by default to pull catalog sources from OCI artifacts.
This protocol requires HTTP/2, which is not supported in some circumstances (example: hosting
the catalog and bundle on a registry deployed on KinD).
To fall back to HTTP-based pull, set the `USE-HTTP` make variable to `true` when building/pushing
the test catalog:

```bash
$ make catalog-push USE-HTTP="true" ...
$ make catalog-run ...
```
