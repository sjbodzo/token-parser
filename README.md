# Overview

This application demoes a simple ETL in Golang using channels. The scenario is receiving batched lists of coins by ID. 

## Running This Demo
TODO

## Architecture
TODO

## Handling Requests
Rules for processing inbound requests:
- Ignore coins we have already seen before
- Ignore coins that do not have valid IDs
- Store valid coins with their exchange data

## Asynchronicity 
All heavy processing of requests happens outside of 
the main response loop in the server handler. 
This ensures client responses are received as quickly
as possible.


To handle the use case of request limiting on
backend APIs, a buffered channel is used as well. 

This means the flow is as follows: 
- A channel is spun up per request by the server
in response to a client request
- That channel marshals and loads the data onto a
raw channel
- The raw channel feeds a buffered channel that implements
throttling to keep its consumer below rate limiting thresholds