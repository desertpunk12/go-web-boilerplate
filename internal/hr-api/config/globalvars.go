package config

import "time"

type LOG_LEVEL_TYPE int8

const (
	LOG_LEVEL_NOTFOUND LOG_LEVEL_TYPE = iota - 1
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_DEBUG
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
)

var (
	BASE_URL        = "localhost:3000"
	IS_PROD         = false
	LOG_LEVEL       = LOG_LEVEL_DEBUG
	SECRET_KEY      = "qweasd123"
	ALLOWED_ORIGINS = ""
	REDIS_KEYS_TTL  = time.Hour * 24 * 7 // 7 days
	TOKEN_TTL       = time.Hour * 5      // 5 hours
	S3BUCKETNAME    = "testbucket"
)
