package k8s

import (
	"context"
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"gopkg.in/yaml.v3"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentInfo struct {
	Namespace string
	Name      string
	Ready     string
	UpToDate  string
	Available string
	Age       string
	Raw       *appsv1.Deployment
	Client    Client
}

func NewDeployment(name, namespace string, k Client) *DeploymentInfo {
	return &DeploymentInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func (d *DeploymentInfo) Fetch() error {
	deployment, err := d.Client.Clientset.AppsV1().Deployments(d.Namespace).Get(
		context.Background(),
		d.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}

	if deployment == nil {
		return fmt.Errorf("deployment not found")
	}

	d.Raw = deployment
	return nil
}

func FetchDeploymentList(client Client, namespace string) ([]string, error) {
	ds, err := client.Clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployments: %v", err)
	}

	deploymentNames := make([]string, 0, len(ds.Items))
	for _, deployment := range ds.Items {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	return deploymentNames, nil
}

func GetDeploymentsTableData(client Client, namespace string) ([]DeploymentInfo, error) {
	deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	var deploymentInfos []DeploymentInfo
	for _, deployment := range deployments.Items {
		status := deployment.Status
		spec := deployment.Spec

		var desiredReplicas int32
		if spec.Replicas != nil {
			desiredReplicas = *spec.Replicas
		}

		readyStr := fmt.Sprintf("%d/%d", status.ReadyReplicas, desiredReplicas)

		deploymentInfos = append(deploymentInfos, DeploymentInfo{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
			Ready:     readyStr,
			UpToDate:  fmt.Sprintf("%d", status.UpdatedReplicas),
			Available: fmt.Sprintf("%d", status.AvailableReplicas),
			Age:       format.FormatAge(deployment.CreationTimestamp.Time),
			Raw:       deployment.DeepCopy(),
			Client:    client,
		})
	}

	return deploymentInfos, nil
}

func (d *DeploymentInfo) GetPods() ([]PodInfo, error) {
	selector, err := d.GetLabelSelector()
	if err != nil {
		return nil, err
	}
	pods, err := FetchPods(d.Client, d.Namespace, selector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (d *DeploymentInfo) GetLabelSelector() (string, error) {
	if d.Raw == nil {
		return "", fmt.Errorf("deployment raw data not available")
	}

	if d.Raw.Spec.Selector == nil {
		return "", fmt.Errorf("deployment has no selector")
	}

	requirements, err := metav1.LabelSelectorAsSelector(d.Raw.Spec.Selector)
	if err != nil {
		return "", fmt.Errorf("failed to convert label selector: %v", err)
	}

	return requirements.String(), nil
}

func (d *DeploymentInfo) Describe() (string, error) {
	if d.Raw == nil {
		if err := d.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch deployment: %v", err)
		}
	}

	if d.Raw == nil {
		return "", fmt.Errorf("deployment raw data is nil")
	}

	// Create a map-based description that filters out empty/null values like kubectl does
	desc := make(map[string]interface{})

	// apiVersion and kind
	apiVersion := d.Raw.APIVersion
	if apiVersion == "" {
		apiVersion = "apps/v1"
	}
	desc["apiVersion"] = apiVersion

	kind := d.Raw.Kind
	if kind == "" {
		kind = "Deployment"
	}
	desc["kind"] = kind

	// Metadata - filter out empty fields
	metadata := make(map[string]interface{})
	objMeta := d.Raw.ObjectMeta

	if objMeta.Name != "" {
		metadata["name"] = objMeta.Name
	}
	if objMeta.GenerateName != "" {
		metadata["generateName"] = objMeta.GenerateName
	}
	if objMeta.Namespace != "" {
		metadata["namespace"] = objMeta.Namespace
	}
	if objMeta.SelfLink != "" {
		metadata["selfLink"] = objMeta.SelfLink
	}
	if objMeta.UID != "" {
		metadata["uid"] = objMeta.UID
	}
	if objMeta.ResourceVersion != "" {
		metadata["resourceVersion"] = objMeta.ResourceVersion
	}
	if objMeta.Generation != 0 {
		metadata["generation"] = objMeta.Generation
	}
	if !objMeta.CreationTimestamp.IsZero() {
		metadata["creationTimestamp"] = objMeta.CreationTimestamp
	}
	if objMeta.DeletionTimestamp != nil {
		metadata["deletionTimestamp"] = objMeta.DeletionTimestamp
	}
	if objMeta.DeletionGracePeriodSeconds != nil {
		metadata["deletionGracePeriodSeconds"] = objMeta.DeletionGracePeriodSeconds
	}
	if len(objMeta.Labels) > 0 {
		metadata["labels"] = objMeta.Labels
	}
	if len(objMeta.Annotations) > 0 {
		metadata["annotations"] = objMeta.Annotations
	}
	if len(objMeta.OwnerReferences) > 0 {
		metadata["ownerReferences"] = objMeta.OwnerReferences
	}
	if len(objMeta.Finalizers) > 0 {
		metadata["finalizers"] = objMeta.Finalizers
	}
	// Skip managedFields as kubectl doesn't show them

	desc["metadata"] = metadata

	// Spec
	spec := make(map[string]interface{})
	deploymentSpec := d.Raw.Spec

	if deploymentSpec.Replicas != nil {
		spec["replicas"] = *deploymentSpec.Replicas
	}
	if deploymentSpec.Selector != nil {
		spec["selector"] = deploymentSpec.Selector
	}
	// Always include template if it has a spec, and ensure metadata is properly structured
	if deploymentSpec.Template.Spec.Containers != nil || len(deploymentSpec.Template.Spec.Containers) > 0 {
		template := make(map[string]interface{})
		templateMeta := make(map[string]interface{})

		// Always include labels if selector exists to prevent validation errors
		if deploymentSpec.Selector != nil && deploymentSpec.Selector.MatchLabels != nil {
			templateMeta["labels"] = deploymentSpec.Selector.MatchLabels
		} else if len(deploymentSpec.Template.ObjectMeta.Labels) > 0 {
			templateMeta["labels"] = deploymentSpec.Template.ObjectMeta.Labels
		}

		if deploymentSpec.Template.ObjectMeta.Name != "" {
			templateMeta["name"] = deploymentSpec.Template.ObjectMeta.Name
		}

		if len(templateMeta) > 0 {
			template["metadata"] = templateMeta
		}

		// Filter the pod spec to exclude empty fields
		podSpec := d.filterPodSpec(deploymentSpec.Template.Spec)
		template["spec"] = podSpec
		spec["template"] = template
	}
	if deploymentSpec.Strategy.Type != "" {
		strategy := make(map[string]interface{})
		strategy["type"] = deploymentSpec.Strategy.Type
		if deploymentSpec.Strategy.RollingUpdate != nil {
			rollingUpdate := make(map[string]interface{})
			if deploymentSpec.Strategy.RollingUpdate.MaxUnavailable != nil {
				rollingUpdate["maxUnavailable"] = deploymentSpec.Strategy.RollingUpdate.MaxUnavailable
			}
			if deploymentSpec.Strategy.RollingUpdate.MaxSurge != nil {
				rollingUpdate["maxSurge"] = deploymentSpec.Strategy.RollingUpdate.MaxSurge
			}
			strategy["rollingUpdate"] = rollingUpdate
		}
		spec["strategy"] = strategy
	}
	if deploymentSpec.ProgressDeadlineSeconds != nil {
		spec["progressDeadlineSeconds"] = *deploymentSpec.ProgressDeadlineSeconds
	}
	if deploymentSpec.RevisionHistoryLimit != nil {
		spec["revisionHistoryLimit"] = *deploymentSpec.RevisionHistoryLimit
	}
	if deploymentSpec.Paused {
		spec["paused"] = deploymentSpec.Paused
	}

	desc["spec"] = spec

	// Status
	status := make(map[string]interface{})
	deploymentStatus := d.Raw.Status

	if deploymentStatus.ObservedGeneration != 0 {
		status["observedGeneration"] = deploymentStatus.ObservedGeneration
	}
	if deploymentStatus.Replicas != 0 {
		status["replicas"] = deploymentStatus.Replicas
	}
	if deploymentStatus.UpdatedReplicas != 0 {
		status["updatedReplicas"] = deploymentStatus.UpdatedReplicas
	}
	if deploymentStatus.ReadyReplicas != 0 {
		status["readyReplicas"] = deploymentStatus.ReadyReplicas
	}
	if deploymentStatus.AvailableReplicas != 0 {
		status["availableReplicas"] = deploymentStatus.AvailableReplicas
	}
	if deploymentStatus.UnavailableReplicas != 0 {
		status["unavailableReplicas"] = deploymentStatus.UnavailableReplicas
	}
	if len(deploymentStatus.Conditions) > 0 {
		status["conditions"] = deploymentStatus.Conditions
	}

	desc["status"] = status

	yamlData, err := yaml.Marshal(desc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (d *DeploymentInfo) filterPodSpec(podSpec corev1.PodSpec) map[string]interface{} {
	spec := make(map[string]interface{})

	if len(podSpec.Volumes) > 0 {
		spec["volumes"] = podSpec.Volumes
	}
	if len(podSpec.InitContainers) > 0 {
		spec["initContainers"] = podSpec.InitContainers
	}
	if len(podSpec.Containers) > 0 {
		// Filter containers to exclude empty fields
		containers := make([]map[string]interface{}, 0, len(podSpec.Containers))
		for _, container := range podSpec.Containers {
			containerMap := d.filterContainer(container)
			containers = append(containers, containerMap)
		}
		spec["containers"] = containers
	}
	if podSpec.RestartPolicy != "" && podSpec.RestartPolicy != corev1.RestartPolicyAlways {
		spec["restartPolicy"] = podSpec.RestartPolicy
	}
	if podSpec.TerminationGracePeriodSeconds != nil {
		spec["terminationGracePeriodSeconds"] = *podSpec.TerminationGracePeriodSeconds
	}
	if podSpec.ActiveDeadlineSeconds != nil {
		spec["activeDeadlineSeconds"] = *podSpec.ActiveDeadlineSeconds
	}
	if podSpec.DNSPolicy != "" && podSpec.DNSPolicy != corev1.DNSClusterFirst {
		spec["dnsPolicy"] = podSpec.DNSPolicy
	}
	if podSpec.NodeSelector != nil && len(podSpec.NodeSelector) > 0 {
		spec["nodeSelector"] = podSpec.NodeSelector
	}
	if podSpec.ServiceAccountName != "" {
		spec["serviceAccountName"] = podSpec.ServiceAccountName
	}
	if podSpec.AutomountServiceAccountToken != nil {
		spec["automountServiceAccountToken"] = *podSpec.AutomountServiceAccountToken
	}
	if len(podSpec.ImagePullSecrets) > 0 {
		spec["imagePullSecrets"] = podSpec.ImagePullSecrets
	}
	if podSpec.HostNetwork {
		spec["hostNetwork"] = podSpec.HostNetwork
	}
	if podSpec.HostPID {
		spec["hostPID"] = podSpec.HostPID
	}
	if podSpec.HostIPC {
		spec["hostIPC"] = podSpec.HostIPC
	}
	if podSpec.ShareProcessNamespace != nil {
		spec["shareProcessNamespace"] = *podSpec.ShareProcessNamespace
	}
	if len(podSpec.HostAliases) > 0 {
		spec["hostAliases"] = podSpec.HostAliases
	}
	if podSpec.PriorityClassName != "" {
		spec["priorityClassName"] = podSpec.PriorityClassName
	}
	if podSpec.Priority != nil {
		spec["priority"] = *podSpec.Priority
	}
	if len(podSpec.ReadinessGates) > 0 {
		spec["readinessGates"] = podSpec.ReadinessGates
	}
	if podSpec.RuntimeClassName != nil {
		spec["runtimeClassName"] = *podSpec.RuntimeClassName
	}
	if len(podSpec.Overhead) > 0 {
		spec["overhead"] = podSpec.Overhead
	}
	if podSpec.EnableServiceLinks != nil {
		spec["enableServiceLinks"] = *podSpec.EnableServiceLinks
	}
	if podSpec.PreemptionPolicy != nil {
		spec["preemptionPolicy"] = *podSpec.PreemptionPolicy
	}
	if len(podSpec.TopologySpreadConstraints) > 0 {
		spec["topologySpreadConstraints"] = podSpec.TopologySpreadConstraints
	}
	if podSpec.OS != nil {
		spec["os"] = podSpec.OS
	}
	if podSpec.HostUsers != nil {
		spec["hostUsers"] = *podSpec.HostUsers
	}
	if len(podSpec.SchedulingGates) > 0 {
		spec["schedulingGates"] = podSpec.SchedulingGates
	}
	if podSpec.ResourceClaims != nil && len(podSpec.ResourceClaims) > 0 {
		spec["resourceClaims"] = podSpec.ResourceClaims
	}

	return spec
}

func (d *DeploymentInfo) filterContainer(container corev1.Container) map[string]interface{} {
	containerMap := make(map[string]interface{})

	containerMap["name"] = container.Name
	containerMap["image"] = container.Image

	if container.Command != nil && len(container.Command) > 0 {
		containerMap["command"] = container.Command
	}
	if container.Args != nil && len(container.Args) > 0 {
		containerMap["args"] = container.Args
	}
	if container.WorkingDir != "" {
		containerMap["workingDir"] = container.WorkingDir
	}
	if len(container.Ports) > 0 {
		// Filter ports to exclude empty fields
		ports := make([]map[string]interface{}, 0, len(container.Ports))
		for _, port := range container.Ports {
			portMap := make(map[string]interface{})
			if port.Name != "" {
				portMap["name"] = port.Name
			}
			if port.HostPort != 0 {
				portMap["hostPort"] = port.HostPort
			}
			portMap["containerPort"] = port.ContainerPort
			if port.Protocol != "" && port.Protocol != corev1.ProtocolTCP {
				portMap["protocol"] = port.Protocol
			}
			if port.HostIP != "" {
				portMap["hostIP"] = port.HostIP
			}
			ports = append(ports, portMap)
		}
		containerMap["ports"] = ports
	}
	if len(container.EnvFrom) > 0 {
		containerMap["envFrom"] = container.EnvFrom
	}
	if len(container.Env) > 0 {
		containerMap["env"] = container.Env
	}
	if len(container.VolumeMounts) > 0 {
		containerMap["volumeMounts"] = container.VolumeMounts
	}
	if container.ImagePullPolicy != "" && container.ImagePullPolicy != corev1.PullIfNotPresent {
		containerMap["imagePullPolicy"] = container.ImagePullPolicy
	}
	if container.TerminationMessagePath != "" && container.TerminationMessagePath != "/dev/termination-log" {
		containerMap["terminationMessagePath"] = container.TerminationMessagePath
	}
	if container.TerminationMessagePolicy != "" && container.TerminationMessagePolicy != corev1.TerminationMessageReadFile {
		containerMap["terminationMessagePolicy"] = container.TerminationMessagePolicy
	}
	if container.Stdin {
		containerMap["stdin"] = container.Stdin
	}
	if container.StdinOnce {
		containerMap["stdinOnce"] = container.StdinOnce
	}
	if container.TTY {
		containerMap["tty"] = container.TTY
	}

	// Resources - filter out empty resource requirements
	if !d.isEmptyResourceRequirements(container.Resources) {
		resources := make(map[string]interface{})
		if !d.isEmptyResourceList(container.Resources.Limits) {
			limits := make(map[string]interface{})
			for k, v := range container.Resources.Limits {
				limits[string(k)] = v.String()
			}
			resources["limits"] = limits
		}
		if !d.isEmptyResourceList(container.Resources.Requests) {
			requests := make(map[string]interface{})
			for k, v := range container.Resources.Requests {
				requests[string(k)] = v.String()
			}
			resources["requests"] = requests
		}
		if len(container.Resources.Claims) > 0 {
			resources["claims"] = container.Resources.Claims
		}
		containerMap["resources"] = resources
	}

	if len(container.ResizePolicy) > 0 {
		containerMap["resizePolicy"] = container.ResizePolicy
	}
	if container.StartupProbe != nil {
		containerMap["startupProbe"] = container.StartupProbe
	}
	if container.LivenessProbe != nil {
		containerMap["livenessProbe"] = container.LivenessProbe
	}
	if container.ReadinessProbe != nil {
		containerMap["readinessProbe"] = container.ReadinessProbe
	}
	if container.Lifecycle != nil {
		containerMap["lifecycle"] = container.Lifecycle
	}
	if container.SecurityContext != nil {
		containerMap["securityContext"] = container.SecurityContext
	}

	return containerMap
}

func (d *DeploymentInfo) isEmptyResourceRequirements(rr corev1.ResourceRequirements) bool {
	return d.isEmptyResourceList(rr.Limits) && d.isEmptyResourceList(rr.Requests) && len(rr.Claims) == 0
}

func (d *DeploymentInfo) isEmptyResourceList(rl corev1.ResourceList) bool {
	return rl == nil || len(rl) == 0
}

func (d *DeploymentInfo) Apply(yamlContent string) error {
	var deployment appsv1.Deployment
	if err := yaml.Unmarshal([]byte(yamlContent), &deployment); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	// Ensure the name and namespace match the current deployment
	deployment.Name = d.Name
	deployment.Namespace = d.Namespace

	// Validate that selector matches template labels to prevent Kubernetes validation errors
	if deployment.Spec.Selector != nil && deployment.Spec.Selector.MatchLabels != nil {
		if deployment.Spec.Template.ObjectMeta.Labels == nil {
			deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string)
		}
		// Ensure template labels include selector labels
		for key, value := range deployment.Spec.Selector.MatchLabels {
			if _, exists := deployment.Spec.Template.ObjectMeta.Labels[key]; !exists {
				deployment.Spec.Template.ObjectMeta.Labels[key] = value
			}
		}
	}

	_, err := d.Client.Clientset.AppsV1().Deployments(d.Namespace).Update(
		context.Background(),
		&deployment,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}

func DeleteDeployment(client Client, namespace string, deploymentName string) error {
	err := client.Clientset.AppsV1().Deployments(namespace).Delete(context.Background(), deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s: %v", deploymentName, err)
	}
	return nil
}
