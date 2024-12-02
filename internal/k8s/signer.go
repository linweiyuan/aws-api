package k8s

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type signerV4 struct {
	client  sts.HTTPPresignerV4
	headers map[string]string
}

func (signerV4 *signerV4) PresignHTTP(
	ctx context.Context,
	credentials aws.Credentials,
	request *http.Request,
	payloadHash string,
	service string,
	region string,
	signingTime time.Time,
	optFns ...func(*v4.SignerOptions),
) (url string, signedHeader http.Header, err error) {
	for key, value := range signerV4.headers {
		request.Header.Add(key, value)
	}

	return signerV4.client.PresignHTTP(ctx, credentials, request, payloadHash, service, region, signingTime, optFns...)
}

func newSignerV4(client sts.HTTPPresignerV4, headers map[string]string) sts.HTTPPresignerV4 {
	return &signerV4{
		client:  client,
		headers: headers}
}
