package models

import (
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

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

type ServiceAccountData struct {
	*k8s.ServiceAccountInfo
}

func (s ServiceAccountData) GetName() string {
	return s.Name
}

func (s ServiceAccountData) GetNamespace() string {
	return s.Namespace
}

func (s ServiceAccountData) GetColumns() table.Row {
	return table.Row{
		s.Namespace,
		s.Name,
		s.Secrets,
		s.Age,
	}
}

type NodeData struct {
	*k8s.NodeInfo
}

func (n NodeData) GetName() string {
	return n.Name
}

func (n NodeData) GetNamespace() string {
	return ""
}

func (n NodeData) GetColumns() table.Row {
	return table.Row{
		n.Name,
		n.Status,
		n.Roles,
		n.Version,
		n.CPU,
		n.Memory,
		n.Pods,
		n.Age,
	}
}

type JobData struct {
	*k8s.JobInfo
}

func (j JobData) GetName() string {
	return j.Name
}

func (j JobData) GetNamespace() string {
	return j.Namespace
}

func (j JobData) GetColumns() table.Row {
	return table.Row{
		j.Namespace,
		j.Name,
		j.Completions,
		j.Duration,
		j.Age,
	}
}

type CronJobData struct {
	*k8s.CronJobInfo
}

func (cj CronJobData) GetName() string {
	return cj.Name
}

func (cj CronJobData) GetNamespace() string {
	return cj.Namespace
}

func (cj CronJobData) GetColumns() table.Row {
	return table.Row{
		cj.Namespace,
		cj.Name,
		cj.Schedule,
		cj.Suspend,
		cj.Active,
		cj.LastSchedule,
		cj.Age,
	}
}

type DaemonSetData struct {
	*k8s.DaemonSetInfo
}

func (ds DaemonSetData) GetName() string {
	return ds.Name
}

func (ds DaemonSetData) GetNamespace() string {
	return ds.Namespace
}

func (ds DaemonSetData) GetColumns() table.Row {
	return table.Row{
		ds.Namespace,
		ds.Name,
		ds.Desired,
		ds.Current,
		ds.Ready,
		ds.UpToDate,
		ds.Available,
		ds.NodeSelector,
		ds.Age,
	}
}

type StatefulSetData struct {
	*k8s.StatefulSetInfo
}

func (ss StatefulSetData) GetName() string {
	return ss.Name
}

func (ss StatefulSetData) GetNamespace() string {
	return ss.Namespace
}

func (ss StatefulSetData) GetColumns() table.Row {
	return table.Row{
		ss.Namespace,
		ss.Name,
		ss.Ready,
		ss.Age,
	}
}
