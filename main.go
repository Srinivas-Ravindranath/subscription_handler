package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"os"
	"regexp"
	"subHandler/src/subscription_handler"
)

type dynamoAttr struct {
	dynamoCli *dynamodb.DynamoDB
	awsRegion string
	tableName string
}

type GetItem struct {
	UserName string
}

func initialize() dynamoAttr {
	awsRegion := os.Getenv("aws_region")
	dynamodbTable := os.Getenv("table_name")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	if err != nil {
		fmt.Println("Error creating sess", err)
		os.Exit(0)
	}

	dynamoClient := dynamodb.New(sess)
	return dynamoAttr{
		dynamoCli: dynamoClient,
		awsRegion: awsRegion,
		tableName: dynamodbTable,
	}
}

func handlerSubscription(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	httpMethod := request.HTTPMethod
	if httpMethod == "GET" {
		var getItem GetItem
		err := json.Unmarshal([]byte(request.Body), &getItem)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
		}
		dynamoCli := initialize()
		subscriptionsResponse := subscriptions.GetSubscriptions(dynamoCli.dynamoCli, dynamoCli.tableName, getItem.UserName)
		responseBody, err := json.Marshal(subscriptionsResponse)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: subscriptionsResponse.StatusCode,
			Body:       string(responseBody),
		}, nil
	}
	if httpMethod == "POST" {
		var addItem subscriptions.AddItem
		err := json.Unmarshal([]byte(request.Body), &addItem)
		finalAddItem := subscriptions.AddItem{
			UUID:       uuid.New().String(),
			UserName:   addItem.UserName,
			VendorName: addItem.VendorName,
			VendorUrl:  addItem.VendorUrl,
			Duration:   addItem.Duration,
			RemindTime: subscriptions.CalcRemindTime(addItem.Duration),
		}
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
		}
		dynamoCli := initialize()
		subscriptionsPostResponse := subscriptions.AddItemToTable(dynamoCli.dynamoCli, dynamoCli.tableName, finalAddItem)
		responseBody, err := json.Marshal(subscriptionsPostResponse)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: subscriptionsPostResponse.StatusCode,
			Body:       string(responseBody),
		}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
}

func handlerSubscriptionID(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	httpMethod := request.HTTPMethod
	if httpMethod == "GET" {
		var getItem GetItem
		subscriptionID := request.PathParameters["subscription_id"]
		err := json.Unmarshal([]byte(request.Body), &getItem)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
		}
		dynamoCli := initialize()
		subscriptionResponse := subscriptions.GetSubscription(dynamoCli.dynamoCli, dynamoCli.tableName, subscriptionID, getItem.UserName)
		responseBody, err := json.Marshal(subscriptionResponse)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: subscriptionResponse.Status,
			Body:       string(responseBody),
		}, nil
	}
	if httpMethod == "DELETE" {
		var deleteItem subscriptions.DeleteItem
		subscriptionID := request.PathParameters["subscription_id"]
		err := json.Unmarshal([]byte(request.Body), &deleteItem)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
		}
		dynamoCli := initialize()
		deleteResponse := subscriptions.DeleteItemFromTable(dynamoCli.dynamoCli, dynamoCli.tableName, subscriptionID, deleteItem)
		responseBody, err := json.Marshal(deleteResponse)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: deleteResponse.StatusCode,
			Body:       string(responseBody),
		}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
}

func handlerUpdate(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	httpMethod := request.HTTPMethod
	if httpMethod == "POST" {
		var updateItem subscriptions.UpdateItem
		subscriptionID := request.PathParameters["subscription_id"]
		err := json.Unmarshal([]byte(request.Body), &updateItem)
		finalUpdateItem := subscriptions.UpdateItem{
			UserName:   updateItem.UserName,
			VendorName: updateItem.VendorName,
			VendorUrl:  updateItem.VendorUrl,
			Duration:   updateItem.Duration,
			RemindTime: subscriptions.CalcRemindTime(updateItem.Duration),
		}
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
		}
		dynamoCli := initialize()
		updateResponse := subscriptions.UpdateSubscription(dynamoCli.dynamoCli, dynamoCli.tableName, subscriptionID, finalUpdateItem)
		responseBody, err := json.Marshal(updateResponse)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: updateResponse.Status,
			Body:       string(responseBody),
		}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}, nil
}

func handlerPath(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := request.Path
	var handlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	fmt.Println(request)
	fmt.Println("Entering Handler path")
	subscriptionIDRegex, err := regexp.Compile(`^\/subscriptions\/([0-9a-zA-Z-]+)$`)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, err
	}
	updateSubscriptionIDRegex, err := regexp.Compile(`^\/subscriptions\/update\/([0-9a-zA-Z-]+)$`)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, err
	}
	switch true {
	case path == "/subscriptions":
		handlerFunc = handlerSubscription
	case subscriptionIDRegex.MatchString(path):
		handlerFunc = handlerSubscriptionID
	case updateSubscriptionIDRegex.MatchString(path):
		handlerFunc = handlerUpdate
	default:
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "Not Found"}, nil
	}
	response, err := handlerFunc(ctx, request)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, err
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseJSON),
	}, nil
}

func main() {
	initialize()

	lambda.Start(handlerPath)
}
