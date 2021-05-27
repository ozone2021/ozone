#!/bin/sh

(cd cli && go mod vendor)

(cd main && go mod vendor)

(cd daemon && go mod vendor)

