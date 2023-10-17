package common

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

// Requeue triggers a object requeue.
func Requeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueOnError triggers requeue when error is not nil.
func RequeueOnError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueWithError triggers a object requeue because the informed error happend.
func RequeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, err
}

// NoRequeue all done, the object does not need reconciliation anymore.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: false}, nil
}

// RequeueAfterWithError allows a Reconciler to override the
// exponential backoff behavior of the Controller, rescheduling the Request at a given time in the future.
func RequeueAfterWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: 5 * time.Second}, err
}
