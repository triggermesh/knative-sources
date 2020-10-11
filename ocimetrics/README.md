# Oracle Cloud Infrastructure Metrics Source for Knative

The OciMetrics Source enables integration of [Zendesk](https://www.zendesk.com/) events with Knative, allowing end-users the ablility to subscribe new `Ticket` events.

## Contents

- [Oracle Cloud Infrastructure Metrics Source for Knative](#oracle-cloud-infrastructure-metrics-source-for-knative)
  - [Contents](#contents)
  - [Building](#building)
  - [Deploy a Controller](#deploy-a-controller)
    - [Deploy an OCI Metrics Source Controller From Code](#deploy-an-oci-metrics-source-controller-from-code)
  - [Create OCI Metrics Integration](#create-oci-metrics-integration)
  - [Deploy an OCI Metrics Source](#deploy-an-oci-metrics-source)
    - [Verify an OCI Metrics Source Deployment](#verify-an-oci-metrics-source-deployment)
    - [Customizing the integration](#customizing-the-integration)
  - [Support](#support)

## Building

The entry point (`main` package) for the target adapter is under `cmd/adapter/`, and can be built
only from the root of the `knative-sources` directory:

```sh
$ make build
```

Binaries will be generated for your current OS and architecture inside the root repo `_output` directory.

Those binaries can also be packaged as container images in order to run inside a Kubernetes cluster:


```sh
$ make image
```

To list the other 'make' functions:

```sh
$ make help
```

## Deploy a Controller

### Deploy an OCI Metrics Source Controller From Code

[ko](https://github.com/google/ko) provides a quick method to build from source and apply the associated Kubernetes configurations.

```sh
$ ko apply -f ./config/
```

Alternatively you can base on the manifests at the config repo to build a set of kubernetes manifests that use your customized images and namespace.

## Deploy an OCI Metrics Source

An instance of the OCI Metrics Source is created by applying a manifest that fullfills its CRD schema. Accepted Spec parameters are:

- `oracleApiPrivateKey` for interacting with the Oracle Cloud REST API
- `oracleApiPrivateKeyPassphrase` for unlocking the `oracleApiPrivateKey`
- `oracleApiPrivateKeyFingerprint` for associating the `oracleApiPrivateKey` with the correct Oracle Cloud user key 
- `oracleTenancy` for the Oracle Cloud tenant being used
- `oracleUser` for the user the `oracleApiPrivateKey` is associated with
- `oracleRegion` for the Oracle Cloud region to collect the metrics data from
- `metricsNamespace` for the type of [metrics](https://docs.cloud.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/MetricData) to use
- `metricsQuery` for the query to run against the Oracle Cloud API
- `metricsPollingFrequency` for how often to run the query

All parameters are required except for the `metricsPollingFrequency` which defaults to `5m`.

Example Secret Deployment:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oraclecreds
type: Opaque
stringData:
  apiPassphrase: ''
  apiKeyFingerprint: '5c:75:c4:67:92:a9:46:2a:01:5b:73:54:6a:b2:74:7d'
  apiKey: |-
    -----BEGIN RSA PRIVATE KEY-----
    MIXEpAIBACKCAQEA2UM2O2lz4D6gN2sAbxUg6VMnGQlrwNbZX7b/wqW6ZEU0Q0BU
    ...
    -----END RSA PRIVATE KEY-----
```

Example Source Deployment:

```yaml
apiVersion: sources.triggermesh.io/v1alpha1
kind: OciMetricsSource
metadata:
  name: metrics-test
spec:
  # required to interact with the Oracle Cloud API
  oracleApiPrivateKey:
    secretKeyRef:
      name: oraclecreds
      key: apiKey
  oracleApiPrivateKeyPassphrase:
    secretKeyRef:
      name: oraclecreds
      key: apiPassphrase
  oracleApiPrivateKeyFingerprint:
    secretKeyRef:
      name: oraclecreds
      key: apiKeyFingerprint
  oracleTenancy: ocid1.tenancy.oc1..aaaaaaaaswr
  oracleUser: ocid1.user.oc1..aaaaaaaaqloc
  oracleRegion: us-ashburn-1

  # required to enable metrics
  metricsNamespace: oci_computeagent
  metricsQuery: CPUUtilization[1m].mean()
  # optional polling frequency. default to 5m
  #metricsPollingFrequency: 5m

  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-display

```

The example relies on an `event-display` service and on the `oraclecreds` secret that should contain `apiKey`, `apiPassphrase`, and `apiKeyFingerprint` secrets.

## Support

This is heavily **Work In Progress** We would love your feedback on this
Operator so don't hesitate to let us know what is wrong and how we could improve
it, just file an [issue](https://github.com/triggermesh/knative-sources/issues/new)

