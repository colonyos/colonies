#!/bin/bash

colonies colony add --name dev --colonyid $COLONIES_COLONY_ID
colonies user add --name="test" --email="test@test.com" --phone="" --userid=$COLONIES_ID
