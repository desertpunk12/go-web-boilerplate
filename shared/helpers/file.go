package helpers

import (
	"context"
	"mime/multipart"
	"web-boilerplate/internal/hr-api/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func FileUploadToS3(file *multipart.FileHeader, bucketname, key string) error {
	src, err := file.Open()

	if err != nil {
		return err
	}

	defer src.Close()

	client := config.GetS3Storage(bucketname)
	// Upload the file to S3
	_, err = client.Conn().PutObject(context.TODO(), &s3.PutObjectInput{
		ACL:                types.ObjectCannedACLPublicRead,
		Bucket:             aws.String(bucketname),
		Key:                aws.String(key),
		ContentType:        aws.String(file.Header.Get("Content-Type")),
		ContentDisposition: aws.String("inline"),
		Body:               src,
	})

	if err != nil {
		return err
	}

	return err
}
