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
  namespace: default
```

It will deploy the Build-Controller in `default` namespace. When `.spec.namespace` is not informed,
it will use the `shipwright-build` namespace instead, this namespace needs to be created before the
actual deployment takes place.

[build-controller]: https://github.com/shipwright-io/build
