package main

import (
	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/log"

	"aws-compliance-scheduler/pkg/controller"

	"github.com/aws/aws-sdk-go-v2/service/sts"

	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"runtime"

	"github.com/spf13/viper"
)

var settings = new(config.Config)

func main() {

	//read from config file if it exists
	initConfig()
	flag.Parse()
	var stsClient *sts.Client

	if runtime.GOOS == "darwin" {
		log.Debug("Tagger::Main::running on darwin")
		awsConfig := config.DefaultAwsConfig(*settings)
		stsClient = awsConfig.StsClient()
	} else {
		log.Debug("Tagger::Main::running on non-darwin OS")
		awsConfig, _ := config.StsAssumedConfig(*settings)
		stsClient = awsConfig.StsClient()
	}

	accMeta, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})

	if err != nil {
		log.Error("Error::conig.Tagger::Main::" + err.Error())

	}
	accountNumber := *accMeta.Account
	log.Debug("Using account :" + accountNumber + "\n")
	ajlocalmeta := config.ReadTagDataFromYaml()
	controller.CollectNonCompliantResourcesAndTakeAction(settings.AwsResource, settings.Action, accountNumber, ajlocalmeta)

}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func init() {
	settings.Verbose = flag.Bool("verbose", false, "Give verbose output")
	settings.OutputFile = flag.String("file", "", "Optional file to save the output to")
	settings.OutputFormat = flag.String("output", "json", "Format for the output, currently supported are csv, json, html")
	settings.AppendToOutput = flag.Bool("append", false, "Add to the provided output file instead of replacing it")
	settings.NameFile = flag.String("namefile", "", "Use this file to provide names")
	settings.Profile = flag.String("profile", "default", "Use a specific profile")
	settings.Region = flag.String("region", "ap-southeast-2", "Use a specific region")
	settings.AwsResource = flag.String("resource", "rds", "AWS resource on which scheduler has run on")
	settings.Environment = flag.String("environment", "labs", "ZIP aws environment to use for the run")
	settings.Action = flag.String("action", "report", "Action to be taken on non compliant resources.\n Options : \n\treport - Reports resources on Confluence.\n\tstop - Stops a running resource [ in case of s3, this would result in removing access\n\t\ttothe bucket]")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".comps") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}

func getName(id string) string {
	if *settings.NameFile != "" {
		nameFile, err := ioutil.ReadFile(*settings.NameFile)
		if err != nil {
			panic(err)
		}
		values := make(map[string]string)
		err = json.Unmarshal(nameFile, &values)
		if err != nil {
			panic(err)
		}
		if val, ok := values[id]; ok {
			return val
		}
	}
	return id
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
