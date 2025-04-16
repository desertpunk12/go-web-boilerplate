package testfunc

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SimpleRequest represents a simple request parameter for the Lambda function
type SimpleRequest struct {
	Message string `json:"message"`
}

// SimpleResponse represents the response from the Lambda function
type SimpleResponse struct {
	Result string `json:"result"`
}

// HandleRequest is the main Lambda handler function
func HandleRequest(ctx context.Context, request SimpleRequest) (SimpleResponse, error) {
	// Initialize AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return SimpleResponse{}, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Define the S3 bucket name
	bucketName := "my-example-bucket"

	// List objects in the bucket as an example operation
	listResult, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return SimpleResponse{}, fmt.Errorf("failed to list objects in bucket: %v", err)
	}

	// Process the request and generate a response
	objectCount := len(listResult.Contents)
	result := fmt.Sprintf("Received message: %s. The bucket contains %d objects.",
		request.Message, objectCount)

	return SimpleResponse{
		Result: result,
	}, nil
}
