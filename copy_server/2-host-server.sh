#!/bin/bash
origin_server=$1
destination_server=$2
origin_tenant=$3

# Copy files from origin server, probably need to be root to access /var/lib/docker/volumes
scp -r ${origin_server}:/var/lib/docker/volumes/${origin_tenant}_db-data/ ~/db-data
scp -r ${origin_server}:/home/user/git/ ~/git
scp -r ${origin_server}:/home/user/avatars/ ~/avatars

# Paste files to destination server
scp -r ~/db-data ${destination_server}:/home/user/db-data
scp -r ~/git ${destination_server}:/home/user/git
scp -r ~/avatars ${destination_server}:/home/user/avatars
