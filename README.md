# putio-operator

Manage your [Put.io](https://app.put.io/) RSS feeds with Kubernetes Custom Resources.

[![Go](https://github.com/SkYNewZ/putio-operator/actions/workflows/main.yml/badge.svg)](https://github.com/SkYNewZ/putio-operator/actions/workflows/main.yml)

## Usage

Before creating an RSS feed resource, you must configure a secret containing an OAuth2 token to interact with Put.io
API. You can get one here https://app.put.io/account/api/apps/new.

Then, place the OAuth2 token into a Kubernetes secret. For example :

```
kubectl -n default create secret generic putio-token --from-literal=token=<your oauth2 token>
```

Then, you can create RSS feed by specifying this secret. You can see examples [here](config/samples/_v1alpha1_feed.yaml)
.

```yaml
apiVersion: skynewz.dev/v1alpha1
kind: Feed
metadata:
  name: house-of-the-dragons
  namespace: default
spec:
  keyword: "House.of.the.Dragon.S01E&.MULTi.1080p.WEB.H264-FW"
  rss_source_url: "https://rss.site.fr/rss?id=2184"
  delete_old_files: false
  title: "House of the Dragon"
  paused: true
  parent_dir_id: 1022542820 # 'House of the Dragon' folder
  dont_process_whole_feed: true
  authSecretRef:
    key: putio-token
    name: token

```

The `kubectl get feeds` output will give you

```
NAME                           KEYWORD                                                                        LAST FETCH   PAUSED   TITLE                           AGE
house-of-the-dragons           House.of.the.Dragon.S01E&.MULTi.1080p.WEB.H264-FW                              20h          true     House of the Dragon             2d2h

```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for
testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever
cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/putio-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/putio-operator:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

## License

Copyright 2022 Quentin Lemaire <quentin@lemairepro.fr>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
