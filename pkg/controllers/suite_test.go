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
	appsv1alpha0 "github.com/safronovD/ppw-operator/pkg/api/v1alpha0"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batch1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgo "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sCl "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testPPW = &appsv1alpha0.Ppw{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testPPW",
			Namespace: "testPPW",
		},
		Spec:       appsv1alpha0.PpwSpec{
			Server:     appsv1alpha0.Server{
				Name:            "testServer",
				Label:           "testLabel",
				Size:            3,
				Image:           "nginx:latest",
				ImagePullSecret: "testSecret",
			},
			Processor:  appsv1alpha0.Processor{
				Name:            "testProcessor",
				Label:           "testLabel",
				Size:            1,
				VolumeClaimName: "testVolume",
				Image:           "nginx:latest",
				ImagePullSecret: "testSecret",
				VolumeMountPath: "testPath",
			},
			Controller: appsv1alpha0.Controller{
				Name:            "testController",
				Lifetime:        10,
				VolumeClaimName: "testVolume",
				Image:           "nginx:latest",
				ImagePullSecret: "testSecret",
				VolumeMountPath: "testPath",
			},
			DataVolume: appsv1alpha0.DataVolume{
				Name:             "testDataVolume",
				Size:             "1Gi",
				StorageClassName: "nfs",
			},
			Service:    appsv1alpha0.Service{
				Name:  "testService",
				Label: "testLabel",
			},
		},
		Status:     appsv1alpha0.PpwStatus{},
	}

	testLog = ctrl.Log.WithName("TEST")
)

func createFakeKubeClient() (client.Client, *runtime.Scheme, error){
	obj := runtime.Object(testPPW.DeepCopy())
	ctx := context.Background()
	schema := runtime.NewScheme()
	utilruntime.Must(clientgo.AddToScheme(schema))
	utilruntime.Must(appsv1alpha0.AddToScheme(schema))
	kubeClient := k8sCl.NewFakeClientWithScheme(schema)
	if err := kubeClient.Create(ctx, obj); err != nil {
		return nil, &runtime.Scheme{}, err
	}
	return kubeClient, schema, nil
}

func TestPpwReconciler(t *testing.T) {
	kubeCl, schemaCl, err := createFakeKubeClient()
	if err != nil {
		t.Error("Bad kubeClient")
	}
	r := &PpwReconciler{
		Client: kubeCl,
		Log:    testLog,
		Scheme: schemaCl,
	}
	if err != nil {
		t.Error("Bad kubeClient")
	}
	t.Run("Reconcile test for k8s. Success", func(t *testing.T) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "testPPW",
				Name:      "testPPW",
			},//
		}
		res, err := r.Reconcile(req)

		assert.Nil(t, err)
		assert.Equal(t, res, reconcile.Result{Requeue: true})
	})
	t.Run("Data Volume Claim Test. Success", func(t *testing.T) {
		ctx := context.Background()
		pvc := r.PersistentVolume(testPPW)
		assert.Nil(t, err)

		pvcObj := new(corev1.PersistentVolumeClaim)
		err = r.Get(ctx, types.NamespacedName{Name: testPPW.Spec.DataVolume.Name, Namespace: testPPW.Namespace}, pvcObj)
		assert.Nil(t, err)
		assert.Equal(t, pvc.Name, pvcObj.Name)
	})

	t.Run("Server Deployment Test. Success", func(t *testing.T) {
		ctx := context.Background()
		serverDep := r.ServerDeployment(testPPW)
		err = kubeCl.Create(ctx, serverDep)
		assert.Nil(t, err)

		serverObj := new(appsv1.Deployment)
		err = r.Get(ctx, types.NamespacedName{Name: testPPW.Spec.Server.Name, Namespace: testPPW.Namespace}, serverObj)
		assert.Nil(t, err)
		assert.Equal(t, serverDep.Name, serverObj.Name)
	})
	t.Run("Service Deployment Test. Success", func(t *testing.T) {
		ctx := context.Background()
		service := r.ServiceDeployment(testPPW)
		err = kubeCl.Create(ctx, service)
		assert.Nil(t, err)

		serviceObj := new(corev1.Service)
		err = r.Get(ctx, types.NamespacedName{Name: testPPW.Spec.Service.Name, Namespace: testPPW.Namespace}, serviceObj)
		assert.Nil(t, err)
		assert.Equal(t, service.Name, serviceObj.Name)
	})
	t.Run("ML Job Test. Success", func(t *testing.T) {
		ctx := context.Background()
		mlctrl := r.MlControllerJob(testPPW)
		err = kubeCl.Create(ctx, mlctrl)
		assert.Nil(t, err)

		mlctrlObj := new(batch1.Job)
		err = r.Get(ctx, types.NamespacedName{Name: testPPW.Spec.Controller.Name, Namespace: testPPW.Namespace}, mlctrlObj)
		assert.Nil(t, err)
		assert.Equal(t, mlctrl.Name, mlctrlObj.Name)
	})
}

//var cfg *rest.Config
//var k8sClient client.Client
//var testEnv *envtest.Environment
//
//func TestAPIs(t *testing.T) {
//	RegisterFailHandler(Fail)
//
//	RunSpecsWithDefaultAndCustomReporters(t,
//		"Controller Suite",
//		[]Reporter{printer.NewlineReporter{}})
//}
//
//var _ = BeforeSuite(func(done Done) {
//	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
//
//	By("bootstrapping test environment")
//	testEnv = &envtest.Environment{
//		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
//	}
//
//	var err error
//	cfg, err = testEnv.Start()
//	Expect(err).ToNot(HaveOccurred())
//	Expect(cfg).ToNot(BeNil())
//
//	err = appsv1alpha0.AddToScheme(scheme.Scheme)
//	Expect(err).NotTo(HaveOccurred())
//
//	// +kubebuilder:scaffold:scheme
//
//	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
//	Expect(err).ToNot(HaveOccurred())
//	Expect(k8sClient).ToNot(BeNil())
//
//	close(done)
//}, 60)
//
//var _ = AfterSuite(func() {
//	By("tearing down the test environment")
//	err := testEnv.Stop()
//	Expect(err).ToNot(HaveOccurred())
//})
