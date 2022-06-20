package services

import (
	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/log"
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CheckTagNone(tagValue string) bool {
	return (strings.ToLower(tagValue) == "none") || (tagValue == "") || (tagValue == "NotDefined")
}
func GetAllDynamoResourcesMissingMandatoryTags(svc *dynamodb.Client, meta config.compsMeta) map[string]map[string]string {
	result := make(map[string]map[string]string)
	validTeam := regexp.MustCompile(meta[0].ValueRegex[0])
	validEnv := regexp.MustCompile(meta[1].ValueRegex[0])
	dynamoList, err := svc.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		panic(err.Error())
	}

	if dynamoList.TableNames == nil {
		log.Fatalf("item not found")
	}
	for _, value := range dynamoList.TableNames {

		item, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: &value,
		})
		tableArn := item.Table.TableArn
		if err != nil {
			panic(err.Error())
		}

		tagr, err := svc.ListTagsOfResource(context.TODO(), &dynamodb.ListTagsOfResourceInput{
			ResourceArn: tableArn,
		})
		if err != nil {
			panic(err.Error())
		}
		teamPresent := false
		environmentPresent := false
		repoPresent := false
		missingTagsInSet := false
		dyTags := make(map[string]string)
		for _, tagSet := range tagr.Tags {
			log.Info("----------------------")
			log.Info("Working with " + value)
			switch *tagSet.Key {
			case "Team":
				{
					// log.Info("Team :" + *tagSet.Value)
					dyTags["team"] = *tagSet.Value
					if !CheckTagNone(*tagSet.Value) {
						if validTeam.MatchString(*tagSet.Value) {
							for _, val := range meta[0].Values {
								if val.Value == *tagSet.Value {
									// log.Info("Comparing " + val.Value + " and " + *tag.Value)
									teamPresent = true
									break
								}
							}
						}
					}

					break
				}
			case "Environment":
				{
					// log.Info("Environment :" + *tagSet.Value)
					dyTags["environment"] = *tagSet.Value
					if !CheckTagNone(*tagSet.Value) {
						if validEnv.MatchString(*tagSet.Value) {
							for _, val := range meta[1].Values {

								if val.Value == *tagSet.Value {
									// log.Info("Comparing " + val.Value + " and " + *tag.Value)
									environmentPresent = true
									break
								}
							}
						}
					}
					break
				}
			case "Repo":
				{
					// log.Info("Repo :" + *tagSet.Value)
					dyTags["repo"] = *tagSet.Value
					if !CheckTagNone(*tagSet.Value) {
						for _, val := range meta[2].ValueRegex {
							if regexp.MustCompile(val).MatchString(*tagSet.Value) {
								repoPresent = true
								break
							}
						}
					}
					break
				}

			default:
				{
					// log.Info("Ignoring " + *tagSet.Key + " with value " + *tagSet.Value)
				}
			} // end of switch

		}
		if !teamPresent {
			// log.Info("Team is not present")
			// dyTags["team"] = "none"
			missingTagsInSet = true

		}
		if !environmentPresent {
			// log.Info("Environment is not present")
			// dyTags["environment"] = "none"
			missingTagsInSet = true

		}
		if !repoPresent {
			// log.Info("Repo is not present")
			// dyTags["repo"] = "none"
			missingTagsInSet = true

		}

		if missingTagsInSet {

			result[value] = dyTags

		} else {
			log.Info("There are no missing tags for " + *tableArn)

		}
		missingTagsInSet = false //reset bool

	}
	return result
}

func BlockAccessToNonComliantDynamoDbTables(svc *dynamodb.Client, dydata map[string]map[string]string) {

}

/*
Updates DynamoDb tags for a secific account. Accepts a map of maps
*/
func UpdateMandatoryTagsOnDynamo(svc *dynamodb.Client, dydata map[string]map[string]string) {
	var tableArn string
	for _, value := range dydata {
		name := value["name"]
		team := value["team"]
		env := value["environment"]
		repo := value["repo"]
		log.Info("services.UpdateMandatorTagsOnDynamo::Table Name being used::" + name)
		log.Info("services.UpdateMandatorTagsOnDynamo::Table Env being used::" + env)
		log.Info("services.UpdateMandatorTagsOnDynamo::Table repo being used::" + repo)
		log.Info("services.UpdateMandatorTagsOnDynamo::Table team being used::" + team)
		item, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(name),
		})
		if err != nil {
			log.Error(err.Error())
		} else {
			tableArn = *item.Table.TableArn

		}
		val, err := svc.TagResource(context.TODO(), &dynamodb.TagResourceInput{
			ResourceArn: aws.String(tableArn),
			Tags: []types.Tag{
				{
					Key:   aws.String("Team"),
					Value: aws.String(team),
				},
				{
					Key:   aws.String("Environment"),
					Value: aws.String(env),
				},
				{
					Key:   aws.String("Repo"),
					Value: aws.String(repo),
				},
			},
		})
		if err != nil {
			log.Error(err.Error())
		} else {
			_ = val
			time.Sleep(2)
		}
		log.Info("services.UpdateMandatorTagsOnDynamo::Table arn for " + name + " :" + tableArn)

	} // end of for loop

	return
}
