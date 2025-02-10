# Getting Started with DevPod

DevPod is a powerful tool used for creating reproducible developer environments. It enables teams to maintain consistency in their development environments by allowing access to a shared project setup. In this guide, we'll walk you through the setup process for DevPod, which is commonly used at Aqua Research to access projects and run experiments in a standardized environment.

## What is DevPod?

DevPod serves two main purposes:

- Access to the Project: DevPod allows developers to access the project's source code and data, ensuring everyone is working in the same development environment.

- Running Experiments: It also provides a platform for running experiments and generating models. The results of these experiments can be compared and visualized on Gitea.

In Aqua Research we use our private server (Hetzner machine) to run the devcontainer instance and ExoScale to manage the actual data access.

## Prerequisites
Before you can set up and use DevPod, make sure you have the following prerequisites:

- Docker installation: On the private server Docker has been installed.

- DevPod Installation: You should have DevPod installed on your local machine. You can find installation instructions at https://devpod.sh/. For this guide, we will only use the CLI of DevPod.

- SSH Access: You need SSH access with a public key to the private machine where you plan to run DevPod. This is typically set up in advance.


## Setup Your DevPod Environment

### Create a Provider (SSH)
To get started, you need to create a provider that points to our private machine using SSH. Use the following commands:

```bash
devpod provider add ssh
# enter root@95.217.101.177
```

This command will create a new provider in DevPod (locally), enabling you to connect to our private machine through SSH using your public key.

## Create a New Workspace
Now, it's time to create a new workspace. This command will open Visual Studio Code and connect it to the DevPod environment inside a Docker container running on our private remote machine. Execute the following command:

```bash
devpod up --provider ssh git@github.com:gradientzero/aqua-research.git --ide vscode --debug
```

Visual Studio Code will open, automatically connecting to the DevPod environment.


## Apply Local Exoscale Credentials (Inside DevContainer)

To access data used in Aqua Research and stored on ExoScale, you need to provide your credentials. Create a file named ```.dvc/config.local``` and add the following content (Note: Replace <1password> with your actual access key and secret access key):

```bash
# create new file: .dvc/config.local
['remote "aquaremote"']
    access_key_id = <1password>
    secret_access_key = <1password>
```
(TODO: havn't found a better way, yet. But at least this has to be done only once)

Now, you can simply use the ```dvc pull``` command to retrieve remote data into this DevPod instance.

## How to Connect to a Workspace
If you've already set up a DevPod workspace and need to reconnect, use the following command:

```bash
# connect to existing devpod on remote machine
devpod up aqua-research --ide vscode --debug
```
(TODO: not sure how to connect from scratch, yet)

## (Optional) Use dvclive to Track Experiments

You can use dvclive, a tool for tracking and visualizing experiments. A new Python file ```test.py``` may have been created for this purpose, which outputs experimental metrics. You can use dvclive to monitor these experiments.

That's it! You are now set up and ready to work with DevPod for your development and experimentation needs at Aqua Research.