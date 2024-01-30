# How to use Portainer Server
Portainer is a lightweight management UI which allows you to easily manage your different Docker environments (Docker hosts or Swarm clusters). Portainer is meant to be as simple to deploy as it is to use.

## Table of Contents

- [Installation](#installation)
- [Add Environment](#add-environment)
- [Create App Templates](#create-app-templates)
- [Deploy an App Template via Stack](#deploy-an-app-template-via-stack)

## Installation

Install the Portainer Server with Docker Compose
```bash
$ docker-compose -f docker-compose.portainer.yml up -d
```

Portainer Server has now been installed. You can check to see whether the Portainer Server container has started by running
```bash
$ docker ps
```

Now that the installation is complete, you can log into your Portainer Server instance by opening a web browser and going to: https://localhost:9443
and then create Creating the first user for Initial setup and following instruction

## Add Environment
In Portainer terms, an environment is an instance that you want to manage through Portainer. Environments can be Docker, Docker Swarm, Kubernetes, ACI, Nomad or a combination. One Portainer Server instance can manage multiple environments.

### Docker Standalone
When connecting a Docker Standalone host to Portainer, there are a few different methods you can use depending on your particular requirements. You can install the Portainer Agent on the Docker Standalone host and connect via the agent, you can connect directly to the Docker API or the Docker socket, or you can deploy the Portainer Edge Agent in standard or async mode.

And this how to add and connect Environments docker standalone with docker agent.
1. First, from the menu select **Environments** then select **Add Environment**.
2. Choose **Docker Standalone** and click **Start Wizard**.
3. from Environment Wizard Choose **Agent** for connect Docker Standalone environment  
4. Copy command docker to run portainer agent and paste it to your machine docker standalone.
5. Fill in the name and Environment address to connect agent.
6. Click connect to save the connection and connect to the environment.
7. Navigate to **Home** and **Live Connect** to the environment to start using it. 

## Create App Templates
An app template lets you deploy a container (or a stack of containers) to an environment with a set of predetermined configuration values while still allowing you to customize the configuration (for example, environment variables).

Here are the steps to use an app template in Portainer, in this step we will create an app template docker-compose.yml from the repository / Gitops.

1. First, from the menu select **App Templates** then select **Custom Templates**.
2. Click **Add Custom Template** then complete the details, using the table below as a guide.

    | Fields            | Overview                                                                                     |
    |-------------------|----------------------------------------------------------------------------------------------|
    | Title             | Give the template a descriptive name.                                                        |
    | Description       | Enter a brief description of what your template includes.                                    |
    | Note              | Note any extra information about the template (optional).                                    |
    | Icon URL          | Enter the URL to an icon to be used for the template when it appears in the list (optional). |
    | Platform          | Select the compatible platform for the template. Options are Linux or Windows.               |
    | Type              | Select the type of template. Options are Standalone or Swarm.                                | 
3. Selecting the build method
   Next, choose the build method that suits your needs. You can use the web editor to manually enter your docker-compose file, upload a docker-compose.yml file from your local computer, or pull the compose file from a Git repository.

   in this step, we will use the **Repository** for get the docker-compose.yml file from git repository.
   Fill in the details for your Git repository.

    | Fields              | Overview                                                                                     |
    |---------------------|----------------------------------------------------------------------------------------------|
    | Authentication      | If your repository requires access authentication, toggle Authentication on then enter the username and personal access token. |
    | Repository URL      | Enter the URL to your Git repository.                                                        |
    | Repository reference| Enter the repository reference to define the branch or tag to pull from. If blank, the default HEAD reference will be used. |
    | Compose path        | Enter the path within the repository to your docker-compose file.                           |


  3. Last, click **Create custom template** for save the template.
     After creating the app template, we will see the app template that we created in the **App Templates** > **Custom Templates** menu.


## Deploy an App Template via Stack
After creating the app template, we can use the app template to deploy the application to the environment. In this step, we will use the app template that we created earlier to deploy the application to the environment.

1. Navigate to **Stacks** from the menu and then select **Add Stack**.
2. Fill in the field stack name form
3. Choose **Custom Template** as the Build Method and select your previously created app template.
3. Set Environment variables as needed
  To add environment variables such as usernames and passwords, switch to advanced mode for the ability to copy and paste multiple variables.
4. Click **Deploy** to deploy the stack. The stack will be deployed and the containers will be started.