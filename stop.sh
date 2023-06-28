#!/bin/bash
set -e
set +x
cd microservices

kill -9 $(pgrep sai-btc) || echo "WARNING: BTC not running, or failed to kill processes"
kill -9 $(pgrep sai-bft) || echo "WARNING: BFT not running, or failed to kill processes"
kill -9 $(pgrep p2p_lin64.sh) || echo "WARNING: P2P proxy not running, or failed to kill processes"
kill -9 $(pgrep sai-p2p) || echo "WARNING: P2P proxy not running, or failed to kill processes"
kill -9 $(pgrep sai-vm1) || echo "WARNING: VM not running, or failed to kill processes"

docker-compose -f docker-compose-windows.yml down