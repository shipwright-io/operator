package controllers

// To minimize the risk of privilege escalation or destructive behavior, the controller is only
// allowed to modify named resources that deploy Shipwright Build.  This is especially true for the
// cluster roles and custom resource definitions included in the release manifest.

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apps,resources=deployments,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,resourceNames=shipwright-build-controller,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,resourceNames=shipwright-build-webhook,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;events;configmaps;secrets;limitranges;namespaces;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,resourceNames=builds.shipwright.io;buildruns.shipwright.io;buildstrategies.shipwright.io;clusterbuildstrategies.shipwright.io,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-aggregate-edit,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-aggregate-view,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,resourceNames=shipwright-build-controller,verbs=update;patch;delete
// +kubebuilder:rbac:groups=shipwright.io,resources=clusterbuildstrategies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.tekton.dev,resources=tektonconfigs,verbs=get;list;create
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io/v1beta1,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,resourceNames=shipwright-build-webhook,verbs=update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,resourceNames=selfsigned-issuer,verbs=update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,resourceNames=shipwright-build-webhook-cert,verbs=update;patch;delete
