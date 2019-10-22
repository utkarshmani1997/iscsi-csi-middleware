package controller

import (
	"github.com/utkarshmani1997/iscsi-operator/pkg/controller/iscsiconnection"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, iscsiconnection.Add)
}
