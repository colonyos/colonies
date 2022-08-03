#!/bin/bash

for i in {1..200}
do
  colonies workflow submit --spec ../examples/workflow.json --insecure
done
