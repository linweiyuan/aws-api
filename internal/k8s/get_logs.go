package k8s

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	clusterEndpoint = ""
	kdToken         = ""
	caData          = ""
)

func GetLogs(c fiber.Ctx) error {
	proxy, _ := url.Parse(kubectlProxy)
	k8sClient, _ := kubernetes.NewForConfig(
		&rest.Config{
			Host:        clusterEndpoint,
			BearerToken: kdToken,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: []byte(caData),
			},
			Proxy: http.ProxyURL(proxy),
		},
	)
	podLogs := k8sClient.CoreV1().Pods(c.Params("namespace")).GetLogs(c.Params("pod"), &v1.PodLogOptions{
		Container: c.Params("container"),
		Follow:    false,
	}).Do(context.TODO())
	raw, _ := podLogs.Raw()

	return c.SendString(string(raw))
}
