# Build stage
FROM golang:1.22 AS build-env

ARG GOPROXY
ENV GOPROXY=${GOPROXY:-direct}

ARG GITEA_VERSION
ARG TAGS="sqlite sqlite_unlock_notify"
ENV TAGS="bindata timetzdata $TAGS"
ARG CGO_EXTRA_CFLAGS

# Install build dependencies
RUN apt-get update && \
    apt-get install -y build-essential git nodejs npm

# Setup repository
COPY . ${GOPATH}/src/code.gitea.io/gitea
WORKDIR ${GOPATH}/src/code.gitea.io/gitea

# Checkout version if set
RUN if [ -n "${GITEA_VERSION}" ]; then git checkout "${GITEA_VERSION}"; fi \
 && make clean-all build

# Begin env-to-ini build
RUN go build contrib/environment-to-ini/environment-to-ini.go

# Final stage
FROM debian:bookworm
LABEL maintainer="maintainers@gitea.io"
ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 22 3000

# Install necessary packages
RUN apt-get update -y && \
    apt-get install -y \
    wget \
    gpg \
    dumb-init \
    bash \
    ca-certificates \
    curl \
    gettext \
    git \
    libpam0g-dev \
    openssh-server \
    s6 \
    sqlite3 \
    sudo \
    gnupg \
    python3 \
    python3-setuptools \
    python3-pip

# Install devpod
RUN curl -L -o devpod "https://github.com/loft-sh/devpod/releases/latest/download/devpod-linux-amd64" && install -c -m 0755 devpod /usr/local/bin && rm -f devpod

# Install dvc and gto
RUN pip3 install --break-system-packages dvc[all] gto

# Install su-exec
RUN  set -ex; \
     \
     curl -o /usr/local/bin/su-exec.c https://raw.githubusercontent.com/ncopa/su-exec/master/su-exec.c; \
     \
     fetch_deps='gcc libc-dev'; \
     apt-get update; \
     apt-get install -y --no-install-recommends $fetch_deps; \
     rm -rf /var/lib/apt/lists/*; \
     gcc -Wall \
         /usr/local/bin/su-exec.c -o/usr/local/bin/su-exec; \
     chown root:root /usr/local/bin/su-exec; \
     chmod 0755 /usr/local/bin/su-exec; \
     rm /usr/local/bin/su-exec.c; \
     \
     apt-get purge -y --auto-remove $fetch_deps

# Create git user
RUN addgroup --gid 1000 git && \
    adduser --system --uid 1000 --ingroup git --home /data/git --shell /bin/bash git && \
    echo "git:*" | chpasswd

# Setup ssh
RUN mkdir -p /var/run/sshd

ENV USER=git
ENV GITEA_CUSTOM=/data/gitea

VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["/bin/s6-svscan", "/etc/s6"]

COPY docker/root /
COPY --from=build-env /go/src/code.gitea.io/gitea/gitea /app/gitea/gitea
COPY --from=build-env /go/src/code.gitea.io/gitea/environment-to-ini /usr/local/bin/environment-to-ini
COPY --from=build-env /go/src/code.gitea.io/gitea/contrib/autocompletion/bash_autocomplete /etc/profile.d/gitea_bash_autocomplete.sh
RUN chmod 755 /usr/bin/entrypoint /app/gitea/gitea /usr/local/bin/gitea /usr/local/bin/environment-to-ini
RUN chmod 755 /etc/s6/gitea/* /etc/s6/openssh/* /etc/s6/.s6-svscan/*
RUN chmod 644 /etc/profile.d/gitea_bash_autocomplete.sh
