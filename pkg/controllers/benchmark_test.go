package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func BenchmarkPpwCreate(b *testing.B) {
	b.StopTimer()
	kubeCl, schemaCl, err := createFakeKubeClient()
	ctx := context.Background()
	if err != nil {
		b.Error("Bad kubeClient")
	}
	r := &PpwReconciler{
		Client: kubeCl,
		Log:    testLog,
		Scheme: schemaCl,
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		server := r.ProcessorDeployment(testPPW)
		err := kubeCl.Create(ctx, server)
		err = kubeCl.Delete(ctx, server)
		if err != nil {
			b.Error(ctx)
		}
	}
}

func BenchmarkPpwDelete(b *testing.B) {
	b.StopTimer()
	kubeCl, schemaCl, err := createFakeKubeClient()
	ctx := context.Background()
	if err != nil {
		b.Error("Bad kubeClient")
	}
	r := &PpwReconciler{
		Client: kubeCl,
		Log:    testLog,
		Scheme: schemaCl,
	}
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		server := r.ServerDeployment(testPPW)
		err = kubeCl.Create(ctx, server)
		if err != nil {
			b.Error("trouble with creation")
		}
		b.StartTimer()
		err = kubeCl.Delete(ctx, server)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkPpwUpdate(b *testing.B) {
	b.StopTimer()
	kubeCl, schemaCl, err := createFakeKubeClient()
	ctx := context.Background()
	if err != nil {
		b.Error("Bad kubeClient")
	}
	r := &PpwReconciler{
		Client: kubeCl,
		Log:    testLog,
		Scheme: schemaCl,
	}
	server := r.ServerDeployment(testPPW)
	err = kubeCl.Create(ctx, server)
	if err != nil {
		b.Error("trouble with creation")
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		err = kubeCl.Update(ctx, server)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkPpwReconciler_Reconcile(b *testing.B) {
	b.StopTimer()
	ctx := context.Background()
	kubeCl, schema, err := createFakeKubeClient()
	if err != nil {
		b.Error("Bad kube client")
	}
	r := &PpwReconciler{
		Client: kubeCl,
		Log: testLog,
		Scheme: schema,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "testPPW",
			Name:      "testPPW",
		},
	}
	server := r.ServerDeployment(testPPW)
	processor := r.ProcessorDeployment(testPPW)
	pvc := r.PersistentVolume(testPPW)
	err = kubeCl.Create(ctx, server)
	err = kubeCl.Create(ctx, processor)
	err = kubeCl.Create(ctx, pvc)
	if err != nil {
		b.Error(err)
	}
	for n := 0; n < b.N; n++ {
		b.StartTimer()
		res, err := r.Reconcile(req)
		if err != nil {
			b.Fail()
		}
		if res != (reconcile.Result{Requeue: true}) && res != (reconcile.Result{Requeue: false}) {
			b.Fail()
		}
		b.StopTimer()
		switch {
		case n % 3 == 0:
			err = kubeCl.Delete(ctx, server)
			if err != nil {
				b.Error("error delete server")
			}
		case n % 4 == 0:
			err = kubeCl.Delete(ctx, processor)
			if err != nil {
				b.Error("error delete proc")
			}
		case n % 5 == 0:
			err = kubeCl.Delete(ctx, pvc)
			if err != nil {
				b.Error("error delete pvc")
			}
		}
	}
}