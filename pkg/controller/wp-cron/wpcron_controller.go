/*
Copyright 2020 Pressinfra SRL.

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

package wpcron

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	wordpressv1alpha1 "github.com/bitpoke/wordpress-operator/pkg/apis/wordpress/v1alpha1"
	"github.com/bitpoke/wordpress-operator/pkg/internal/wordpress"
)

const (
	controllerName      = "wp-cron-controller"
	cronTriggerInterval = 30 * time.Second
	cronTriggerTimeout  = 30 * time.Second
)

var errHTTP = errors.New("HTTP error")

// Add creates a new Wordpress Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r := &ReconcileWordpress{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName(controllerName).WithValues("controller", controllerName),
		recorder: mgr.GetEventRecorderFor(controllerName),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&wordpressv1alpha1.Wordpress{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 32}).
		Complete(r)
}

var _ reconcile.Reconciler = &ReconcileWordpress{}

// ReconcileWordpress reconciles a Wordpress object.
type ReconcileWordpress struct {
	client.Client
	Log      logr.Logger
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Wordpress object and makes changes based on the state read
// and what is in the Wordpress.Spec.
//
// Automatically generate RBAC rules to allow the Controller to read and write Deployments.
// +kubebuilder:rbac:groups=wordpress.presslabs.org,resources=wordpresses;wordpresses/status,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileWordpress) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Wordpress instance
	wp := wordpress.New(&wordpressv1alpha1.Wordpress{})

	err := r.Get(ctx, request.NamespacedName, wp.Unwrap())
	if err != nil {
		return reconcile.Result{}, ignoreNotFound(err)
	}

	wp.SetDefaults()

	log := r.Log.WithValues("key", request.NamespacedName)

	requeue := reconcile.Result{
		Requeue:      true,
		RequeueAfter: cronTriggerInterval,
	}

	svcHostname := fmt.Sprintf("%s.%s.svc", wp.Name, wp.Namespace)
	u := wp.SiteURL("wp-cron.php") + "?doing_wp_cron"

	_u, err := url.Parse(u)
	if err != nil {
		log.Error(err, "error parsing url", "url", u)

		return requeue, nil
	}

	_u.Scheme = "http"
	_u.Host = svcHostname

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), cronTriggerTimeout)
	defer cancel()

	err = r.pingURL(ctxWithTimeout, _u.String(), wp.MainDomain())
	// pingURL failures are recorded as a status condition (WPCronTriggering=False)
	// rather than spamming the operator log every poll interval. Use
	// `kubectl get wordpress` or the WPCronTriggering condition to observe state.

	if updateErr := r.updateWPCronStatus(ctx, wp, err); updateErr != nil && !k8serrors.IsConflict(updateErr) {
		// Optimistic-concurrency conflicts are expected when many controllers update
		// status concurrently and resolve naturally on the next reconcile.
		log.Error(updateErr, "error updating wordpress wp-cron status")
	}

	return requeue, nil
}

func maybeUpdateWPCronCondition(cond wordpressv1alpha1.WordpressCondition, err error) (wordpressv1alpha1.WordpressCondition, bool) {
	needsUpdate := false

	// We've got an error, but WPCronTriggering is true/unknown/empty
	if err != nil && (cond.Status == corev1.ConditionTrue || cond.Status == corev1.ConditionUnknown || cond.Status == "") {
		needsUpdate = true
	}

	// The error message has changed
	if err != nil && cond.Message != err.Error() {
		needsUpdate = true
	}

	// WPCronTriggering is resuming normal operation
	if err == nil && (cond.Status == corev1.ConditionFalse || cond.Status == corev1.ConditionUnknown || cond.Status == "") {
		needsUpdate = true
	}

	if needsUpdate {
		now := metav1.Now()
		cond.LastUpdateTime = now
		cond.LastTransitionTime = now

		if err == nil {
			cond.Status = corev1.ConditionTrue
			cond.Reason = wordpressv1alpha1.WPCronTriggeringReason
			cond.Message = "wp-cron is triggering"
		} else {
			cond.Status = corev1.ConditionFalse
			cond.Reason = wordpressv1alpha1.WPCronTriggerErrorReason
			cond.Message = err.Error()
		}
	}

	return cond, needsUpdate
}

func (r *ReconcileWordpress) updateWPCronStatus(ctx context.Context, wp *wordpress.Wordpress, e error) error {
	var needsUpdate bool

	idx := -1

	for i := range wp.Status.Conditions {
		if wp.Status.Conditions[i].Type == wordpressv1alpha1.WPCronTriggeringCondition {
			idx = i
		}
	}

	if idx == -1 {
		wp.Status.Conditions = append(wp.Status.Conditions, wordpressv1alpha1.WordpressCondition{
			Type: wordpressv1alpha1.WPCronTriggeringCondition,
		})
		idx = len(wp.Status.Conditions) - 1
	}

	wp.Status.Conditions[idx], needsUpdate = maybeUpdateWPCronCondition(wp.Status.Conditions[idx], e)

	if needsUpdate {
		err := r.Client.Status().Update(ctx, wp.Unwrap())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileWordpress) pingURL(ctx context.Context, url, hostOverride string) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Host = hostOverride

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			r.Log.Error(err, "unexpected error while closing HTTP response body")
		}
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%w: %v, %v", errHTTP, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func ignoreNotFound(err error) error {
	if k8serrors.IsNotFound(err) {
		return nil
	}

	return err
}
