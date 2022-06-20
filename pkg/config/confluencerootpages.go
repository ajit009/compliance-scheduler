package config

import "strings"

func GetRootIdForResource(res string) string {
	switch strings.ToLower(res) {
	case "rds":
		return "xxxxxxx"
	case "dynamo":
		return "xxxxxxx"
	case "ec2":
		return "xxxxxx"
	case "s3":
		return "xxxxxxx"
	case "redis":
		return "xxxxxxxx"
	default:
		return ""
	}

}
