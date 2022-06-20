package config

import "time"

func GetLocalTime() string {

	offSet, err := time.ParseDuration("+10.00h")
	if err != nil {
		panic(err)
	}
	t := time.Now().UTC().Add(offSet)
	return string(t.Format("2006-01-02 15:04:05"))
}
