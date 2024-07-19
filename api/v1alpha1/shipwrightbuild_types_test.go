package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionReady    = "Ready"
	ConditionNotReady = "NotReady"
)

// TestIsReady tests IsReady condition status function
func TestIsReady(t *testing.T) {
	testCases := map[string]struct {
		status         ShipwrightBuildStatus
		expectedOutput bool
	}{
		"ready": {
			status: ShipwrightBuildStatus{
				Conditions: []metav1.Condition{
					metav1.Condition{
						Type:   ConditionReady,
						Status: metav1.ConditionTrue,
						Reason: "Good",
					},
					metav1.Condition{
						Type:   ConditionNotReady,
						Status: metav1.ConditionFalse,
						Reason: "Good",
					},
				},
			},
			expectedOutput: true,
		},
		"notReady": {
			status: ShipwrightBuildStatus{
				Conditions: []metav1.Condition{
					metav1.Condition{
						Type:   ConditionReady,
						Status: metav1.ConditionFalse,
						Reason: "NotGood",
					},
					metav1.Condition{
						Type:   ConditionNotReady,
						Status: metav1.ConditionFalse,
						Reason: "Good",
					},
				},
			},
			expectedOutput: false,
		},
	}

	for tcName, tc := range testCases {
		if output := tc.status.IsReady(); output != tc.expectedOutput {
			t.Errorf("%s Got %t while expecting %t", tcName, output, tc.expectedOutput)
		}
	}

}
