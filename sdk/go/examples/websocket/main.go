package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	wlteopenapi "github.com/svnwi/wlte-openapi-sdk/sdk/go"
	"github.com/svnwi/wlte-openapi-sdk/sdk/go/examples/common"
)

func main() {
	if err := common.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client, err := wlteopenapi.NewClient(wlteopenapi.ClientOptions{
		ClientID:     os.Getenv("WLTE_CLIENT_ID"),
		ClientSecret: os.Getenv("WLTE_CLIENT_SECRET"),
		BaseURL:      os.Getenv("WLTE_BASE_URL"),
	})
	if err != nil {
		log.Fatal(err)
	}

	session, err := client.WebSocket.Connect(ctx, wlteopenapi.WebSocketConnectOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	device, err := session.GetDeviceState(ctx, os.Getenv("WLTE_DEVICE_ID"))
	if err != nil {
		log.Fatal(err)
	}
	if err := common.PrintJSON(device); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-session.Done():
			if err := session.Err(); err != nil {
				log.Fatal(err)
			}
			return
		case event, ok := <-session.Events():
			if !ok {
				return
			}
			log.Printf("event topic=%s data=%s", event.Topic, event.Data)
		}
	}
}
