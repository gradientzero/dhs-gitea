# Deployment with docker-compose
Follow the instructions to deploy on a bare metal:
https://gradient0.atlassian.net/wiki/spaces/PROJ/pages/2138144772/Hetzner+Docker+Deployment+Template

## Build configuration
On MacOS Docker Desktop ensure "Allow privileged port mapping" is checked in the Docker Desktop settings. This allows the container to bind to port 22 or any other port below 1024.

## dhs-gitea
For build image latest and with specified version:
```bash
docker build -t gradient0/dhs-gitea:latest -t gradient0/dhs-gitea:<$version> .
```

To push image to registry docker hub, you need to log in to Docker Hub first and then push, with the following command:
```bash
docker login --username gradient0 # 1password
docker push gradient0/dhs-gitea:latest
docker push gradient0/dhs-gitea:{$version}
```

## Copy data to remote machine
Copy the following files to the destination server:
```bash
scp docker-compose.yml \
    .env \
    root@sandbox.gradient0.com:/root/sandbox
```

## Prepare server for dhs-gitea

### OS user 'git'
```bash
adduser git
```

### Create folders
Since 1.19+, Gitea requires the .ssh folder to be present in the home directory of the git user. Create the folder and set the correct permissions:
```bash
mkdir /home/git/.ssh
chown git:git /home/git/.ssh
chmod 700 /home/git/.ssh
```

Lets map the new data folder to the server:
```bash
mkdir /home/git/sandbox/data
chown git:git /home/git/sandbox/data
chmod 700 /home/git/sandbox/data
```

Ensure to map with docker-compose.yml

### systemd service
```bash
# /etc/systemd/system/sandbox.service
[Unit]
Description=sandbox service
Requires=docker.service
After=docker.service

[Service]
User=root
Group=root
WorkingDirectory=/home/git/sandbox
ExecStart=/usr/bin/docker compose -p sandbox -f docker-compose.yml up
ExecStop=/usr/bin/docker compose -p sandbox -f docker-compose.yml down

[Install]
WantedBy=multi-user.target
```

Run the following commands to enable and start the sandbox service:
```bash
systemctl enable sandbox
systemctl start sandbox
```

### nginx configuration for sandbox
Create a new configuration file for nginx:
```bash
# /etc/nginx/conf.d/sandbox.gradient0.com.conf
server {
  listen 80;
  server_name sandbox.gradient0.com;
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
certbot certonly --webroot -w /usr/share/nginx/html -d sandbox.gradient0.com
```

Once the certificate has been retrieved, replace the temporary configuration file with the final configuration file:

```bash
# /etc/nginx/conf.d/sandbox.gradient0.com.conf
upstream sandbox {
  server 127.0.0.1:3001;
}

server {
  listen 80;
  server_name sandbox.gradient0.com;
  return 301 https://$server_name$request_uri;
}

server {
  listen 443 ssl http2;
  listen [::]:443 ssl http2;

  server_name sandbox.gradient0.com;
  charset utf-8;

  root /usr/share/nginx/html;
  index index.html;

  client_max_body_size 200M;

  ssl_certificate /etc/letsencrypt/live/sandbox.gradient0.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/sandbox.gradient0.com/privkey.pem;

  location ~ /.well-known/acme-challenge/ {
    allow all;
  }

  location / {
    #auth_basic "Restricted Access!";
    #auth_basic_user_file /etc/nginx/conf.d/.htpasswd;
    proxy_pass http://sandbox/;
    include proxy_params;
  }
}
```

Create a new user for basic auth:
```bash
# create new entry in password file (user: portainer, pass: <add to 1password>)
echo -n 'sandbox:' >> /etc/nginx/conf.d/.htpasswd
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
- new OS user 'git'
- project related files uploaded to the server
- systemd service portainer up and running


# Set the new server

Follow instructions: https://gradient0.atlassian.net/wiki/spaces/D/pages/2138832922/10+-+Run+Single+Sandbox+Instance

Short:
- `SSH Server Port` maps with you environment variable `GITEA_SSH_PORT` (default: 2221)
- `Gitea HTTP Listen Port` stays with value `3000` - it is the internal port within Docker Compose!
- `Server Domain` stays `sandbox.gradient0.com`, but can be adapted to your requirements if run on a server.
- `Gitea Base URL` stays `https://sandbox.gradient0.com/`, but the port must map your `GITEA_PORT`. Can be adapted to your requirements if run on a server.


# Connect Remote Machine

Der Teil besteht aus mehreren Schritten:
- Zunächst muss eine organistion angelegt werden
- Anschließend kann eine neues Repository (zB aus dem Template) erstellt werden
- Danach erstellen wir in der Organisation ein SSH Key. Wichtig: Dieser SSH Key dient NICHT dem clonen des Repositories, sondern dem Zugriff über SSH (public) auf die Maschine.
- Der neue Public Key des SSH Keys wird in destination Maschine unter `<user-home>/.ssh/authorized_keys` abgelegt. Bei root ist das `/root/.ssh/authorized_keys`. Der Nutzer muss docker Rechte haben.
- Dann erstellen wir eine neue Maschine und nutzen den neuen SSH Key
- Wenn das Repository und die Organization public sind, kann das Repository ohne weitere Schritte geclont und Compute ausgeführt werden
- Wenn das Repository und/oder die Organization aber privat ist, muss zunächst ein Nutzer, der Zugriff auf das zu klonende Repo in der Sandbox hat, ein Access Token erstellen. User Settings -> Applications -> Generate Token. Mit diesem Token kann das Repository geclont werden.
- Als letzter Schritt muss dieser neue Nutzer Access Token als Gitea Token in der Organization -> Settings -> Gitea Token hinterlegt werden. Als Nutzer kann zum Beispiel: user und als Token kann das zuvor erstellte Token verwendet werden. Dieses Token wird benötigt, um das Repository zu clonen.
- Wenn das Repository DVC verwendet, so müssen zusätzlich `Devpod Credentials` in der Organization -> Settings -> Devpod Credentials hinterlegt werden. Hierbei handelt es sich um die S3 Credentials, die für das DVC Repository benötigt werden. Format:

-- Remote => aquaremote
-- Key => access_key_id
-- Value => <1password>

-- Remote => aquaremote
-- Key => secret_access_key
-- Value => <1password>

# Troubleshooting
```bash
# delete all volumes
docker compose down --volumes
```


bash into running service:
```bash
docker compose -p sandbox exec -it server bash
```
