package config

import "github.com/gofiber/storage/s3"

var (
	s3Storage   *s3.Storage
	s3Endpoint  = ""
	s3Region    = ""
	s3AccessKey = ""
	s3SecretKey = ""
)

func GetS3Storage(bucketname string) *s3.Storage {
	if s3Storage == nil {
		s3Storage = s3.New(s3.Config{
			Endpoint: s3Endpoint,
			Bucket:   bucketname,
			Region:   s3Region,
			Credentials: s3.Credentials{
				AccessKey:       s3AccessKey,
				SecretAccessKey: s3SecretKey,
			},
		})
	}

	return s3Storage
}
