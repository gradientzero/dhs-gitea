# Development
Source: https://docs.gitea.com/installation/install-from-source

```bash
# ensure to use node 18.18+
nvm use 18.20
```

```bash
TAGS="bindata sqlite sqlite_unlock_notify" make build
```

```bash
./gitea web --custom-path /Users/me/Documents/gradient0/repos/dhs-gitea/local-workpath
```

```bash
# open http://localhost:3000/ in your browser
```

Environment variables:
```bash
# /custom/conf/app.ini

[service]
# disable user registration
DISABLE_REGISTRATION = true
# disallow users to create new organizations
DISABLE_ORGANIZATION_CREATION = true

# activate dark mode by default
[ui]
THEMES = gitea-dark
DEFAULT_THEME = gitea-dark

# remove footer content
[other]
SHOW_FOOTER_VERSION = false
SHOW_FOOTER_TEMPLATE_LOAD_TIME = false
SHOW_FOOTER_POWERED_BY = false
```

## Custom Remote Machine with Docker

Build docker image with sshd and sudo:
```bash
docker build -t docker-ssh - <<EOF
FROM docker:dind

RUN apk add --no-cache openssh openrc sudo

RUN adduser -D -u 1000 -s /bin/sh myuser && \
    adduser myuser docker && \
    echo "myuser:password" | chpasswd && \
    echo "myuser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN mkdir -p /home/myuser/.ssh && chmod 700 /home/myuser/.ssh && \
    echo "" > /home/myuser/.ssh/authorized_keys && \
    chmod 600 /home/myuser/.ssh/authorized_keys && chown -R myuser:myuser /home/myuser/.ssh

RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh && \
    echo "" > /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys

RUN mkdir -p /run/openrc && touch /run/openrc/softlevel && \
    ssh-keygen -A && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

RUN chown root:docker /var/run/docker.sock && chmod 660 /var/run/docker.sock

RUN sed -i 's/#PubkeyAuthentication no/PubkeyAuthentication yes/' /etc/ssh/sshd_config

RUN chown -R myuser:myuser /home/myuser
RUN mkdir -p /home/myuser/.devpod/agent/contexts/default/workspaces && \
    chown -R myuser:myuser /home/myuser/.devpod

CMD ["/bin/sh", "-c", "/usr/sbin/sshd -D & dockerd-entrypoint.sh"]
EOF
```

Create container:
```bash
docker run --privileged --name my-docker-ssh \
    -p 2222:22 \
    -e DOCKER_TLS_CERTDIR="" \
    docker-ssh
```

Generate SSH Key in Sandbox, for example:
```bash
# private
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAL7HtbW7F4DRqJFSsSuD2I0ipLhV47ZXLcS87R1M6lXoAoGCCqGSM49
AwEHoUQDQgAEUplz9BX1aJbpKyIfoKzNQFtlfYkQ5f+0T/94PFdRIzZhDwtn9h3w
xdbqtGtbKv8yG1VKW2vGa+KqeukpnRvllQ==
-----END EC PRIVATE KEY-----

# public
ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFKZc/QV9WiW6SsiH6CszUBbZX2JEOX/tE//eDxXUSM2YQ8LZ/Yd8MXW6rRrWyr/MhtVSltrxmviqnrpKZ0b5ZU=
```

Connect with SSH:
```bash
ssh myuser@localhost -p 2222

# add in Sandbox generated SSH public
docker exec -it my-docker-ssh sh -c 'echo "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFKZc/QV9WiW6SsiH6CszUBbZX2JEOX/tE//eDxXUSM2YQ8LZ/Yd8MXW6rRrWyr/MhtVSltrxmviqnrpKZ0b5ZU=" >> /home/myuser/.ssh/authorized_keys'
```

Add new machine:
```bash
# name: docker-in-docker-localhost
# user: myuser
# ssh key: <select generated one>
# host: localhost
# port: 2222
```

Add DevPod provider:
```bash
echo "-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAL7HtbW7F4DRqJFSsSuD2I0ipLhV47ZXLcS87R1M6lXoAoGCCqGSM49
AwEHoUQDQgAEUplz9BX1aJbpKyIfoKzNQFtlfYkQ5f+0T/94PFdRIzZhDwtn9h3w
xdbqtGtbKv8yG1VKW2vGa+KqeukpnRvllQ==
-----END EC PRIVATE KEY-----" > /tmp/temp_ssh_key
chmod 600 /tmp/temp_ssh_key

# directly with SSH:
ssh -i /tmp/temp_ssh_key myuser@localhost -p 2222

# using devpod:
devpod provider add ssh --name my-provider \
  -o HOST=root@localhost \
  -o PORT=2222 \
  -o EXTRA_FLAGS="-i /tmp/temp_ssh_key"

# cleanuo
rm -f /tmp/temp_ssh_key
```

devpod up my-workspace \
  --source=git:http://host.docker.internal:3001/arturs-org/arturs-repo \
  --provider=my-provider \
  --ide none \
  --id my-workspace \
  --debug
