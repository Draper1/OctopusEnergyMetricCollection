# Octopus Energy Collection

This project is a rewrite of the [octopusenergy-consumption-metrics](https://github.com/ainsey11/octopusenergy-consumption-metrics) project but written into Go.

This project can be deployed via docker and simply needs a config file to be created to start.

Octopus API documentation can be found at [here](https://developer.octopus.energy/docs/api/#api-end-points).

This requires InfluxDB 2.x in order to work.

## Setup Steps

Add a new file into the base directory called `config.json` and add the following keys replacing the example values with information from InfluxDB and Octopus:

Note, you can get your Octopus Energy API and Meter information at the [here](https://octopus.energy/dashboard/developer).

```json
{
  "OctoApiKey": "sk_live_khasjkhasiuywqh",
  "OctoElectricMpan": "159105721782",
  "OctoElectricSn": "Z11K91827",
  "OctoGasMprn": "13761827",
  "OctoGasSn": "G9N129812J8",
  "OctoGasCost": 3.27,
  "OctoElectricCost": 18.78,
  "InfluxdbUrl": "http://localhost:8086",
  "InfluxdbToken": "yRhHOS8BakoM-Yaskjy7asd7qw906R2SR4Rqe-Wkg6CiQc287qwekjqwn878KAsD5dErA::",
  "InfluxdbOrg": "home",
  "InfluxdbBucket": "octopus",
  "LoopTime": 600,
  "PageSize": 100,
  "VolumeCorrection": 1.02264,
  "CalorificValue": 37.5,
  "JoulesConversion": 3.6
}
```

## Deploy Steps

To deploy, build using Docker:

```bash
docker build
```
