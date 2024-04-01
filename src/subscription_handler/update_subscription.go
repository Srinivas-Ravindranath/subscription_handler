// Package subscriptions
/*
Used to update the Item in a DynamoDB table
*/
package subscriptions

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

type UpdateItem struct {
	UserName             string
	VendorName           string
	VendorUrl            string
	SubscriptionDuration string
	RemindTime           string
}

type UpdateResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func UpdateSubscription(dynamoClient *dynamodb.DynamoDB, tableName, uuid string, item UpdateItem) UpdateResponse {
	/*
		Updates the given Item in the DynamoDB table based on the UUID
		and the username provided.
		Params: dynamoClient *dynamodb.DynamoDB
			    tableName string
				uuid string
				item UpdateItem
		Return: UpdateResponse
	*/

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vu": {
				S: aws.String(item.VendorUrl),
			},
			":dur": {
				S: aws.String(item.SubscriptionDuration),
			},
			":rt": {
				S: aws.String(item.RemindTime),
			},
			":vn": {
				S: aws.String(item.VendorName),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UUID": {
				S: aws.String(uuid),
			},
			"UserName": {
				S: aws.String(item.UserName),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("SET VendorUrl = :vu, SubscriptionDuration = :dur, RemindTime = :rt, VendorName = :vn"),
	}

	_, err := dynamoClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Printf("Conditional Check failed for item addition")
				return UpdateResponse{
					Message: "Conditional Check failed for item addition",
					Status:  400,
				}
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				log.Printf("Throughput exceeded for table %s please try again after sometime", tableName)
				return UpdateResponse{
					Message: "Throughput exceeded for table, try again after sometime",
					Status:  400,
				}
			case dynamodb.ErrCodeResourceNotFoundException:
				log.Printf("Table %s not found, please check tablename", tableName)
				return UpdateResponse{
					Message: "Table not found, please check the tablename",
					Status:  400,
				}
			case dynamodb.ErrCodeTransactionConflictException:
				log.Printf("Transaction already in progress for itme, please try again after sometime")
				return UpdateResponse{
					Message: "Transaction in progress for item, try again after sometime",
					Status:  400,
				}
			}
		} else {
			log.Printf("Unable to update item %s", err)
			return UpdateResponse{
				Status: 400,
			}
		}
	}
	log.Printf("Succesfully update table items with UUID %s", uuid)
	return UpdateResponse{
		Status: 200,
	}
}
