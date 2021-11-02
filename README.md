# Shipwright Operator

An operator to install and configure [Shipwright](https://shipwright.io) on Kubernetes clusters.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to build, test, and submit
contributions to the operator.

## Usage

To deploy and manage [Shipwright Builds](https://github.com/shipwright-io/build) in your cluster,
first make sure this operator is installed and running on your cluster.

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

_Note: this namespace needs to be created before the actual deployment takes place._
