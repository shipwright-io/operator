# Shipwright Operator

An operator to install and configure [Shipwright](https://shipwright.io) on Kubernetes clusters.

## Installation

The Shipwright operator is designed to be installed with the [Operator Lifecycle Manager](https://olm.operatorframework.io/) ("OLM").
Before installation, ensure that OLM has been deployed on your cluster by following the [OLM installation instructions](https://olm.operatorframework.io/docs/getting-started/#installing-olm-in-your-cluster).

Once OLM has been deployed, use the following command to install the latest operator release from [operatorhub.io](https://operatorhub.io/operator/shipwright-operator):

```sh
$ kubectl apply -f https://operatorhub.io/install/shipwright-operator.yaml
```

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to build, test, and submit
contributions to the operator.
