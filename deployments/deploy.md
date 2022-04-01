## Deploying To A Local Cluster

### Overview 

Here are the steps involved to run this locally: 
- Get a running Kubernetes cluster
- Deploy a postgres db
- Build and deploy the application

What follows is an opinionated way to set you up with everything you need. 

### Get A Running Cluster
There are many ways to get a Kubernetes cluster. 
One of the simplest is to use [minikube](https://github.com/kubernetes/minikube).

Minikube supports several different driver (including `docker` and `podman`). 

For example, using [podman](https://podman.io/getting-started/installation.html) as the backend:
```shell
# Set up a podman machine locally 
podman machine init --cpus 4 --memory 8096
podman machine set --rootful
podman machine start

# Create a minikube cluster using a podman backend
minikube start --driver=podman --container-runtime=docker
```

The above commands create a virtualized server using podman, and initialize Kubernetes onto it.
If everything goes well, your local `kubectl get pods -A` will show you the cluster. 

### Deploy A Postgres DB
We can deploy Postgres using `helm`. 
 
There is also `ConfigMap` we need to deploy before postgres, as it is used
to boostrap the database at deploy time. 

```shell
# Deploy the bootstrap config map
kubectl apply -f ./seed.yaml
 
# Now deploy the chart
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm install coinparserpg \
      --set auth.database="crypto" \
      --set image.repository="postgres" \
      --set image.tag="latest" \
      --set primary.initdb.scriptsConfigMap="seed-db" \
      -- bitnami/postgresql 
```

### Build And Deploy The App
One of the neat features in `minikube` is that you can build images locally, and
ship them into your cluster without pushing them to a registry. This can really speed
up local development!

This points your local Docker to the agent running in the `podman machine`:
```shell
eval $(minikube -p minikube docker-env)
```

Now you simply build the image with docker, and your cluster can see it!
```shell
make docker-build VERSION=latest
```

To deploy the app:
```shell
kubectl apply -f ./deployment.yaml
```

### Sending Real Traffic
You can use `kubectl port-forward` to forward traffic to your deployed pods. 

However, we can also use `minikube tunnel`:
- In a separate terminal window run `minikube tunnel`
- Create a `LoadBalancer` type service for our app
  ```shell
  kubectl expose deployment coinparser --type=LoadBalancer --port=8080
  ```
- Your application is now accessible at the container port (8080 in this case)
  ```shell
  http://127.0.0.1:8080/api/v1/parse  
  ```
  For example, to use the `coin-spewer` to send traffic:
  ```shell
  coin-spewer -e http://127.0.0.1:8080/api/v1/parse -p 100
  ```