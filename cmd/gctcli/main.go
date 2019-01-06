package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/thrasher-/gocryptotrader/core"
	"github.com/thrasher-/gocryptotrader/gctrpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

const (
	defaultHost = ":4444"
)

func jsonOutput(in interface{}) {
	j, err := json.MarshalIndent(in, "", " ")
	if err != nil {
		return
	}
	fmt.Print(string(j))
}

func getExchanges(c *cli.Context) error {
	conn, err := grpc.Dial(defaultHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect to gRPC server. Err: %s", err)
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.GetExchanges(context.Background(),
		&gctrpc.GetExchangesRequest{
			Enabled: c.IsSet("enabled"),
		},
	)

	if err != nil {
		return err
	}

	jsonOutput(result)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gctcli"
	app.Version = core.Version(true)
	app.Usage = "command line interface for managing the gocryptotrader daemon"
	app.Commands = []cli.Command{
		{
			Name:      "getexchanges",
			Usage:     "gets a list of enabled or available exchanges",
			ArgsUsage: "<enabled>",
			Action:    getExchanges,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "enabled",
					Usage: "whether to list enabled exchanges or not",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
