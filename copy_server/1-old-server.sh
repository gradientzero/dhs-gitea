#!/bin/bash
docker_container="5f651399aab2"

docker cp ${docker_container}:/data/git /home/user/git
docker cp ${docker_container}:/data/gitea/avatars /home/user/avatars

