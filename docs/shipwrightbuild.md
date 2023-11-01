# ShipwrightBuild Custom Resource

When the Shipwright Operator is installed with the Operator Lifecycle Manager, the
`ShipwrightBuild` [custom resource definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) is added to your cluster.
This custom resource is used to install and configure Shipwright Builds on your cluster.
The current operator will install version `0.12.0` of Builds.

When the `ShipwrightBuild` instance is created, the following components are installed:

- The custom resources to run Shipwright builds (`ClusterBuildStrategy`, `BuildStrategy`, `Build`,
  `BuildRun`).
- Shipwright Build's controller, conversion webhook, and associated CA certificates.
- The following example `ClusterBuildStrategies`:
  - `buildah-shipwright-managed-push`
  - `buildah-strategy-managed-push`
  - `buildkit`
  - `buildpacks-v3`
  - `buildpacks-v3-heroku`
  - `kaniko`
  - `kaniko-trivy`
  - `ko`
  - `source-to-image`


## ShipwrightBuild Reference

| Field | Description |
| ----- | ----------- |
| spec.targetNamespace | The target namespace where Shipwright Build will be deployed. If omitted, this will default to `shipwright-build` |
| status.conditions | Conditions which report the status of Shipwright Build. Current reported conditions:<br><br>- `Ready` |
