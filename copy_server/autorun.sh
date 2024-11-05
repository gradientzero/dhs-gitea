#!/bin/bash

origin_server=$1 #ex: root@0.0.0.0
origin_docker_container=$2 #ex: 5f651399aab2
origin_tenant=$3 #ex: usb
destination_server=$4 #ex: root@0.0.0.0   
destination_docker_container=$5 #ex: bfd6e62fa2d0
destination_tenant=$6 #ex: dhcs

ssh ${origin_server} ./1-old-server.sh ${origin_docker_container}
./2-host-server.sh ${origin_server} ${destination_server} ${origin_tenant}
ssh ${destination_server} ./3-new-server.sh ${destination_docker_container} ${destination_tenant}

# Clean all files
ssh ${origin_server} "rm -rf /home/user/git /home/user/avatars"
rm -rf ~/git ~/avatars ~/db-data
ssh ${destination_server} "rm -rf /home/user/git /home/user/avatars /home/user/db-data"