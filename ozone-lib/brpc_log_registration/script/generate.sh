#!/bin/bash

mkdir -p log_registration_pb

protoc --go_out=./log_registration_pb --go_opt=paths=source_relative \
    --go-grpc_out=./log_registration_pb --go-grpc_opt=paths=source_relative \
    log_registration.proto