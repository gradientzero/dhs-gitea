#!/bin/bash
docker_container="bfd6e62fa2d0"

# Copy mysql docker volume
cp -r /home/user/dhcs/usb_db-data/_data /var/lib/docker/volumes/dhcs_db-data/_data

# Copy /data/gi/respositories and /data/gitea/avatars
docker cp /home/user/dhcs/git/repositories ${docker_container}:/data/git/
docker cp /home/user/dhcs/avatars ${docker_container}:/data/gitea/

# Change owner from root:root to git:git
docker exec ${docker_container} chown -R git:git /data/git/repositories
docker exec ${docker_container} chown -R git:git /data/gitea/avatars
