package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceInfo struct {
	Namespace  string
	Name       string
	Type       string
	ClusterIP  string
	ExternalIP string
	Ports      string
	Age        string
	Raw        *corev1.Service
	Client     Client
}

func NewService(name, namespace string, k Client) *ServiceInfo {
	return &ServiceInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchServiceList(client Client, namespace string) ([]string, error) {
	services, err := client.Clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %v", err)
	}

	serviceNames := make([]string, 0, len(services.Items))
	for _, service := range services.Items {
		serviceNames = append(serviceNames, service.Name)
	}

	return serviceNames, nil
}

func GetServicesTableData(client Client, namespace string) ([]ServiceInfo, error) {
	services, err := client.Clientset.CoreV1().Services(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	var serviceInfos []ServiceInfo
	for _, service := range services.Items {
		serviceType := string(service.Spec.Type)

		clusterIP := service.Spec.ClusterIP
		if clusterIP == "" {
			clusterIP = "<none>"
		}

		var externalIPs []string
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			for _, ingress := range service.Status.LoadBalancer.Ingress {
				if ingress.IP != "" {
					externalIPs = append(externalIPs, ingress.IP)
				} else if ingress.Hostname != "" {
					externalIPs = append(externalIPs, ingress.Hostname)
				}
			}
		}
		externalIP := ""
		if len(externalIPs) > 0 {
			externalIP = externalIPs[0]
			if len(externalIPs) > 1 {
				externalIP += fmt.Sprintf(" +%d more", len(externalIPs)-1)
			}
		} else {
			externalIP = "<none>"
		}

		var ports []string
		for _, port := range service.Spec.Ports {
			var portStr string
			if port.Name != "" {
				portStr = fmt.Sprintf("%s:%d/%s", port.Name, port.Port, string(port.Protocol))
			} else {
				portStr = fmt.Sprintf("%d:%s/%s", port.Port, port.TargetPort.String(), string(port.Protocol))
			}
			ports = append(ports, portStr)
		}
		portsStr := ""
		if len(ports) > 0 {
			portsStr = ports[0]
			if len(ports) > 1 {
				portsStr += fmt.Sprintf(" +%d more", len(ports)-1)
			}
		}

		serviceInfos = append(serviceInfos, ServiceInfo{
			Namespace:  service.Namespace,
			Name:       service.Name,
			Type:       serviceType,
			ClusterIP:  clusterIP,
			ExternalIP: externalIP,
			Ports:      portsStr,
			Age:        utils.FormatAge(service.CreationTimestamp.Time),
			Raw:        service.DeepCopy(),
			Client:     client,
		})
	}

	return serviceInfos, nil
}

func (s *ServiceInfo) Fetch() error {
	service, err := s.Client.Clientset.CoreV1().Services(s.Namespace).Get(context.Background(), s.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service: %v", err)
	}
	s.Raw = service
	return nil
}

func (s *ServiceInfo) Describe() (string, error) {
	if s.Raw == nil {
		if err := s.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch service: %v", err)
		}
	}

	events, err := s.Client.Clientset.CoreV1().Events(s.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Service", s.Name, s.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get service events: %v", err)
	}

	data, err := s.DescribeService(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe service: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (s *ServiceInfo) DescribeService(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        s.Name,
		"namespace":   s.Namespace,
		"labels":      s.Raw.Labels,
		"annotations": s.Raw.Annotations,
		"created":     formatTime(s.Raw.CreationTimestamp),
	}

	desc["type"] = string(s.Raw.Spec.Type)

	if s.Raw.Spec.ClusterIP != "" {
		desc["clusterIP"] = s.Raw.Spec.ClusterIP
	}

	if len(s.Raw.Spec.ExternalIPs) > 0 {
		desc["externalIPs"] = s.Raw.Spec.ExternalIPs
	}

	if len(s.Raw.Status.LoadBalancer.Ingress) > 0 {
		lbIngress := make([]map[string]any, 0, len(s.Raw.Status.LoadBalancer.Ingress))
		for _, lb := range s.Raw.Status.LoadBalancer.Ingress {
			lbDesc := map[string]any{}
			if lb.IP != "" {
				lbDesc["ip"] = lb.IP
			}
			if lb.Hostname != "" {
				lbDesc["hostname"] = lb.Hostname
			}
			lbIngress = append(lbIngress, lbDesc)
		}
		desc["loadBalancer"] = map[string]any{"ingress": lbIngress}
	}

	if len(s.Raw.Spec.Selector) > 0 {
		desc["selector"] = s.Raw.Spec.Selector
	}

	if len(s.Raw.Spec.Ports) > 0 {
		ports := make([]map[string]any, 0, len(s.Raw.Spec.Ports))
		for _, port := range s.Raw.Spec.Ports {
			portDesc := map[string]any{
				"port":     port.Port,
				"protocol": string(port.Protocol),
			}
			if port.Name != "" {
				portDesc["name"] = port.Name
			}
			if port.TargetPort.Type == 1 {
				portDesc["targetPort"] = port.TargetPort.IntVal
			} else {
				portDesc["targetPort"] = port.TargetPort.StrVal
			}
			if port.NodePort != 0 {
				portDesc["nodePort"] = port.NodePort
			}
			ports = append(ports, portDesc)
		}
		desc["ports"] = ports
	}

	if s.Raw.Spec.SessionAffinity != "" {
		desc["sessionAffinity"] = string(s.Raw.Spec.SessionAffinity)
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

func DeleteService(client Client, namespace string, serviceName string) error {
	err := client.Clientset.CoreV1().Services(namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service %s: %v", serviceName, err)
	}
	return nil
}
