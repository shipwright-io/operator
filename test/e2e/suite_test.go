// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operatorv1alpha1 "github.com/shipwright-io/operator/api/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var ctx context.Context
var testEnv *envtest.Environment
var restTimeout = 5 * time.Minute
var restRetry = 1 * time.Second
var log logr.Logger

func TestOperator(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(restTimeout)
	SetDefaultEventuallyPollingInterval(restRetry)
	config.DefaultReporterConfig.SlowSpecThreshold = (1 * time.Minute).Seconds()
	RunSpecsWithDefaultAndCustomReporters(t,
		"Operator e2e",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	log = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(log)

	By("setting up KUBECONFIG")
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))

	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = operatorv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiextv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	ctx = ctrl.SetupSignalHandler()

	By("waiting for operator to be deployed")
	err = waitForOperator(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

}, 60)

var _ = AfterSuite(func() {
	By("dumping operator deployment state")
	deploymentJSON, err := getOperatorDeploymentJSON(ctx, k8sClient)
	if err != nil {
		log.Error(err, "failed to get operator deployment")
		return
	}
	log.Info(fmt.Sprintf("operator deployment state: %s", deploymentJSON))
	podsJSON, err := getOperatorPodsJSON(ctx, k8sClient)
	if err != nil {
		log.Error(err, "failed to get operator pods")
	}
	log.Info(fmt.Sprintf("operator pods state: %s", podsJSON))
})
