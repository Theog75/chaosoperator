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
	"reflect"

	"github.com/go-logr/logr"

	// v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1alpha1 "chaosoperator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	// "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChaoskubeReconciler reconciles a Chaoskube object
type ChaoskubeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.redhat.com,resources=chaoskubes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.redhat.com,resources=chaoskubes/status,verbs=get;update;patch

func (r *ChaoskubeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("chaoskube", req.NamespacedName)

	// Fetch the chaoskube instance
	chaoskube := &cachev1alpha1.Chaoskube{}
	err := r.Get(ctx, req.NamespacedName, chaoskube)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Chaoskube resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Chaoskube")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &v1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: chaoskube.Name, Namespace: chaoskube.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForChaoskube(chaoskube)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// your logic here

	// Ensure the deployment size is the same as the spec
	size := chaoskube.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Chaoskube status with the pod names
	// List the pods for this chaoskube's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(chaoskube.Namespace),
		client.MatchingLabels(labelsForChaoskube(chaoskube.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Chaoskube.Namespace", chaoskube.Namespace, "Chaoskube.Name", chaoskube.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, chaoskube.Status.Nodes) {
		chaoskube.Status.Nodes = podNames
		err := r.Status().Update(ctx, chaoskube)
		if err != nil {
			log.Error(err, "Failed to update Chaoskube status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ChaoskubeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Chaoskube{}).
		Complete(r)
}

// deploymentForchaoskube returns a chaoskube Deployment object
func (r *ChaoskubeReconciler) deploymentForChaoskube(m *cachev1alpha1.Chaoskube) *appsv1.Deployment {
	ls := labelsForChaoskube(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "docker.rct.co.il/chaoskube:v0.21.0",
						Name:  "chaoskube",
						// Command: []string{"chaoskube", "-m=64", "-o", "modern", "-v"},
						// Ports: []corev1.ContainerPort{{
						// 	ContainerPort: 11211,
						// 	Name:          "chaoskube",
						// }},
					}},
				},
			},
		},
	}
	// Set chaoskube instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForChaoskube returns the labels for selecting the resources
// belonging to the given chaoskube CR name.
func labelsForChaoskube(name string) map[string]string {
	return map[string]string{"app": "chaoskube", "chaoskube_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
