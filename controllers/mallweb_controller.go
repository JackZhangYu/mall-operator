/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mallwebv1 "mall-operator/api/v1"
)

// MallWebReconciler reconciles a MallWeb object
type MallWebReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mallweb.mall.com,resources=mallwebs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mallweb.mall.com,resources=mallwebs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mallweb.mall.com,resources=mallwebs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MallWeb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *MallWebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &mallwebv1.MallWeb{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info(fmt.Sprintf("instance:%s", instance.String()))

	//get Deployment
	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil {
		if errors.IsNotFound(err) {
			// If deployment Not search that ,then will need create it
			logger.Info("deploy not exists")

			if *instance.Spec.TotalQPS < 1 {
				logger.Info("not need deployment")
				return ctrl.Result{}, nil
			}

			// create service
			if err := CreateServiceIfNotExists(ctx, r, instance, req); err != nil {
				return ctrl.Result{}, err
			}

			//create deployment
			if err := CreateDeployment(ctx, r, instance); err != nil {
				return ctrl.Result{}, err
			}

			//update status
			if err := updateStatus(ctx, r, instance); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get deploy")
		return ctrl.Result{}, err
	}

	// accroding to the single pod to compute the expect pod numbers
	expectReplicas := getExpectReplicas(instance)

	// get the current deployment pod replicas
	realReplicas := deploy.Spec.Replicas

	if expectReplicas == *realReplicas {
		logger.Info("not need to reconcile")
		return ctrl.Result{}, nil
	}

	//restart fine number
	deploy.Spec.Replicas = &expectReplicas
	// update deployment
	if err := r.Update(ctx, deploy); err != nil {
		logger.Error(err, "update deploy replcias error")
		return ctrl.Result{}, err
	}

	//update Status
	if err := updateStatus(ctx, r, instance); err != nil {
		logger.Error(err, "update status error")
		return ctrl.Result{}, err
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MallWebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		For(&mallwebv1.MallWeb{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
