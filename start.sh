#!/bin/bash
set -e
set +x
cd microservices

dir=$(pwd)
p2p_d="saiP2p/build"
p2p_f="p2p_lin64"
p2p_fs="p2p_lin64.sh"
p2pp_d="saiP2pProxy/build"
p2pp_f="sai-p2p"
btc_d="saiBtc/build"
btc_f="sai-btc"
vm_d="saiVM1/build"
vm_f="sai-vm1"
bft_d="saiBft/build"
bft_f="sai-bft"

echo "Starting the node"
docker-compose -f docker-compose.yml up -d

echo "Starting the btc"
if [ -f "$btc_d/$btc_f" ]
then
  cd $btc_d
  chmod +x $btc_f
  ./$btc_f 1>"$btc_f"".log" 2>&1 &
  cd "$dir"
  echo "Done"
fi

echo "Starting the p2p"
if [ -f "$p2p_d/$p2p_f" ]
then
  cd $p2p_d
  chmod +x $p2p_fs
  chmod +x $p2p_f
  ./$p2p_fs $p2p_f 1>"$p2p_f"".log" 2>&1 &
  cd "$dir"
  echo "Done"
fi

echo "Starting the p2pp"
if [ -f "$p2pp_d/$p2pp_f" ]
then
  cd $p2pp_d
  chmod +x $p2pp_f
  ./$p2pp_f 1>"$p2pp_f"".log" 2>&1 &
  cd "$dir"
  echo "Done"
fi

echo "Starting the vm"
if [ -f "$vm_d/$vm_f" ]
then
  cd $vm_d
  chmod +x $vm_f
  ./$vm_f start 1>"$vm_f"".log" 2>&1 &
  cd "$dir"
  echo "Done"
fi

echo "Starting the bft"
if [ -f "$bft_d/$bft_f" ]
then
  cd $bft_d
  ./$bft_f start 1>"$bft_f"".log" 2>&1 &
  cd "$dir"
  echo "Done"
fi
