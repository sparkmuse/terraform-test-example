package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestTerraformAwsDynamoDBExample(t *testing.T) {
	t.Parallel()

	awsRegion := aws.GetRandomStableRegion(t, []string{"us-east-1"}, nil)

	expectedTableName := fmt.Sprintf("test-table-%s", random.UniqueId())

	// //////////////////////////////////////////
	// Initialize localstack
	// //////////////////////////////////////////
	//ctx := context.Background()
	//req := testcontainers.ContainerRequest{
	//	Image:        "localstack/localstack",
	//	ExposedPorts: []string{"4566/tcp"},
	//	WaitingFor:   wait.ForLog("Ready."),
	//}
	//localStack, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	//	ContainerRequest: req,
	//	Started:          true,
	//})
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//defer func(localStack testcontainers.Container, ctx context.Context) {
	//	err := localStack.Terminate(ctx)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}(localStack, ctx)
	//
	//containerPort, err := localStack.MappedPort(ctx, "4566")
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Println("containerPort = " + containerPort)

	// //////////////////////////////////////////
	// Get terraform data and apply it
	// //////////////////////////////////////////
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../myDynamoDb",
		Vars: map[string]interface{}{
			"table_name": expectedTableName,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// //////////////////////////////////////////
	// Create client
	// //////////////////////////////////////////
	customResolver := awsSdk.EndpointResolverFunc(func(service, region string) (awsSdk.Endpoint, error) {
		return awsSdk.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("mock_key", "mock_secret", "mock_token")),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		t.Error(err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	// //////////////////////////////////////////
	// Do some testing
	// //////////////////////////////////////////
	resp, err := svc.ListTables(context.TODO(), &dynamodb.ListTablesInput{
		Limit: awsSdk.Int32(5),
	})
	if err != nil {
		log.Fatalf("failed to list tables, %v", err)
	}

	fmt.Println("Tables:")
	for _, tableName := range resp.TableNames {
		fmt.Println(tableName)
	}
}
