package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	customstyles "github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Pod struct {
	Name      string
	Namespace string
	YAML      string
	Raw       *corev1.Pod
	Client    kubernetes.Interface
	Config    *rest.Config
}

func NewPod(name, namespace string, k Client) *Pod {
	return &Pod{
		Name:      name,
		Namespace: namespace,
		Client:    k.Clientset,
		Config:    k.Config,
	}
}

func (p *Pod) Fetch() error {
	pod, err := p.Client.CoreV1().Pods(p.Namespace).Get(context.Background(), p.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %v", err)
	}
	p.Raw = pod
	return nil
}

func (p *Pod) Describe() (string, error) {
	if p.Raw == nil {
		if err := p.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch pod: %v", err)
		}
	}

	events, err := p.Client.CoreV1().Events(p.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", p.Name, p.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get pod events: %v", err)
	}

	data, err := p.DescribePod(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe pod: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod to YAML: %v", err)
	}

	p.YAML = string(yamlData)
	return p.YAML, nil
}
func (pod Pod) DescribePod(events *corev1.EventList) (any, error) {
	type ContainerDescription struct {
		Name         string            `yaml:"name"`
		ContainerID  string            `yaml:"containerID,omitempty"`
		Image        string            `yaml:"image"`
		ImageID      string            `yaml:"imageID,omitempty"`
		Ports        []string          `yaml:"ports,omitempty"`
		State        any               `yaml:"state"`
		LastState    any               `yaml:"lastState,omitempty"`
		Ready        bool              `yaml:"ready"`
		RestartCount int32             `yaml:"restartCount"`
		Environment  map[string]string `yaml:"environment,omitempty"`
		Mounts       []string          `yaml:"mounts,omitempty"`
	}

	type Condition struct {
		Type   string `yaml:"type"`
		Status string `yaml:"status"`
	}

	type VolumeDescription struct {
		Name string `yaml:"name"`
		Type string `yaml:"type,omitempty"`
	}

	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":           pod.Name,
		"namespace":      pod.Namespace,
		"priority":       *pod.Raw.Spec.Priority,
		"serviceAccount": pod.Raw.Spec.ServiceAccountName,
		"node":           fmt.Sprintf("%s/%s", pod.Raw.Spec.NodeName, pod.Raw.Status.HostIP),
		"startTime":      formatTime(*pod.Raw.Status.StartTime),
		"labels":         pod.Raw.Labels,
		"annotations":    pod.Raw.Annotations,
		"status":         string(pod.Raw.Status.Phase),
		"IPs":            pod.Raw.Status.PodIPs,
	}

	containers := make([]ContainerDescription, 0)
	for _, container := range pod.Raw.Spec.Containers {
		var status corev1.ContainerStatus
		for _, s := range pod.Raw.Status.ContainerStatuses {
			if s.Name == container.Name {
				status = s
				break
			}
		}

		containerDesc := ContainerDescription{
			Name:         container.Name,
			ContainerID:  status.ContainerID,
			Image:        container.Image,
			ImageID:      status.ImageID,
			Ready:        status.Ready,
			RestartCount: status.RestartCount,
		}

		if len(container.Ports) > 0 {
			ports := make([]string, 0)
			for _, port := range container.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol))
			}
			containerDesc.Ports = ports
		}

		if status.State.Running != nil {
			containerDesc.State = map[string]string{
				"state":   "Running",
				"started": formatTime(status.State.Running.StartedAt),
			}
		} else if status.State.Terminated != nil {
			containerDesc.State = map[string]any{
				"state":    "Terminated",
				"reason":   status.State.Terminated.Reason,
				"exitCode": status.State.Terminated.ExitCode,
				"started":  formatTime(status.State.Terminated.StartedAt),
				"finished": formatTime(status.State.Terminated.FinishedAt),
			}
		} else if status.State.Waiting != nil {
			containerDesc.State = map[string]string{
				"state":  "Waiting",
				"reason": status.State.Waiting.Reason,
			}
		}

		if status.LastTerminationState.Terminated != nil {
			containerDesc.LastState = map[string]any{
				"state":    "Terminated",
				"reason":   status.LastTerminationState.Terminated.Reason,
				"exitCode": status.LastTerminationState.Terminated.ExitCode,
				"started":  formatTime(status.LastTerminationState.Terminated.StartedAt),
				"finished": formatTime(status.LastTerminationState.Terminated.FinishedAt),
			}
		}

		if len(container.Env) > 0 {
			env := make(map[string]string)
			for _, e := range container.Env {
				env[e.Name] = e.Value
			}
			containerDesc.Environment = env
		}

		if len(container.VolumeMounts) > 0 {
			mounts := make([]string, 0)
			for _, mount := range container.VolumeMounts {
				mounts = append(mounts, fmt.Sprintf("%s from %s (%v)", mount.MountPath, mount.Name, mount.ReadOnly))
			}
			containerDesc.Mounts = mounts
		}

		containers = append(containers, containerDesc)
	}
	desc["containers"] = containers

	conditions := make([]Condition, 0)
	for _, condition := range pod.Raw.Status.Conditions {
		conditions = append(conditions, Condition{
			Type:   string(condition.Type),
			Status: string(condition.Status),
		})
	}
	desc["conditions"] = conditions

	volumes := make([]VolumeDescription, 0)
	for _, volume := range pod.Raw.Spec.Volumes {
		volDesc := VolumeDescription{Name: volume.Name}
		if volume.Projected != nil {
			volDesc.Type = "Projected"
		}
		volumes = append(volumes, volDesc)
	}
	desc["volumes"] = volumes

	desc["qosClass"] = pod.Raw.Status.QOSClass

	if len(pod.Raw.Spec.NodeSelector) > 0 {
		desc["nodeSelectors"] = pod.Raw.Spec.NodeSelector
	}

	if len(pod.Raw.Spec.Tolerations) > 0 {
		desc["tolerations"] = pod.Raw.Spec.Tolerations
	}

	if len(events.Items) > 0 {
		eventList := make([]Event, 0)
		for _, event := range events.Items {
			age := time.Since(event.LastTimestamp.Time).Round(time.Second)
			eventList = append(eventList, Event{
				Type:    event.Type,
				Reason:  event.Reason,
				Age:     age.String(),
				From:    event.Source.Component,
				Message: event.Message,
			})
		}
		desc["events"] = eventList
	}

	return desc, nil
}

func (p *Pod) GetLogs() (string, error) {
	req := p.Client.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &corev1.PodLogOptions{})
	logs, err := req.DoRaw(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}
	return string(logs), nil
}

func (p *Pod) Exec(command []string) (string, string, error) {
	req := p.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(p.Name).
		Namespace(p.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.Config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), err
}

func (p *Pod) ExecWithTTY(command []string) error {
	req := p.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(p.Name).
		Namespace(p.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	borderedStdout := newBorderedWriter(os.Stdout)

	return exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: borderedStdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
}

func (p *Pod) GetStatus() (corev1.PodStatus, error) {
	if p.Raw == nil {
		if err := p.Fetch(); err != nil {
			return corev1.PodStatus{}, err
		}
	}
	return p.Raw.Status, nil
}

func (p *Pod) Delete() error {
	return p.Client.CoreV1().Pods(p.Namespace).Delete(context.Background(), p.Name, metav1.DeleteOptions{})
}

func (p *Pod) GetContainers() ([]string, error) {
	if p.Raw == nil {
		if err := p.Fetch(); err != nil {
			return nil, err
		}
	}

	containers := make([]string, 0, len(p.Raw.Spec.Containers))
	for _, container := range p.Raw.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return containers, nil
}

func formatTime(t metav1.Time) string {
	return t.Format(time.RFC1123)
}

type borderedWriter struct {
	innerWriter io.Writer
}

func newBorderedWriter(w io.Writer) *borderedWriter {
	return &borderedWriter{innerWriter: w}
}

func (b *borderedWriter) Write(p []byte) (n int, err error) {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor))

	styledText := style.Render(string(p))
	return b.innerWriter.Write([]byte(styledText))
}
