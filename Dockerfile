# Build stage
FROM golang:1.22-bullseye AS build-env

ARG GOPROXY
ENV GOPROXY=${GOPROXY:-direct}

ARG GITEA_VERSION
ARG TAGS="sqlite sqlite_unlock_notify"
ENV TAGS="bindata timetzdata $TAGS"
ARG CGO_EXTRA_CFLAGS

# Build deps
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get update && apt-get install -y \
    build-essential \
    git \
    nodejs \
    patchelf \
    && rm -rf /var/lib/apt/lists/*

# Setup repo
COPY . ${GOPATH}/src/code.gitea.io/gitea
WORKDIR ${GOPATH}/src/code.gitea.io/gitea

# Checkout version if set
RUN if [ -n "${GITEA_VERSION}" ]; then git checkout "${GITEA_VERSION}"; fi \
 && make clean-all build -j1

# Begin env-to-ini build
RUN go build contrib/environment-to-ini/environment-to-ini.go

# Copy local files
COPY docker/root /tmp/local

# Set permissions
RUN chmod 755 /tmp/local/usr/bin/entrypoint \
              /tmp/local/usr/local/bin/gitea \
              /tmp/local/etc/s6/gitea/* \
              /tmp/local/etc/s6/openssh/* \
              /tmp/local/etc/s6/.s6-svscan/* \
              /go/src/code.gitea.io/gitea/gitea \
              /go/src/code.gitea.io/gitea/environment-to-ini
RUN chmod 644 /go/src/code.gitea.io/gitea/contrib/autocompletion/bash_autocomplete


# Final stage
FROM ubuntu:jammy
LABEL maintainer="maintainers@gitea.io"

EXPOSE 22 3000

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    gettext-base \
    git \
    libpam0g \
    openssh-server \
    sqlite3 \
    gnupg \
    sudo \
    dumb-init \
    build-essential \
    gcc \
    g++ \
    libffi-dev \
    libssl-dev \
    cmake \
    autoconf \
    automake \
    libtool \
    s6 \
    python3 \
    python3-dev \
    python3-setuptools \
    python3-pip \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

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

# Install devpod
RUN curl -L -o devpod "https://github.com/loft-sh/devpod/releases/download/v0.6.11/devpod-linux-amd64" && \
    install -c -m 0755 devpod /usr/local/bin && rm -f devpod

# Install dvc and gto
RUN pip3 install dvc[all]==3.59.0 gto

# Create user and group
RUN groupadd --gid 1000 git && \
    useradd --uid 1000 --gid 1000 --home /data/git --create-home --shell /bin/bash git && \
    echo "git:*" | chpasswd

# Create the /run/sshd directory and set permissions
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

ENV USER=git
ENV GITEA_CUSTOM=/data/gitea

VOLUME ["/data"]

ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["/usr/bin/s6-svscan", "/etc/s6"]

COPY --from=build-env /tmp/local /
COPY --from=build-env /go/src/code.gitea.io/gitea/gitea /app/gitea/gitea
COPY --from=build-env /go/src/code.gitea.io/gitea/environment-to-ini /usr/local/bin/environment-to-ini
COPY --from=build-env /go/src/code.gitea.io/gitea/contrib/autocompletion/bash_autocomplete /etc/profile.d/gitea_bash_autocomplete.sh
RUN chmod 755 /usr/bin/entrypoint /app/gitea/gitea /usr/local/bin/gitea /usr/local/bin/environment-to-ini
RUN chmod 755 /etc/s6/gitea/* /etc/s6/openssh/* /etc/s6/.s6-svscan/*
RUN chmod 644 /etc/profile.d/gitea_bash_autocomplete.sh
