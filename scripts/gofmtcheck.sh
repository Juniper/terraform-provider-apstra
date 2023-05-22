#!/usr/bin/env bash

echo -n "Check formatting... "
require_formatting=$(gofmt -s -l .)

if [[ -n "${require_formatting}" ]]; then
  echo "FAILED"
  echo "${require_formatting}"
  exit 1
else
  echo "OK"
  exit 0
fi
