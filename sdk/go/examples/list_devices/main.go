package main

import (
	"context"
	"log"
	"os"

	wlteopenapi "github.com/svnwi/wlte-openapi-sdk/sdk/go"
	"github.com/svnwi/wlte-openapi-sdk/sdk/go/examples/common"
)

func main() {
	if err := common.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	client, err := wlteopenapi.NewClient(wlteopenapi.ClientOptions{
		ClientID:     os.Getenv("WLTE_CLIENT_ID"),
		ClientSecret: os.Getenv("WLTE_CLIENT_SECRET"),
		BaseURL:      os.Getenv("WLTE_BASE_URL"),
	})
	if err != nil {
		log.Fatal(err)
	}

	devices, err := client.Devices.List(context.Background(), wlteopenapi.ListDevicesOptions{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := common.PrintJSON(devices); err != nil {
		log.Fatal(err)
	}
}
