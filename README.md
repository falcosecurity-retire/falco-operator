# falco-operator

[falco-operator](http://github.com/mumoshu/falco-operator) is a [Kubernetes operator](https://coreos.com/operators/)
for [Sysdig Falco](https://www.sysdig.com/opensource/falco/).
 

To know more about the original Sysdig Falco and its Helm chart, have a look at [stable/falco](https://github.com/helm/charts/tree/master/stable/falco).

## Introduction

In simple workds, `falco-operator` helps [`DevSecOps`](https://www.redhat.com/en/topics/devops/what-is-devsecops).

With it, you can delegate writing a bunch of application-specific container behavioral monitoring rules to
your application developer.

As a cluster administrator, all you have to do is:

- Deploy a `falco-operator` into your cluster by using the `helm` chart
- Provide application developers correct RBAC roles and bindings to allow access to `falcorules` within their namespaces  

After that, application developers can write a `FalcoRule` in their own namespaces:

```yaml
apiVersion: "mumoshu.github.io/v1alpha1"
kind: "FalcoRule"
metadata:
  name: "bash"
  namespace: "default"
spec:
  rule: shell_in_container
  desc: notice shell activity within a container
  condition: container.id != host and proc.name = bash
  output: shell in a container (user=%user.name container_id=%container.id container_name=%container.name shell=%proc.name parent=%proc.pname cmdline=%proc.cmdline)
  priority: WARNING
```

Then, `falco-operator` takes care of the rest. It:

- Watches for `FalcoRule`s, group by namespaces,
- Creates a [Falco Rules file](https://github.com/falcosecurity/falco/wiki/Falco-Rules) per namespace
- Restart `falco` running on each node in your cluster

## How it works

If you are familiar with falco rules files, the above `FalcoRule` is translated to a rules file like:

`/var/falco-operator/rules/default.yaml`:

```yaml
- rule: shell_in_container
  desc: notice shell activity within a container
  condition: container.id != host and proc.name = bash
  output: shell in a container (user=%user.name container_id=%container.id container_name=%container.name shell=%proc.name parent=%proc.pname cmdline=%proc.cmdline)
  priority: WARNING
```

`falco-operator` automatically clones `/etc/falco/falco.yaml` to `/var/falco-operator/falco.yaml`, adding the generated rules files to `rules:` that looks:

`/var/falco-operator/falco.yaml`:

```yaml
rules:
- /var/falco-operator/rules/default.yaml
```

The operator points `falco` to the `falco.yaml` and (re)start it, so that the generated configuration is taken into account:

```console
/usr/bin/falco -c /var/falco-operator/falco.yaml
```

## Getting Started

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
