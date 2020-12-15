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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	appsv1alpha0 "github.com/safronovD/ppw-operator/pkg/api/v1alpha0"
)

// PpwReconciler reconciles a Ppw object
type PpwReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.ppw.example.com,resources=ppws,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.ppw.example.com,resources=ppws/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;

func (r *PpwReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("ppw", req.NamespacedName)

	// Fetch the Ppw instance
	ppw := &appsv1alpha0.Ppw{}
	err := r.Get(ctx, req.NamespacedName, ppw)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Ppw resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Ppw")
		return ctrl.Result{}, err
	}

	// Check if the server deployment already exists, if not create a new one
	serverDep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Name, Namespace: ppw.Namespace}, serverDep)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.ServerDeployment(ppw)
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

	// Ensure the deployment size is the same as the spec
	size := ppw.Spec.Server.Size
	if *serverDep.Spec.Replicas != size {
		serverDep.Spec.Replicas = &size
		err = r.Update(ctx, serverDep)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", serverDep.Namespace, "Deployment.Name", serverDep.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	//// Check if the server deployment already exists, if not create a new one
	//processorDep := &appsv1.Deployment{}
	//err = r.Get(ctx, types.NamespacedName{Name: ppw.Name, Namespace: ppw.Namespace}, processorDep)
	//if err != nil && errors.IsNotFound(err) {
	//	// Define a new deployment
	//	dep := r.ProcessorDeployment(ppw)
	//	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	//	err = r.Create(ctx, dep)
	//	if err != nil {
	//		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	//		return ctrl.Result{}, err
	//	}
	//	// Deployment created successfully - return and requeue
	//	return ctrl.Result{Requeue: true}, nil
	//} else if err != nil {
	//	log.Error(err, "Failed to get Deployment")
	//	return ctrl.Result{}, err
	//}
	//
	//size = ppw.Spec.Processor.Size
	//if *processorDep.Spec.Replicas != size {
	//	processorDep.Spec.Replicas = &size
	//	err = r.Update(ctx, processorDep)
	//	if err != nil {
	//		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", processorDep.Namespace, "Deployment.Name", processorDep.Name)
	//		return ctrl.Result{}, err
	//	}
	//	// Spec updated - return and requeue
	//	return ctrl.Result{Requeue: true}, nil
	//}

	// Update the Ppw status with the pod names
	// List the pods for this ppw's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(ppw.Namespace),
		client.MatchingLabels(labelsForPpw(ppw.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Ppw.Namespace", ppw.Namespace, "Ppw.Name", ppw.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, ppw.Status.Nodes) {
		ppw.Status.Nodes = podNames
		err := r.Status().Update(ctx, ppw)
		if err != nil {
			log.Error(err, "Failed to update Ppw status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PpwReconciler) ServerDeployment(ppw *appsv1alpha0.Ppw) *appsv1.Deployment {
	ls := labelsForPpw(ppw.Name)
	replicas := ppw.Spec.Server.Size
	image := ppw.Spec.Server.Image
	imsec := ppw.Spec.Server.ImagePullSecret

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ppw.Name,
			Namespace: ppw.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
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
						Image:           image,
						Name:            "ppw-server",
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 666,
						}},
					}},
					ImagePullSecrets: []corev1.LocalObjectReference{{
						Name: imsec,
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(ppw, dep, r.Scheme)
	return dep
}

func (r *PpwReconciler) ProcessorDeployment(ppw *appsv1alpha0.Ppw) *appsv1.Deployment {
	ls := labelsForPpw(ppw.Name)
	replicas := ppw.Spec.Processor.Size
	image := ppw.Spec.Processor.Image
	pvc := ppw.Spec.Processor.PvcName

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ppw.Name,
			Namespace: ppw.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "data-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvc,
								ReadOnly:  false,
							}},
					}},
					Containers: []corev1.Container{{
						Image: image,
						Name:  "ppw-processor",
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "data-volume",
							MountPath: "/usr/src/app/data",
						}},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(ppw, dep, r.Scheme)
	return dep
}

func labelsForPpw(name string) map[string]string {
	return map[string]string{"app": "ppw", "ppw_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *PpwReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha0.Ppw{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
