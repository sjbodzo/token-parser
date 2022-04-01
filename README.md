# Overview

This application demos a simple ETL in Golang using channels. The scenario is receiving batched lists of coins by ID. 

## Running This Demo
To run the app locally using `cobra`:
```shell
go run main.go server
```

This application requires a local postgres database with the schema defined in `./deployments/docker-entrypoint-initb.d/init.sql`. 
For more information about the deployment, including how to run this locally in Kubernetes see the [deploy doc](./deployments/deploy.md).

## Handling Requests
Rules for processing inbound requests:
- Ignore coins we have already seen before
- Ignore coins that do not have valid IDs
- Store valid coins with their exchange data
