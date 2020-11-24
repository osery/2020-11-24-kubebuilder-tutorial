/*


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
	"time"

	"github.com/go-logr/logr"
	"github.com/osery/coffee-maker/pkg/model"
	"github.com/osery/coffee-maker/pkg/rest"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	beveragev1 "coffee.demo.purestorage.com/api/v1"
)

// CoffeeReconciler reconciles a Coffee object
type CoffeeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=beverage.coffee.demo.purestorage.com,resources=coffees,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=beverage.coffee.demo.purestorage.com,resources=coffees/status,verbs=get;update;patch

func (r *CoffeeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("coffee", req.NamespacedName)

	var k8sCoffee beveragev1.Coffee
	if err := r.Get(ctx, req.NamespacedName, &k8sCoffee); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	c := rest.NewClient("http://coffee-maker.default:3334")
	restCoffee, err := c.GetByName(ctx, k8sCoffee.Name)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if restCoffee == nil {
		restCoffee, err = c.Create(ctx, k8sCoffee.Name, model.CoffeeType(k8sCoffee.Spec.Type), k8sCoffee.Spec.ExtraSugar)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	k8sCoffee.Status.Status = string(restCoffee.Status)
	err = r.Status().Update(ctx, &k8sCoffee)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	if restCoffee.Status != model.Done {
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	return ctrl.Result{}, nil
}

func (r *CoffeeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&beveragev1.Coffee{}).
		Complete(r)
}
