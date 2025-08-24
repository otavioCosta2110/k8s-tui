package k8s

import (
	"context"
	"fmt"
	"otaviocosta2110/k8s-tui/utils"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressInfo struct {
	Namespace string
	Name      string
	Class     string
	Hosts     string
	Address   string
	Ports     string
	Age       string
	Raw       *networkingv1.Ingress
	Client    Client
}

func NewIngress(name, namespace string, k Client) *IngressInfo {
	return &IngressInfo{
		Name:      name,
		Namespace: namespace,
		Client:    k,
	}
}

func FetchIngressList(client Client, namespace string) ([]string, error) {
	ingresses, err := client.Clientset.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ingresses: %v", err)
	}

	ingressNames := make([]string, 0, len(ingresses.Items))
	for _, ingress := range ingresses.Items {
		ingressNames = append(ingressNames, ingress.Name)
	}

	return ingressNames, nil
}

func GetIngressesTableData(client Client, namespace string) ([]IngressInfo, error) {
	ingresses, err := client.Clientset.NetworkingV1().Ingresses(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %v", err)
	}

	var ingressInfos []IngressInfo
	for _, ingress := range ingresses.Items {
		ingressClass := ""
		if ingress.Spec.IngressClassName != nil {
			ingressClass = *ingress.Spec.IngressClassName
		}

		var hosts []string
		for _, rule := range ingress.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}
		hostsStr := ""
		if len(hosts) > 0 {
			hostsStr = hosts[0]
			if len(hosts) > 1 {
				hostsStr += fmt.Sprintf(" +%d more", len(hosts)-1)
			}
		}

		address := ""
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].IP
			} else if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		var ports []string
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Port.Number != 0 {
						ports = append(ports, fmt.Sprintf("%d", path.Backend.Service.Port.Number))
					}
				}
			}
		}
		portsStr := ""
		if len(ports) > 0 {
			portsStr = ports[0]
			if len(ports) > 1 {
				portsStr += fmt.Sprintf(" +%d more", len(ports)-1)
			}
		}

		ingressInfos = append(ingressInfos, IngressInfo{
			Namespace: ingress.Namespace,
			Name:      ingress.Name,
			Class:     ingressClass,
			Hosts:     hostsStr,
			Address:   address,
			Ports:     portsStr,
			Age:       utils.FormatAge(ingress.CreationTimestamp.Time),
			Raw:       ingress.DeepCopy(),
			Client:    client,
		})
	}

	return ingressInfos, nil
}

func DeleteIngress(client Client, namespace string, ingressName string) error {
	err := client.Clientset.NetworkingV1().Ingresses(namespace).Delete(context.Background(), ingressName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ingress %s: %v", ingressName, err)
	}
	return nil
}

func (i *IngressInfo) Fetch() error {
	ingress, err := i.Client.Clientset.NetworkingV1().Ingresses(i.Namespace).Get(context.Background(), i.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}
	i.Raw = ingress
	return nil
}

func (i *IngressInfo) Describe() (string, error) {
	if i.Raw == nil {
		if err := i.Fetch(); err != nil {
			return "", fmt.Errorf("failed to fetch ingress: %v", err)
		}
	}

	events, err := i.Client.Clientset.CoreV1().Events(i.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Ingress", i.Name, i.Namespace),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get ingress events: %v", err)
	}

	data, err := i.DescribeIngress(events)
	if err != nil {
		return "", fmt.Errorf("failed to describe ingress: %v", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ingress to YAML: %v", err)
	}

	return string(yamlData), nil
}

func (i *IngressInfo) DescribeIngress(events *corev1.EventList) (map[string]any, error) {
	type Event struct {
		Type    string `yaml:"type"`
		Reason  string `yaml:"reason"`
		Age     string `yaml:"age"`
		From    string `yaml:"from"`
		Message string `yaml:"message"`
	}

	desc := map[string]any{
		"name":        i.Name,
		"namespace":   i.Namespace,
		"labels":      i.Raw.Labels,
		"annotations": i.Raw.Annotations,
		"created":     formatTime(i.Raw.CreationTimestamp),
	}

	if i.Raw.Spec.IngressClassName != nil {
		desc["ingressClass"] = *i.Raw.Spec.IngressClassName
	}

	if len(i.Raw.Spec.Rules) > 0 {
		rules := make([]map[string]any, 0, len(i.Raw.Spec.Rules))
		for _, rule := range i.Raw.Spec.Rules {
			ruleDesc := map[string]any{}

			if rule.Host != "" {
				ruleDesc["host"] = rule.Host
			}

			if rule.HTTP != nil {
				paths := make([]map[string]any, 0, len(rule.HTTP.Paths))
				for _, path := range rule.HTTP.Paths {
					pathDesc := map[string]any{
						"path":     path.Path,
						"pathType": string(*path.PathType),
					}

					if path.Backend.Service != nil {
						backendDesc := map[string]any{
							"service": map[string]any{
								"name": path.Backend.Service.Name,
							},
						}
						if path.Backend.Service.Port.Number != 0 {
							backendDesc["service"].(map[string]any)["port"] = path.Backend.Service.Port.Number
						}
						if path.Backend.Service.Port.Name != "" {
							backendDesc["service"].(map[string]any)["portName"] = path.Backend.Service.Port.Name
						}
						pathDesc["backend"] = backendDesc
					}

					paths = append(paths, pathDesc)
				}
				ruleDesc["http"] = map[string]any{"paths": paths}
			}

			rules = append(rules, ruleDesc)
		}
		desc["rules"] = rules
	}

	if len(i.Raw.Spec.TLS) > 0 {
		tls := make([]map[string]any, 0, len(i.Raw.Spec.TLS))
		for _, tlsSpec := range i.Raw.Spec.TLS {
			tlsDesc := map[string]any{}
			if len(tlsSpec.Hosts) > 0 {
				tlsDesc["hosts"] = tlsSpec.Hosts
			}
			if tlsSpec.SecretName != "" {
				tlsDesc["secretName"] = tlsSpec.SecretName
			}
			tls = append(tls, tlsDesc)
		}
		desc["tls"] = tls
	}

	if len(i.Raw.Status.LoadBalancer.Ingress) > 0 {
		lbIngress := make([]map[string]any, 0, len(i.Raw.Status.LoadBalancer.Ingress))
		for _, lb := range i.Raw.Status.LoadBalancer.Ingress {
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
