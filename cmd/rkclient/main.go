package main

import (
	"fmt"

	"github.com/zinic/forculus/apitools"
	"github.com/zinic/forculus/recordkeeper/model"
	"github.com/zinic/forculus/recordkeeper/rkapi"
)

func main() {
	eventRecord := model.EventRecord{
		StorageTarget: "aws_s3",
		StorageKey:    "event-12345.tar.gz",
		AccessToken:   "access_token",
		Tags: map[string]string{
			"Name": "Some Name",
		},
	}

	credentials := rkapi.Credentials{
		Username: "test",
		Password: "test",
	}

	endpoint := apitools.NewEndpoint("http", "localhost", 8080, "")

	client := rkapi.NewClient(credentials, endpoint)
	if newRecord, err := client.CreateEventRecord(eventRecord); err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("New record created - ID:%d\n", newRecord.ID)
	}
}
