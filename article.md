

Docker Swarm

Install docker

```bash
docker swarm init
docker build -t username/blog
docker stack deploy --compose-file docker-compose.yml blogapp

```

Check that the docker stack is running

```bash
docker stack services blogapp
```

You should see the following result:

```txt
ID             NAME           MODE         REPLICAS   IMAGE                  PORTS
qt7f7zoum6rs   blogapp_blog   replicated   1/1        cuongld2/blog:latest   *:8081->8081/tcp
tn48ag0q8euv   blogapp_db     replicated   1/1        postgres:latest        *:5432->5432/tcp
```



## Kubernetes

Install minikube

```bash
minikube start

```


```bash
mkdir ~/kubernetes/postgres/docker-pg-vol/data -p
```

Deploy postgres database
```bash
kubectl apply -f postgres-volume.yml
kubectl apply -f postgres-pvc.yml
kubectl apply -f postgres-initdb-config.yml
kubectl apply -f postgres-deployment.yml
kubectl apply -f postgres-service.yml

```

Deploy blogapp

Create docker secret

```bash
kubectl create secret docker-registry dockerhub-secret \
  --docker-server=docker.io \
  --docker-username=username \
  --docker-password=password \
  --docker-email=email
```

```bash
kubectl apply -f deployment.yml
kubectl apply -f service.yml
```

```bash
minikube tunnel
```

