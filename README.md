![Deployment Google Cloud](https://github.com/safronovD/ppw-operator/workflows/Deployment%20Google%20Cloud/badge.svg?branch=dev&event=push)
![Build and test development](https://github.com/safronovD/ppw-operator/workflows/Build%20and%20test%20development/badge.svg?branch=dev&event=push)
# ppw-operator

Operator for https://github.com/safronovD/python-pravega-writer

## Setup
```ShellSession
# Go 1.13
...

# gcc
sudo apt update
sudo apt install build-essential

# Configure Go
mkdir -p ~/go_projects/{bin,src,pkg}
export GOPATH="$HOME/go_projects"
export GOBIN="$GOPATH/bin"
export PATH=$PATH:"$GOBIN"
```

## Dependencies
```ShellSession
make all-dependencies
```

## Deploy controller
```ShellSession
make create-dir

# Generate addition files
# With controller-gen
make generate-crds
make generate-deepcopy

# With kustomize
make kustomize-crds
make kustomize-controller IMG=${IMG}


# Docker image
make docker-build IMG=${IMG}
make docker-push IMG=${IMG}


# Deploy into a cluster
make install-crds
make deploy-controller
```

## Check (deploy ppw)
```ShellSession
sudo nano config/samples/apps_v1alpha0_ppw.yaml
kubectl apply -f config/samples/apps_v1alpha0_ppw.yaml
```
