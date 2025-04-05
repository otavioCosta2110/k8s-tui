package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"time"

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
	Pod       *corev1.Pod
	Client    *kubernetes.Clientset
	Config    *rest.Config
}

func NewPod(name, namespace string, client *kubernetes.Clientset, config *rest.Config) *Pod {
	return &Pod{
		Name:      name,
		Namespace: namespace,
		Client:    client,
		Config:    config,
	}
}

func (p *Pod) DescribePod() (string, error) {
	if p.Pod == nil {
		if err := p.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch pod: %v", err)
		}
	}

	events, err := p.Client.CoreV1().Events(p.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", p.Name, p.Namespace),
	})
	if err != nil {
		panic(err.Error())
	}

	yamlData, err := yaml.Marshal(describePod(p.Pod, events))
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod to YAML: %v", err)
	}

	p.YAML = string(yamlData)
	return p.YAML, nil
}

func (p *Pod) Fetch() error {
	pod, err := p.Client.CoreV1().Pods(p.Namespace).Get(context.Background(), p.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod: %v", err)
	}
	p.Pod = pod
	return nil
}

func (p *Pod) Exec(command []string) (string, string, error) {
	req := p.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(p.Name).
		Namespace(p.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.Config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), err
}

func (p *Pod) GetLogs() (string, error) {
	req := p.Client.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &corev1.PodLogOptions{})
	logs, err := req.DoRaw(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}
	return string(logs), nil
}

func (p *Pod) GetStatus() (corev1.PodStatus, error) {
	if p.Pod == nil {
		if err := p.Fetch(); err != nil {
			return corev1.PodStatus{}, err
		}
	}
	return p.Pod.Status, nil
}

func formatTime(t metav1.Time) string {
	return t.Format(time.RFC1123)
}

func describePod(pod *corev1.Pod, events *corev1.EventList) string {
	desc := fmt.Sprintf("Name:             %s\n", pod.Name)
	desc += fmt.Sprintf("Namespace:        %s\n", pod.Namespace)
	desc += fmt.Sprintf("Priority:         %d\n", *pod.Spec.Priority)
	desc += fmt.Sprintf("Service Account:  %s\n", pod.Spec.ServiceAccountName)
	desc += fmt.Sprintf("Node:             %s/%s\n", pod.Spec.NodeName, pod.Status.HostIP)
	desc += fmt.Sprintf("Start Time:       %s\n", formatTime(*pod.Status.StartTime))
	
	if len(pod.Labels) == 0 {
		desc += "Labels:           <none>\n"
	} else {
		desc += "Labels:\n"
		for k, v := range pod.Labels {
			desc += fmt.Sprintf("  %s=%s\n", k, v)
		}
	}

	if len(pod.Annotations) == 0 {
		desc += "Annotations:      <none>\n"
	} else {
		desc += "Annotations:\n"
		for k, v := range pod.Annotations {
			desc += fmt.Sprintf("  %s=%s\n", k, v)
		}
	}

	desc += fmt.Sprintf("Status:           %s\n", pod.Status.Phase)
	desc += fmt.Sprintf("IP:               %s\n", pod.Status.PodIP)
	desc += "IPs:\n"
	for _, ip := range pod.Status.PodIPs {
		desc += fmt.Sprintf("  IP:  %s\n", ip.IP)
	}

	desc += "Containers:\n"
	for _, container := range pod.Spec.Containers {
		desc += fmt.Sprintf("  %s:\n", container.Name)
		
		var status corev1.ContainerStatus
		for _, s := range pod.Status.ContainerStatuses {
			if s.Name == container.Name {
				status = s
				break
			}
		}

		desc += fmt.Sprintf("    Container ID:   %s\n", status.ContainerID)
		desc += fmt.Sprintf("    Image:          %s\n", container.Image)
		desc += fmt.Sprintf("    Image ID:       %s\n", status.ImageID)
		
		if len(container.Ports) > 0 {
			for _, port := range container.Ports {
				desc += fmt.Sprintf("    Port:           %d/%s\n", port.ContainerPort, port.Protocol)
				desc += fmt.Sprintf("    Host Port:      %d/%s\n", port.HostPort, port.Protocol)
			}
		} else {
			desc += "    Port:           <none>\n"
		}

		desc += "    State:          "
		if status.State.Running != nil {
			desc += fmt.Sprintf("Running\n      Started:      %s\n", formatTime(status.State.Running.StartedAt))
		} else if status.State.Terminated != nil {
			desc += fmt.Sprintf("Terminated\n      Reason:       %s\n", status.State.Terminated.Reason)
			desc += fmt.Sprintf("      Exit Code:    %d\n", status.State.Terminated.ExitCode)
			desc += fmt.Sprintf("      Started:      %s\n", formatTime(status.State.Terminated.StartedAt))
			desc += fmt.Sprintf("      Finished:     %s\n", formatTime(status.State.Terminated.FinishedAt))
		} else if status.State.Waiting != nil {
			desc += fmt.Sprintf("Waiting\n      Reason:       %s\n", status.State.Waiting.Reason)
		}

		if status.LastTerminationState.Terminated != nil {
			desc += "    Last State:     Terminated\n"
			desc += fmt.Sprintf("      Reason:       %s\n", status.LastTerminationState.Terminated.Reason)
			desc += fmt.Sprintf("      Exit Code:    %d\n", status.LastTerminationState.Terminated.ExitCode)
			desc += fmt.Sprintf("      Started:      %s\n", formatTime(status.LastTerminationState.Terminated.StartedAt))
			desc += fmt.Sprintf("      Finished:     %s\n", formatTime(status.LastTerminationState.Terminated.FinishedAt))
		}

		desc += fmt.Sprintf("    Ready:          %t\n", status.Ready)
		desc += fmt.Sprintf("    Restart Count:  %d\n", status.RestartCount)
		
		if len(container.Env) == 0 {
			desc += "    Environment:    <none>\n"
		} else {
			desc += "    Environment:\n"
			for _, env := range container.Env {
				desc += fmt.Sprintf("      %s=%s\n", env.Name, env.Value)
			}
		}

		if len(container.VolumeMounts) == 0 {
			desc += "    Mounts:         <none>\n"
		} else {
			desc += "    Mounts:\n"
			for _, mount := range container.VolumeMounts {
				desc += fmt.Sprintf("      %s from %s (%s)\n", mount.MountPath, mount.Name, mount.ReadOnly)
			}
		}
	}

	desc += "Conditions:\n"
	desc += "  Type                        Status\n"
	for _, condition := range pod.Status.Conditions {
		desc += fmt.Sprintf("  %-28s %v\n", condition.Type, condition.Status)
	}

	desc += "Volumes:\n"
	for _, volume := range pod.Spec.Volumes {
		desc += fmt.Sprintf("  %s:\n", volume.Name)
		if volume.Projected != nil {
			desc += "    Type:                    Projected (a volume that contains injected data from multiple sources)\n"
			if volume.Projected.Sources[0].ServiceAccountToken != nil {
				desc += fmt.Sprintf("    TokenExpirationSeconds:  %d\n", volume.Projected.Sources[0].ServiceAccountToken.ExpirationSeconds)
			}
			for _, source := range volume.Projected.Sources {
				if source.ConfigMap != nil {
					desc += fmt.Sprintf("    ConfigMapName:          %s\n", source.ConfigMap.Name)
					desc += fmt.Sprintf("    ConfigMapOptional:      %v\n", source.ConfigMap.Optional)
				}
				if source.DownwardAPI != nil {
					desc += "    DownwardAPI:            true\n"
				}
			}
		}
	}

	desc += fmt.Sprintf("QoS Class:                   %s\n", pod.Status.QOSClass)

	if len(pod.Spec.NodeSelector) == 0 {
		desc += "Node-Selectors:              <none>\n"
	} else {
		desc += "Node-Selectors:\n"
		for k, v := range pod.Spec.NodeSelector {
			desc += fmt.Sprintf("  %s=%s\n", k, v)
		}
	}

	if len(pod.Spec.Tolerations) == 0 {
		desc += "Tolerations:                 <none>\n"
	} else {
		desc += "Tolerations:\n"
		for _, tol := range pod.Spec.Tolerations {
			desc += fmt.Sprintf("  %s:%s op=%s for %s\n", tol.Key, tol.Operator, tol.Effect, tol.TolerationSeconds)
		}
	}

	desc += "Events:\n"
	if len(events.Items) == 0 {
		desc += "  <none>\n"
	} else {
		desc += "  Type    Reason          Age   From               Message\n"
		desc += "  ----    ------          ----  ----               -------\n"
		for _, event := range events.Items {
			age := time.Since(event.LastTimestamp.Time).Round(time.Second)
			desc += fmt.Sprintf("  %-7s %-15s %-5s %-18s %s\n",
				event.Type,
				event.Reason,
				age.String(),
				event.Source.Component,
				event.Message)
		}
	}

	return desc
}
