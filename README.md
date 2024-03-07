# kubectl-reverse-proxy

`kubectl-reverse-proxy` is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) that starts a reverse proxy locally and routes traffic to all the pods behind a given service object.

Kubectl has a port-forward command, which basically select one random pod to connect to, and sends traffic to just that pod.

This plugin is different in that it sends the traffic to all the pods behind a given service.

# Installing via krew
- install `krew` using instructions [here](https://github.com/kubernetes-sigs/krew#installation)
- run `kubectl krew update`
- run `kubectl krew install reverse-proxy`


# Usage

```bash
-> kubectl reverse-proxy simple-app

starting reverse proxy listening on localhost:9090
2024/03/07 15:26:02.392 INFO    admin   admin endpoint started  {"address": "localhost:2019", "enforce_origin": false, "origins": ["//[::1]:2019", "//127.0.0.1:2019", "//localhost:2019"]}
2024/03/07 15:26:02.392 INFO    autosaved config (load with --resume flag)      {"file": "/Users/rajatjindal/Library/Application Support/Caddy/autosave.json"}
2024/03/07 15:26:02.402 INFO    admin.api       received request        {"method": "POST", "host": "localhost:2019", "uri": "/load", "remote_ip": "127.0.0.1", "remote_port": "56843", "headers": {"Accept-Encoding":["gzip"],"Content-Length":["836"],"Content-Type":["application/json"],"User-Agent":["Go-http-client/1.1"]}}
2024/03/07 15:26:02.402 INFO    redirected default logger       {"from": "stderr", "to": "discard"}
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.833080210Z 
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841354835Z Serving http://0.0.0.0:80
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841360710Z Available Routes:
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841361501Z   hello: http://0.0.0.0:80/hello
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841363043Z   go-hello: http://0.0.0.0:80/go-hello
```

with dashboard enabled

```shell
-> kubectl reverse-proxy simple-app --dashboard

starting reverse proxy listening on localhost:9090
starting metrics dashboard at http://localhost:9092
2024/03/07 15:26:02.392 INFO    admin   admin endpoint started  {"address": "localhost:2019", "enforce_origin": false, "origins": ["//[::1]:2019", "//127.0.0.1:2019", "//localhost:2019"]}
2024/03/07 15:26:02.392 INFO    autosaved config (load with --resume flag)      {"file": "/Users/rajatjindal/Library/Application Support/Caddy/autosave.json"}
2024/03/07 15:26:02.402 INFO    admin.api       received request        {"method": "POST", "host": "localhost:2019", "uri": "/load", "remote_ip": "127.0.0.1", "remote_port": "56843", "headers": {"Accept-Encoding":["gzip"],"Content-Length":["836"],"Content-Type":["application/json"],"User-Agent":["Go-http-client/1.1"]}}
2024/03/07 15:26:02.402 INFO    redirected default logger       {"from": "stderr", "to": "discard"}
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.833080210Z 
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841354835Z Serving http://0.0.0.0:80
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841360710Z Available Routes:
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841361501Z   hello: http://0.0.0.0:80/hello
[pod/simple-app-6c4fc84dd6-b49lb/simple-app] 2024-03-07T14:27:31.841363043Z   go-hello: http://0.0.0.0:80/go-hello
```

provide namespace explicitly

```shell
-> kubectl reverse-proxy simple-app -n foo-bar
```

use different kubeconfig file

```shell
-> kubectl reverse-proxy simple-app --kubeconfig /path/to/different/kube/config
```


