package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
}

func NewClient(kubeconfigPath string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{clientset: clientset}, nil
}

// func (c *Client) GetPods(ctx context.Context, namespace string) ([]Pod, error) {
// 	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	
// 	// Convert to your own Pod type
// 	var result []Pod
// 	for _, p := range pods.Items {
// 		result = append(result, Pod{
// 			Name:      p.Name,
// 			Namespace: p.Namespace,
// 			Status:    string(p.Status.Phase),
// 		})
// 	}
// 	
// 	return result, nil
// }
