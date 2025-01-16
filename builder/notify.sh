#!/bin/sh

. /$HOME/lib/cloudfare_webhook.env

curl -s -X POST "https://api.cloudflare.com/client/v4/pages/webhooks/deploy_hooks/$CF_WEBHOOKURL"