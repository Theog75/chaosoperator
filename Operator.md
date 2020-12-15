#### initialize the sdk:
```
operator-sdk init --domain=redhat.com --repo=chaosoperator

operator-sdk create api --group=cache --version=v1alpha1 --kind=Chaoskube
```

#### Build the operator

```
make docker-build IMG=docker.rct.co.il/chaosoperator:V0.1.0
make docker-push IMG=docker.rct.co.il/chaosoperator:V0.1.0
make deploy IMG=docker.rct.co.il/chaosoperator:V0.1.0
```