package config

import (
	"aws-compliance-scheduler/pkg/log"
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	external "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

//AWSConfig is a holder for AWS Config type information
type AWSConfig struct {
	Config aws.Config
}

// DefaultAwsConfig loads default AWS Config
func StsAssumedConfig(config Config) (AWSConfig, aws.Credentials) {
	awsConfig := AWSConfig{}
	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", os.Getenv("AWS_ACCOUNT_ID"), os.Getenv("AWS_ROLE"))
	ctx := context.TODO()
	cfg, err := external.LoadDefaultConfig(ctx,
		external.WithRegion("ap-southeast-2"),
		//config.WithClientLogMode(aws.LogSigning),
	)
	if err != nil {
		log.Fatal(err)
	}
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)
	// without the following, I'm getting an error message: api error SignatureDoesNotMatch:
	// The request signature we calculated does not match the signature you provided.
	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(creds.SecretAccessKey)
	awsConfig.Config = cfg
	// if *config.Region != "" {
	// 	awsConfig.Config.Region = *config.Region
	// }

	return awsConfig, creds
}

/*
default config for aws
*/
func DefaultAwsConfig(config Config) AWSConfig {
	awsConfig := AWSConfig{}
	cfg, err := external.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err.Error())
	}
	awsConfig.Config = cfg
	if *config.Region != "" {
		awsConfig.Config.Region = *config.Region
	}
	return awsConfig
}

func MapAccountToConfluenceResource(acc string, resource string) (string, string, string) {
	accountNumber := acc

	
	default:
		{
			return "nil", "nil", "nil" //return labs
		}
	} //end of switch

}

// StsClient returns an STS Client
func (config *AWSConfig) StsClient() *sts.Client {
	return sts.NewFromConfig(config.Config)
}

func AssumedRoleForS3() *s3.Client {
	cfg := GetConfig()
	return s3.NewFromConfig(cfg)
}

func AssumedRoleForDynamoDb() *dynamodb.Client {
	cfg := GetConfig()
	return dynamodb.NewFromConfig(cfg)
}

//Ec2Client returns an ec2 Client
func Ec2Client() *ec2.Client {
	cfg := GetConfig()

	return ec2.NewFromConfig(cfg)
}

/*
Retirns RDS client with creds assumed from the role passed in from ENV
*/
func AssumedRoleForRDS() *rds.Client {
	cfg := GetConfig()
	return rds.NewFromConfig(cfg)
}

func GetConfig() aws.Config {
	if runtime.GOOS == "darwin" {
		cfg, err := external.LoadDefaultConfig(context.TODO(), external.WithRegion(os.Getenv("AWS_DEFAULT_REGION")))
		if err != nil {
			panic(err)
		}
		return cfg
	} else {
		cfg, err := external.LoadDefaultConfig(context.TODO(), external.WithRegion(os.Getenv("AWS_DEFAULT_REGION")))
		if err != nil {
			panic(err)
		}
		role := fmt.Sprintf("arn:aws:iam::%s:role/%s", os.Getenv("AWS_ACCOUNT_ID"), os.Getenv("AWS_ROLE"))
		cfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), role)
		return cfg
	}

}

//IamClient returns an IAM Client
func (config *AWSConfig) IamClient() *iam.Client {
	return iam.NewFromConfig(config.Config)
}

//OrganizationsClient returns an organizations Client
func (config *AWSConfig) OrganizationsClient() *organizations.Client {
	return organizations.NewFromConfig(config.Config)
}
