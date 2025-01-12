#!/bin/sh

. $HOME/lib/netlify_webhook.env

curl -s  -v -X POST -d {} $NETLIFY_WOOKURL | nobs
