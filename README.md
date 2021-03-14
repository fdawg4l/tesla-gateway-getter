# tesla-gateway-getter

The Tesla Powerwall Gateway exports an https + JSON endpoint which you can use
to build dashboards with relevant energy information over time.  The gateway
uses an `AuthToken` which you present for every API call in order to return the
raw data.  This `AuthToken` is calculated and returned in a `Cookie` which
makes `GET`ing the API endpoints odd with something like `telegraf` since
`telegraf` has no way of specifying which endpoint to auth with first in order
to cache the resulting `Cookie`.  I mean, there might be a way, but this seemed
pretty easy.

This project grabs data from the gateway and pushes it to influxdb.
Configuration is passed via environment variables, and a sample k8s deployment
yaml is provided.



```
# To build
$ make

# To make the container and export the resulting image
$ make docker

# And because I'm lazy and don't want to run a registry on my k8s cluster
$  microk8s.ctr images import /tmp/tesla-gateway-getter-###.tar

# Create your secrets
$ microk8s kubectl create secret generic tesla-gw-getter-creds \
  --from-literal=TESLA_INFLUXBUCKET=<bucket name> \
  --from-literal=TESLA_INFLUXTOKEN=<user auth token with write access to bucket> \
  --from-literal=TESLA_INFLUXORG=<org id> \
  --from-literal=TESLA_INFLUXHOST=http://influxdb:8086  \
  --from-literal=TESLA_EMAIL=<the email address you registered with your gateway> \
  --from-literal=TESLA_PASSWORD=<password for the above> \
  --from-literal=TESLA_GATEWAY=https://<gateway ip>  \
  -n <your namespace>

# Create your deployment
$ microk8s.kubectl -n <your namespace> apply -f ./tesla-gateway-getter-deployment.yaml
```


There are probably easier ways of doing all of this like using `helm` or some
tweaking of `telegraf` yaml.  I find `yaml` infuriating.  I need a T-square to
know what is under what and even then it's a coin toss if the result will do
the right thing.  This felt way easier than arguing with `yaml`.  Anyway, PRs
are welcome if _the easier way_ is implemented by someone else.

In the end you'll get something like this.  The grafana dashboard is included.

![grafana](https://github.com/fdawg4l/tesla-gateway-getter/blob/main/grafana/sample.png)
