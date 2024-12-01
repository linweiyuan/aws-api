package ecs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/gofiber/fiber/v3"

	myAws "github.com/linweiyuan/aws-api/internal/aws"
	"github.com/linweiyuan/aws-api/internal/common"
	"github.com/linweiyuan/aws-api/internal/proxy"
)

const (
	// ecs
	cluster     = ""
	serviceName = ""

	// cloudwatch
	logGroupName        = ""
	logStreamNamePrefix = ""
)

type GetLogsRequest struct {
	myAws.AssumeRoleRequest
}

type LogInfo struct {
	Timestamp *int64 `json:"timestamp"`
	Message   string `json:"message"`
}

func GetLogs(c fiber.Ctx) error {
	ctx := context.TODO()

	getLogsRequest := new(GetLogsRequest)
	if err := c.Bind().JSON(getLogsRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to parse get logs request"})
	}

	assumeRoleResponse, err := getLogsRequest.AssumeRoleRequest.Assume()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "failed to assume role"})
	}

	requestCredentials := assumeRoleResponse.Credentials
	httpClient := proxy.NewClientWithAuth(getLogsRequest.Username, getLogsRequest.Password)
	awsCredentials := credentials.NewStaticCredentialsProvider(
		*requestCredentials.AccessKeyId,
		*requestCredentials.SecretAccessKey,
		*requestCredentials.SessionToken,
	)

	ecsClient := ecs.NewFromConfig(aws.Config{
		Region:     common.RegionAP,
		HTTPClient: httpClient,
	}, func(options *ecs.Options) {
		options.Credentials = awsCredentials
	})

	tasks, err := ecsClient.ListTasks(ctx, &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(serviceName),
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	task := tasks.TaskArns[0]
	taskId := task[strings.LastIndex(task, "/")+1:]

	cloudwatchLogsClient := cloudwatchlogs.NewFromConfig(aws.Config{
		Region:     common.RegionAP,
		HTTPClient: httpClient,
	}, func(options *cloudwatchlogs.Options) {
		options.Credentials = awsCredentials
	})

	getLogEvents, err := cloudwatchLogsClient.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(fmt.Sprintf("%s/%s", logStreamNamePrefix, taskId)),
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	var logInfoList []LogInfo
	for _, event := range getLogEvents.Events {
		logInfoList = append(logInfoList, LogInfo{
			Timestamp: event.Timestamp,
			Message:   *event.Message,
		})
	}

	return c.JSON(logInfoList)
}
