package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/redhatinsights/platform-go-middlewares/v2/logging/cloudwatch"
	"github.com/sirupsen/logrus"
)

var cwHook *cloudwatch.LogrusHook

func SetupCloudWatch(cfg *config.ChromeServiceConfig) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg: "message",
		},
	})

	cwCfg := cfg.CloudWatch
	if cwCfg.AccessKeyId == "" || cwCfg.SecretAccessKey == "" || cwCfg.Region == "" || cwCfg.LogGroup == "" {
		logrus.Info("CloudWatch credentials not configured, logging to stdout only")
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	awsCfg := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(
			cwCfg.AccessKeyId,
			cwCfg.SecretAccessKey,
			"",
		)).
		WithRegion(cwCfg.Region)

	writer, err := cloudwatch.NewBatchWriterWithDuration(
		cwCfg.LogGroup,
		hostname,
		awsCfg,
		5*time.Second,
	)
	if err != nil {
		logrus.Errorf("Failed to initialize CloudWatch writer: %v", err)
		return
	}

	cwHook = cloudwatch.NewLogrusHook(writer)
	logrus.AddHook(cwHook)
	logrus.Info("CloudWatch logging initialized")
}

func FlushCloudWatch() {
	if cwHook != nil {
		if err := cwHook.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush CloudWatch logs: %v\n", err)
		}
	}
}
