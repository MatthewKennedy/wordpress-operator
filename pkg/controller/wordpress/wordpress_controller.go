/*
Copyright 2018 Pressinfra SRL.

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

package wordpress

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/presslabs/controller-util/pkg/syncer"

	wordpressv1alpha1 "github.com/bitpoke/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/bitpoke/wordpress-operator/pkg/controller/wordpress/internal/sync"
	"github.com/bitpoke/wordpress-operator/pkg/internal/wordpress"
)

const controllerName = "wordpress-controller"

// Add creates a new Wordpress Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler.
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWordpress{Client: mgr.GetClient(), recorder: mgr.GetEventRecorderFor(controllerName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wordpressv1alpha1.Wordpress{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		WithOptions(controller.Options{}).
		Complete(r)
}

var _ reconcile.Reconciler = &ReconcileWordpress{}

// ReconcileWordpress reconciles a Wordpress object.
type ReconcileWordpress struct {
	client.Client
	recorder record.EventRecorder
}

// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=core,resources=secrets;services;persistentvolumeclaims;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses;wordpresses/status,verbs=get;list;watch;create;update;patch;delete

// Reconcile reads that state of the cluster for a Wordpress object and makes changes based on the state read
// and what is in the Wordpress.Spec.
func (r *ReconcileWordpress) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Wordpress instance
	wp := wordpress.New(&wordpressv1alpha1.Wordpress{})

	err := r.Get(ctx, request.NamespacedName, wp.Unwrap())
	if err != nil {
		return reconcile.Result{}, ignoreNotFound(err)
	}

	wp.SetDefaults()

	secretSyncer := sync.NewSecretSyncer(wp, r.Client)
	deploySyncer := sync.NewDeploymentSyncer(wp, secretSyncer.Object().(*corev1.Secret), r.Client)
	syncers := []syncer.Interface{
		secretSyncer,
		deploySyncer,
		sync.NewServiceSyncer(wp, r.Client),
		// sync.NewDBUpgradeJobSyncer(wp, r.Client),
	}

	if wp.Spec.CodeVolumeSpec != nil && wp.Spec.CodeVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewCodePVCSyncer(wp, r.Client))
	}

	if wp.Spec.MediaVolumeSpec != nil && wp.Spec.MediaVolumeSpec.PersistentVolumeClaim != nil {
		syncers = append(syncers, sync.NewMediaPVCSyncer(wp, r.Client))
	}

	if err = r.sync(ctx, syncers); err != nil {
		return reconcile.Result{}, err
	}

	oldStatus := wp.Status.DeepCopy()
	wp.Status.Replicas = deploySyncer.Object().(*appsv1.Deployment).Status.Replicas

	if oldStatus.Replicas != wp.Status.Replicas {
		if errUp := r.Status().Update(ctx, wp.Unwrap()); errUp != nil {
			return reconcile.Result{}, errUp
		}
	}

	return reconcile.Result{}, nil
}

func ignoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}

	return err
}

func (r *ReconcileWordpress) sync(ctx context.Context, syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(ctx, s, r.recorder); err != nil {
			return err
		}
	}

	return nil
}
