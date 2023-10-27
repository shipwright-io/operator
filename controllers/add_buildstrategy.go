// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"github.com/shipwright-io/operator/controllers/buildstrategy"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, buildstrategy.Add)
}
