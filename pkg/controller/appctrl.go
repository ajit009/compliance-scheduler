package controller

import (
	"aws-compliance-scheduler/pkg/comms"
	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/confluence"
	"aws-compliance-scheduler/pkg/log"
	"aws-compliance-scheduler/pkg/services"
	"strconv"
	"strings"
)

const CHANNEL_NAME string = "tagging-support"
const WEB_HOOK_URL string = "https://hooks.slack.com/services/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"
const TEST_HOOK string = "https://hooks.slack.com/services/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"

type ControllerData struct {
	Resource          *string
	RemediationAction *string
}

/*
Collect all non compliant resources defined by the string pointer <em>resource</em>
and exetutes the action defined by the string pointer <em>action</em> and returns a
boolen result.
*/

func CollectNonCompliantResourcesAndTakeAction(resource *string, action *string, accountNumber string, meta config.compsMeta) {

	switch *resource {
	case "rds":
		{
			log.Info("appctrl::Controller::CollectNonCompliantResourcesAndTakeAction::RDS request. Action is " + *action)
			CollectNonCompliantRDSAndTakeAction(action, accountNumber, meta)
		}
	case "s3":
		{
			log.Info("appctrl::Controller::CollectNonCompliantResourcesAndTakeAction::S3 request. Action is " + *action)
			CollectNonCompliantS3AndTakeAction(action, accountNumber, meta)
		}
	case "dynamodb":
		{
			log.Info("appctrl::Controller::CollectNonCompliantResourcesAndTakeAction::DynamoDB request. Action is " + *action)
			CollectNonCompliantDynamoDbAndTakeAction(action, accountNumber, meta)
		}
	case "test":
		{
			log.Info("appctrl::Controller::CollectNonCompliantResourcesAndTakeAction::Test request. Action is " + *action)
			TestFeaturesOnLocal(action, accountNumber, meta)
		}
	}

}

func CollectNonCompliantRDSAndTakeAction(action *string, accountNumber string, meta config.compsMeta) {
	// var nonCompliantResourceList map[string]map[string]string
	nonCompliantResourceList := services.GetAllRdsResourcesMissingMandatoryTags(config.AssumedRoleForRDS(), meta)
	//get the AWS account details
	_, accTitle, _ := config.MapAccountToConfluenceResource(accountNumber, "rds")
	if *action == "stop" {
		log.Info("appctrl::Controller::CollectNonCompliantRDSAndTakeAction:: Stop instance requested")
		rdsdata, err := services.StopInstancesWithMissingTags(
			config.AssumedRoleForRDS(),
			nonCompliantResourceList,
		)

		if err != nil {

			log.Error(err)

		}
		//Report kills to slack

		SendSlackNotification(rdsdata, accTitle, "Following RDS instances were stopped today", "warning")

	} else {
		log.Info("appctrl::Controller::CollectNonCompliantRDSAndTakeAction:: " + *action + " instance requested")

	}
	confluence.UpdateConfluencePage(
		nonCompliantResourceList,
		accountNumber,
		"rds")

}

func CollectNonCompliantS3AndTakeAction(action *string, accountNumber string, meta config.compsMeta) {
	// var nonCompliantResourceList map[string]map[string]string
	nonCompliantResourceList := services.GetAllS3ResourcesMissingMandatoryTags(config.AssumedRoleForS3(), meta)
	if strings.ToLower(*action) == "block" {
		log.Info("appctrl::Controller::CollectNonCompliantS3AndTakeAction:: Block bucket requested")
		services.BlockAccessToNonComliantS3Buckets(
			config.AssumedRoleForS3(),
			nonCompliantResourceList,
		)
	}

	confluence.UpdateConfluencePage(
		nonCompliantResourceList,
		accountNumber,
		"s3")
}

func CollectNonCompliantDynamoDbAndTakeAction(action *string, accountNumber string, meta config.compsMeta) {
	// var nonCompliantResourceList map[string]map[string]string
	nonCompliantResourceList := services.GetAllDynamoResourcesMissingMandatoryTags(config.AssumedRoleForDynamoDb(), meta)
	if strings.ToLower(*action) == "block" {
		log.Info("appctrl::Controller::CollectNonCompliantDynamoDbAndTakeAction:: Block table requested")
		services.BlockAccessToNonComliantDynamoDbTables(
			config.AssumedRoleForDynamoDb(),
			nonCompliantResourceList,
		)
	}

	confluence.UpdateConfluencePage(
		nonCompliantResourceList,
		accountNumber,
		"dynamo")
}

func TestFeaturesOnLocal(action *string, accountNumber string, meta config.compsMeta) {

	log.Info("appctrl::Controller::TestFeaturesOnLocal::Running for " + *action)
	nonCompliantResourceList := services.GetAllRdsResourcesMissingMandatoryTags(config.AssumedRoleForRDS(), meta)
	if strings.ToLower(*action) == "slacktests" {
		SendSlackNotification(nonCompliantResourceList, accountNumber, "Generic", "warning")
	}
}

func SendSlackNotification(data map[string]map[string]string, acc string, message string, msgType string) {
	log.Info("appctrl::Controller::TestFeaturesOnLocal::Invoke slack")

	sc := comms.SlackClient{
		WebHookUrl: WEB_HOOK_URL,
		Channel:    CHANNEL_NAME,
	}
	totalResources := len(data)
	if totalResources != 0 {
		messageToBeSent := "Following " + strconv.Itoa(totalResources) + " resources were stopped with missing/non-compliant tags. "
		for index, _ := range data {
			messageToBeSent = messageToBeSent + "\n # " + index
		}
		sr := comms.SlackJobNotification{
			Text:      "RDS tag compliance in " + acc,
			Details:   messageToBeSent,
			Color:     msgType,
			IconEmoji: ":octagonal_sign",
		}
		err := sc.SendJobNotification(sr)
		if err != nil {
			log.Error(err)
		}
	}

}
