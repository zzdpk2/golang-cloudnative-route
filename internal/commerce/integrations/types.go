// Package k8s is a hand-rolled model of a Kubernetes operator.
//
// **This is a simulation, not the real API.** There is no client-go, no
// controller-runtime, no API server. What it teaches is the *shape* of an
// operator — desired state, reconciliation, finalizers, owner references — so
// that when you open the real thing in L17.4 the concepts are already familiar
// and only the API surface is new.
//
// Be clear about the boundary. Nothing you write here will compile against a
// real cluster, and some of the real API's hardest parts (informers, caches,
// watch semantics, RBAC, admission) are absent entirely.
//
// # The one idea that matters
//
// A controller does not execute commands. It repeatedly compares **desired
// state** against **actual state** and nudges reality one step closer. That is
// why Reconcile takes only a name: it re-reads everything, every time, and must
// reach the same conclusion no matter how often it runs.
//
// The consequence is that reconciliation must be **idempotent**. Running it
// three times must leave the world exactly as running it once did. Most
// operator bugs are a violation of that sentence.
package integrations

import "time"

// ============================================================
// CRD Types
// ============================================================

// NamespacedName identifies an object. Kubernetes objects are unique per
// namespace, not globally, so the pair is the real key.
type NamespacedName struct {
	Namespace string
	Name      string
}

// String renders the canonical form.
//
//	NamespacedName{Namespace: "default", Name: "db"} → "default/db"
//	NamespacedName{Name: "cluster"}                  → "cluster"
//
// The second case is a *cluster-scoped* object — a Node, a
// CustomResourceDefinition — which has no namespace. Emitting "/cluster" for it
// would be wrong, and this is the kind of small formatting rule that shows up
// in log greps forever.
func (n NamespacedName) String() string {
	panic("TODO")
}

// ObjectMeta is the metadata every Kubernetes object carries.
type ObjectMeta struct {
	Name      string
	Namespace string
	// ResourceVersion is the optimistic-concurrency token. An update carrying a
	// stale version is rejected, which is how two controllers writing the same
	// object avoid clobbering each other. The same idea as the version field in
	// L6's optimistic locking.
	ResourceVersion string
	// DeletionTimestamp is set when a delete is requested. The object is *not*
	// removed while finalizers remain — see IsDeleting.
	DeletionTimestamp *time.Time
	// Finalizers block deletion until each one is removed by whoever added it.
	Finalizers []string
	// OwnerReferences make the garbage collector delete this object when its
	// owner goes away.
	OwnerReferences []OwnerReference
	Labels          map[string]string
}

// OwnerReference points at the object that owns this one.
//
// This is how an operator avoids writing cleanup code: mark the StatefulSet as
// owned by the MySQLCluster, and deleting the cluster deletes the StatefulSet
// automatically. Controller=true means this is the *managing* owner, of which
// there may be only one.
type OwnerReference struct {
	APIVersion string
	Kind       string
	Name       string
	UID        string
	Controller bool
}

// MySQLClusterSpec is the **desired** state — what the user asked for.
type MySQLClusterSpec struct {
	Replicas  int
	Version   string
	StorageGB int
}

// MySQLClusterStatus is the **observed** state — what the controller found.
//
// The spec/status split is the core convention: users write spec, controllers
// write status, and neither writes the other's half. A controller that modifies
// spec is fighting the user, and the user usually wins by reapplying their YAML.
type MySQLClusterStatus struct {
	Phase         string
	ReadyReplicas int
	Message       string
}

const (
	PhasePending = "Pending"
	PhaseRunning = "Running"
	PhaseFailed  = "Failed"

	// MySQLFinalizer is namespaced with a domain so two operators cannot
	// collide on the same string. That convention is not optional in practice.
	MySQLFinalizer = "database.example.com/mysql-finalizer"
)

// MySQLCluster is the custom resource.
//
// The +kubebuilder markers are how the real toolchain generates the CRD YAML
// and the deep-copy code from this struct. They do nothing here; they are
// present so you recognise them later.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type MySQLCluster struct {
	APIVersion string
	Kind       string
	Metadata   ObjectMeta
	Spec       MySQLClusterSpec
	Status     MySQLClusterStatus
}

// NewMySQLCluster builds a cluster in its initial state.
//
//	NewMySQLCluster("default", "db", 3)
//	  → Metadata.Name "db", Metadata.Namespace "default"
//	  → Spec.Replicas 3
//	  → Status.Phase "Pending"
//	  → no finalizers
//
// Set APIVersion and Kind too — a real object always carries them, and
// DesiredStatefulSet needs them for the owner reference.
func NewMySQLCluster(namespace, name string, replicas int) *MySQLCluster {
	panic("TODO")
}

// NamespacedName returns this cluster's key.
func (m *MySQLCluster) NamespacedName() NamespacedName {
	panic("TODO")
}

// MarkRunning records that the cluster is healthy with `ready` replicas.
//
//	MarkRunning(3) → Phase "Running", ReadyReplicas 3, Message cleared
//
// Clearing the message matters: a stale "failed to create statefulset" left
// next to Phase Running is how an operator lies to its users.
func (m *MySQLCluster) MarkRunning(ready int) {
	panic("TODO")
}

// MarkFailed records a failure and why.
//
//	MarkFailed("image pull failed") → Phase "Failed", Message set
func (m *MySQLCluster) MarkFailed(message string) {
	panic("TODO")
}

// HasFinalizer reports whether the named finalizer is present.
func (m *MySQLCluster) HasFinalizer(finalizer string) bool {
	panic("TODO")
}

// AddFinalizer adds the finalizer, **idempotently**.
//
//	AddFinalizer(f); AddFinalizer(f) → exactly one entry
//
// The test calls it twice and checks the length. Idempotence is not a nicety
// here: Reconcile runs repeatedly, so a naive append grows the list without
// bound and the object eventually fails to update.
func (m *MySQLCluster) AddFinalizer(finalizer string) {
	panic("TODO")
}

// RemoveFinalizer removes the finalizer if present, and does nothing if not.
//
// Removing the last finalizer is what actually lets Kubernetes delete the
// object. Fail to remove it and the resource is stuck in Terminating forever —
// the single most common way an operator ruins someone's afternoon.
func (m *MySQLCluster) RemoveFinalizer(finalizer string) {
	panic("TODO")
}

// IsDeleting reports whether deletion has been requested.
//
//	DeletionTimestamp == nil → false
//	DeletionTimestamp set    → true, even though the object still exists
//
// That second line is the point. A "deleted" object with finalizers is still
// readable, still reconciled, and still has to be cleaned up by hand — deletion
// in Kubernetes is a *request*, not an event.
func (m *MySQLCluster) IsDeleting() bool {
	panic("TODO")
}

// StatefulSet is the workload the operator manages on the cluster's behalf.
type StatefulSet struct {
	Metadata ObjectMeta
	Replicas int
	Image    string
}

// Service is the stable network endpoint in front of it.
type Service struct {
	Metadata ObjectMeta
	Port     int
	Selector map[string]string
}
