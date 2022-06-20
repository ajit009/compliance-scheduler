package services

import (
	"aws-compliance-scheduler/pkg/comms"
	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/log"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

const WEBHOOK_URL string = "https://hooks.slack.com/services/xxxxx"
const WEBHOOK_SRE string = "https://hooks.slack.com/services/xxxxxx"

// GetRDSName returns the name of the provided RDS Resource
func GetRDSName(rdsname *string, svc *rds.Client) string {
	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: rdsname,
	}
	resp, err := svc.DescribeDBInstances(context.TODO(), params)

	if err != nil {
		panic(err)
	}

	for _, instance := range resp.DBInstances {
		return aws.ToString(instance.DBInstanceIdentifier)
	}
	return ""
}

//GetAllRdsResourceNames gets a list of all names for RDS objects
//TODO: clusters, subnet groups, parameter groups, option groups
func GetAllRdsResourceNames(svc *rds.Client) map[string]string {
	result := make(map[string]string)
	result = addAllInstanceNames(svc, result)
	return result
}

func CheckifValid(tagValue string) bool {

	return (strings.ToLower(tagValue) == "none") || (tagValue == "")
}
func GetAllRdsResourcesMissingMandatoryTags(svc *rds.Client, meta config.compsMeta) map[string]map[string]string {
	result := make(map[string]map[string]string)
	validTeam := regexp.MustCompile(meta[0].ValueRegex[0])
	validEnv := regexp.MustCompile(meta[1].ValueRegex[0])
	resp, err := svc.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Panic(err)
	}

	for _, dbinstance := range resp.DBInstances {
		teamPresent := false
		environmentPresent := false
		repoPresent := false
		missingTagsInSet := false
		dbName := ""
		dbTags := make(map[string]string)
		if dbinstance.TagList != nil && *dbinstance.DBInstanceStatus == "available" {
			// log.Info("--------------------------------------------")
			// log.Info("Working on " + *dbinstance.DBInstanceArn)
			for _, tag := range dbinstance.TagList {
				switch *tag.Key {
				case "Team":
					{

						dbTags["team"] = *tag.Value
						if !CheckifValid(*tag.Value) {
							if validTeam.MatchString(*tag.Value) {
								for _, val := range meta[0].Values {
									if val.Value == *tag.Value {
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

						dbTags["environment"] = *tag.Value
						if !CheckifValid(*tag.Value) {
							if validEnv.MatchString(*tag.Value) {
								// log.Info("regex match done for ENV")
								for _, val := range meta[1].Values {
									// log.Info("Comparing " + val.Value + " and " + *tag.Value)
									if val.Value == *tag.Value {
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

						dbTags["repo"] = *tag.Value
						if !CheckifValid(*tag.Value) {
							for _, val := range meta[2].ValueRegex {
								if regexp.MustCompile(val).MatchString(*tag.Value) {
									repoPresent = true
									break
								}

							}

						}

						break
					}
				case "Name":
					{

						dbName = *tag.Value

					}
				default:
					{
						// log.Debug("Ignoring " + *tag.Key + " with value " + *tag.Value)
					}
				}

			}
			if !teamPresent {
				// log.Info("Team is not present")
				// dbTags["team"] = "none"
				dbTags["team"] = dbTags["team"] + "\nNon-Compliant tag value.\nPls ping #tagging-support "
				missingTagsInSet = true

			}
			if !environmentPresent {
				// log.Info("Environment is not present")
				// dbTags["environment"] = "none"
				dbTags["environment"] = dbTags["environment"] + "\nNon-Compliant tag value.\nPls use one of \nprod|perf|sand|stag|dev|labs"
				missingTagsInSet = true

			}
			if !repoPresent {
				// log.Info("Repo is not present")
				dbTags["repo"] = dbTags["repo"] + "\nNon-Compliant tag value.\nPls ping #tagging-support"
				missingTagsInSet = true

			}

			if missingTagsInSet {
				// log.Info("_______________________________|||" + *dbinstance.DBInstanceIdentifier + "has been added")
				dbTags["id"] = *dbinstance.DBInstanceIdentifier
				if dbName == "" {

					result[*dbinstance.DBInstanceIdentifier] = dbTags
				} else {
					result[*dbinstance.DBInstanceIdentifier] = dbTags

				}
			} else {
				log.Info("There are no missing tags for " + *dbinstance.DBInstanceArn)

			}
			missingTagsInSet = false //reset bool

		}

	}
	return result
}

func addAllInstanceNames(svc *rds.Client, result map[string]string) map[string]string {
	resp, err := svc.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		panic(err)
	}
	for _, dbinstance := range resp.DBInstances {
		result[*dbinstance.DbiResourceId] = *dbinstance.DBInstanceIdentifier
		if dbinstance.TagList != nil {
			for _, tag := range dbinstance.TagList {
				if *tag.Key == "Name" {
					result[*dbinstance.DbiResourceId] = *tag.Value
					break
				}
			}
		}
	}
	return result
}

/*
NO-TAG-NO-RUN : Stop instances with missing Team and Environment Tags
*/
func StopInstancesWithMissingTags(svc *rds.Client, rdsdata map[string]map[string]string) (map[string]map[string]string, error) {
	sc := comms.SlackClient{
		WebHookUrl: WEBHOOK_URL,
		Channel:    "CHANNEL_NAME",
	}
	resp, err := svc.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Panic(err)
	}

	for _, dbinstance := range resp.DBInstances {
		for index, value := range rdsdata {
			if *dbinstance.DBInstanceIdentifier == value["id"] {
				log.Info("---------------------------------")
				log.Info(*dbinstance.DBInstanceIdentifier + " would be stopped")
				if dbinstance.DBClusterIdentifier == nil {
					log.Info("No cluster")
					input := &rds.StopDBInstanceInput{
						DBInstanceIdentifier: aws.String(*dbinstance.DBInstanceIdentifier),
					}
					result, err := svc.StopDBInstance(context.TODO(), input)

					if err != nil {
						log.Error(err)
						sc.WebHookUrl = WEBHOOK_SRE
						sc.SendError("Error while doing StopDBInstance for "+*dbinstance.DBInstanceIdentifier+"\n"+err.Error(), ":rotating_light:")
						value["comments"] = "@SRE action required." + config.GetLocalTime()
						delete(rdsdata, index)

					} else {
						log.Info(*result.DBInstance.DBInstanceStatus)
						value["comments"] = "Instance stopped at " + config.GetLocalTime()
						sc.WebHookUrl = WEBHOOK_SRE
						sc.SendWarning(*dbinstance.DBInstanceIdentifier + " was stopped with non-compliant tags")
					}

				} else {
					log.Info("Cluster id " + *dbinstance.DBClusterIdentifier)
					result, err := svc.StopDBCluster(context.TODO(), &rds.StopDBClusterInput{
						DBClusterIdentifier: aws.String(*dbinstance.DBClusterIdentifier),
					})
					if err != nil {
						log.Error(err)
						value["comments"] = config.GetLocalTime() + " - @SRE action required."
						sc.WebHookUrl = WEBHOOK_SRE
						sc.SendError("Error while doing StopDBCluster for "+*dbinstance.DBClusterIdentifier+"\n"+err.Error(), ":rotating_light:")
						delete(rdsdata, index)

					} else {
						log.Info(*result.DBCluster.Status)
						value["comments"] = "Instance stopped at " + config.GetLocalTime()
						sc.WebHookUrl = WEBHOOK_SRE
						sc.SendWarning(*dbinstance.DBClusterIdentifier + " was stopped with non-compliant tags")
					}

				}

				// time.Sleep(2 * time.Second)
				log.Info("---------------------------------")

			} else {
				// xac log.Info("Skipping as names not match::" + *dbinstance.DBInstanceIdentifier + " and " + value["DBInstanceIdentifier"])

			}
		}
	}
	return rdsdata, err
}

/*
Update the tags for the resources mentioned on the xlsx
*/

func UpdateMandatoryTags(svc *rds.Client, rdsdata map[string]map[string]string) {

	resp, err := svc.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Panic(err)
	}
	for _, dbinstance := range resp.DBInstances {
		for _, value := range rdsdata {

			if *dbinstance.DBInstanceIdentifier == value["name"] {
				log.Info("---------------------------------")
				log.Info(*dbinstance.DBInstanceIdentifier + " and " + value["name"])
				input := &rds.AddTagsToResourceInput{
					ResourceName: aws.String(*dbinstance.DBInstanceArn),
					Tags: []types.Tag{
						{
							Key:   aws.String("Team"),
							Value: aws.String(value["team"]),
						}, {
							Key:   aws.String("Environment"),
							Value: aws.String(value["environment"]),
						},
						{
							Key:   aws.String("Repo"),
							Value: aws.String(value["repo"]),
						},
					},
				}
				result, err := svc.AddTagsToResource(context.TODO(), input)
				if err != nil {
					log.Panic(err.Error())
				}
				fmt.Print(result.ResultMetadata.Get(0))
				// time.Sleep(2 * time.Second)
				log.Info("---------------------------------")

			} else {
				// log.Info("Skipping as names not match::" + *dbinstance.DBInstanceIdentifier + " and " + value["name"])

			}
		}
	}

	return
}
