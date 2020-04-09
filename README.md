# Quay Github Actions Dispatch

A mini-service for securely forwarding Quay [build notifications](https://docs.quay.io/guides/notifications.html) to Github Actions' [obscure](https://help.github.com/en/actions/reference/events-that-trigger-workflows#external-events-repository_dispatch) `repository_dispatch` webhook to trigger any workflow in your repository _remotely_ on successful image build.

It combines the speed and caching efficiency of Quay.io for building your production Docker images with the _"I don't ever need another CI/CD tool"_ promise of Github Actions.

Here's how it works:

```
+-------------+                                      +---------+
| GitHub Repo +--------+build trigger👷‍+------------>+ Quay.io |
+------+------+                                      +-----+---+
       ^                                                   |
       |                                                   |
       |                                                   |
       |                                      Webhook      |
       |                                      success      |
       |  Triggers                            notification |
       |  Github Actions                      over HTTPS   |
       |  workflow👌                                       |
       |                                                   |
       |               +-----------------+                 |
       +---------------+ This service :) +<----------------+
                       +-----------------+

                        Verifies authenticity
                        through Quay's SSL cert 🔑
```


## Usage

### Running your own quay-github-actions-dispatch endpoint

It is recommended to run this service on a dedicated VPS with **no load balancer or reverse proxy** and an open 443 port.
You should also obtain a certificate for a _domain name you will put your endpoint_ on as Quay needs to send you a webhook over HTTPS (otherwise you will not be able to authenticate through SSL) and point this domain name to your VPS' IP address.

:warn: `quay-github-actions-dispatch` performs it's own SSL termination to read Quay's certificate and it needs access to relevant `.pem` and `.key` files.

On the server with a 443 port exposed and certificates for domain name set up:

```sh
 docker run -d --rm -p 443:443 \
 -v /certs:/certs \                # /certs folder on your host needs to have .key and .pem files
 -e GH_TOKEN=YOUR_GITHUB_TOKEN \   # token needs only `repo:read` access
 -e DEBUG=true \                  # optional to log incoming and outgoing requests and responses
 lewagon/quay-github-actions-dispatch:0.2
```
In your Quay repository settings, under "Events and Notifications" create a "Webhook POST" notification for "Dockerfile Build Successfully Completed" and provide `https://YOUR_CERTIFIED_DOMAIN.com/incoming` endpoint.

### Incoming payload

```json
{
  "build_id": "fake-build-id",
  "trigger_kind": "GitHub",
  "name": "your_repo_name",
  "repository": "your_github/your_repo_name",
  "namespace": "your_repo_name",
  "docker_url": "quay.io/your_github/your_repo_name",
  "trigger_id": "1245634",
  "docker_tags": [
    "latest",
    "foo",
    "bar"
  ],
  "build_name": "some-fake-build",
  "image_id": "1245657346",
  "trigger_metadata": {
    "default_branch": "master",
    "ref": "refs/heads/somebranch",
    "commit": "42d4a62c53350993ea41069e9f2cfdefb0df097d"
  },
  "homepage": "https://quay.io/your_github/your_repo_name/your_repo_name/build/fake-build-id"
}
```

### Outgoing payload

Will be sent to `https://api.github.com/repos/your_github/your_repo/dispatches`

```json
{
  "event_type": "QUAY_BUILD_SUCCESS",
  "client_payload": {
    "text": "42d4a62"
  }
}
```

`"text"` field will contain first 7 chars of your commit SHA from the incoming payload

### Workflow example to trigger on QUAY_BUILD_SUCCESS event

Deploy to Digital Ocenan managed Kubernetes with Helm 3

```yml
name: Helm deploy all things
on:
  repository_dispatch:
    types: [QUAY_BUILD_SUCCESS]

jobs:
  build:
    if: startsWith(github.sha, github.event.client_payload.text)
    name: Helm update
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master

      - name: Save DigitalOcean kubeconfig
        uses: digitalocean/action-doctl@master
        env:
          DIGITALOCEAN_ACCESS_TOKEN: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
        with:
          args: kubernetes cluster kubeconfig show CLUSTER_NAME > $GITHUB_WORKSPACE/.kubeconfig

      - name: Upgrade/install chart
        run: |
        export KUBECONFIG=$GITHUB_WORKSPACE/.kubeconfig && \
        helm upgrade linkedin charts/linkedin --install \
        --atomic --cleanup-on-fail \
        --set-string image.tag=$(echo $GITHUB_SHA | head -c7) \
```

## Why?

* Containers are awesome. Kubernetes makes them even more awesome.
* Github Actions are awesome, but it's not possible to cache Docker builds, so bigger production images can take long (10 mins+) to build.
* Quay is awesome: easy to trigger a build from a push to any branch on GitHub and label an image either with a branch name or a commit SHA. Builds are cached and it takes very little time to build a new image, if the only layer changed is your `COPY . /myawesomeapp`. There is even a possibility to send a POST request to any webhook if the build has started/succeeded/failed. However, there is no control over headers or a body of that POST request.
* Github Actions workflows can be triggered by the external [repository_dispatch](https://help.github.com/en/actions/reference/events-that-trigger-workflows#external-events-repository_dispatch) event, but GitHub HTTP API  expects very certain payload with very certain headers.
* Imagine the world where Quay could send a webhook POST directly to GitHub and trigger a workflow in your repo if the image was successfully built.

`quay-github-actions-dispatch` is exactly that :point_up: missing link.

## How

Easy, you think, just let me code an endpoint that receives a webhook payload from Quay and transforms it into a webhook request for GitHub.

_Not so fast!_

Quay notifications bare no authentication, so _anyone_ simulating the Quay payload could trigger events directly in your repo and this is definitely not what you want. However:

> When the URL is HTTPS, the call will have an SSL client certificate set from Quay.io. Verification of this certificate will prove the call originated from Quay.io. —Quay.io Repository Notifications [Manual](https://docs.quay.io/guides/notifications.html)

This is exactly what `quay-github-actions-dispatch` does: **verifies that Quay is indeed Quay by looking at it's TLS peer certificate**

## Build your own Docker image from source

Use latest versions of Go with built-in [modules](https://github.com/golang/go/wiki/Modules#example) support

```sh
git clone git@github.com:lewagon/quay-github-actions-dispatch.git
cd quay-github-actions-dispatch
GOOS=linux GOARCH=amd64 go build .  # assuming you use a Linux server
docker build -t IMAGE:TAG .
```
