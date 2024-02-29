# Deployment with docker-compose and portainer (preferred way)
Follow the isntructions to deploy on a bare metal:
https://gradient0.atlassian.net/wiki/spaces/PROJ/pages/2138144772/Hetzner+Docker+Deployment+Template

We use `portainer` to deploy new Sandbox instances (`dhs-gitea`)). Here we find the instructions for setting up `portainer` and a short guide on how to start new instances with `portainer`.

The logical structure on one server would then look like this:
- remote machine
-- portainer
--- dhs-gitea-001
--- dhs-gitea-002
--- ...

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

# Desrired Outcome
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





### Setup portainer to manage dhs-gitea instances

Now that the installation is complete, you can log into your Portainer Server instance by opening a web browser and going to: https://localhost:9443 and then create the first user for initial setup and follow instruction

#### Add Environment
1. First, from the menu select **Environments** then select **Add Environment**.
2. Choose **Docker Standalone** and click **Start Wizard**.
3. from Environment Wizard Choose **Agent** for connect Docker Standalone environment
4. Copy command docker to run portainer agent and paste it to your machine docker standalone.
5. Fill in the name and Environment address to connect agent.
6. Click connect to save the connection and connect to the environment.
7. Navigate to **Home** and **Live Connect** to the environment to start using it.

#### Create App Templates
1. First, from the menu select **App Templates** then select **Custom Templates**.
2. Click **Add Custom Template** then complete the details
3. Selecting the build method
   I suggest to use web editor to manually enter docker-compose.yml content.
4. Click **Create custom template** for save the template.
   Then, we will see the app template that we created in the **App Templates** > **Custom Templates** menu.

#### Deploy an App Template via Stack
After creating the app template, we can use the app template to deploy the application to the environment.
1. Navigate to **Stacks** from the menu and then select **Add Stack**.
2. Fill in the field stack name form
3. Choose **Custom Template** as the Build Method and select your previously created app template.
4. Set Environment variables as needed
  To add environment variables such as usernames and passwords, switch to advanced mode for the ability to copy and paste multiple variables.
5. Click **Deploy** to deploy the stack. The stack will be deployed and the containers will be started.


# Deployment with docker-compose and without portainer
We usually use portainer to deploy new instances. However, if you want to set up a new sandbox instance (dhs-gitea), you can follow this section and set up a new independent instance using docker compose.


## Build configuration
On MacOS Docker Desktop ensure "Allow privileged port mapping" is checked in the Docker Desktop settings. This allows the container to bind to port 22 or any other port below 1024.

## Build docker
For build image latest and with specified version:

```bash
docker build -t gradient0/dhs-gitea:latest -t gradient0/dhs-gitea:<$version> .
```

## Push image to docker hub
To push image to registry docker hub, you need to log in to Docker Hub first and then push, with the following command:

```bash
docker login --username gradient0
# 1password

docker push gradient0/dhs-gitea:latest
docker push gradient0/dhs-gitea:{$version}
```

## Copy data to remote machine

Prepare environment variables:
```bash
cp stack.env.example stack.env
```

Copy the following files to the destination server:

```bash
- docker-compose.yml
- stack.env
```

```bash
scp docker-compose.yml stack.env root@dhs.detabord.com:/home/user/dhs
```

## Deploy with docker compose
For deploy with docker compose, you need to create `stack.env` file and change environment needed, you can check `stack.env.example` file for example, with the following command to run docker compose:

```bash
docker compose -f docker-compose.yml -p <my-project-name>--env-file stack.env up
```



# Deployment without docker-compose and without portainer (bare-metal)


## Install go 1.21.x
- cd ~
- curl -OL https://golang.org/dl/go1.21.3.linux-amd64.tar.gz
- sudo tar -C /usr/local -xvf go1.21.3.linux-amd64.tar.gz
- echo ‘export PATH=$PATH:/usr/local/go/bin’ > ~/.profile
- source ~/.profile

## Install mysql
- sudo apt update
- sudo apt install mysql-server
- sudo systemctl start mysql.service

## Create mysql user and db
- ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'password';
- CREATE DATABASE gitea;

## Create system user “dev”
- sudo adduser dev
- sudo usermod -aG sudo dev

## Clone dhs_gitea repo inside /home/dev folder

## Build go application
- TAGS="bindata" make build

## Set system service
- sudo nano /lib/systemd/system/gitea.service
- Copy following code inside the file.

```
[Unit]
Description=gitea

[Service]
Type=simple
User=dev
Restart=always
RestartSec=5s
ExecStart=/home/dev/dhs-gitea/gitea web

[Install]
WantedBy=multi-user.target
```
- sudo service gitea start

## Install nginx

- sudo apt update
- sudo apt install nginx
- sudo nano /etc/nginx/sites-available/dhs.detabord.com
- Put following configuration

```
server {
        listen 80 default_server;
        server_name <server_name>;

        location / {
                proxy_set_header   X-Forwarded-For $remote_addr;
                proxy_set_header   Host $http_host;
                proxy_pass         http://<server_ip>:3000;
        }
}
```
- sudo ln -s /etc/nginx/sites-available/dhs.detabord.com /etc/nginx/sites-enabled
- sudo service nginx restart


## Install gitea
- Open http://<server_ip> from browser
- Configure db credentials on the page and install.

## Protect signup with basic auth
- sudo apt install apache2-utils
- htpasswd -c /etc/nginx/conf.d/.htpasswd admin
- Chage nginx config like following
```
server {
        listen 80 default_server;
        server_name <server_name>;

        location / {
                proxy_set_header   X-Forwarded-For $remote_addr;
                proxy_set_header   Host $http_host;
                proxy_pass         http://<server_ip>:3000;

                location /user/sign_up {
                        auth_basic "Restricted Access!";
                        auth_basic_user_file /etc/nginx/conf.d/.htpasswd;
                        proxy_set_header   X-Forwarded-For $remote_addr;
                        proxy_set_header   Host $http_host;
                        proxy_pass         http://<server_ip>:3000;
                }
        }
}
```

## Install TLS with letsencrypt
- apt install snapd
- snap install --classic certbot
- certbot --nginx -d dhs.detabord.com
- systemctl enable --now snapd.socket
- ln -s /var/lib/snapd/snap /snap
- systemctl restart snapd
- snap install core
- snap refresh core
- snap install --classic certbot
- ln -s /snap/bin/certbot /usr/bin/certbot
- certbot certonly --nginx
- Change nginx config like following
```
server {
        listen 80;
        server_name dhs.detabord.com;
        return 301 https://$server_name$request_uri;
}

server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        server_name dhs.detabord.com;
        charset utf-8;

        ssl_certificate /etc/letsencrypt/live/dhs.detabord.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/dhs.detabord.com/privkey.pem;

        location / {
                proxy_set_header   X-Forwarded-For $remote_addr;
                proxy_set_header   Host $http_host;
                proxy_pass         http://95.217.101.177:3000;

                location /user/sign_up {
                        auth_basic "Restricted Access!";
                        auth_basic_user_file /etc/nginx/conf.d/.htpasswd;
                        proxy_set_header   X-Forwarded-For $remote_addr;
                        proxy_set_header   Host $http_host;
                        proxy_pass         http://95.217.101.177:3000;
                }
        }
}

```
