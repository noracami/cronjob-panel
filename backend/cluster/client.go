// backend/cluster/client.go
package cluster

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClient(endpoint, authType string, authData []byte) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	switch authType {
	case "token":
		config = &rest.Config{
			Host:            endpoint,
			BearerToken:     string(authData),
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		}
	case "kubeconfig":
		config, err = clientcmd.RESTConfigFromKubeConfig(authData)
		if err != nil {
			return nil, fmt.Errorf("parse kubeconfig: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown auth type: %s", authType)
	}

	return kubernetes.NewForConfig(config)
}
