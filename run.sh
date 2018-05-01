#!/bin/bash

set -e

make

./prometurbo -stderrthreshold=FATAL -log_dir=./log -v=4

