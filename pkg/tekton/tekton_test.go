package tekton

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/shipwright-io/operator/pkg/common"

	o "github.com/onsi/gomega"

	tektonoperatorv1alpha1 "github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	tektonoperatorfake "github.com/tektoncd/operator/pkg/client/clientset/versioned/fake"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	ktesting "k8s.io/client-go/testing"
)

func TestReconcileTekton(t *testing.T) {
	cases := []struct {
		name                           string
		taskRunCRD                     *apiextensionsv1.CustomResourceDefinition
		tektonConfigCRD                *apiextensionsv1.CustomResourceDefinition
		tektonConfigObj                *tektonoperatorv1alpha1.TektonConfig
		createTektonConfigErr          error
		expectError                    bool
		expectRequeue                  bool
		expectTektonConfigCreateAction bool
	}{
		{
			name:        "No Tekton Objects",
			expectError: true,
		},
		{
			name: "Tekton TaskRuns defined",
			taskRunCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "taskruns.tekton.dev",
				},
			},
		},
		{
			name: "TektonConfig defined no version",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
				},
			},
			expectRequeue: true,
			expectError:   true,
		},
		{
			name: "TektonConfig defined insufficient version",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": "v0.36.0",
					},
				},
			},
			expectRequeue: true,
			expectError:   true,
		},
		{
			name: "TektonConfig defined sufficient version",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": common.TektonOpMinSupportedVersion,
					},
				},
			},
			expectTektonConfigCreateAction: true,
		},
		{
			name: "Create TektonConfig error",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": common.TektonOpMinSupportedVersion,
					},
				},
			},
			createTektonConfigErr:          fmt.Errorf("internal error CREATE"),
			expectError:                    true,
			expectRequeue:                  true,
			expectTektonConfigCreateAction: true,
		},
		{
			name: "TektonConfig already created",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": common.TektonOpMinSupportedVersion,
					},
				},
			},
			tektonConfigObj: &tektonoperatorv1alpha1.TektonConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "config",
				},
			},
			expectRequeue: false,
			expectError:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)
			ctx := context.TODO()
			crds := []runtime.Object{}
			if tc.taskRunCRD != nil {
				crds = append(crds, tc.taskRunCRD)
			}
			if tc.tektonConfigCRD != nil {
				crds = append(crds, tc.tektonConfigCRD)
			}
			crdClient := apiextensionsfake.NewSimpleClientset(crds...)
			tektonConfigs := []runtime.Object{}
			if tc.tektonConfigObj != nil {
				tektonConfigs = append(tektonConfigs, tc.tektonConfigObj)
			}
			tektonOperatorClient := tektonoperatorfake.NewSimpleClientset(tektonConfigs...)
			if tc.createTektonConfigErr != nil {
				tektonOperatorClient.PrependReactor("create", "tektonconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tc.createTektonConfigErr
				})
			}
			tektonConfig, requeue, err := ReconcileTekton(ctx, crdClient.ApiextensionsV1(), tektonOperatorClient.OperatorV1alpha1())
			if tc.expectError {
				g.Expect(err).To(o.HaveOccurred())
			} else {
				g.Expect(err).NotTo(o.HaveOccurred())
			}
			g.Expect(requeue).To(o.Equal(tc.expectRequeue))
			createOccurred := false
			for _, action := range tektonOperatorClient.Actions() {
				if action.Matches("create", "tektonconfigs") {
					createOccurred = true
				}
			}
			g.Expect(createOccurred).To(o.Equal(tc.expectTektonConfigCreateAction))
			if tc.expectTektonConfigCreateAction && tc.createTektonConfigErr == nil {
				g.Expect(tektonConfig).NotTo(o.BeNil())
				g.Expect(tektonConfig.Name).To(o.Equal("config"))
				g.Expect(tektonConfig.Spec.Profile).To(o.Equal("lite"))
				g.Expect(tektonConfig.Spec.TargetNamespace).To(o.Equal("tekton-pipelines"))
				g.Expect(tektonConfig.Spec.Pruner.Disabled).To(o.Equal(false))
				g.Expect(tektonConfig.Spec.Pruner.Keep).NotTo(o.BeNil())
			}
		})
	}
}

func TestIsTektonPipelinesInstalled(t *testing.T) {
	cases := []struct {
		name            string
		expectInstalled bool
		expectError     bool
	}{
		{
			name: "Not present",
		},
		{
			name:            "Tekton present",
			expectInstalled: true,
		},
		{
			name:        "Error getting TaskRuns",
			expectError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			objects := []runtime.Object{}
			if tc.expectInstalled {
				objects = append(objects, &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "taskruns.tekton.dev",
					},
				})
			}
			client := apiextensionsfake.NewSimpleClientset(objects...)
			if tc.expectError {
				client.PrependReactor("get", "customresourcedefinitions", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("test error!")
				})
			}
			installed, err := IsTektonPipelinesInstalled(ctx, client.ApiextensionsV1())
			if tc.expectError {
				if err == nil {
					t.Error("expected error to be raised, but none was received")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if installed != tc.expectInstalled {
				t.Errorf("expected installed to be %t, got %t", tc.expectInstalled, installed)
			}

		})
	}
}

func TestIsTektonOperatorInstalled(t *testing.T) {
	cases := []struct {
		name            string
		expectInstalled bool
		expectError     bool
	}{
		{
			name: "Not present",
		},
		{
			name:            "Tekton Operator present",
			expectInstalled: true,
		},
		{
			name:        "Error getting TetkonConfigs",
			expectError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			objects := []runtime.Object{}
			if tc.expectInstalled {
				objects = append(objects, &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tektonconfigs.operator.tekton.dev",
					},
				})
			}
			client := apiextensionsfake.NewSimpleClientset(objects...)
			if tc.expectError {
				client.PrependReactor("get", "customresourcedefinitions", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("test error!")
				})
			}
			installed, err := IsTektonOperatorInstalled(ctx, client.ApiextensionsV1())
			if tc.expectError {
				if err == nil {
					t.Error("expected error to be raised, but none was received")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if installed != tc.expectInstalled {
				t.Errorf("expected installed to be %t, got %t", tc.expectInstalled, installed)
			}

		})
	}
}

func TestGetTektonOperatorVersion(t *testing.T) {
	cases := []struct {
		name            string
		tektonConfigCRD *apiextensionsv1.CustomResourceDefinition
		expectedVersion *version.Version
		thrownError     error
		expectError     bool
	}{
		{
			name:        "No CRD",
			expectError: true,
		},
		{
			name:        "Get CRD error",
			thrownError: fmt.Errorf("failed to GET"),
			expectError: true,
		},
		{
			name: "No labels on TektonConfig CRD",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
				},
			},
			expectError: true,
		},
		{
			name: "No release label on TektonConfig CRD",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"some-label": "value",
					},
				},
			},
			expectError: true,
		},
		{
			name: "release label on TektonConfig CRD is not a semver",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": "value",
					},
				},
			},
			expectError: true,
		},
		{
			name: "Valid TektonConfig CRD",
			tektonConfigCRD: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tektonconfigs.operator.tekton.dev",
					Labels: map[string]string{
						"operator.tekton.dev/release": common.TektonOpMinSupportedVersion,
					},
				},
			},
			expectedVersion: version.MustParseSemantic(common.TektonOpMinSupportedVersion),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			objects := []runtime.Object{}
			if tc.tektonConfigCRD != nil {
				objects = append(objects, tc.tektonConfigCRD)
			}
			client := apiextensionsfake.NewSimpleClientset(objects...)
			if tc.thrownError != nil {
				client.PrependReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, tc.thrownError
				})
			}
			version, err := GetTektonOperatorVersion(ctx, client.ApiextensionsV1())
			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(version, tc.expectedVersion) {
				t.Errorf("expected version %s, got %s", tc.expectedVersion, version)
			}
		})
	}
}

func TestIsTektonConfigPresent(t *testing.T) {
	cases := []struct {
		name          string
		tektonConfig  *tektonoperatorv1alpha1.TektonConfig
		thrownError   error
		expectPresent bool
		expectError   bool
	}{
		{
			name: "No TektonConfig",
		},
		{
			name:        "List TektonConfig error",
			thrownError: fmt.Errorf("failed to LIST"),
			expectError: true,
		},
		{
			name: "TektonConfig present",
			tektonConfig: &tektonoperatorv1alpha1.TektonConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "config",
				},
			},
			expectPresent: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			objects := []runtime.Object{}
			if tc.tektonConfig != nil {
				objects = append(objects, tc.tektonConfig)
			}
			client := tektonoperatorfake.NewSimpleClientset(objects...)
			if tc.thrownError != nil {
				client.PrependReactor("*", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, tc.thrownError
				})
			}
			isPresent, err := IsTektonConfigPresent(ctx, client.OperatorV1alpha1())
			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if isPresent != tc.expectPresent {
				t.Errorf("expected TektonConfig presence to be %t, got %t", tc.expectPresent, isPresent)
			}
		})
	}
}

func TestCreateTektonConfig(t *testing.T) {
	ctx := context.TODO()
	client := tektonoperatorfake.NewSimpleClientset()
	expectedProfile := "test"
	expectedNamespace := "test-namespace"
	tektonConfig, err := CreateTektonConfigWithProfileAndTargetNamespace(ctx,
		client.OperatorV1alpha1(),
		expectedProfile,
		expectedNamespace)
	if len(client.Actions()) != 1 {
		t.Errorf("expected 1 client action, got %d", len(client.Actions()))
	}
	for _, action := range client.Actions() {
		if !action.Matches("create", "tektonconfigs") {
			t.Errorf("expected action %s/%s, got %s/%s", "create", "tektonconfigs", action.GetVerb(), action.GetResource().Resource)
		}
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tektonConfig == nil {
		t.Fatal("received nil TektonConfig")
	}
	if tektonConfig.Name != "config" {
		t.Errorf("expected TektonConfig name %s, got %s", "config", tektonConfig.Name)
	}
	if tektonConfig.Spec.Profile != expectedProfile {
		t.Errorf("expected profile %s, got %s", expectedProfile, tektonConfig.Spec.Profile)
	}
	if tektonConfig.Spec.TargetNamespace != expectedNamespace {
		t.Errorf("expected target namespace %s, got %s", expectedNamespace, tektonConfig.Spec.TargetNamespace)
	}
}
