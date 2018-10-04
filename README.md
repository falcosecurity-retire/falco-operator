Installation:

```console
helm tiller run -- \
  helm upgrade --install fo1 charts/falco-operator \
  --recreate-pods \
  --values charts/falco-operator/values.yaml \
  --namespace kube-system

$ kubectl create -f examples/bash.falcorule.yaml
```

Verity that falco-operator triggers an alert for the rule created from the custom resource:

```console
$ kubectl run --image redis --restart Never myredis

$ kubectl exec -it myredis bash
```

```console
$ ks logs fo1-falco-operator-falco-v97rc
/var/falco-operator/rules/test1: test1
/var/falco-operator/rules: rules

Watching 2 files
ignoring dir of /var/falco-operator/rules/..2018_10_04_14_02_09.952277388/test1
ignoring /var/falco-operator/rules/..data
/var/falco-operator/rules/test1 has been updated
starting app...
[dancer-crack] 2018/10/04 14:02:10 Started with PID 10
[dancer-crack] 2018/10/04 14:02:10 out: * Setting up /usr/src links from host
[dancer-crack] 2018/10/04 14:02:10 out: ls: cannot access '/host/usr/src': No such file or directory
[dancer-crack] 2018/10/04 14:02:10 out: * Mounting debugfs
[dancer-crack] 2018/10/04 14:02:10 out: Found kernel config at /proc/config.gz
[dancer-crack] 2018/10/04 14:02:10 out: * Minikube detected (v0.28.1), downloading and setting up kernel headers
[dancer-crack] 2018/10/04 14:02:10 out: * Downloading http://mirrors.edge.kernel.org/pub/linux/kernel/v4.x/linux-4.15.tar.gz
kuoka-yusuke-3:falco-operator kuoka-yusuke$ ks logs fo1-falco-operator-falco-v97rc -f
/var/falco-operator/rules/test1: test1
/var/falco-operator/rules: rules

Watching 2 files
ignoring dir of /var/falco-operator/rules/..2018_10_04_14_02_09.952277388/test1
ignoring /var/falco-operator/rules/..data
/var/falco-operator/rules/test1 has been updated
starting app...
[dancer-crack] 2018/10/04 14:02:10 Started with PID 10
[dancer-crack] 2018/10/04 14:02:10 out: * Setting up /usr/src links from host
[dancer-crack] 2018/10/04 14:02:10 out: ls: cannot access '/host/usr/src': No such file or directory
[dancer-crack] 2018/10/04 14:02:10 out: * Mounting debugfs
[dancer-crack] 2018/10/04 14:02:10 out: Found kernel config at /proc/config.gz
[dancer-crack] 2018/10/04 14:02:10 out: * Minikube detected (v0.28.1), downloading and setting up kernel headers
[dancer-crack] 2018/10/04 14:02:10 out: * Downloading http://mirrors.edge.kernel.org/pub/linux/kernel/v4.x/linux-4.15.tar.gz
```
