package controllers

// To minimize the risk of privilege escalation or destructive behavior, the controller is only
// allowed to modify named resources that deploy Shipwright Build.  This is especially true for the
// cluster roles and custom resource definitions included in the release manifest.

// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=shipwright.io,resources=*,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=core,resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resourceNames=shipwright-build,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,resourceNames=builds.shipwright.io;buildruns.shipwright.io;buildstrategies.shipwright.io;clusterbuildstrategies.shipwright.io,verbs=update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=tekton.dev,resources=tasks;taskruns,verbs=create;delete;get;list;patch;update;watch
