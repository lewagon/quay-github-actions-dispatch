# Quay Github Actions Dispatch

A mini-service for securely forwarding Quay [build notifications](https://docs.quay.io/guides/notifications.html) to Github Actions' [obscure](https://help.github.com/en/actions/reference/events-that-trigger-workflows#external-events-repository_dispatch) `repository_dispatch` webhook to trigger any workflow in your repository _remotely_ on successful image build.

It allows you to combine the speed and caching efficiency of Quay.io for building your production Docker images with the "I don't ever need another CI/CD tool" promise of Github Actions.

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


> When the URL is HTTPS, the call will have an SSL client certificate set from Quay.io. Verification of this certificate will prove the call originated from Quay.io. —Quay.io Repository Notifications [Doc](https://docs.quay.io/guides/notifications.html)
