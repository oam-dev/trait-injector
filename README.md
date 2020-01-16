# Trait Injector

Trait injector is a k8s admission webhook to intercept component workload operation and inject traits information.

## Build

```bash
make
```

## Test

```bash
make test
```

## SSL

the `ssl/` dir contains a script to create a self-signed certificate, not sure this will even work when running in k8s but that's part of figuring this out I guess

_NOTE: the app expects the cert/key to be in `ssl/` dir relative to where the app is running/started and currently is hardcoded to `mutateme.{key,pem}`_

```bash
cd ssl/ 
make 
```

## Docker

```bash
make docker-build
```

## Deploy to Minikube

```bash
make minikube
```

## Quickstart

```bash
# Create ServiceBinding
kubectl create -f ./example/servicebinding.yaml

# Create Deployment and Secret
kubectl create -f ./example/noenv.yaml

# Verify the envFrom field has been injected successfully
kubectl get deploy busybox1 -o json | jq -r '.spec.template.spec.containers[0]'
```
![alt text](./doc/img/envFrom.png)
