package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/0xste/validator-stats/internal/validator"
	"github.com/0xste/validator-stats/pkg/beacon"
	"github.com/0xste/validator-stats/pkg/prom"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	// common config
	configOutFile   = "OUT_FILE"
	configInfoFile  = "INFO_FILE"
	configTimeRange = "TIME_RANGE"
	configMode      = "RUN_MODE"

	// mode == file
	configFile = "CONFIG_FILE"

	// mode == prom
	configPromUser     = "PROM_USER"
	configPromPassword = "PROM_PASSWORD"
	configPromEndpoint = "PROM_ENDPOINT"
)

func init() {
	viper.SetDefault(configMode, "prom") // or "prom"
	viper.SetDefault(configFile, "./pubkeys.yml")
	viper.SetDefault(configOutFile, "./out.csv")
	viper.SetDefault(configInfoFile, "./info.csv")
	viper.SetDefault(configTimeRange, time.Hour*24*90)
}

func main() {
	processStart := time.Now()

	viper.AutomaticEnv()
	file, err := os.ReadFile(viper.GetString(configFile))
	if err != nil {
		log.Fatal(err)
	}

	var promClient *prom.Client
	var pubkeys []string
	switch viper.GetString(configMode) {
	case "file":
		if viper.GetString(configFile) == "" {
			log.Fatal("missing file config")
		}
		if err := yaml.Unmarshal(file, &pubkeys); err != nil {
			log.Fatal(err)
		}
	case "prom":
		if viper.GetString(configPromEndpoint) == "" || viper.GetString(configPromUser) == "" || viper.GetString(configPromPassword) == "" {
			log.Fatal("missing prom config")
		}
		promClient, err = prom.New(
			prom.WithAddress(viper.GetString(configPromEndpoint)),
			prom.WithBasicAuth(viper.GetString(configPromUser), viper.GetString(configPromPassword)),
		)
		if err != nil {
			log.Fatal(err)
		}
		pubkeys, err = promClient.GetValidatorPubkeys(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("there are %d validators to check\n", len(pubkeys))

	beaconClient := beacon.NewClient(http.DefaultClient, "", 0, 0)

	client := validator.NewClient(beaconClient, promClient)
	if err := run(client, pubkeys, processStart); err != nil {
		log.Fatal(err)
	}
}

func run(client *validator.Client, pubkeys []string, processStart time.Time) error {
	start := time.Now()
	log.Printf("retrieving pubkeys took %s\n", time.Since(processStart))

	client.GetEstimatedDuration(len(pubkeys))

	err := writeValidators(client, pubkeys)
	log.Printf("write took %s\n", time.Since(start))
	return err
}

func writeValidators(client *validator.Client, pubkeys []string) error {
	// manage outfile
	outFile, err := os.Create(viper.GetString(configOutFile))
	if err != nil {
		return errors.Wrap(err, "failed to create out file")
	}
	defer outFile.Close()
	outWriter := csv.NewWriter(outFile)
	defer outWriter.Flush()

	// write headers
	err = outWriter.Write([]string{"pubkey", "issue_type", "count", "timestamp", "status", "withdrawal_credentials"})
	if err != nil {
		return err
	}
	outWriter.Flush()

	// manage info file
	infoFile, err := os.Create(viper.GetString(configInfoFile))
	if err != nil {
		return errors.Wrap(err, "failed to create info file")
	}
	defer infoFile.Close()
	infoWriter := csv.NewWriter(infoFile)
	defer infoWriter.Flush()
	err = infoWriter.Write([]string{"pubkey", "status", "withdrawal", "slashed", "name", "index", "timestamp"})
	if err != nil {
		return err
	}
	infoWriter.Flush()

	// make a request and immediately write to file
	lookback := -viper.GetDuration(configTimeRange)
	for _, pubkey := range pubkeys {
		var lines [][]string
		health, err := client.GetValidatorHealth(pubkey, lookback)
		if err != nil {
			log.Printf("skipping %s\n", pubkey)
			continue
		}
		info := health.Info.Data

		err = infoWriter.Write([]string{info.Pubkey, info.Status, info.Withdrawalcredentials, strconv.FormatBool(info.Slashed), info.Name, strconv.Itoa(info.Validatorindex), time.Now().String()})
		if err != nil {
			return err
		}
		infoWriter.Flush()

		// write health conditions file
		for _, conditions := range health.Conditions {
			for _, condition := range conditions {
				lines = append(lines, []string{
					health.Info.Data.Pubkey,
					string(condition.IssueType),
					fmt.Sprintf("%d",
						condition.Count,
					), condition.Day.String(),
					health.Info.Data.Status,
					health.Info.Data.Withdrawalcredentials,
				})
			}
			if err := outWriter.WriteAll(lines); err != nil {
				return errors.Wrap(err, "error writing record to file")
			}

		}

	}
	return nil
}
