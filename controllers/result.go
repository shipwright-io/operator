package controllers

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Requeue triggers a object requeue.
func Requeue() (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
}

// RequeueOnError triggers requeue when error is not nil.
func RequeueOnError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueWithError triggers a object requeue because the informed error happend.
func RequeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: 1 * time.Second}, err
}

// NoRequeue all done, the object does not need reconciliation anymore.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
