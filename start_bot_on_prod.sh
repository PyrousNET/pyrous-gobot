#!/bin/bash
cd /srv/pyrous-gobot/daemon
ENV="prod" go run . >> /var/log/pyrous-gobot.log 2>&1
