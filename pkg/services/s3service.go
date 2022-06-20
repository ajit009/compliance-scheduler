package services

import (
	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/log"
	"context"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func CheckS3NoneValues(tagValue string) bool {
	return (strings.ToLower(tagValue) == "none") || (tagValue == "") || (tagValue == "NotDefined")
}
func GetAllS3ResourcesMissingMandatoryTags(svc *s3.Client, meta config.compsMeta) map[string]map[string]string {
	result := make(map[string]map[string]string)
	bucketList, err := svc.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	validTeam := regexp.MustCompile(meta[0].ValueRegex[0])
	log.Info("Team validation rule: ")
	log.Info(validTeam)
	validEnv := regexp.MustCompile(meta[1].ValueRegex[0])
	if err != nil {
		panic("GetAllS3ResourcesMissingMandatoryTags failed")
	}
	for _, bucket := range bucketList.Buckets {
		resp, err := svc.GetBucketTagging(context.TODO(), &s3.GetBucketTaggingInput{
			Bucket: aws.String(*bucket.Name),
		})
		if err != nil {
			log.Error(*bucket.Name + ":::" + err.Error())
		} else {
			log.Info("--------------------------------------------")
			log.Info("Working on bucket " + *bucket.Name)
			teamPresent := false
			environmentPresent := false
			repoPresent := false
			missingTagsInSet := false
			s3Name := ""
			s3Tags := make(map[string]string)
			for _, tag := range resp.TagSet {

				if resp.TagSet != nil {
					switch *tag.Key {
					case "Team":
						{
							// log.Info("Team :" + *tag.Value)
							s3Tags["team"] = *tag.Value
							if !CheckS3NoneValues(*tag.Value) {
								if validTeam.MatchString(*tag.Value) {
									// log.Info("Regex match done")
									for _, val := range meta[0].Values {
										// log.Info("Comparing " + val.Value + " with " + *tag.Value)
										if val.Value == *tag.Value {
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
							// log.Info("Environment :" + *tag.Value)
							s3Tags["environment"] = *tag.Value
							if !CheckS3NoneValues(*tag.Value) {
								if validEnv.MatchString(*tag.Value) {
									// log.Info("Regex match done")
									// log.Info(validEnv)
									for _, val := range meta[1].Values {
										// log.Info("Comparing " + val.Value + " with " + *tag.Value)
										if val.Value == *tag.Value {
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
							// log.Info("Repo :" + *tag.Value)
							s3Tags["repo"] = *tag.Value
							if !CheckS3NoneValues(*tag.Value) {
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

							s3Name = *tag.Value
							break
						}
					default:
						{
							// fmt.Println("Ignoring " + *tag.Key + " with value " + *tag.Value)
						}
					}

				}

			}
			if !teamPresent {
				// s3Tags["team"] = "none"
				// log.Info(" team is not as expected")
				missingTagsInSet = true

			}
			if !environmentPresent {
				// log.Info(" Environment is not as expected")
				// s3Tags["environment"] = "none"
				missingTagsInSet = true

			}
			if !repoPresent {
				// log.Info(" Repo is not as expected")
				// s3Tags["repo"] = "none"
				missingTagsInSet = true

			}

			if missingTagsInSet {
				// log.Info("adding" + *bucket.Name)
				// log.Info(teamPresent)
				// log.Info(environmentPresent)
				// log.Info(repoPresent)
				if s3Name == "" {
					result[*bucket.Name] = s3Tags
				} else {
					result[s3Name] = s3Tags
				}
			} else {
				log.Info("There are no missing tags for " + *bucket.Name)
			}
			missingTagsInSet = false //reset bool

		}

	}
	return result
}

func BlockAccessToNonComliantS3Buckets(svc *s3.Client, s3data map[string]map[string]string) {

}

/*
Updates the tags for the resources mentioned on the xlsx
*/

func UpdateMandatoryTagsOnS3(svc *s3.Client, s3data map[string]map[string]string) {

	for _, bucket := range s3data {
		log.Info("services.UpdateMandatoryTagsOnS3::working on " + bucket["name"])
		log.Info("services.UpdateMandatoryTagsOnS3::Environment is " + bucket["environment"])

		resp, err := svc.GetBucketTagging(context.TODO(), &s3.GetBucketTaggingInput{
			Bucket: aws.String(bucket["name"]),
		})

		if err != nil {

			log.Error(err.Error())
		} else {
			tagSt := resp.TagSet
			foundTeam := false
			foundEnv := false
			foundRepo := false
			for _, tag := range tagSt {
				if *tag.Key == "Team" {
					*tag.Value = bucket["team"]
					foundTeam = true
				}
				if *tag.Key == "Environment" {
					*tag.Value = bucket["environment"]
					foundEnv = true
				}
				if *tag.Key == "Repo" {
					*tag.Value = bucket["repo"]
					foundRepo = true
				}

			}
			if !foundTeam {
				tagSt = append(tagSt, types.Tag{Key: aws.String("Team"), Value: aws.String(bucket["team"])})
			}
			if !foundEnv {
				tagSt = append(tagSt, types.Tag{Key: aws.String("Environment"), Value: aws.String(bucket["environment"])})
			}
			if !foundRepo {
				tagSt = append(tagSt, types.Tag{Key: aws.String("Repo"), Value: aws.String(bucket["repo"])})
			}
			for _, tag := range tagSt {
				log.Info("Updated Tag" + *tag.Key + "=" + *tag.Value)
			}

			delCo, err := svc.DeleteBucketTagging(context.TODO(), &s3.DeleteBucketTaggingInput{
				Bucket: aws.String(bucket["name"]),
			})

			if err != nil {
				log.Error(err.Error())
			} else {
				log.Debug(delCo)
			}

			for _, tag := range tagSt {
				log.Info("services.UpdateMandatoryTagsOnS3:: Updated Tag: " + *tag.Key + "=" + *tag.Value)
			}
			val, err := svc.PutBucketTagging(context.TODO(), &s3.PutBucketTaggingInput{

				Bucket: aws.String(bucket["name"]),
				Tagging: &types.Tagging{
					TagSet: tagSt,
				},
			})
			if err != nil {
				log.Error(err.Error())
			} else {
				_ = val
			}

		}
		log.Info("--------------------------------------------")
	}

	return
}
