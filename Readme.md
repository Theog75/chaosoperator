# Chaoskube Operator

The Chaoskube operator allows developers to consume a chaoskube object into their environment on top of Kubernetes. the operator will deploy a chaoskube instance in the relevant namespace along with the requested configuration as listed in the Chaoskube object request (CR) yaml. for example:

```
apiVersion: cache.redhat.com/v1alpha1
kind: Chaoskube
metadata:
  name: chaoskube-nightly
spec:
  # Add fields here
  size: 1
  args:
  - --interval=1m
  - --namespaces=sosivio-test
  - --minimum-age=5m
  - --no-dry-run
```

The above example will instantiate a chaoskube instance in the chaoskube-nightly namespace which will terminate a random pod every 1 minute with a minimum pod age of 5 minutes.

[All Chaoskube configurations](https://github.com/linki/chaoskube) can be applied as args in the CR above to configure chaoskube for each namespace (or possibly several of them running in a namespace) individually.

## deploying the Operator

1. clone this git repo
3. run the build.sh script as cluster-admin



