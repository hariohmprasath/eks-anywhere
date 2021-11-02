
---
title: "EKS-A on vSphere with GitOps"
linkTitle: "GitOps on vSphere"
weight: 00
date: 2017-01-05
description: >
  End to end workshop for deploying EKS-A with GitOps on vSphere
---

## Provision a cluster

The first thing we need to do is provision a cluster.
Here is the getting started guide for provisioning.

{{< content "../../../getting-started/production-environment/_index.md" >}}

The next step will be to set up kube-vip to add a floating IP address.

[next]({{< relref "kube-vip.md" >}})