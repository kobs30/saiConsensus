#!/bin/bash

while true
do
  echo run "$@"
  "$@"
  sleep 1
done