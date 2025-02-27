---
title: "Registry Mirror configuration"
linkTitle: "Registry Mirror"
weight: 90
description: >
  EKS Anywhere cluster yaml specification for registry mirror configuration
---

## Registry Mirror Support (optional)
You can configure EKS Anywhere to use a private registry as a mirror for pulling the required images.

The following cluster spec shows an example of how to configure registry mirror:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
  registryMirrorConfiguration:
    endpoint: <private registry IP or hostname>
    caCertContent: |
      -----BEGIN CERTIFICATE-----
      MIIF1DCCA...
      ...
      es6RXmsCj...
      -----END CERTIFICATE-----  
```
## Registry Mirror Configuration Spec Details
### __registryMirrorConfiguration__ (required)
* __Description__: top level key; required to use a private registry.
* __Type__: object

### __endpoint__ (required)
* __Description__: IP address or hostname of the private registry for pulling images
* __Type__: string
* __Example__: ```endpoint: 192.168.0.1```
### __caCertContent__ (optional)
* __Description__: Certificate Authority (CA) Certificate for the private registry . When using 
  self-signed certificates it is necessary to pass this parameter in the cluster spec.<br/>
  It is also possible to configure CACertContent by exporting an environment variable:<br/>
  `export EKSA_REGISTRY_MIRROR_CA="/path/to/certificate-file"`
* __Type__: string
* __Example__: <br/>
  ```yaml
  CACertContent: |
    -----BEGIN CERTIFICATE-----
    MIIF1DCCA...
    ...
    es6RXmsCj...
    -----END CERTIFICATE-----
  ```

## Import images into a private registry
You can use the `import-images` command to pull images from `public.ecr.aws` and push them to your
private registry.

```bash
docker login https://<private registry endpoint>
...
eksctl anywhere import-images -f cluster-spec.yaml
```
## Docker configurations
It is necessary to add the private registry's CA Certificate
to the list of CA certificates on the admin machine if your registry uses self-signed certificates.

For [Linux](https://docs.docker.com/engine/security/certificates/), you can place your certificate here: `/etc/docker/certs.d/<private-registry-endpoint>/ca.crt`

For [Mac](https://docs.docker.com/desktop/mac/#add-tls-certificates), you can follow this guide to add the certificate to your keychain: https://docs.docker.com/desktop/mac/#add-tls-certificates

{{% alert title="Note" color="primary" %}}
  You may need to restart Docker after adding the certificates.
{{% /alert %}}

## Registry configurations
Depending on what registry you decide to use, you will need to create the following projects:

```
bottlerocket
eks-anywhere
eks-distro
isovalent
```

For example, if a registry is available at `private-registry.local`, then the following 
projects will have to be created:

```
https://private-registry.local/bottlerocket
https://private-registry.local/eks-anywhere
https://private-registry.local/eks-distro
https://private-registry.local/isovalent
```
