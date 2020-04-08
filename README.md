# Quay Github Actions Dispatch

A mini-service for securely forwarding Quay [build notifications](https://docs.quay.io/guides/notifications.html) to Github Action's  `repository_dispatch` webhook.

It allows you to combine the speed and caching efficiency of Quay.io for building your production Docker images, and the "I don't ever need another CI/CD tool" promise of Github Actions.

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
       |  Triggers       |                    notification |
       |  Github Actions |                    over HTTPS   |
       |  workflow👌     |                                 |
       |                                                   |
       |               +-----------------+                 |
       +---------------+ This service :) +<----------------+
                       +-----------------+

                        Verifies authenticity
                        through Quay's SSL cert 🔑
```

 use Quay.io for building your Docker images on a push to your GitHub branch, and then receiving notification in the same repository that the build has succeeded, triggering any workflow you want (e.g., update Helm chart to deploy new image to production Kubernetes cluster).
