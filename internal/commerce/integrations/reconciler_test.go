package integrations

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMySQLClusterFinalizers(t *testing.T) {
	c := NewMySQLCluster("default", "db", 3)

	if c.HasFinalizer(MySQLFinalizer) {
		t.Fatal("new cluster should not have finalizer")
	}

	c.AddFinalizer(MySQLFinalizer)
	c.AddFinalizer(MySQLFinalizer)
	if len(c.Metadata.Finalizers) != 1 {
		t.Fatalf("finalizer should be idempotent: %v", c.Metadata.Finalizers)
	}

	c.RemoveFinalizer(MySQLFinalizer)
	if c.HasFinalizer(MySQLFinalizer) {
		t.Fatal("finalizer should be removed")
	}
}

func TestNamespacedNameString(t *testing.T) {
	if (NamespacedName{Namespace: "default", Name: "db"}).String() != "default/db" {
		t.Fatal("namespaced string mismatch")
	}
	if (NamespacedName{Name: "cluster"}).String() != "cluster" {
		t.Fatal("cluster-scoped string mismatch")
	}
}

func TestDesiredObjects(t *testing.T) {
	c := NewMySQLCluster("default", "db", 3)
	c.Spec.Version = "8.0"

	sts := DesiredStatefulSet(c)
	if sts.Metadata.Name != "db" || sts.Metadata.Namespace != "default" {
		t.Fatalf("sts metadata = %+v", sts.Metadata)
	}
	if sts.Replicas != 3 {
		t.Fatalf("replicas = %d", sts.Replicas)
	}
	if sts.Image != "mysql:8.0" {
		t.Fatalf("image = %s", sts.Image)
	}
	if len(sts.Metadata.OwnerReferences) != 1 || !sts.Metadata.OwnerReferences[0].Controller {
		t.Fatal("statefulset should have controller owner reference")
	}

	svc := DesiredService(c)
	if svc.Port != 3306 {
		t.Fatalf("port = %d", svc.Port)
	}
	if svc.Selector["mysqlcluster"] != "db" {
		t.Fatalf("selector = %v", svc.Selector)
	}
}

func TestNeedsUpdate(t *testing.T) {
	current := &StatefulSet{Replicas: 1, Image: "mysql:5.7", Metadata: ObjectMeta{Labels: map[string]string{"a": "b"}}}
	desired := &StatefulSet{Replicas: 3, Image: "mysql:8.0", Metadata: ObjectMeta{Labels: map[string]string{"a": "b"}}}
	if !NeedsStatefulSetUpdate(current, desired) {
		t.Fatal("should need update")
	}

	current = desired
	if NeedsStatefulSetUpdate(current, desired) {
		t.Fatal("same object should not need update")
	}

	svcA := &Service{Port: 80, Selector: map[string]string{"app": "a"}}
	svcB := &Service{Port: 3306, Selector: map[string]string{"app": "a"}}
	if !NeedsServiceUpdate(svcA, svcB) {
		t.Fatal("service port differs")
	}
}

func TestWorkQueueDedup(t *testing.T) {
	q := NewWorkQueue()
	key := NamespacedName{Namespace: "default", Name: "db"}

	q.Add(key)
	q.Add(key)
	if q.Len() != 1 {
		t.Fatalf("queue should deduplicate, len=%d", q.Len())
	}

	got, ok := q.Get(context.Background())
	if !ok || got != key {
		t.Fatalf("got=%v ok=%v", got, ok)
	}
}

func TestInformer(t *testing.T) {
	q := NewWorkQueue()
	informer := NewInformer(q)

	oldObj := NewMySQLCluster("default", "db", 1)
	oldObj.Metadata.ResourceVersion = "1"
	newObj := NewMySQLCluster("default", "db", 1)
	newObj.Metadata.ResourceVersion = "2"

	informer.OnUpdate(oldObj, newObj)
	if q.Len() != 1 {
		t.Fatalf("update should enqueue")
	}
}

func TestRateLimiter(t *testing.T) {
	key := NamespacedName{Name: "db"}
	rl := NewRateLimiter(time.Second, 10*time.Second)

	if d := rl.When(key); d != time.Second {
		t.Fatalf("first delay = %v", d)
	}
	if d := rl.When(key); d != 2*time.Second {
		t.Fatalf("second delay = %v", d)
	}
	if rl.NumRequeues(key) != 2 {
		t.Fatalf("requeues = %d", rl.NumRequeues(key))
	}
	rl.Forget(key)
	if rl.NumRequeues(key) != 0 {
		t.Fatal("Forget should reset")
	}
}

func TestFakeClientCRUD(t *testing.T) {
	ctx := context.Background()
	client := NewFakeClient()
	key := NamespacedName{Namespace: "default", Name: "db"}

	if _, err := client.GetMySQLCluster(ctx, key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	cluster := NewMySQLCluster("default", "db", 1)
	if err := client.UpdateMySQLCluster(ctx, cluster); err != nil {
		t.Fatal(err)
	}

	found, err := client.GetMySQLCluster(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if found.Metadata.Name != "db" {
		t.Fatalf("found = %+v", found)
	}
}

func TestReconcileCreatesChildren(t *testing.T) {
	ctx := context.Background()
	client := NewFakeClient()
	cluster := NewMySQLCluster("default", "db", 2)
	if err := client.UpdateMySQLCluster(ctx, cluster); err != nil {
		t.Fatal(err)
	}

	r := NewMySQLReconciler(client)
	if _, err := r.Reconcile(ctx, cluster.NamespacedName()); err != nil {
		t.Fatal(err)
	}

	if _, err := client.GetStatefulSet(ctx, cluster.NamespacedName()); err != nil {
		t.Fatalf("statefulset should be created: %v", err)
	}
	if _, err := client.GetService(ctx, cluster.NamespacedName()); err != nil {
		t.Fatalf("service should be created: %v", err)
	}

	updated, _ := client.GetMySQLCluster(ctx, cluster.NamespacedName())
	if !updated.HasFinalizer(MySQLFinalizer) {
		t.Fatal("finalizer should be added")
	}
}
