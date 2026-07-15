package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	execution, err := client.Relays.Set(context.Background(), os.Getenv("WLTE_DEVICE_ID"), wlteopenapi.RelaySetOptions{
		Index:          1,
		On:             true,
		IdempotencyKey: fmt.Sprintf("example-%d", time.Now().UnixNano()),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := common.PrintJSON(execution); err != nil {
		log.Fatal(err)
	}
}
