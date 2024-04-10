package subscriptions

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DeleteItem struct {
	UserName string
}

type DeleteResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func DeleteItemFromTable(dynamoClient *dynamodb.DynamoDB, tableName, uuid string, userName string) DeleteResponse {
	/*
		Deletes the given Item in the DynamoDB table based on the UUID
		and the username provided.
		Params: dynamoClient *dynamodb.DynamoDB
			    tableName string
				uuid string
				username string
				item DeleteItem
		Return: DeleteResponse
	*/
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"UUID": {
				S: aws.String(uuid),
			},
			"UserName": {
				S: aws.String(userName),
			},
		},
		TableName: aws.String(tableName),
	}

	_, err := dynamoClient.DeleteItem(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Printf("Conditional Check failed for item deletion")
				return DeleteResponse{
					Message:    "Conditional Check failed for item deletion",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				log.Printf("Throughput exceeded for table %s please try again after sometime", tableName)
				return DeleteResponse{
					Message:    "Throughput exceeded for table, try again after sometime",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeResourceNotFoundException:
				log.Printf("Table %s not found, or item not found please check tablename/item", tableName)
				return DeleteResponse{
					Message:    "Table/Item not found, please check the tablename/item",
					StatusCode: 400,
				}
			case dynamodb.ErrCodeTransactionConflictException:
				log.Printf("Transaction already in progress for itme, please try again after sometime")
				return DeleteResponse{
					Message:    "Transaction in progress for item, try again after sometime",
					StatusCode: 400,
				}
			}
		} else {
			log.Printf("Got error calling DeleteItem: %s", err)
			return DeleteResponse{
				StatusCode: 400,
			}
		}
	}

	log.Printf("Succesfully Deleted Item from table %s", tableName)
	return DeleteResponse{
		StatusCode: 200,
	}
}
