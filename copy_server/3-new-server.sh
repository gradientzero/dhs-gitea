#!/bin/bash
docker_container=$1
destination_tenant=$2

# Copy mysql docker volume, probably need to be root to access /var/lib/docker/volumes
cp -r /home/user/db-data/_data /var/lib/docker/volumes/${destination_tenant}_db-data/_data

# Copy /data/git/respositories and /data/gitea/avatars
docker cp /home/user/git/repositories ${docker_container}:/data/git/
docker cp /home/user/avatars ${docker_container}:/data/gitea/

# Change owner from root:root to git:git
docker exec ${docker_container} chown -R git:git /data/git/repositories
docker exec ${docker_container} chown -R git:git /data/gitea/avatars
