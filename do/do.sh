#!/bin/bash

go build -o bookgen
./bookgen $@
rm -rf bookgen

