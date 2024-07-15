# Deployment with docker-compose and portainer (preferred way)
Follow the instructions to deploy on a bare metal:
https://gradient0.atlassian.net/wiki/spaces/PROJ/pages/2138144772/Hetzner+Docker+Deployment+Template

We use `portainer` to deploy new Sandbox instances (`dhs-gitea`). Here we find the instructions for setting up `portainer` and a short guide on how to start new instances with `portainer`.

The logical structure on one server would then look like this:
```bash
- remote machine
-- portainer
--- dhs-gitea-001
--- dhs-gitea-002
--- ...
```

`portainer` can also manage Sandbox (`dhs-gitea`) instances remotely on different servers, but that's out of scope for now.

## Build configuration
On MacOS Docker Desktop ensure "Allow privileged port mapping" is checked in the Docker Desktop settings. This allows the container to bind to port 22 or any other port below 1024.

## dhs-gitea
`dhs-gitea` is the Sandbox and is controlled by `portainer`. We can also run `dgs-gitea` manually with docker-compose, but portainer enables us to  create, starte and stop running instances automatically.

For build image latest and with specified version:
```bash
docker build -t gradient0/dhs-gitea:latest -t gradient0/dhs-gitea:<$version> .
```
Note: `portainer` does not need to be built and can be obtained directly in docker-compose.

To push image to registry docker hub, you need to log in to Docker Hub first and then push, with the following command:
```bash
docker login --username gradient0 # 1password
docker push gradient0/dhs-gitea:latest
docker push gradient0/dhs-gitea:{$version}
```

## portainer
`portainer` is a management service which allows us to easily manage  Docker compose services. In our case `portainer` manages all Sandbox instances, or in technical form: `dhs-gitea` instances.

## Copy data to remote machine
Prepare environment variables:
```bash
cp portainer.env.example portainer.env
cp stack.env.example stack.env
```

Copy the following files to the destination server:
```bash
scp docker-compose.yml \
    docker-compose.portainer.yml \
    portainer.env \
    stack.env \
    root@dhs.detabord.com:/home/user/dhs
```

## Prepare server for dhs-gitea

### OS user 'git'
```bash
adduser git

# Note: not sure if you really these steps
# add user 'git' to add to following groups: sudo, docker and user
usermod -aG sudo git
usermod -aG docker git
usermod -aG user git
```

### systemd service for portainer
Create a new systemd service for portainer. This will allow us to start and stop portainer as a service.
```bash
# /etc/systemd/system/portainer.service
[Unit]
Description=portainer service
Requires=docker.service
After=docker.service

[Service]
User=root
Group=root
WorkingDirectory=/home/user/dhs
ExecStart=/usr/bin/docker compose -p portainer -f docker-compose.portainer.yml --env-file portainer.env up
ExecStop=/usr/bin/docker compose -p portainer -f docker-compose.portainer.yml --env-file portainer.env down

[Install]
WantedBy=multi-user.target
```

Run the following commands to enable and start the portainer service:
```bash
systemctl enable portainer
systemctl start portainer
```

### nginx configuration for portainer
Create a nginx configuration file for portainer. This will allow us to access portainer via a web browser. Because we need to retrieve the certificate first, we need to create a temporary configuration file and then replace it with the final configuration file:

```bash
# /etc/nginx/conf.d/portainer.detabord.com
server {
  listen 80;
  server_name portainer.detabord.com;
  root /usr/share/nginx/html;
  index index.html;
  location ~ /.well-known/acme-challenge {
    allow all;
  }
}
```

Rstart nginx:
```bash
systemctl restart nginx
```

Get the certificate:
```bash
certbot certonly --webroot -w /usr/share/nginx/html -d portainer.detabord.com
```

Once the certificate has been retrieved, replace the temporary configuration file with the final configuration file:

```bash
# /etc/nginx/conf.d/portainer.detabord.com
upstream portainer-server {
  # sync with portainer.env (PORTAINER_HTTPS_PORT)
  server 127.0.0.1:9443 fail_timeout=0;
}

upstream portainer-tunnel {
  # sync with portainer.env (PORTAINER_TUNNEL_PORT)
  server 127.0.0.1:8000 fail_timeout=0;
}

server {
  listen 80;
  server_name portainer.detabord.com;
  return 301 https://$server_name$request_uri;
}

server {
  listen 443 ssl http2;
  listen [::]:443 ssl http2;

  server_name portainer.detabord.com;
  charset utf-8;

  root /usr/share/nginx/html;
  index index.html;

  client_max_body_size 200M;

  ssl_certificate /etc/letsencrypt/live/portainer.detabord.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/portainer.detabord.com/privkey.pem;

  location ~ /.well-known/acme-challenge/ {
    allow all;
  }

  location / {
    auth_basic "Restricted Access!";
    auth_basic_user_file /etc/nginx/conf.d/.htpasswd;
    proxy_pass http://portainer-server/;
    include proxy_params;
  }
}
```

Create a new user for basic auth:
```bash
# create new entry in password file (user: portainer, pass: <add to 1password>)
echo -n 'portainer:' >> /etc/nginx/conf.d/.htpasswd
openssl passwd -apr1 >> /etc/nginx/conf.d/.htpasswd
```

Finally, restart nginx:
```bash
systemctl restart nginx
```

# Desired Outcome
After following the instructions above, we now need to have a server with the following setup:
- ssh restricted access via public key only
- nginx installed
- certbot installed
- docker installed
- new OS user 'user'
- new OS user 'git'
- project related files uploaded to the server
- systemd service portainer up and running
- nginx configured to serve portainer dashboard

