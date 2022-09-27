#!/bin/bash
cd daemon
ENV="prod" go run . >> /var/log/pyrous-gobot.log 2>&1 &
