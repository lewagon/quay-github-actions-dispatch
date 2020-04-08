# Quay Github Actions Dispatch

A mini-service for securely forwarding Quay [build notifications](https://docs.quay.io/guides/notifications.html) to Github Action's  `repository_dispatch` webhook.

It allows you to use Quay.io for building your Docker images on a push to your GitHub branch, and then receiving notification in the same repository that the build has succeeded, triggering any workflow you want (e.g., update Helm chart to deploy new image to production Kubernetes cluster). 
