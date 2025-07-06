package k8s

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaces struct {
	k Client
}

func newNamespaces() *Namespaces {
	return &Namespaces{}
}

func FetchNamespaces(k Client)  ([]string, error) {
	if k.clientset == nil {
		return []string{}, errors.New("clientset is nil")
	}

	namespaces, err := k.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}

	namespacesArray := make([]string, 0, len(namespaces.Items))
	for _, nm := range namespaces.Items {
		namespacesArray = append(namespacesArray, nm.Name)
	}
	return namespacesArray, nil
}
