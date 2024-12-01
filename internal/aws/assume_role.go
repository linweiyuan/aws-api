package aws

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gofiber/fiber/v3"

	"github.com/linweiyuan/aws-api/internal/common"
	"github.com/linweiyuan/aws-api/internal/proxy"
)

type AssumeRoleRequest struct {
	UserInfo
	PrincipalArn  string `json:"principalArn"`
	RoleArn       string `json:"roleArn"`
	SAMLAssertion string `json:"samlAssertion"`
}

type Credentials struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}

type AssumeRoleResponse struct {
	Credentials
	Expiration *time.Time `json:"expiration"`
}

func AssumeRole(c fiber.Ctx) error {
	assumeRoleRequest := new(AssumeRoleRequest)
	if err := c.Bind().JSON(assumeRoleRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to parse assume role request"})
	}

	assumeRoleResponse, err := assumeRoleRequest.Assume()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	responseCredentials := assumeRoleResponse.Credentials
	location, _ := time.LoadLocation(common.TZ)
	expiration := assumeRoleResponse.Credentials.Expiration.In(location)

	return c.JSON(AssumeRoleResponse{
		Credentials: Credentials{
			AccessKeyId:     *responseCredentials.AccessKeyId,
			SecretAccessKey: *responseCredentials.SecretAccessKey,
			SessionToken:    *responseCredentials.SessionToken,
		},
		Expiration: &expiration,
	})
}

func (assumeRoleRequest AssumeRoleRequest) Assume() (*sts.AssumeRoleWithSAMLOutput, error) {
	staClient := sts.NewFromConfig(aws.Config{
		Region:     common.RegionAP,
		HTTPClient: proxy.NewClientWithAuth(assumeRoleRequest.Username, assumeRoleRequest.Password),
	})
	assumeROleResponse, err := staClient.AssumeRoleWithSAML(context.TODO(), &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  &assumeRoleRequest.PrincipalArn,
		RoleArn:       &assumeRoleRequest.RoleArn,
		SAMLAssertion: &assumeRoleRequest.SAMLAssertion,
	})
	if err != nil {
		return nil, err
	}

	return assumeROleResponse, nil
}
