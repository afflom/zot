# Pulumi IaC | Zot on Kubernetes

This README outlines the steps to clone the Zot repository, navigate to the iac directory, and deploy Zot to a local Kind Kubernetes cluster using Pulumi written in Golang.

## Table of Contents

- Prerequisites
- Cloning the Repository
- Navigate to the IAC Directory
- Install Dependencies
- Pulumi Setup
- Golang Hygiene
- Deploy and Destroy Zot


### Install Dependencies

* Git
* Go (Golang)
* Pulumi
* Kind (Kubernetes in Docker)

### Clone the Repository

Clone the Zot repository to your local machine & navigate to the IAC Directory

```bash
git clone https://github.com/emporous/uor-zot && cd ./uor-zot/iac
```

### Init Pulumi Stack & Set Configuration Values

Download Dependencies

```bash
go get -u ./...
go mod download
go mod tidy

```

Initialize a new Pulumi stack

```bash
pulumi stack init dev
```

Set the Zot username:password credentials

```bash
pulumi config set zotUser yourUsername "admin"
pulumi config set --secret zotPasswd superSecretPasswd
```

Optional: To push the image to a registry, set the pushImage flag:

```bash
pulumi config set pushImage true
```

### Build & Deploy Zot

Run the following command to build the zot image from source and deploy to [KinD](https://kind.sigs.k8s.io):

```bash
pulumi up
```

Review the preview and select yes to proceed with the deployment.

### To destroy the deployed stack, execute:

```bash
pulumi destroy
```

Review the resources to be destroyed and select yes to proceed with the destruction.

-------------------------------------------------------

## Local Development

```bash
export GOFLAGS='-replace=github.com/emporous/uor-zot=.'
```