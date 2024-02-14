# kubectl-reverse-proxy

`kubectl-reverse-proxy` is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) that starts a reverse proxy locally and routes traffic to all the pods behind a given service object.

Kubectl has a port-forward command, which basically select one random pod to connect to, and sends traffic to just that pod.

This plugin is different in that it sends the traffic to all the pods behind a given service.
