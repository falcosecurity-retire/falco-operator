# Falco Operator

The Falco Operator for Kubernetes provides an easy way to deploy Falco.

[Falco](https://falco.org) is a behavioral activity monitor designed to detect
anomalous activity in your applications. You can use Falco to monitor run-time
security of your Kubernetes applications and internal components.

To know more about Falco have a look at:

- [Kubernetes security logging with Falco & Fluentd
](https://sysdig.com/blog/kubernetes-security-logging-fluentd-falco/)
- [Active Kubernetes security with Sysdig Falco, NATS, and kubeless](https://sysdig.com/blog/active-kubernetes-security-falco-nats-kubeless/)
- [Detecting cryptojacking with Sysdig’s Falco
](https://sysdig.com/blog/detecting-cryptojacking-with-sysdigs-falco/)

This operator adds Falco to all nodes in your cluster using a DaemonSet. It
also provides a Deployment for generating Falco alerts. This is useful for
testing purposes.

## Quickstart

To quickly try out **just** the Falco Operator inside a cluster, run the
following command:

```shell
kubectl apply -f https://raw.githubusercontent.com/falcosecurity/falco-operator/helm-based-operator/bundle.yaml
```

This command deploys the operator itself and its dependencies: Custom Resource
Definitions, ServiceAccount and its ClusterRoleBinding.

## Settings

This operator, uses the same options than the
[Helm Chart](https://hub.helm.sh/charts/stable/falco), please take a look to
all the options in the following table:

| Parameter                                       | Description                                                          | Default                                                                                |
| ---                                             | ---                                                                  | ---                                                                                    |
| `image.repository`                              | The image repository to pull from                                    | `falcosecurity/falco`                                                                  |
| `image.tag`                                     | The image tag to pull                                                | `0.13.0`                                                                               |
| `image.pullPolicy`                              | The image pull policy                                                | `IfNotPresent`                                                                         |
| `resources`                                     | Specify container resources                                          | `{}`                                                                                   |
| `extraArgs`                                     | Specify additional container args                                    | `[]`                                                                                   |
| `rbac.create`                                   | If true, create & use RBAC resources                                 | `true`                                                                                 |
| `serviceAccount.create`                         | Create serviceAccount                                                | `true`                                                                                 |
| `serviceAccount.name`                           | Use this value as serviceAccountName                                 | ` `                                                                                    |
| `fakeEventGenerator.enabled`                    | Run falco-event-generator for sample events                          | `false`                                                                                |
| `fakeEventGenerator.replicas`                   | How many replicas of falco-event-generator to run                    | `1`                                                                                    |
| `proxy.httpProxy`                               | Set the Proxy server if is behind a firewall                         | ``                                                                                     |
| `proxy.httpsProxy`                              | Set the Proxy server if is behind a firewall                         | ``                                                                                     |
| `proxy.noProxy`                                 | Set the Proxy server if is behind a firewall                         | ``                                                                                     |
| `ebpf.enabled`                                  | Enable eBPF support for Falco instead of `falco-probe` kernel module | `false`                                                                                |
| `ebpf.settings.hostNetwork`                     | Needed to enable eBPF JIT at runtime for performance reasons         | `true`                                                                                 |
| `ebpf.settings.mountEtcVolume`                  | Needed to detect which kernel version are running in Google COS      | `true`                                                                                 |
| `falco.rulesFile`                               | The location of the rules files                                      | `[/etc/falco/falco_rules.yaml, /etc/falco/falco_rules.local.yaml, /etc/falco/rules.d]` |
| `falco.jsonOutput`                              | Output events in json or text                                        | `false`                                                                                |
| `falco.jsonIncludeOutputProperty`               | Include output property in json output                               | `true`                                                                                 |
| `falco.logStderr`                               | Send Falco debugging information logs to stderr                      | `true`                                                                                 |
| `falco.logSyslog`                               | Send Falco debugging information logs to syslog                      | `true`                                                                                 |
| `falco.logLevel`                                | The minimum level of Falco debugging information to include in logs  | `info`                                                                                 |
| `falco.priority`                                | The minimum rule priority level to load and run                      | `debug`                                                                                |
| `falco.bufferedOutputs`                         | Use buffered outputs to channels                                     | `false`                                                                                |
| `falco.outputs.rate`                            | Number of tokens gained per second                                   | `1`                                                                                    |
| `falco.outputs.maxBurst`                        | Maximum number of tokens outstanding                                 | `1000`                                                                                 |
| `falco.syslogOutput.enabled`                    | Enable syslog output for security notifications                      | `true`                                                                                 |
| `falco.fileOutput.enabled`                      | Enable file output for security notifications                        | `false`                                                                                |
| `falco.fileOutput.keepAlive`                    | Open file once or every time a new notification arrives              | `false`                                                                                |
| `falco.fileOutput.filename`                     | The filename for logging notifications                               | `./events.txt`                                                                         |
| `falco.stdoutOutput.enabled`                    | Enable stdout output for security notifications                      | `true`                                                                                 |
| `falco.programOutput.enabled`                   | Enable program output for security notifications                     | `false`                                                                                |
| `falco.programOutput.keepAlive`                 | Start the program once or re-spawn when a notification arrives       | `false`                                                                                |
| `falco.programOutput.program`                   | Command to execute for program output                                | `mail -s "Falco Notification" someone@example.com`                                     |
| `customRules`                                   | Third party rules enabled for Falco                                  | `{}`                                                                                   |
| `integrations.gcscc.enabled`                    | Enable Google Cloud Security Command Center integration              | `false`                                                                                |
| `integrations.gcscc.webhookUrl`                 | The URL where sysdig-gcscc-connector webhook is listening            | `http://sysdig-gcscc-connector.default.svc.cluster.local:8080/events`                  |
| `integrations.gcscc.webhookAuthenticationToken` | Token used for authentication and webhook                            | `b27511f86e911f20b9e0f9c8104b4ec4`                                                     |
| `integrations.natsOutput.enabled`               | Enable NATS Output integration                                       | `false`                                                                                |
| `integrations.natsOutput.natsUrl`               | The NATS' URL where Falco is going to publish security alerts        | `nats://nats.nats-io.svc.cluster.local:4222`                                           |
| `integrations.snsOutput.enabled`                | Enable Amazon SNS Output integration                                 | `false`                                                                                |
| `integrations.snsOutput.topic`                  | The SNS topic where Falco is going to publish security alerts        | ` `                                                                                    |
| `integrations.snsOutput.aws_access_key_id`      | The AWS Access Key Id credentials for access to SNS n                | ` `                                                                                    |
| `integrations.snsOutput.aws_secret_access_key`  | The AWS Secret Access Key credential to access to SNS                | ` `                                                                                    |
| `integrations.snsOutput.aws_default_region`     | The AWS region where SNS is deployed                                 | ` `                                                                                    |
| `tolerations`                                   | The tolerations for scheduling                                       | `node-role.kubernetes.io/master:NoSchedule`                                            |


For example, if you want to deploy a DaemonSet with eBPF enabled:

```yaml
apiVersion: falco.org/v1alpha1
kind: Falco
metadata:
  name: falco-with-ebpf
spec:
  ebpf:
    enabled: true
```

And you can apply this file with `kubectl apply -f`

And then you can see you pods with Falco deployed:

```shell
$ kubectl get pods
NAME                                              READY   STATUS    RESTARTS   AGE
falco-operator-5577677484-tbj68                   1/1     Running   0          16s
falco-with-ebpf-djp216w2844g850nzy4qaa9rs-9p7rc   1/1     Running   0          3s
falco-with-ebpf-djp216w2844g850nzy4qaa9rs-c2gll   1/1     Running   0          3s
falco-with-ebpf-djp216w2844g850nzy4qaa9rs-hf62x   1/1     Running   0          3s
```

## Removal

Just run the `kubectl delete` command to remove the operator and its dependencies.

```shell
kubectl delete -f https://raw.githubusercontent.com/falcosecurity/falco-operator/helm-based-operator/bundle.yaml
```
