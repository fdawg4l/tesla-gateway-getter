package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/fdawg4l/tesla-gateway-getter/pkg/build"
	"github.com/fdawg4l/tesla-gateway-getter/pkg/gateway"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Email        string `mapstructure: "TESLA_EMAIL"`
	Password     string `mapstructure: "TESLA_PASSWORD"`
	Gateway      string `mapstructure: "TESLA_GATEWAY"`
	InfluxHost   string `mapstructure: "TESLA_INFLUXHOST"`
	InfluxBucket string `mapstructure: "TESLA_INFLUXBUCKET"`
	InfluxOrg    string `mapstructure: "TESLA_INFLUXORG"`
	InfluxToken  string `mapstructure: "TESLA_INFLUXTOKEN"`
	Interval     uint   `mapstructure: "TESLA_INTERVAL"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config := new(Config)

	viper.SetEnvPrefix("TESLA")
	viper.BindEnv("gateway")
	viper.BindEnv("password")
	viper.BindEnv("email")
	viper.BindEnv("influxhost")
	viper.BindEnv("influxbucket")
	viper.BindEnv("influxorg")
	viper.BindEnv("influxtoken")
	viper.BindEnv("interval")
	viper.AutomaticEnv()

	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("error: %s", err.Error())
	}

	interval := time.Duration(config.Interval)
	if interval == 0 {
		interval = 30
	}

	log.Printf("%s -- build %s", os.Args[0], build.GitCommitID)
	log.Printf("Using influx host=%s bucket=%s every=%d", config.InfluxHost, config.InfluxBucket, interval)
	log.Printf("Using gateway host=%s", config.Gateway)

	// Create a new client using an InfluxDB server base URL and an authentication token
	influxClient := influxdb2.NewClient(config.InfluxHost, config.InfluxToken)
	defer influxClient.Close()

	// Use blocking write client for writes to desired bucket
	writeAPI := influxClient.WriteAPIBlocking(config.InfluxOrg, config.InfluxBucket)

	u, err := url.Parse(config.Gateway)
	if err != nil {
		log.Fatalf("error connecting to gateway: %s", err.Error())
	}
	log.Printf("Using gateway=%s", u.String())

	gatewayClient, err := gateway.NewClient(u, config.Email, config.Password)
	if err != nil {
		log.Fatalf("error connecting to gateway: %s", err.Error())
	}

	t := time.NewTicker(interval * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("exiting")
			return

		case <-t.C:
			agg, err := gatewayClient.Aggregates()
			if err != nil {
				log.Fatalf("error getting aggregates: %s", err.Error())
			}

			soe, err := gatewayClient.SOE()
			if err != nil {
				log.Fatalf("error getting soe: %s", err.Error())
			}

			// Create point using full params constructor
			// write point immediately
			p := influxdb2.NewPoint("http", nil, agg.Values, time.Now())
			if err := writeAPI.WritePoint(context.Background(), p); err != nil {
				log.Fatalf("error getting writing aggregates to influx %s", err.Error())
			}

			p = influxdb2.NewPoint("http", nil, map[string]interface{}{"percentage": soe.Percentage}, time.Now())
			if err := writeAPI.WritePoint(context.Background(), p); err != nil {
				log.Fatalf("error getting writing soe to influx %s", err.Error())
			}
		}
	}
}
