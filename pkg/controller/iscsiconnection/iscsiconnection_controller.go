package iscsiconnection

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
	ic "github.com/utkarshmani1997/iscsi-operator/pkg/apis/openebs/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_iscsiconnection")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ISCSIConnection Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileISCSIConnection{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("iscsiconnection-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ISCSIConnection
	err = c.Watch(&source.Kind{Type: &ic.ISCSIConnection{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileISCSIConnection implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileISCSIConnection{}

// ReconcileISCSIConnection reconciles a ISCSIConnection object
type ReconcileISCSIConnection struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ISCSIConnection object and makes changes based on the state read
// and what is in the ISCSIConnection.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileISCSIConnection) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ISCSIConnection")

	// Fetch the ISCSIConnection instance
	instance := &ic.ISCSIConnection{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Spec.NodeName == "" {
		return reconcile.Result{}, fmt.Errorf("login failed as node_name is empty")
	}

	if !isEligibleNode(instance.Spec.NodeName, reqLogger) {
		return reconcile.Result{}, nil
	}

	switch instance.Status.Phase {
	case ic.ISCSIConnectionPhasePending:
		return reconcile.Result{}, r.connect(instance, reqLogger)
	case ic.ISCSIConnectionPhaseLoginSuccess:
		return reconcile.Result{}, nil
	case ic.ISCSIConnectionPhaseLogoutStart:
		return reconcile.Result{}, r.disconnect(instance, reqLogger)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileISCSIConnection) disconnect(cr *ic.ISCSIConnection, reqLog logr.Logger) (err error) {
	defer func() {
		if err != nil {
			if err := r.updateISCSIConnectionPhase(cr, ic.ISCSIConnectionPhaseLogoutFailed); err != nil {
				reqLog.Error(err, "failed to update ISCSIConnection phase", "ISCSIConnection CR: ", cr)
			}
		} else {
			if err := r.updateISCSIConnectionPhase(cr, ic.ISCSIConnectionPhaseLogoutSuccess); err != nil {
				reqLog.Error(err, "failed to update ISCSIConnection phase", "ISCSIConnection CR: ", cr)
			}
		}
	}()

	err = iscsi.Disconnect(cr.Spec.TargetIqn, cr.Spec.TargetPortals)
	return err

}

func (r *ReconcileISCSIConnection) updateISCSIConnectionPhase(cr *ic.ISCSIConnection, phase ic.ISCSIConnectionPhase) error {
	cr.Status.Phase = phase
	if err := r.client.Update(context.TODO(), cr); err != nil {
		return fmt.Errorf("failed to update JivaVolume CR: {%v} with targetIP, err: %v", cr, err)
	}
	return nil
}

func isEligibleNode(name string, reqLog logr.Logger) bool {
	nodeName := os.Getenv("NODE_NAME")
	if name != nodeName {
		reqLog.Info("NodeName doesn't match with env provided, ignore login on this node", "ISCSIConnection.NodeName", nodeName, "NODE_NAME", nodeName)
		return false
	}
	return true
}

func (r *ReconcileISCSIConnection) connect(cr *ic.ISCSIConnection, reqLog logr.Logger) (err error) {
	var devicePath string
	defer func() {
		if err != nil {
			if err := r.updateISCSIConnectionPhase(cr, ic.ISCSIConnectionPhaseLoginFailed); err != nil {
				reqLog.Error(err, "failed to update ISCSIConnection phase", "ISCSIConnection CR: ", cr)
			}
		} else {
			if err := r.updateISCSIConnectionPhase(cr, ic.ISCSIConnectionPhaseLoginSuccess); err != nil {
				reqLog.Error(err, "failed to update ISCSIConnection phase", "ISCSIConnection CR: ", cr)
			}
		}
	}()

	devicePath, err = r.discoveryAndLogin(cr)
	if err != nil {
		return err
	}

	if devicePath == "" {
		err = fmt.Errorf("connection successful but returned empty device path")
		return err
	}
	return nil
}

func (r *ReconcileISCSIConnection) discoveryAndLogin(cr *ic.ISCSIConnection) (string, error) {
	conn := &iscsi.Connector{}
	b, err := json.Marshal(cr.Spec)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(b, conn); err != nil {
		return "", err
	}

	devicePath, err := iscsi.Connect(*conn)
	if err != nil {
		return "", err
	}

	return devicePath, nil
}
