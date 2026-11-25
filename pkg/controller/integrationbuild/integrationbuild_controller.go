/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integrationbuild

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/events"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"

	"github.com/apache/camel-k/v2/pkg/client"

	camelevent "github.com/apache/camel-k/v2/pkg/event"
	"github.com/apache/camel-k/v2/pkg/util/log"
	"github.com/apache/camel-k/v2/pkg/util/monitoring"
)

// Add creates a new IntegrationBuild Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(ctx context.Context, mgr manager.Manager, c client.Client) error {
	return add(mgr, newReconciler(mgr, c), c)
}

func newReconciler(mgr manager.Manager, c client.Client) reconcile.Reconciler {
	return monitoring.NewInstrumentedReconciler(
		&ReconcileIntegrationBuild{
			client:   c,
			scheme:   mgr.GetScheme(),
			recorder: mgr.GetEventRecorder("camel-k-integrationbuild-controller"),
			actions: []Action{
				NewInitializeAction(),
				NewMonitorAction(),
			},
		},
		schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    v1alpha1.IntegrationBuildKind,
		},
	)
}

func add(mgr manager.Manager, r reconcile.Reconciler, c client.Client) error {
	b := builder.ControllerManagedBy(mgr).
		Named("integrationbuild-controller").
		// TODO
		// WithOptions(ctrlcontroller.Options{
		// 	MaxConcurrentReconciles: platform.GetIntegrationMaxConcurrentReconciles(),
		// }).
		// Watch for changes to primary resource IntegrationBuild owning Jobs
		For(&v1alpha1.IntegrationBuild{}).
		Owns(&batchv1.Job{})

	return b.Complete(r)
}

var _ reconcile.Reconciler = &ReconcileIntegrationBuild{}

// ReconcileIntegrationBuild reconciles an IntegrationBuild object.
type ReconcileIntegrationBuild struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the API server
	client   client.Client
	scheme   *runtime.Scheme
	recorder events.EventRecorder
	actions  []Action
}

// Reconcile reads that state of the cluster for a IntegrationBuild object and makes changes based
// on the state read and what is in the IntegrationBuild.Spec
func (r *ReconcileIntegrationBuild) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	rlog := Log.WithValues("request-namespace", request.Namespace, "request-name", request.Name)
	rlog.Debug("Reconciling IntegrationBuild")

	var instance v1alpha1.IntegrationBuild

	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var err error

	target := instance.DeepCopy()
	targetLog := rlog.ForIntegrationBuild(target)

	for _, a := range r.actions {
		a.InjectClient(r.client)
		a.InjectLogger(targetLog)

		if !a.CanHandle(target) {
			continue
		}

		targetLog.Debugf("Invoking action %s", a.Name())

		target, err = a.Handle(ctx, target)
		if err != nil {
			camelevent.NotifyError(r.recorder, &instance, target, instance.Name, instance.Kind, err)
			// Update the phase if the new instance is returned
			if target != nil {
				_ = r.update(ctx, &instance, target, &targetLog)
			}

			return reconcile.Result{}, err
		}

		if target != nil {
			if err := r.update(ctx, &instance, target, &targetLog); err != nil {
				camelevent.NotifyError(r.recorder, &instance, target, target.Name, target.Kind, err)

				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIntegrationBuild) update(
	ctx context.Context, base *v1alpha1.IntegrationBuild, target *v1alpha1.IntegrationBuild, log *log.Logger) error {

	if err := r.client.Status().Patch(ctx, target, ctrl.MergeFrom(base)); err != nil {
		camelevent.NotifyError(r.recorder, base, target, target.Name, target.Kind, err)

		return err
	}

	if target.Status.Phase != base.Status.Phase {
		log.Info(
			"State transition",
			"phase-from", base.Status.Phase,
			"phase-to", target.Status.Phase,
		)
	}

	return nil
}
