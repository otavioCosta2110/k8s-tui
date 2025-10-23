package k8s

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaces struct {
	k Client
}

func NewNamespaces(k Client) *Namespaces {
	return &Namespaces{k: k}
}

func (n *Namespaces) Create(name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err := n.k.Clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	return err
}

func FetchNamespaces(k Client) ([]string, error) {
	if k.Clientset == nil {
		return []string{}, errors.New("clientset is nil")
	}

	namespaces, err := k.Clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}

	namespacesArray := make([]string, 0, len(namespaces.Items))
	for _, nm := range namespaces.Items {
		namespacesArray = append(namespacesArray, nm.Name)
	}
	return namespacesArray, nil
}
