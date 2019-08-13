[![Build Status](https://travis-ci.org/C123R/kubectl-getenv.svg?branch=master)](https://travis-ci.org/C123R/kubectl-getenv)

# kubectl-getenv

This is a kubectl plugin to get the all the environment variables for the containers that run in the Pod.

## Installation

To use this kubectl-getenv plugin, you can follow the official Kubernetes Plugin [documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/#using-a-plugin).

- Make it executable:

```sh
chmod u+x kubectl-getenv
```

- Place it in your PATH:

```sh
 mv kubectl-aks /usr/local/bin
```

- Now it can be access using `kubectl` command:

```sh
$ kubectl getenv
The kubectl-getenv plugin gets all the environment variables for the containers that run in the Pod.

Usage:
  getenv [flags]

Flags:
  -h, --help               help for getenv
  -n, --namespace string   name of the namespace (default "default")
```

## Usage

- Get list of environment variables for specific pod:

```sh
$ kubectl getenv -n nginx nginx-pod
```