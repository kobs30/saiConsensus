#!/bin/bash
set -e
set +x
cd microservices

killall p2p_lin64
killall sai-p2p
killall sai-btc
killall sai-vm1
killall sai-bft

docker-compose -f docker-compose-windows.yml down