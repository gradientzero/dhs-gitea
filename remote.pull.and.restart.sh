# -----------------------------------
# local
# -----------------------------------
# load env file into this shell
set -a; source .env.dockerhub; set +a

#ssh root@usb.detabord.com /bin/bash << EOF
ssh root@sandbox.gradient0.com /bin/bash << EOF
  # docker login
  #d ocker login --username $DOCKER_HUB_USER --password-stdin <<< "$DOCKER_HUB_PERSONAL_ACCESS_TOKEN"

  # pull new images
  docker pull gradient0/dhs-gitea:latest

  # disable systemd service and stop
  systemctl restart sandbox

  # docker logout
  # docker logout
EOF
