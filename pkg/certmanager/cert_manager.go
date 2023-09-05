package certmanager

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/shipwright-io/operator/pkg/common"
)

func ReconcileCertManager(ctx context.Context, crdClient crdclientv1.ApiextensionsV1Interface, client client.Client, logger logr.Logger, namespace string) (bool, error) {
	certificatesInstalled, err := isCertificatesInstalled(ctx, crdClient)
	if err != nil {
		return true, err
	}
	if !certificatesInstalled {
		certManagerOperatorInstalled, err := isCertManagerOperatorInstalled(ctx, crdClient)
		if err != nil {
			return true, err
		}
		if !certManagerOperatorInstalled {
			return false, fmt.Errorf("cert-manager operator not installed")
		}
	}

	manifest, err := common.SetupManifestival(client, "certificates.yaml", logger)
	if err != nil {
		return true, fmt.Errorf("Error creating inital certificates manifest")
	}
	manifest, err = manifest.
		Filter(manifestival.Not(manifestival.ByKind("Namespace"))).
		Transform(manifestival.InjectNamespace(namespace))
	if err != nil {
		return true, fmt.Errorf("Error transorming manifest using target namespace")
	}

	if err = manifest.Apply(); err != nil {
		return true, err
	}

	return false, nil
}

func isCertificatesInstalled(ctx context.Context, client crdclientv1.ApiextensionsV1Interface) (bool, error) {
	return common.CRDExist(ctx, client, "certificates.cert-manager.io")
}

func isCertManagerOperatorInstalled(ctx context.Context, client crdclientv1.ApiextensionsV1Interface) (bool, error) {
	return common.CRDExist(ctx, client, "certmanagers.operator.openshift.io")
}
