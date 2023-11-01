package certmanager

import (
	"context"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/go-logr/logr"
	mf "github.com/manifestival/manifestival"
	"github.com/shipwright-io/operator/pkg/common"
	crdclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	certDomainsTemplate = []string{
		"shp-build-webhook.%s.svc",
	}
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

	manifest, err := common.SetupManifestival(client, "certificates.yaml", false, logger)
	if err != nil {
		return true, fmt.Errorf("Error creating inital certificates manifest")
	}
	manifest, err = manifest.
		Filter(mf.Not(mf.ByKind("Namespace"))).
		Transform(mf.InjectNamespace(namespace), injectDnsNames(buildCertDomains(namespace)))

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

// injectDnsNames adds DnsNames to Certificate.spec manifest
func injectDnsNames(domains []string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		kind := u.GetKind()
		if u.GetKind() != "Certificate" {
			return nil
		}

		for _, domain := range domains {
			if !govalidator.IsDNSName(domain) {
				return fmt.Errorf("'%s' is not a valid dns name", domain)
			}
		}

		err := unstructured.SetNestedStringSlice(u.Object, domains, "spec", "dnsNames")
		if err != nil {
			return fmt.Errorf("error updating dnsNames for %s:%s, %s", kind, u.GetName(), err)
		}
		return nil
	}
}

// buildCertDomains injects namespace and returns a slice of svc dnsdomains
func buildCertDomains(targetNamespace string) []string {
	domains := []string{}
	for _, t := range certDomainsTemplate {
		domains = append(domains, fmt.Sprintf(t, targetNamespace))
	}
	return domains
}
