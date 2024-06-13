# -----------------------------------
# local
# -----------------------------------
# load env file into this shell
set -a; source .env.dockerhub; set +a

ssh root@dhs.detabord.com /bin/bash << EOF
  # docker login
  docker login --username $DOCKER_HUB_USER --password-stdin <<< "$DOCKER_HUB_PERSONAL_ACCESS_TOKEN"

  # pull new images
  docker pull gradient0/dhs-gitea:latest

  # disable systemd service and stop
  systemctl restart sentinel

  # docker logout
  docker logout
EOF
