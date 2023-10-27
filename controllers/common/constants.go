package common

const (
	// FinalizerAnnotation annotation string appended on finalizer slice.
	FinalizerAnnotation = "finalizer.operator.shipwright.io"
	// defaultTargetNamespace fallback namespace when `.spec.namepace` is not informed.
	DefaultTargetNamespace = "shipwright-build"

	// Ready object is providing service.
	ConditionReady = "Ready"

	// UseManagedWebhookCerts is an env Var that controls wether we install the webhook certs
	UseManagedWebhookCerts = "USE_MANAGED_WEBHOOK_CERTS"

	CertManagerInjectAnnotationKey           = "cert-manager.io/inject-ca-from"
	CertManagerInjectAnnotationValueTemplate = "%s/shipwright-build-webhook-cert"
)
