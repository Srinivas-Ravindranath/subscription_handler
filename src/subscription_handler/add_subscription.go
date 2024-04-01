package subscriptions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"strconv"
	"time"
)

type AddResponse struct {
	UUID       string `json:"uuid"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

type AddItem struct {
	UUID                 string
	UserName             string
	VendorName           string
	VendorUrl            string
	SubscriptionDuration string
	RemindTime           string
}

func CalcRemindTime(addtime string) string {
	/*
		Computes the remaining time for the given duration of
		the subscription
		Params: addTime string
		Return: string
	*/
	intTime, err := strconv.Atoi(addtime)
	if err != nil {
		log.Fatalf("Invalid Time format to add: %s", err)
	}
	finalTime := time.Now().AddDate(0, 0, intTime)
	formattedDate := finalTime.Format("2006-01-02")
	return formattedDate
}

func AddItemToTable(dynamoClient *dynamodb.DynamoDB, tableName string, items AddItem) AddResponse {
	/*
		Adds a given Item to the DynamoDB table.
		Params: dynamoClient *dynamodb.DynamoDB
			    tableName string
				items AddItem
		Return: AddResponse
	*/
	mappedItem, _ := dynamodbattribute.MarshalMap(items)
	tableInput := &dynamodb.PutItemInput{
		Item:      mappedItem,
		TableName: aws.String(tableName),
	}
	_, err := dynamoClient.PutItem(tableInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Printf("Conditional Check failed for item addition")
				return AddResponse{
					Message:    "Conditional Check failed for item addition",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				log.Printf("Throughput exceeded for table %s please try again after sometime", tableName)
				return AddResponse{
					Message:    "Throughput exceeded for table, try again after sometime",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeResourceNotFoundException:
				log.Printf("Table %s not found, please check tablename", tableName)
				return AddResponse{
					Message:    "Table not found, please check the tablename",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeTransactionConflictException:
				log.Printf("Transaction already in progress for itme, please try again after sometime")
				return AddResponse{
					Message:    "Transaction in progress for item, try again after sometime",
					StatusCode: 400,
				}
			}
		} else {
			log.Printf("Unable to add item to table: %s\n", err)
			return AddResponse{
				StatusCode: 400,
			}
		}
	}
	log.Printf("Successfully added the items into the table %s", tableName)
	return AddResponse{
		UUID:       items.UUID,
		StatusCode: 200,
	}
}
