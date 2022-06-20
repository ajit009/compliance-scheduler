package config

import (
	"strings"
)


// Config holds the global configuration settings
type Config struct {
	Verbose        *bool
	OutputFile     *string
	OutputFormat   *string
	AppendToOutput *bool
	NameFile       *string
	AwsResource *string
	Profile *string
	Region  *string
	Environment *string
	Action *string
	
}

// GetOutputFormat returns the output format
func (config *Config) GetOutputFormat() string {
	return strings.ToLower(*config.OutputFormat)
}

// ShouldAppend returns if the output should append
func (config *Config) ShouldAppend() bool {
	return *config.AppendToOutput
}

// ShouldCombineAndAppend returns if the output should be combined
func (config *Config) ShouldCombineAndAppend() bool {
	if !config.ShouldAppend() {
		return false
	}
	if config.GetOutputFormat() == "html" {
		return false
	}
	return true
}
