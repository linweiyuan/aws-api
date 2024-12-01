package db

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gofiber/fiber/v3"

	myAws "github.com/linweiyuan/aws-api/internal/aws"
	"github.com/linweiyuan/aws-api/internal/common"
	"github.com/linweiyuan/aws-api/internal/proxy"
)

type GetDbTokenRequest struct {
	myAws.AssumeRoleRequest
	DbInfo
}

type DbInfo struct {
	IamRole  string `json:"iamRole"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

func GetToken(c fiber.Ctx) error {
	getDbTokenRequest := new(GetDbTokenRequest)
	if err := c.Bind().JSON(getDbTokenRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to parse get db token request"})
	}

	assumeRoleResponse, err := getDbTokenRequest.AssumeRoleRequest.Assume()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to assume role"})
	}

	requestCredentials := assumeRoleResponse.Credentials
	stsClient := sts.NewFromConfig(aws.Config{
		Region:     common.RegionAP,
		HTTPClient: proxy.NewClientWithAuth(getDbTokenRequest.Username, getDbTokenRequest.Password),
	}, func(options *sts.Options) {
		options.Credentials = credentials.NewStaticCredentialsProvider(
			*requestCredentials.AccessKeyId,
			*requestCredentials.SecretAccessKey,
			*requestCredentials.SessionToken,
		)
	})

	dbInfo := getDbTokenRequest.DbInfo
	assumeRdsRoleResponse, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(dbInfo.IamRole),
		RoleSessionName: aws.String("rds"),
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to assume rds role"})
	}

	responseCredentials := assumeRdsRoleResponse.Credentials
	token, err := auth.BuildAuthToken(context.TODO(), fmt.Sprintf("%s:%d", dbInfo.Host, dbInfo.Port), common.RegionAP, dbInfo.Username, credentials.NewStaticCredentialsProvider(
		*responseCredentials.AccessKeyId,
		*responseCredentials.SecretAccessKey,
		*responseCredentials.SessionToken,
	))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to get db token"})
	}

	return c.JSON(fiber.Map{"token": token})
}
