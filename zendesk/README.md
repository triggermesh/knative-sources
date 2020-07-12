# Zendesk Source for Knative

Zendesk Source enables integration between zendesk messages using the Events API and Knative Eventing.

## Contents
- [Zendesk Source for Knative](#zendesk-source-for-knative)
  - [Contents](#contents)
  - [Building](#building)
  - [Deploy controller](#deploy-controller)
    - [Deploy Zendesk Source Controller](#deploy-zendesk-source-controller)
  - [Create Zendesk Integration](#create-zendesk-integration)
    - [Deploy Zendesk Source](#deploy-zendesk-source)
    - [Configure Zendesk Events API App](#configure-zendesk-events-api-app)
    - [Secure the Zendesk Source](#secure-the-zendesk-source)
  - [Events](#events)
  - [Support](#support)

## Building

###### The entry point (`main` package) for the controller and target adapterunder `cmd/controller/` and `cmd/adapter/,` respectively . Both these programs can be built using the Go toolchain

To create binaries for your current OS and architecture inside the root repo `_output` directory: 
```sh
knative-sources/zendesk$ make build
```

To create container images:
 
```sh
knative-sources/zendesk$ make image
```

To list the other 'make' functions:

```sh
$ make help
```

## Deploy controller

### Deploying a Zendesk Source Controller From Code

To apply the associated Kuberneties configurations and build from source. [ko](https://github.com/google/ko) 

```sh
knative-sources/zendesk$ ko apply -f ./config/
```

Alternatively you can base on the manifests at the config repo to build a set of kubernetes manifests that use your customized images and namespace.

### Deploy Zendesk Source

An instance of the Zendesk Source is created by applying a manifest that fullfills its CRD schema. Accepted Spec parameters are:

- `email` : The email associated with a valid Zendesk account. 
- `username` : Used for basic authentication between Zendesk and the Source
- `password` : Used for basic authentication between Zendesk and the Source
- `subdomain` : The Zendesk Subdomain 

Example:

```yaml
apiVersion: sources.triggermesh.io/v1alpha1
kind: ZendeskSource
metadata:
  name: zendesksource
spec:
  email: '' # Example: jeffthenaef@gmail.com 
  username: '' # Example: jeff 
  subdomain: '' # Example: tmdev2 
  token:
            secretKeyRef:
              name: zendesksource
              key: token
  password:
            secretKeyRef:
              name: zendesksource
              key: password            
  ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-display

```

Once created wait for the source to be ready and take note of the URL (`status.address.url`):

``` sh
 kubectl get zendesksource -n odacremolbap zendesk-source
NAME                READY   REASON   URL                                                              SINK                                                  AGE
zendesksource       True             https://zendesksource-triggermesh.odacremolbap.dev.munu.io      http://event-display.odacremolbap.svc.cluster.local    25h
```

## Events

Below you can find an example Cloudevent from a Zendesk Source.

cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.zendesk.new
  source: jeffthenaef.zendesksource-zsrc.tmdev2
  subject: New Zendesk Ticket
  id: 62
  time: 2020-07-12T05:15:43.43054774Z
  datacontenttype: application/json
Data,
  {
    "id": "62",
    "description": "----------------------------------------------\n\njeff naef, Jul 12, 2020, 2:15 AM\n\nHello World",
    "created_at": "0001-01-01T00:00:00Z",
    "due_at": "0001-01-01T00:00:00Z",
    "via": {
      "source": {}
    },
    "satisfaction_rating": {}
  }
