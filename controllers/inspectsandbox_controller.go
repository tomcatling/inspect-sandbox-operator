package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	inspectv1alpha1 "github.com/example/inspect-operator/api/v1alpha1"
)

// CiliumNetworkPolicy is a simplified representation of the Cilium network policy
type CiliumNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              map[string]interface{} `json:"spec,omitempty"`
}

// DeepCopyObject implements runtime.Object interface
func (p *CiliumNetworkPolicy) DeepCopyObject() runtime.Object {
	c := &CiliumNetworkPolicy{
		TypeMeta:   p.TypeMeta,
		ObjectMeta: *p.ObjectMeta.DeepCopy(),
	}

	if p.Spec != nil {
		c.Spec = make(map[string]interface{}, len(p.Spec))
		for k, v := range p.Spec {
			c.Spec[k] = v
		}
	}

	return c
}

// InspectSandboxReconciler reconciles an InspectSandbox object
type InspectSandboxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inspect.example.com,resources=inspectsandboxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inspect.example.com,resources=inspectsandboxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inspect.example.com,resources=inspectsandboxes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services;configmaps;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cilium.io,resources=ciliumnetworkpolicies,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles the reconciliation loop for InspectSandbox resources
func (r *InspectSandboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling InspectSandbox", "request", req.NamespacedName)

	// Fetch the InspectSandbox instance
	var sandbox inspectv1alpha1.InspectSandbox
	if err := r.Get(ctx, req.NamespacedName, &sandbox); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request
			return ctrl.Result{}, nil
		}
		// Error reading the object
		return ctrl.Result{}, err
	}

	// Initialize status if not already
	if sandbox.Status.Services == nil {
		sandbox.Status.Services = make(map[string]inspectv1alpha1.ServiceStatus)
	}

	// Reconcile volumes if defined
	for volName, volSpec := range sandbox.Spec.Volumes {
		if err := r.reconcileVolume(ctx, &sandbox, volName, volSpec); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile services
	for svcName, svcSpec := range sandbox.Spec.Services {
		if err := r.reconcileService(ctx, &sandbox, svcName, svcSpec); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile network policies
	if err := r.reconcileNetworkPolicies(ctx, &sandbox); err != nil {
		logger.Error(err, "Failed to reconcile network policies")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.Status().Update(ctx, &sandbox); err != nil {
		logger.Error(err, "Failed to update InspectSandbox status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// reconcileVolume ensures a PVC exists for the specified volume
func (r *InspectSandboxReconciler) reconcileVolume(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
	volName string,
	volSpec inspectv1alpha1.VolumeSpec,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling volume", "name", volName)

	// Implement PVC creation logic here
	// This is a placeholder for actual implementation

	return nil
}

// reconcileService ensures a StatefulSet and Service exist for the specified service
func (r *InspectSandboxReconciler) reconcileService(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
	svcName string,
	svcSpec inspectv1alpha1.ServiceSpec,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling service", "name", svcName)

	// Create or update the StatefulSet
	sts, err := r.reconcileStatefulSet(ctx, sandbox, svcName, svcSpec)
	if err != nil {
		sandbox.Status.Services[svcName] = inspectv1alpha1.ServiceStatus{
			Ready:   false,
			Message: fmt.Sprintf("Failed to reconcile StatefulSet: %v", err),
		}
		return err
	}

	// Create or update the Service if DNS is enabled
	if svcSpec.DNSRecord || len(svcSpec.AdditionalDNSRecords) > 0 {
		if err := r.reconcileKubeService(ctx, sandbox, svcName, svcSpec); err != nil {
			sandbox.Status.Services[svcName] = inspectv1alpha1.ServiceStatus{
				Ready:   false,
				Message: fmt.Sprintf("Failed to reconcile Service: %v", err),
			}
			return err
		}
	}

	// Update service status
	ready := sts.Status.ReadyReplicas > 0
	sandbox.Status.Services[svcName] = inspectv1alpha1.ServiceStatus{
		Ready:   ready,
		Message: getStatusMessage(sts),
	}

	return nil
}

// reconcileStatefulSet ensures a StatefulSet exists for the service
func (r *InspectSandboxReconciler) reconcileStatefulSet(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
	svcName string,
	svcSpec inspectv1alpha1.ServiceSpec,
) (*appsv1.StatefulSet, error) {
	stsName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-%s", sandbox.Name, svcName),
		Namespace: sandbox.Namespace,
	}

	// Check if StatefulSet already exists
	var sts appsv1.StatefulSet
	err := r.Get(ctx, stsName, &sts)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// Create new StatefulSet if it doesn't exist
	if errors.IsNotFound(err) {
		sts = buildStatefulSet(sandbox, svcName, svcSpec)
		if err := controllerutil.SetControllerReference(sandbox, &sts, r.Scheme); err != nil {
			return nil, err
		}
		if err := r.Create(ctx, &sts); err != nil {
			return nil, err
		}
	} else {
		// Update existing StatefulSet if needed
		newSts := buildStatefulSet(sandbox, svcName, svcSpec)
		sts.Spec = newSts.Spec
		if err := r.Update(ctx, &sts); err != nil {
			return nil, err
		}
	}

	// Fetch updated StatefulSet with status
	if err := r.Get(ctx, stsName, &sts); err != nil {
		return nil, err
	}

	return &sts, nil
}

// reconcileKubeService ensures a Kubernetes Service exists for the service
func (r *InspectSandboxReconciler) reconcileKubeService(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
	svcName string,
	svcSpec inspectv1alpha1.ServiceSpec,
) error {
	serviceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-%s", sandbox.Name, svcName),
		Namespace: sandbox.Namespace,
	}

	// Check if Service already exists
	var service corev1.Service
	err := r.Get(ctx, serviceName, &service)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Create new Service if it doesn't exist
	if errors.IsNotFound(err) {
		service = buildKubeService(sandbox, svcName, svcSpec)
		if err := controllerutil.SetControllerReference(sandbox, &service, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, &service); err != nil {
			return err
		}
	} else {
		// Update existing Service if needed
		newService := buildKubeService(sandbox, svcName, svcSpec)
		service.Spec.Ports = newService.Spec.Ports
		service.Spec.Selector = newService.Spec.Selector
		if err := r.Update(ctx, &service); err != nil {
			return err
		}
	}

	return nil
}

// buildStatefulSet constructs a StatefulSet for the service
func buildStatefulSet(sandbox *inspectv1alpha1.InspectSandbox, svcName string, svcSpec inspectv1alpha1.ServiceSpec) appsv1.StatefulSet {
	name := fmt.Sprintf("%s-%s", sandbox.Name, svcName)
	labels := map[string]string{
		"app.kubernetes.io/name":       "inspectsandbox",
		"app.kubernetes.io/instance":   sandbox.Name,
		"app.kubernetes.io/component":  svcName,
		"app.kubernetes.io/managed-by": "inspect-operator",
		"inspect/service":              svcName,
	}

	// Add network labels if specified
	for _, network := range svcSpec.Networks {
		labels[fmt.Sprintf("inspect.example.com/network-%s", network)] = "true"
	}

	// Create pod template
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			EnableServiceLinks: pointer(false),
			Containers: []corev1.Container{
				{
					Name:       svcName,
					Image:      svcSpec.Image,
					Command:    svcSpec.Command,
					Args:       svcSpec.Args,
					WorkingDir: svcSpec.WorkingDir,
					Env:        append([]corev1.EnvVar{{Name: "AGENT_ENV", Value: sandbox.Name}}, svcSpec.Env...),
					Resources:  svcSpec.Resources,
				},
			},
		},
	}

	// Set runtime class if specified
	if svcSpec.RuntimeClassName != "" && svcSpec.RuntimeClassName != "CLUSTER_DEFAULT" {
		podTemplate.Spec.RuntimeClassName = pointer(svcSpec.RuntimeClassName)
	}

	// Create statefulset
	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sandbox.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: fmt.Sprintf("%s-service", svcName),
			Replicas:    pointer(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/instance":  sandbox.Name,
					"app.kubernetes.io/component": svcName,
					"inspect/service":             svcName,
				},
			},
			Template: podTemplate,
		},
	}
}

// buildKubeService constructs a Kubernetes Service for the service
func buildKubeService(sandbox *inspectv1alpha1.InspectSandbox, svcName string, svcSpec inspectv1alpha1.ServiceSpec) corev1.Service {
	name := fmt.Sprintf("%s-%s", sandbox.Name, svcName)
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sandbox.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "inspectsandbox",
				"app.kubernetes.io/instance":   sandbox.Name,
				"app.kubernetes.io/component":  svcName,
				"app.kubernetes.io/managed-by": "inspect-operator",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None", // Headless service
			Selector: map[string]string{
				"app.kubernetes.io/instance":  sandbox.Name,
				"app.kubernetes.io/component": svcName,
				"inspect/service":             svcName,
			},
		},
	}
}

// getStatusMessage returns a status message for the service
func getStatusMessage(sts *appsv1.StatefulSet) string {
	if sts.Status.ReadyReplicas > 0 {
		return "Service is ready"
	}
	if sts.Status.Replicas == 0 {
		return "Service is starting"
	}
	return "Service is not ready"
}

// pointer returns a pointer to the provided value
func pointer[T any](v T) *T {
	return &v
}

// reconcileNetworkPolicies ensures Cilium Network Policies exist for the InspectSandbox
func (r *InspectSandboxReconciler) reconcileNetworkPolicies(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling network policies", "sandbox", sandbox.Name)

	// Reconcile default egress policy (for allowed domains)
	if err := r.reconcileSandboxEgressPolicy(ctx, sandbox); err != nil {
		return err
	}

	// Reconcile default deny policy (to deny all ingress by default)
	if err := r.reconcileDefaultDenyIngressPolicy(ctx, sandbox); err != nil {
		return err
	}

	// Reconcile network-specific ingress policies for each network
	for networkName := range sandbox.Spec.Networks {
		if err := r.reconcileNetworkIngressPolicy(ctx, sandbox, networkName); err != nil {
			return err
		}
	}

	return nil
}

// reconcileSandboxEgressPolicy ensures an egress policy exists that allows:
// - DNS lookups for allowed domains
// - Communication between pods in the same sandbox
// - Communication to any allowed domains
func (r *InspectSandboxReconciler) reconcileSandboxEgressPolicy(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
) error {
	policyName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-egress", sandbox.Name),
		Namespace: sandbox.Namespace,
	}

	// Define the policy
	policy := buildSandboxEgressPolicy(sandbox)

	// Check if policy exists
	var existingPolicy CiliumNetworkPolicy
	err := r.Get(ctx, policyName, &existingPolicy)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Create policy if it doesn't exist, update otherwise
	if errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(sandbox, &policy, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, &policy)
	}

	// Update the policy spec
	existingPolicy.Spec = policy.Spec
	return r.Update(ctx, &existingPolicy)
}

// reconcileDefaultDenyIngressPolicy ensures a default deny ingress policy exists
func (r *InspectSandboxReconciler) reconcileDefaultDenyIngressPolicy(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
) error {
	policyName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-default-deny-ingress", sandbox.Name),
		Namespace: sandbox.Namespace,
	}

	// Define the policy
	policy := buildDefaultDenyIngressPolicy(sandbox)

	// Check if policy exists
	var existingPolicy CiliumNetworkPolicy
	err := r.Get(ctx, policyName, &existingPolicy)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Create policy if it doesn't exist, update otherwise
	if errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(sandbox, &policy, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, &policy)
	}

	// Update the policy spec
	existingPolicy.Spec = policy.Spec
	return r.Update(ctx, &existingPolicy)
}

// reconcileNetworkIngressPolicy ensures a network-specific ingress policy exists
func (r *InspectSandboxReconciler) reconcileNetworkIngressPolicy(
	ctx context.Context,
	sandbox *inspectv1alpha1.InspectSandbox,
	networkName string,
) error {
	policyName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-network-%s-ingress", sandbox.Name, networkName),
		Namespace: sandbox.Namespace,
	}

	// Define the policy
	policy := buildNetworkIngressPolicy(sandbox, networkName)

	// Check if policy exists
	var existingPolicy CiliumNetworkPolicy
	err := r.Get(ctx, policyName, &existingPolicy)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Create policy if it doesn't exist, update otherwise
	if errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(sandbox, &policy, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, &policy)
	}

	// Update the policy spec
	existingPolicy.Spec = policy.Spec
	return r.Update(ctx, &existingPolicy)
}

// buildSandboxEgressPolicy constructs an egress policy for the sandbox
func buildSandboxEgressPolicy(sandbox *inspectv1alpha1.InspectSandbox) CiliumNetworkPolicy {
	// Build the policy spec
	egressRules := []map[string]interface{}{
		// Allow DNS lookups
		{
			"toEndpoints": []map[string]interface{}{
				{
					"matchLabels": map[string]string{
						"io.kubernetes.pod.namespace": "kube-system",
						"k8s-app":                     "kube-dns",
					},
				},
			},
			"toPorts": []map[string]interface{}{
				{
					"ports": []map[string]interface{}{
						{
							"port":     "53",
							"protocol": "UDP",
						},
						{
							"port":     "53",
							"protocol": "TCP",
						},
					},
					"rules": map[string]interface{}{
						"dns": []map[string]interface{}{
							{
								"matchPattern": "*",
							},
						},
					},
				},
			},
		},
		// Allow communication within the sandbox
		{
			"toEndpoints": []map[string]interface{}{
				{
					"matchLabels": map[string]string{
						"app.kubernetes.io/instance": sandbox.Name,
					},
				},
			},
		},
	}

	// Allow specific domains if specified
	if len(sandbox.Spec.AllowDomains) > 0 {
		// Try a different approach for domain rules
		domainRule := map[string]interface{}{
			"toFQDNs": []map[string]interface{}{},
		}

		for _, domain := range sandbox.Spec.AllowDomains {
			domainRule["toFQDNs"] = append(
				domainRule["toFQDNs"].([]map[string]interface{}),
				map[string]interface{}{
					"matchName": domain,
				},
			)

			// Add wildcard subdomain
			domainRule["toFQDNs"] = append(
				domainRule["toFQDNs"].([]map[string]interface{}),
				map[string]interface{}{
					"matchPattern": fmt.Sprintf("*.%s", domain),
				},
			)
		}

		egressRules = append(egressRules, domainRule)
	}

	return CiliumNetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cilium.io/v2",
			Kind:       "CiliumNetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-egress", sandbox.Name),
			Namespace: sandbox.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "inspectsandbox",
				"app.kubernetes.io/instance":   sandbox.Name,
				"app.kubernetes.io/managed-by": "inspect-operator",
			},
		},
		Spec: map[string]interface{}{
			"endpointSelector": map[string]interface{}{
				"matchLabels": map[string]string{
					"app.kubernetes.io/instance": sandbox.Name,
				},
			},
			"egress": egressRules,
		},
	}
}

// buildDefaultDenyIngressPolicy constructs a default deny ingress policy
func buildDefaultDenyIngressPolicy(sandbox *inspectv1alpha1.InspectSandbox) CiliumNetworkPolicy {
	return CiliumNetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cilium.io/v2",
			Kind:       "CiliumNetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-default-deny-ingress", sandbox.Name),
			Namespace: sandbox.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "inspectsandbox",
				"app.kubernetes.io/instance":   sandbox.Name,
				"app.kubernetes.io/managed-by": "inspect-operator",
			},
		},
		Spec: map[string]interface{}{
			"endpointSelector": map[string]interface{}{
				"matchLabels": map[string]string{
					"app.kubernetes.io/instance": sandbox.Name,
				},
			},
			// Empty ingress rules to deny all incoming traffic
			"ingress": []map[string]interface{}{},
		},
	}
}

// buildNetworkIngressPolicy constructs a network-specific ingress policy
func buildNetworkIngressPolicy(sandbox *inspectv1alpha1.InspectSandbox, networkName string) CiliumNetworkPolicy {
	networkLabel := fmt.Sprintf("inspect.example.com/network-%s", networkName)

	return CiliumNetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cilium.io/v2",
			Kind:       "CiliumNetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-network-%s-ingress", sandbox.Name, networkName),
			Namespace: sandbox.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "inspectsandbox",
				"app.kubernetes.io/instance":   sandbox.Name,
				"app.kubernetes.io/managed-by": "inspect-operator",
				"inspect.example.com/network":  networkName,
			},
		},
		Spec: map[string]interface{}{
			"endpointSelector": map[string]interface{}{
				"matchLabels": map[string]string{
					"app.kubernetes.io/instance": sandbox.Name,
					networkLabel:                 "true",
				},
			},
			"ingress": []map[string]interface{}{
				{
					"fromEndpoints": []map[string]interface{}{
						{
							"matchLabels": map[string]string{
								"app.kubernetes.io/instance": sandbox.Name,
								networkLabel:                 "true",
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *InspectSandboxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register our custom CiliumNetworkPolicy type with the scheme
	scheme := mgr.GetScheme()

	// Define the GroupVersion
	groupVersion := schema.GroupVersion{Group: "cilium.io", Version: "v2"}

	// Create a SchemeBuilder that adds both types
	schemeBuilder := runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(groupVersion,
			&CiliumNetworkPolicy{},
			&CiliumNetworkPolicyList{},
		)
		metav1.AddToGroupVersion(scheme, groupVersion)
		return nil
	})

	// Add the types to the scheme
	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&inspectv1alpha1.InspectSandbox{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// CiliumNetworkPolicyList contains a list of CiliumNetworkPolicy
type CiliumNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CiliumNetworkPolicy `json:"items"`
}

// DeepCopyObject implements runtime.Object interface
func (p *CiliumNetworkPolicyList) DeepCopyObject() runtime.Object {
	c := &CiliumNetworkPolicyList{
		TypeMeta: p.TypeMeta,
		ListMeta: *p.ListMeta.DeepCopy(),
		Items:    make([]CiliumNetworkPolicy, len(p.Items)),
	}

	for i, item := range p.Items {
		c.Items[i] = *item.DeepCopyObject().(*CiliumNetworkPolicy)
	}

	return c
}
