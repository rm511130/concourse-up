#!/bin/bash

set -eu

mkdir -p $GOPATH/src/github.com/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/engineerbetter/concourse-up
cd $GOPATH/src/github.com/engineerbetter/concourse-up

deployment="system-test-$RANDOM"

if [[ -n $ROOT_DOMAIN ]]; then
  domain="$deployment.$ROOT_DOMAIN"

  go run main.go deploy $deployment --domain $domain

  config=$(go run main.go info $deployment)

  username=$(echo $config | jq -r '.config.concourse_username')
  password=$(echo $config | jq -r '.config.concourse_password')
else
  go run main.go deploy $deployment

  config=$(go run main.go info $deployment)

  domain=$(echo $config | jq -r '.terraform.elb_dns_name.value')
  username=$(echo $config | jq -r '.config.concourse_username')
  password=$(echo $config | jq -r '.config.concourse_password')
fi

fly --target system-test login --insecure --concourse-url https://$domain --username $username --password $password
fly --target system-test workers

go run main.go --non-interactive destroy $deployment