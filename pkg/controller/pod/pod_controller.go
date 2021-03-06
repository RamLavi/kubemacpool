/*
Copyright 2018 The KubeMacPool Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import (
	"context"
	"fmt"
	"github.com/intel/multus-cni/logging"
	"github.com/k8snetworkplumbingwg/kubemacpool/pkg/pool-manager"
	corev1 "k8s.io/api/core/v1"
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

var log = logf.Log.WithName("Pod Controller")

// Add creates a new Policy Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, poolManager *pool_manager.PoolManager) error {
	return add(mgr, newReconciler(mgr, poolManager))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, poolManager *pool_manager.PoolManager) reconcile.Reconciler {
	return &ReconcilePolicy{Client: mgr.GetClient(), scheme: mgr.GetScheme(), poolManager: poolManager}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Pod
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePolicy{}

// ReconcilePolicy reconciles a Policy object
type ReconcilePolicy struct {
	client.Client
	scheme      *runtime.Scheme
	poolManager *pool_manager.PoolManager
}

// Reconcile reads that state of the cluster for a Pod object and makes changes based on the state
func (r *ReconcilePolicy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.V(1).Info("got a pod event in the controller")
	instance := &corev1.Pod{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			err := r.poolManager.ReleasePodMac(fmt.Sprintf("%s/%s", request.Namespace, request.Name))
			if err != nil {
				logging.Printf(logging.ErrorLevel, "failed to release mac for pod %s: %v", request.NamespacedName, err)
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// allocate only when the pod is ready
	if instance.Annotations == nil {
		return reconcile.Result{}, nil
	}

	err = r.poolManager.AllocatePodMac(instance)
	if err != nil {
		log.Error(err, "failed to allocate mac for pod")
	}

	return reconcile.Result{}, nil
}
