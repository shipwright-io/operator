# Shipwright Operator

An operator to install and configure [Shipwright](https://shipwright.io) on Kubernetes clusters.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to build, test, and submit
contributions to the operator.

## Usage

To deploy and manage instances of [Shipwright Build-Controller][build-controller], make sure this
operator is up-and-running, and then create the following:

```yml
---
apiVersion: operator.shipwright.io/v1alpha1
kind: ShipwrightBuild
metadata:
  name: shipwright-operator
spec:
  targetNamespace: shipwright-build
  namespace: default
```

It will deploy the Build-Controller in `shipwright-build` namespace. When `.spec.namespace` is not
set, it will use the `shipwright-build` namespace, this namespace needs to be created before the
actual deployment takes place.

It will also ensure Tekton is sufficient installed at the required level if the minimally required version
of the [Tekton Operator](https://github.com/tektoncd/operator) is installed on the cluster.

[build-controller]: https://github.com/shipwright-io/build
