#!/usr/bin/env bash

CHECK_THIS=main.go

echo -n "Strict check formatting... "
require_formatting=$(go run mvdan.cc/gofumpt -l $CHECK_THIS)

if [[ -n "${require_formatting}" ]]; then
  echo "FAILED"
  echo "${require_formatting}"
  exit 1
else
  echo "OK"
  exit 0
fi
