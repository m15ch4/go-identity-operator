/*
Copyright 2024.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	idmv1 "github.com/m15ch4/go-identity-operator/api/v1"
	idmsvc "github.com/m15ch4/go-identity-operator/internal/service"
)

const userFinalizer = "micze.io/user-finalizer"

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=idm.micze.io,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idm.micze.io,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idm.micze.io,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Start reconciliation")

	// Fetch the User instance
	user := &idmv1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("User resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get User")
		return ctrl.Result{}, nil
	}

	// Check if the instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		// If finalizer is present, run finalization logic
		// then remove the finalizer from the list and update the object
		if containsString(user.GetFinalizers(), userFinalizer) {
			err := r.finalizeUser(ctx, user)
			if err != nil {
				return ctrl.Result{}, err
			}

			user.SetFinalizers(removeString(user.GetFinalizers(), userFinalizer))
			err = r.Update(ctx, user)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// If ID field is not set, create a new user
	if user.Status.ID == "" {
		log.Info("Creating user")
		extUser, err := r.createUser(ctx, user)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Update the user status with the ID and State
		user.Status.State = "Created"
		user.Status.ID = extUser.ID
		err = r.Status().Update(ctx, user)
		if err != nil {
			log.Info("Failed to update user status")
			return ctrl.Result{}, err
		}

		log.Info("User created")
		return ctrl.Result{}, nil
	} else {
		//Get the external user
		extUser, err := r.getUser(ctx, user.Status.ID)
		if err != nil {
			return ctrl.Result{}, err
		}

		// compare fields of the external user with the spec fields of user in the cluster (do not compare the status fields)
		if extUser.Name != user.Spec.Name || extUser.Firstname != user.Spec.Firstname || extUser.Lastname != user.Spec.Lastname || extUser.Role != user.Spec.Role || extUser.Age != user.Spec.Age {
			log.Info("Updating user")
			_, err = r.updateUser(ctx, user, extUser)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Add finalizer for this CR
	if !containsString(user.GetFinalizers(), userFinalizer) {
		if err := r.addFinalizer(ctx, user); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Reconciliation finished")

	return ctrl.Result{}, nil
}

// finalizeUser removes object from external system
func (r *UserReconciler) finalizeUser(ctx context.Context, user *idmv1.User) error {
	_ = log.FromContext(ctx)

	cfg := idmsvc.NewIdentityConfig()
	svc := idmsvc.NewIdentityService(&cfg)

	_, err := svc.GetToken()
	if err != nil {
		return err
	}

	err = svc.DeleteUser(user.Status.ID)
	if err != nil {
		return err
	}

	return nil
}

// createUser creates a new user in external system
func (r *UserReconciler) createUser(ctx context.Context, user *idmv1.User) (*idmsvc.IdentityUser, error) {
	_ = log.FromContext(ctx)

	cfg := idmsvc.NewIdentityConfig()
	svc := idmsvc.NewIdentityService(&cfg)

	_, err := svc.GetToken()
	if err != nil {
		return nil, err
	}

	usr, err := svc.CreateUser(&user.Spec)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// getUser gets an existing user from external system
func (r *UserReconciler) getUser(ctx context.Context, id string) (*idmsvc.IdentityUser, error) {
	_ = log.FromContext(ctx)

	cfg := idmsvc.NewIdentityConfig()
	svc := idmsvc.NewIdentityService(&cfg)

	_, err := svc.GetToken()
	if err != nil {
		return nil, err
	}

	usr, err := svc.GetUser(id)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// updateUser updates an existing user in external system
func (r *UserReconciler) updateUser(ctx context.Context, user *idmv1.User, extUser *idmsvc.IdentityUser) (*idmsvc.IdentityUser, error) {
	_ = log.FromContext(ctx)

	cfg := idmsvc.NewIdentityConfig()
	svc := idmsvc.NewIdentityService(&cfg)

	_, err := svc.GetToken()
	if err != nil {
		return nil, err
	}

	usr, err := svc.UpdateUser(extUser.ID, &user.Spec)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (r *UserReconciler) addFinalizer(ctx context.Context, user *idmv1.User) error {
	log := log.FromContext(ctx)
	log.Info("Adding finalizer")
	user.SetFinalizers(append(user.GetFinalizers(), userFinalizer))
	return r.Update(ctx, user)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&idmv1.User{}).
		Complete(r)
}
