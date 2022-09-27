#!/bin/bash
cd daemon
ENV="prod" go run . >> /var/log/bot.log 2>&1 &
