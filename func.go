package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	_ "os"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/objectstorage"
	"github.com/Sugi275/serless_objectstorage_pre-auth/loglib"
	fdk "github.com/fnproject/fdk-go"
)

const (
	envBucketName = "OCI_BUCKETNAME"
	envSourceRegion = "OCI_SOURCE_REGION"
	envDestinationRegions = "OCI_DESTINATION_REGIONS"
	actionTypeCreate = "com.oraclecloud.objectstorage.createobject"
	actionTypeUpdate = "com.oraclecloud.objectstorage.updateobject"
	actionTypeDelete = "com.oraclecloud.objectstorage.deleteobject"
)

// EventsInput EventsInput
type EventsInput struct {
	CloudEventsVersion string      `json:"cloudEventsVersion"`
	EventID            string      `json:"eventID"`
	EventType          string      `json:"eventType"`
	Source             string      `json:"source"`
	EventTypeVersion   string      `json:"eventTypeVersion"`
	EventTime          time.Time   `json:"eventTime"`
	SchemaURL          interface{} `json:"schemaURL"`
	ContentType        string      `json:"contentType"`
	Extensions         struct {
		CompartmentID string `json:"compartmentId"`
	} `json:"extensions"`
	Data struct {
		CompartmentID      string `json:"compartmentId"`
		CompartmentName    string `json:"compartmentName"`
		ResourceName       string `json:"resourceName"`
		ResourceID         string `json:"resourceId"`
		AvailabilityDomain string `json:"availabilityDomain"`
		FreeFormTags       struct {
			Department string `json:"Department"`
		} `json:"freeFormTags"`
		DefinedTags struct {
			Operations struct {
				CostCenter string `json:"CostCenter"`
			} `json:"Operations"`
		} `json:"definedTags"`
		AdditionalDetails struct {
			Namespace        string `json:"namespace"`
			PublicAccessType string `json:"publicAccessType"`
			ETag             string `json:"eTag"`
		} `json:"additionalDetails"`
	} `json:"data"`
}

func main() {
	fdk.Handle(fdk.HandlerFunc(fnMain))

	// ------- local development ---------
	// reader := os.Stdin
	// writer := os.Stdout
	// fnMain(context.TODO(), reader, writer)
}

func fnMain(ctx context.Context, in io.Reader, out io.Writer) {
	// Events から受け取るパラメータ
	input := &EventsInput{}
	json.NewDecoder(in).Decode(input)
	outputJSON, _ := json.Marshal(&input)
	fmt.Println(string(outputJSON))

	loglib.InitSugar()
	defer loglib.Sugar.Sync()

	provider, err := auth.ResourcePrincipalConfigurationProvider()
	if err != nil {
		loglib.Sugar.Error(err)
		return
	}

	// provider := common.DefaultConfigProvider()

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
    if err != nil {
	 	loglib.Sugar.Error(err)
	 	return
	}

	client.SetRegion(string(common.RegionAPTokyo1)) 

	test := common.SDKTime{
		time.Now().Add(24 * time.Hour),
	}

	detail := objectstorage.CreatePreauthenticatedRequestDetails{
		Name: common.String("serverless_demo"),
		AccessType: objectstorage.CreatePreauthenticatedRequestDetailsAccessTypeAnyobjectwrite,
		TimeExpires: &test,
	}

	request := objectstorage.CreatePreauthenticatedRequestRequest{
		NamespaceName: common.String("orasejapan"),
		BucketName: common.String("serverless_movie"),
		CreatePreauthenticatedRequestDetails: detail,
	}

	response, err := client.CreatePreauthenticatedRequest(ctx, request)
	
	if err != nil {
		loglib.Sugar.Error(err)
		return
	}

	funcResponse := struct {
		RequestID string `json:"request_id"`
		AccsessURI string `json:"access_uri"`
	}{
		RequestID: *response.Id,
		AccsessURI: "https://objectstorage.ap-tokyo-1.oraclecloud.com" + *response.AccessUri,
	}

	json.NewEncoder(out).Encode(&funcResponse)
}
