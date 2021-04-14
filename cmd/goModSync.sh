#!/bin/bash

(cd main && go mod vendor)

(cd daemon && go mod vendor)

(cd cli && go mod vendor)