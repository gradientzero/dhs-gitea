# -----------------------------------
# local
# -----------------------------------
# load env file into this shell
set -a; source .env.dockerhub; set +a

echo "-- login into dockerhub using .env.dockerhub"
docker logout $DOCKER_HUB_REGISTRY
docker login --username $DOCKER_HUB_USER --password-stdin <<< "$DOCKER_HUB_PERSONAL_ACCESS_TOKEN" $DOCKER_HUB_REGISTRY

# rebuild docker image
VERSION=$(<VERSION)
echo "-- rebuilding docker image, version $VERSION.."

docker build \
    --platform linux/amd64 \
    --build-arg GITEA_VERSION=main \
    -t gradient0/dhs-gitea:$VERSION \
    -t gradient0/dhs-gitea:latest \
    .

echo "-- pushing docker image, ensure you are logged in.. (docker login, pw see 1password)"
docker push gradient0/dhs-gitea:latest
docker push gradient0/dhs-gitea:$VERSION

# docker logout
docker logout

