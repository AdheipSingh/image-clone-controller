## fuck this. Controller, utter time water fuck kubermatic
# image-clone-controller

- Based on controller-runtime.
- Language used Golang.
- Watches deployments and daemonsets. In case image not found in backup registery, 
  the image is pulled > tagged > pushed to backupregistery.
- The deployments and daemonsets are patched with the image in backup registery and a annotations
  ```backup: true```.
- Authentication to registery is basic auth, set env to USERNAME and PASSWORD.

## Run locally

```
- export USERNAME=   #dockerhub username
- export PASSWORD=   #dockerhub password
- export REGISTERY=docker.io/imageclonecontroller/backup-registery #dockerhub registery
- export DENY_LIST=kube-system
- export WATCH_NAMESPACE=nginx
- kubectl create ns nginx
- go build -o controller
- ./controller
```

## RUN in cluster
```
- kubectl create ns nginx
- kubectl apply -f configs/
```
