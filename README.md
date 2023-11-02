# Shipwright Operator

An operator to install and configure [Shipwright](https://shipwright.io) on Kubernetes clusters.

## Installation

The Shipwright operator is designed to be installed with the [Operator Lifecycle Manager](https://olm.operatorframework.io/) ("OLM").
Before installation, ensure that OLM has been deployed on your cluster by following the [OLM installation instructions](https://olm.operatorframework.io/docs/getting-started/#installing-olm-in-your-cluster).

Once OLM has been deployed, use the following command to install the latest operator release from [operatorhub.io](https://operatorhub.io/operator/shipwright-operator):

```sh
$ kubectl apply -f https://operatorhub.io/install/shipwright-operator.yaml
```

## OLM Dependencies
When installed via OLM using the provided Shipwright Operator Bundle, the Shipwright operator will ask OLM to deploy the following operators:
- The [Tekton operator](https://tekton.dev/docs/operator/) to deploy and manage Tekton Pipelines.
- The [Cert-Manager operator](https://cert-manager.io/docs/installation/operator-lifecycle-manager/) to provision certificates for admission/conversion webhooks.
For this to work, the Shipwright operator must be included in a catalog that includes these other operators.

## Usage

To deploy and manage [Shipwright Builds](https://github.com/shipwright-io/build) in your cluster,
first make sure this operator is installed and running.

Next, create the following:

```yaml
---
apiVersion: operator.shipwright.io/v1alpha1
kind: ShipwrightBuild
metadata:
  name: shipwright-operator
spec:
  targetNamespace: shipwright-build
```

The operator will deploy Shipwright Builds in the provided `targetNamespace`.
When `.spec.targetNamespace` is not set, the namespace will default to `shipwright-build`.
Refer to the [ShipwrightBuild documentation](docs/shipwrightbuild.md) for more information about this custom resource.

The operator handles differents environment variables to customize Shiprwright controller installation:
- KO_DATA_PATH : defines the shipwright controller manifest to install
- USE_MANAGED_WEBHOOK_CERTS: defines wether the webook ssl certificate is installed by the operator. It requires cert-manager to be installed in the cluster.
- IMAGE_SHIPWRIGHT_SHIPWRIGHT_BUILD : defines the Shipwright Build Controller Image to use
- IMAGE_SHIPWRIGHT_GIT_CONTAINER_IMAGE: defines the Shipwright Git Container Image to use
- IMAGE_SHIPWRIGHT_IMAGE_PROCESSING_CONTAINER_IMAGE:  defines the Shipwright Processing Image to use
- IMAGE_SHIPWRIGHT_BUNDLE_CONTAINER_IMAGE: defines the Shipwright Bundle Image to use
- IMAGE_SHIPWRIGHT_WAITER_CONTAINER_IMAGE: defines the Shipwright Waiter Image to use
- IMAGE_SHIPWRIGHT_SHIPWRIGHT_BUILD_WEBHOOK: defines the Shipwright Build Webhook Image to use

For more information about the function of these images, please consider the Shipwright Build doc https://github.com/shipwright-io/build/blob/main/docs/configuration.md

## Contributing

Please review the overall project
[Contributing Guide](https://github.com/shipwright-io/.github/blob/main/CONTRIBUTING.md) before
submitting bug reports, feature requests, or code.

Want to start hacking on the operator? Refer to the [Development Guilde](DEVELOPMENT.md) to learn
how to build, test, and deploy the operator.
