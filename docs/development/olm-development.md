# OLM Development

The Shipwright operator is meant to be deployed on a cluster using the
[Operator Lifecycle Manager](https://olm.operatorframework.io/) (OLM).
OLM provides mechanisms to support over the air upgrades and automatically deploy related operators
that are packaged in operator [catalogs](https://olm.operatorframework.io/).
Additional steps need to be taken to ensure the operator can be deployed with OLM.

## Prerequisites

* Ensure you have access to a Kubernetes cluster via `kubectl` with cluster admin permissions.
* Install Go version 1.15 or higher.
* Install OLM on your cluster. This can be done using the `make install-olm` command.
* Ability to push to a container registry that is accessible inside your Kubernetes cluster.

## Step 1: Push the operator image to a registry

Run `make ko-publish IMAGE_REPO=<your-registry>`, pushing to a container registry that is accessible inside your Kubernetes cluster.
Using `ko.local` or `kind.local` is not recommended, as this will not push the resulting image to a container registry.

If you are using [KinD](https://kind.sigs.k8s.io/), follow the instructions on how to configure a
[local registry](https://kind.sigs.k8s.io/docs/user/local-registry/).
This will let you use `localhost:<port>` as your container registry.

## Step 2: Build and push the operator bundle

Next, run `make bundle-push IMAGE_REPO=<your-registry>`.
This will push the [operator bundle](https://olm.operatorframework.io/docs/tasks/creating-operator-bundle/)
to the container registry.
An operator bundle is an OCI aritfact that tells OLM how to deploy your operator.
Be sure to use the same container registry for the `IMAGE_REPO` argument.

## Step 3: Build and push an operator catalog

Next, run `make catalog-push IMAGE_REPO=<your-registry>`.
This will build and push an [operator catalog](https://olm.operatorframework.io/docs/tasks/creating-a-catalog/),
which packages your test operator bundle with the other operators available on [operatorhub.io](https://operatorhub.io).
As in step 2, be sure to use the same container registry for the `IMAGE_REPO` argument.

## Step 4: Deploy the operator using the catalog image

Finally, deploy the operator using `make catalog-run IMAGE_REPO=<your-registry>`, using the same
value for `IMAGE_REPO` as in the previous steps.
This will run a script that does the following:

1. Creates a custom [CatalogSource](https://olm.operatorframework.io/docs/tasks/make-catalog-available-on-cluster/)
   and [OperatorGroup](https://olm.operatorframework.io/docs/advanced-tasks/operator-scoping-with-operatorgroups/),
   which allows the operators in step 3's catalog to be installed anywhere on the clsuter.
2. Creates a [Subscription](https://olm.operatorframework.io/docs/tasks/install-operator-with-olm/),
   which instructs OLM to install the operator and any dependent operators.
3. Checks that the operator has successfully been installed and rolled out.

Once the script completes, the Shipwright and Tekton operators will be installed on the cluster.

_Note:_

Scripts in `hack` folder may require `sed` (GNU), therefore in platforms other than Linux you may have it with a different name. For instance, on macOS it's usually named `gsed`, in this case provide the `SED_BIN` make variable with the alternative name.

```bash
$ make catalog-run SED_BIN=gsed ...
```
