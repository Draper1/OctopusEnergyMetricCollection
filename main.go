package main

import (
	"context"
	"encoding/json"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type OctopusConfig struct {
	OctoAccountNumber string
	OctoApiKey        string
	OctoElectricMpan  string
	OctoElectricSn    string
	OctoGasMprn       string
	OctoGasSn         string
	OctoGasCost       float32
	OctoElectricCost  float32
	InfluxdbUrl       string
	InfluxdbToken     string
	InfluxdbOrg       string
	InfluxdbBucket    string
	PageSize          int
	VolumeCorrection  float32
	CalorificValue    float32
	JoulesConversion  float32
}

type OctopusMeterConsumption struct {
	Count    int                             `json:"count"`
	Next     string                          `json:"next"`
	Previous string                          `json:"previous"`
	Results  []OctopusMeterConsumptionMetric `json:"results"`
}

type OctopusMeterConsumptionMetric struct {
	Consumption   float32 `json:"consumption"`
	IntervalStart string  `json:"interval_start"`
	IntervalEnd   string  `json:"interval_end"`
}

func processElectricMetricPoints(results []OctopusMeterConsumptionMetric, writeAPI api.WriteAPI) {
	for _, metric := range results {
		// Parse the date string
		t, err := time.Parse(time.RFC3339, metric.IntervalEnd)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Processing Electric Metrics from " + metric.IntervalEnd)

		// Create a new influxDB point
		point := influxdb2.NewPointWithMeasurement("electricity").AddTag("consumption", "electric").AddField(
			"consumption", metric.Consumption,
		).SetTime(t)

		// Create a new influxDB cost point
		costPoint := influxdb2.NewPointWithMeasurement("electricity_cost").AddTag("consumption", "electric").AddField(
			"consumption", metric.Consumption,
		).SetTime(t)

		writeAPI.WritePoint(point)
		writeAPI.WritePoint(costPoint)
	}
}

func processGasMetricPoints(results []OctopusMeterConsumptionMetric, writeAPI api.WriteAPI, config OctopusConfig) {
	for _, metric := range results {

		// Parse the date string
		t, err := time.Parse(time.RFC3339, metric.IntervalEnd)
		if err != nil {
			log.Fatal(err)
		}

		// Create a new influxDB point
		point := influxdb2.NewPointWithMeasurement("gas").AddTag("consumption", "gas").AddField(
			"consumption", metric.Consumption,
		).SetTime(t)

		// define Gas Cost Calculation
		var kilowatts = (metric.Consumption * config.VolumeCorrection * config.CalorificValue) / config.JoulesConversion

		// Create a new influxDB cost point
		kwhGasPoint := influxdb2.NewPointWithMeasurement("gaskwh").AddTag("consumption", "gas").AddField(
			"consumption_kwh", kilowatts,
		).SetTime(t)

		var gasCost = kilowatts * config.OctoGasCost / 100

		// Create a new influxDB cost point
		gasCostPoint := influxdb2.NewPointWithMeasurement("gas_cost").AddTag("consumption", "gas").AddField(
			"price", gasCost,
		).SetTime(t)

		writeAPI.WritePoint(point)
		writeAPI.WritePoint(kwhGasPoint)
		writeAPI.WritePoint(gasCostPoint)
	}
}

func main() {
	go forever()
	// block forever since we want this to run indefinitely
	select {}
}

func forever() {
	for {
		ctx := context.Background()

		// Open the config file
		file, err := os.Open("config.json")
		if err != nil {
			panic(err)
		}

		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				panic(err)
			}
		}(file)

		// Decode the file's contents into a OctopusConfig struct
		var config OctopusConfig
		if err := json.NewDecoder(file).Decode(&config); err != nil {
			panic(err)
		}

		// setup InfluxDB Connection
		influxClient := influxdb2.NewClient(config.InfluxdbUrl, config.InfluxdbToken)
		writeAPI := influxClient.WriteAPI(config.InfluxdbOrg, config.InfluxdbBucket)

		pageSizeStr := strconv.Itoa(config.PageSize)

		// Retrieve data from octopus API
		var electricResponse, gasResponse *http.Response
		httpClient := http.Client{}
		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"https://api.octopus.energy/v1/electricity-meter-points/"+config.OctoElectricMpan+"/meters/"+config.OctoElectricSn+"/consumption?page_size="+pageSizeStr,
			nil,
		)
		req.SetBasicAuth(config.OctoApiKey, "")
		electricResponse, _ = httpClient.Do(req)
		req, _ = http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"https://api.octopus.energy/v1/gas-meter-points/"+config.OctoGasMprn+"/meters/"+config.OctoGasSn+"/consumption?page_size="+pageSizeStr,
			nil,
		)
		req.SetBasicAuth(config.OctoApiKey, "")
		gasResponse, _ = httpClient.Do(req)

		// Unmarshal Gas JSON response
		var gasConsumption OctopusMeterConsumption
		gasBody, _ := ioutil.ReadAll(gasResponse.Body)

		err = json.Unmarshal(gasBody, &gasConsumption)
		if err != nil {
			log.Fatal(err)
		}

		// Unmarshal Electric JSON response
		var electricConsumption OctopusMeterConsumption
		electricBody, _ := ioutil.ReadAll(electricResponse.Body)
		err = json.Unmarshal(electricBody, &electricConsumption)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Electric Record Count", electricConsumption.Count)
		log.Println("Gas Record Count", gasConsumption.Count)

		// Add metrics to influxDB is there is data returned
		if electricConsumption.Count > 0 {
			processElectricMetricPoints(electricConsumption.Results, writeAPI)
		}

		// Add metrics to influxDB is there is data returned
		if gasConsumption.Count > 0 {
			processGasMetricPoints(gasConsumption.Results, writeAPI, config)
		}

		fmt.Println("Finished collection at", time.Now().UTC())
		fmt.Println("Sleeping for 1 day")
		time.Sleep(time.Hour * 24)
	}
}
