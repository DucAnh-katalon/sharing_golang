#!/bin/bash

cd /Users/anh.dao/Desktop/courses/master-HCMUS/distributed_system/
# Accessing argument for loop range
node=$1
password="ducanh123zz"


echo "[SYS]:: Removing lab_2_$((node))"
rm -rf lab_2_$((node))
cp -r lab_2 lab_2_$((node))
echo "[SYS]:: Copied lab_2 to lab_2_$((node))"
echo "[SYS]:: Starting server on lab_2_$((node))"
cd lab_2_$((node))/main
echo $password | sudo -S go run main_server.go 
