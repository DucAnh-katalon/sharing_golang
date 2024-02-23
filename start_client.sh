#!/bin/bash

cd /Users/anh.dao/Desktop/courses/master-HCMUS/distributed_system/lab_2/main
# Accessing argument for loop range
password="ducanh123zz"
echo $password | sudo -S go run main_client.go 
