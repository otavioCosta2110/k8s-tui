package models

import (
	"fmt"

	"otaviocosta2110/k8s-tui/internal/k8s"

	"github.com/charmbracelet/bubbles/table"
)

type PodData struct {
	*k8s.PodInfo
}

func (p PodData) GetName() string {
	return p.Name
}

func (p PodData) GetNamespace() string {
	return p.Namespace
}

func (p PodData) GetColumns() table.Row {
	return table.Row{
		p.Namespace,
		p.Name,
		p.Ready,
		p.Status,
		fmt.Sprintf("%d", p.Restarts),
		p.Age,
	}
}

type DeploymentData struct {
	*k8s.DeploymentInfo
}

func (d DeploymentData) GetName() string {
	return d.Name
}

func (d DeploymentData) GetNamespace() string {
	return d.Namespace
}

func (d DeploymentData) GetColumns() table.Row {
	return table.Row{
		d.Namespace,
		d.Name,
		d.Ready,
		d.UpToDate,
		d.Available,
		d.Age,
	}
}

type ReplicaSetData struct {
	*k8s.ReplicaSetInfo
}

func (r ReplicaSetData) GetName() string {
	return r.Name
}

func (r ReplicaSetData) GetNamespace() string {
	return r.Namespace
}

func (r ReplicaSetData) GetColumns() table.Row {
	return table.Row{
		r.Namespace,
		r.Name,
		r.Desired,
		r.Current,
		r.Ready,
		r.Age,
	}
}

type ConfigMapData struct {
	*k8s.Configmap
}

func (c ConfigMapData) GetName() string {
	return c.Name
}

func (c ConfigMapData) GetNamespace() string {
	return c.Namespace
}

func (c ConfigMapData) GetColumns() table.Row {
	return table.Row{
		c.Namespace,
		c.Name,
		c.Data,
		c.Age,
	}
}

type IngressData struct {
	*k8s.IngressInfo
}

func (i IngressData) GetName() string {
	return i.Name
}

func (i IngressData) GetNamespace() string {
	return i.Namespace
}

func (i IngressData) GetColumns() table.Row {
	return table.Row{
		i.Namespace,
		i.Name,
		i.Class,
		i.Hosts,
		i.Address,
		i.Ports,
		i.Age,
	}
}

type ServiceData struct {
	*k8s.ServiceInfo
}

func (s ServiceData) GetName() string {
	return s.Name
}

func (s ServiceData) GetNamespace() string {
	return s.Namespace
}

func (s ServiceData) GetColumns() table.Row {
	return table.Row{
		s.Namespace,
		s.Name,
		s.Type,
		s.ClusterIP,
		s.ExternalIP,
		s.Ports,
		s.Age,
	}
}

type SecretData struct {
	*k8s.SecretInfo
}

func (s SecretData) GetName() string {
	return s.Name
}

func (s SecretData) GetNamespace() string {
	return s.Namespace
}

func (s SecretData) GetColumns() table.Row {
	return table.Row{
		s.Namespace,
		s.Name,
		s.Type,
		s.Data,
		s.Age,
	}
}
