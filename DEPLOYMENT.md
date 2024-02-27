# Deployment with docker-compose and portainer (preferred way)
We use portainer to deploy new Sandbox instances. Here we find the instructions for setting up portainer and a short guide on how to start new instances with portainer.

TODO: Add instructions for setting up portainer and deploying new instances.


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
