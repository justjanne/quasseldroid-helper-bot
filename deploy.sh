#!/bin/sh
IMAGE=k8r.eu/justjanne/quasseldroid-helper-bot
TAGS=$(git describe --always --tags HEAD)
DEPLOYMENT=quasseldroid-helper-bot
POD=quasseldroid-helper-bot

kubectl set image deployment/$DEPLOYMENT $POD=$IMAGE:$TAGS