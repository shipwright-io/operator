# Developer Guide

Want to add a new feature (or fix bugs) in the operator? This document will help you understand how
to build, test, and deploy the operator.

## Prerequisites

Before you begin, ensure you have have a development environment well suited for Go and Kubernetes
development:

- Install [Go](https://go.dev/doc/install) version 1.21 or higher. We recommend using an
  interactive development environment (IDE) that supports go, such as
  [GoLand](https://www.jetbrains.com/go/) or [VSCode](https://code.visualstudio.com/).
- Install a container engine that can build and run container images. We recommend
  [docker](https://docs.docker.com/get-docker/) or [podman](https://podman.io/).
- Obtain access to a Kubernetes cluster that you have administrative permissions on. This can be a
  "micro" distribution such as [kind](https://kind.sigs.k8s.io/), or a fully managed cluster like
  [GKE](https://cloud.google.com/kubernetes-engine). We recommend `kind` for local development.


## Operator Foundations

This project builds a Kubernetes [operator](https://operatorframework.io/what/) for Shipwright
projects, such as [build](https://github.com/shipwright-io/build). It is scaffolded with
[Operator SDK](https://sdk.operatorframework.io/) and its related Kubernetes development framework,
[kubebuilder](https://book.kubebuilder.io/). Those who are new to Kubernetes concepts like
controllers and custom resources are encouraged to try  the
[Operator SDK go tutorial](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)
first.

Operator SDK and kubebuilder are strongly opinionated when it comes to project structure.
Contributors should use `operator-sdk` [commands](https://sdk.operatorframework.io/docs/cli/) or
`kubebuilder` [commands](https://github.com/kubernetes-sigs/kubebuilder) when adding new custom
resources or controllers. Code within the `pkg/` directory is generally safe from these
generators and should be reserved for core reconciliation logic ("business logic").

This project also generates an
[operator bundle](https://sdk.operatorframework.io/docs/olm-integration/tutorial-bundle/), which is
used to deploy the operator on a cluster with
[Operator Lifecycle Manager](https://olm.operatorframework.io/) (OLM). Refer to the
[OLM core tasks](https://olm.operatorframework.io/docs/tasks/) for the steps needed to deploy an
operator with OLM.


## Development Flows

Most features and bug fixes only require the operator and its related custom resource definitions
to be deployed on Kubernetes. The [local development guide](/docs/development/local-development.md)
describes how to build, deploy, and test your changes against a Kubernetes cluster.

Changes that modify the `bundle` directory, however, may require the operator to be packaged and
deployed with Operator Lifecycle Manager. Refer to the
[OLM development guide](/docs/development/olm-development.md) for instructions on how to build,
deploy, and test the operator with OLM.
