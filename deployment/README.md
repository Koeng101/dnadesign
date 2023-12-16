# DnaDesign Deployments

This directory contains deployment infrastructure for DnaDesign services. Tests use docker [testcontainers](https://testcontainers.com/) running [k3s](https://k3s.io/). Deployment is on a single-host k3s server. Even though we emphasize maintainability over all else, in practice, this means using a standard deployment platform that many people can understand. Kubernetes is used as this standard platform because it can largely be defined in software upfront and rapidly re-deployed.

Test take about 50s-100s to run.
