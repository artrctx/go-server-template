package env

import (
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Port              int
	JwksUrl           string
	AuthUrl           string
	DatabaseUrl       string
	PostHogToken      string
	AppEnv            string
	AppName           string
	R2AccountID       string
	R2BucketName      string
	R2AccessKey       string
	R2AccessSecretKey string
)

func init() {
	// set env for testing
	if os.Getenv("SKIP_ENV_INIT") == "true" {
		log.Println("env initialization blocked")
		return
	}
	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatalln("Missing 'PORT' env")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid 'PORT' env got %s expected int", portStr)
	} else {
		Port = port
	}

	DatabaseUrl = os.Getenv("DATABASE_URL")
	if DatabaseUrl == "" {
		log.Fatalln("Missing 'DATABASE_URL' env")
	}

	JwksUrl = os.Getenv("JWKS_URL")
	if JwksUrl == "" {
		log.Fatalln("Missing 'JWKS_URL' env")
	}

	AuthUrl = os.Getenv("AUTH_URL")
	if AuthUrl == "" {
		log.Fatalln("Missing 'AUTH_URL' env")
	}

	PostHogToken = os.Getenv("POSTHOG_TOKEN")
	if PostHogToken == "" {
		log.Fatalln("Missing 'POSTHOG_TOKEN' env")
	}

	R2AccountID = os.Getenv("R2_ACCOUNT_ID")
	if R2AccountID == "" {
		log.Fatalln("Missing 'R2_ACCOUNT_ID' env")
	}

	R2BucketName = os.Getenv("R2_BUCKET_NAME")
	if R2BucketName == "" {
		log.Fatalln("Missing 'R2_BUCKET_NAME' env")
	}

	R2AccessKey = os.Getenv("R2_ACCESS_KEY")
	if R2AccessKey == "" {
		log.Fatalln("Missing 'R2_ACCESS_KEY' env")
	}

	R2AccessSecretKey = os.Getenv("R2_ACCESS_SECRET_KEY")
	if R2AccessKey == "" {
		log.Fatalln("Missing 'R2_ACCESS_SECRET_KEY' env")
	}

	AppEnv = os.Getenv("APP_ENV")
	if AppEnv == "" {
		log.Fatalln("Missing 'APP_ENV' env")
	}

	AppName = fmt.Sprintf("shuffle-%s", AppEnv)

}
