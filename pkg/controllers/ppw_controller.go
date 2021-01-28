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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	batch1 "k8s.io/api/batch/v1"
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
// +kubebuilder:rbac:groups=core,resources=persistentvolumecla ims,verbs=get;list;watch;create;update;patch;delete;

func (r *PpwReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) { //TODO: Добавить реализацию деплоймента для процессора

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

	//Check data volume and create a new one if need it
	datavolume := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Spec.DataVolume.Name, Namespace: ppw.Namespace}, datavolume)
	if err != nil && errors.IsNotFound(err) {
		pvc := r.PersistentVolume(ppw)
		log.Info("Creating a new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err,"Failed to create PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		return ctrl.Result{}, err
	}

	//Check ml job and create a new one
	mlcontroller := &batch1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Spec.Controller.Name, Namespace: ppw.Namespace}, mlcontroller)
	if err != nil && errors.IsNotFound(err) {
		mljob := r.MlControllerJob(ppw)
		log.Info("Creating ML Job", "Job.Namespace", mljob.Namespace, "Jab.Name", mljob.Name)
		err = r.Create(ctx, mljob)
		if err != nil {
			log.Error(err, "Failed to create ML Job", "Job.Namespace", mljob.Namespace, "Job.Name", mljob.Name)
			return ctrl.Result{}, err
		}
		return  ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ML Job")
		return ctrl.Result{}, err
	}

	// Check if the ppw-server deployment already exists, if not create a new one
	server := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Spec.Server.Name, Namespace: ppw.Namespace}, server)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		serv_dep := r.ServerDeployment(ppw)
		log.Info("Creating a new PPW-Server deployment", "Deployment.Namespace", serv_dep.Namespace, "Deployment.Name", serv_dep.Name)
		err = r.Create(ctx, serv_dep)
		if err != nil {
			log.Error(err, "Failed to create new PPW-Server deployment", "Deployment.Namespace", serv_dep.Namespace, "Deployment.Name", serv_dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PPW-Server deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := ppw.Spec.Server.Size
	if *server.Spec.Replicas != size {
		server.Spec.Replicas = &size
		err = r.Update(ctx, server)
		if err != nil {
			log.Error(err, "Failed to update PPW-Server deployment", "Deployment.Namespace", server.Namespace, "Deployment.Name", server.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	//Check ppw-processor avaliability and create new one if need it
	processor := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Spec.Processor.Name, Namespace: ppw.Namespace}, processor)
	if err != nil && errors.IsNotFound(err) {
		proc_dep := r.ProcessorDeployment(ppw)
		log.Info("Creating a new PPW-Processor deployment", "Deployment.Namespace", proc_dep.Namespace, "Deployment.Name", proc_dep.Name)
		err = r.Create(ctx, proc_dep)
		if err != nil {
			log.Error(err, "Failed to create new PPW-Processor deployment", "Deployment.Namespace", proc_dep.Namespace, "Deployment.Name", proc_dep.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil{
		log.Error(err, "Failed to get PPW-Processor deployment")
		return ctrl.Result{}, err
	}

	size = ppw.Spec.Processor.Size
	if *processor.Spec.Replicas != size {
		processor.Spec.Replicas = &size
		err = r.Update(ctx, processor)
		if err != nil {
			log.Error(err, "Failed to update PPW-Processor deployment", "Deployment.Namespace", processor.Namespace, "Deployment.Name", processor.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	//Check service availability and create onw if need it
	service := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: ppw.Spec.Service.Name, Namespace: ppw.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		srv := r.ServiceDeployment(ppw)
		log.Info("Creating a new PPW-Server service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
		err = r.Create(ctx, srv)
		if err != nil {
			log.Error(err, "Failed to create new PPW-Server service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PPW-Server service")
		return ctrl.Result{}, err
	}

	// Update the Ppw status with the pod names
	// List the pods for this ppw's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(ppw.Namespace),
		client.MatchingLabels(map[string]string{"app": "ppw-server"}),
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
	name := ppw.Spec.Server.Name
	ls := map[string]string{"app": ppw.Spec.Server.Label}
	replicas := ppw.Spec.Server.Size
	image := ppw.Spec.Server.Image
	imsec := ppw.Spec.Server.ImagePullSecret

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
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
						Name:            "ppw-server",
						Image:           image,
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

func (r* PpwReconciler) ProcessorDeployment(ppw *appsv1alpha0.Ppw) *appsv1.Deployment {
	name := ppw.Spec.Processor.Name
	ls := map[string]string{"app": ppw.Spec.Processor.Label}
	replicas := ppw.Spec.Processor.Size
	image := ppw.Spec.Processor.Image
	imsec := ppw.Spec.Processor.ImagePullSecret

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
								ClaimName: "data-claim",
							},
						},
					}},
					Containers: []corev1.Container{{
						Name: "ppw-processor",
						Image: image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts: []corev1.VolumeMount{{
							Name: "data-volume",
							MountPath: "/usr/src/app/data",
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

func (r* PpwReconciler) PersistentVolume(ppw *appsv1alpha0.Ppw) *corev1.PersistentVolumeClaim {
	name := ppw.Spec.DataVolume.Name
	scn := ppw.Spec.DataVolume.StorageClassName
	size := ppw.Spec.DataVolume.Size

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: ppw.Namespace,
		},
		Spec:       corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: &scn,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
	ctrl.SetControllerReference(ppw, pvc, r.Scheme)
	return pvc
}

func (r* PpwReconciler) ServiceDeployment(ppw *appsv1alpha0.Ppw) *corev1.Service {
	name := ppw.Spec.Service.Name
	ls := map[string]string{"app": ppw.Spec.Service.Label}
	tport := intstr.IntOrString{
		Type:   0,
		IntVal: 666,
		StrVal: "666",
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: ppw.Namespace,
			Labels: ls,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "http",
				Protocol: corev1.ProtocolTCP,
				TargetPort: tport,
				Port: 666,
			}},
			Selector: ls,
			Type: corev1.ServiceTypeNodePort,
		},
	}
	ctrl.SetControllerReference(ppw, svc, r.Scheme)
	return svc
}

func (r *PpwReconciler) MlControllerJob(ppw *appsv1alpha0.Ppw) *batch1.Job {
	name := ppw.Spec.Controller.Name
	time := ppw.Spec.Controller.Lifetime
	volname := ppw.Spec.Controller.VolumeClaimName
	image := ppw.Spec.Controller.Image
	imsec := ppw.Spec.Controller.ImagePullSecret
	volpath := ppw.Spec.Controller.VolumeMountPath

	jb := &batch1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: ppw.Namespace,
		},
		Spec: batch1.JobSpec{
			TTLSecondsAfterFinished: &time,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "data-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: volname,
								ReadOnly:  false,
							},
						},
					}},
					Containers: []corev1.Container{{
						Name: "ppw-connector",
						Image: image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts: []corev1.VolumeMount{{
							Name: "data-volume",
							MountPath: volpath,
						}},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
					ImagePullSecrets: []corev1.LocalObjectReference{{
						Name: imsec,
					}},
				},
			},
		},
		Status: batch1.JobStatus{},
	}

	ctrl.SetControllerReference(ppw, jb, r.Scheme)
	return jb
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
