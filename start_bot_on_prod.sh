#!/bin/bash
cd /srv/purous-gobot/daemon
ENV="prod" go run . >> /var/log/pyrous-gobot.log 2>&1
