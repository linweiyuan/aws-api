package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gofiber/fiber/v3"

	myAws "github.com/linweiyuan/aws-api/internal/aws"
	"github.com/linweiyuan/aws-api/internal/common"
	"github.com/linweiyuan/aws-api/internal/proxy"
)

const (
	clusterName        = ""
	kubectlProxy       = ""
	clusterSet         = ""
	kubeConfigTemplate = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
	proxy-url: %s
	server: %s
  name: %s
contexts:
- context:
    cluster: %s
	user: %s
  name: %s
current-context: %s
users:
- name: %s
  user:
    token: %s
`
)

type GetK8sTokenRequest struct {
	myAws.AssumeRoleRequest
}

type GetK8sTokenResponse struct {
	KdToken    string `json:"kdToken"`
	KubeConfig string `json:"kubeConfig"`
}

func GetToken(c fiber.Ctx) error {
	getK8sTokenRequest := new(GetK8sTokenRequest)
	if err := c.Bind().JSON(getK8sTokenRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to parse get k8s token request"})
	}

	assumeRoleResponse, err := getK8sTokenRequest.AssumeRoleRequest.Assume()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to assume role"})
	}

	requestCredentials := assumeRoleResponse.Credentials
	awsCredentials := credentials.NewStaticCredentialsProvider(
		*requestCredentials.AccessKeyId,
		*requestCredentials.SecretAccessKey,
		*requestCredentials.SessionToken,
	)

	httpClient := proxy.NewClientWithAuth(getK8sTokenRequest.Username, getK8sTokenRequest.Password)
	stsClient := sts.NewFromConfig(aws.Config{
		Region:     common.RegionAP,
		HTTPClient: httpClient,
	}, func(options *sts.Options) {
		options.Credentials = awsCredentials
	})

	ctx := context.TODO()

	presignedHttpRequest, _ := sts.NewPresignClient(stsClient).PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(options *sts.PresignOptions) {
		options.Presigner = newSignerV4(options.Presigner, map[string]string{
			"X-K8s-Aws-Id":  clusterName,
			"X-Amz-Expires": "900",
		})
	})

	kdToken := fmt.Sprintf("k8s-aws-v1.%s", base64.RawStdEncoding.EncodeToString([]byte(presignedHttpRequest.URL)))

	eksCluster := getEksCluster(ctx, awsCredentials, httpClient)
	kubeConfig := fmt.Sprintf(
		kdToken,
		*eksCluster.CertificateAuthority.Data,
		kubectlProxy,
		*eksCluster.Endpoint,
		*eksCluster.Arn,
		*eksCluster.Arn,
		*eksCluster.Arn,
		clusterSet,
		clusterSet,
		*eksCluster.Arn,
		kdToken,
	)

	return c.JSON(&GetK8sTokenResponse{
		KdToken:    kdToken,
		KubeConfig: kubeConfig,
	})
}

func getEksCluster(ctx context.Context, credentials aws.CredentialsProvider, httpClient aws.HTTPClient) *types.Cluster {
	eksClient := eks.New(
		eks.Options{
			Region:      common.RegionAP,
			Credentials: credentials,
			HTTPClient:  httpClient,
		},
	)
	result, _ := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	return result.Cluster
}
