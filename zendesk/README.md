# Zendesk Source for Knative

The Zendesk Source enables integration of [Zendesk](https://www.zendesk.com/) events with Knative, allowing end-users the ablility to subscribe new `Ticket` events.

## Contents

- [Zendesk Source for Knative](#zendesk-source-for-knative)
  - [Contents](#contents)
  - [Building](#building)
  - [Deploy a Controller](#deploy-a-controller)
    - [Deploy a Zendesk Source Controller From Code](#deploy-a-zendesk-source-controller-from-code)
  - [Create Zendesk Integration](#create-zendesk-integration)
  - [Deploy a Zendesk Source](#deploy-a-zendesk-source)
    - [Verify a Zendesk Source Deployment](#verify-a-zendesk-source-deployment)
    - [Customizing the integration](#customizing-the-integration)
  - [Events](#events)
  - [Support](#support)

## Building

The entry point (`main` package) for the controller and target adapter are respectively under
`cmd/controller/` and `cmd/adapter/`. Both these programs can be built using
the Go toolchain from the `knative-sources/zendesk` directory:

```sh
$ make build
```

Binaries will be generated for your current OS and architecture inside the root repo `_output` thdirectory.

Those binaries can also be packaged as container images in order to run inside a Kubernetes cluster:


```sh
$ make image
```

To list the other 'make' functions:

```sh
$ make help
```

## Deploy a Controller

### Deploy a Zendesk Source Controller From Code

[ko](https://github.com/google/ko) provides a quick method to build from source and apply the associated Kuberneties configurations.

```sh
$ ko apply -f ./config/
```

Alternatively you can base on the manifests at the config repo to build a set of kubernetes manifests that use your customized images and namespace.

## Deploy a Zendesk Source

An instance of the Zendesk Source is created by applying a manifest that fullfills its CRD schema. Accepted Spec parameters are:

- `subdomain` for the Zendesk tenant being used.
- `email` associated with THE Zendesk account.
- `token` generated from Zendesk admin site for the integration.
- `webhookUsername` that will be used to verify event callbacks.
- `webhookPassword` that will be used to verify event callbacks.

All parameters are required.

Note that `webhookUsername` and `webhookPassword` are arbitrary values and will be used from zendesk to sign requests, and at the Zendesk source to verify them.

Example Secret Deployment:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: zendesksource
type: Opaque
stringData:
  token: 'tHpUJ2ieiXsxEvBotczR99EwpETeQOiUU07KovBJ'
  password: 'Pa$$sw0rd'
```

Example Source Deployment:

```yaml
apiVersion: sources.triggermesh.io/v1alpha1
kind: ZendeskSource
metadata:
  name: zendesksource
spec:
  email: coyote@acmeanvils.com
  subdomain: 'acmeanvils'
  token:
    secretKeyRef:
      name: zendesksource
      key: token
  webhookUsername: 'webhookuser'
  webhookPassword:
    secretKeyRef:
      name: zendesksource
      key: webhookPassword
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-display

```

The example relies on an `event-display` service and on the `zendesksource` secret that should contains `token` and `webhookPassword` keys.

## Events

An example Cloudevent from a Zendesk Source:

```yaml
☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.zendesk.ticket.created
  source: triggermesh.zendesksource-zd-pr-test
  id: 0bb02a50-e4ee-4b69-b612-0a03d6e78267
  time: 2020-07-20T13:02:30.332708062Z
  datacontenttype: application/json
Data,
  {
    "current_user": {
      "details": "",
      "email": "dev@triggermesh.com",
      "external_id": "",
      "first_name": "Triggermesh",
      "language": "English",
      "name": "Triggermesh Developer",
      "notes": "",
      "organization": {
        "details": "",
        "name": "",
        "notes": ""
      },
      "phone": ""
    },
    "satisfaction": {
      "current_comment": "",
      "current_rating": ""
    },
    "ticket": {
      "account": "TriggerMesh",
      "assignee": {
        "email": "dev@triggermesh.com",
        "first_name": "Triggermesh",
        "last_name": "Developer",
        "name": "Triggermesh Developer"
      },
      "brand_name": "TriggerMesh",
      "cc_names": "",
      "ccs": "[]",
      "current_holiday_name": "Liquid error: internal",
      "description": "----------------------------------------------\n\nTriggermesh Developer, Jul 20, 2020, 10:02 AM\n\ne s t",
      "due_date": "",
      "external_id": "",
      "group_name": "Support",
      "id": "86",
      "organization": {
        "details": "",
        "external_id": "",
        "name": "",
        "notes": ""
      },
      "priority": "",
      "requester": {
        "details": "",
        "email": "dev@triggermesh.com",
        "external_id": "",
        "field": "",
        "first_name": "Triggermesh",
        "language": "English",
        "last_name": "Developer",
        "name": "Triggermesh Developer",
        "phone": ""
      },
      "status": "Open",
      "tags": "",
      "ticket_field_id": "",
      "ticket_field_option_title_id": "",
      "ticket_type": "Ticket",
      "title": "t",
      "url": "triggermesh.zendesk.com/agent/tickets/86",
      "via": "Web Form"
    }
  }
```
## Support

This is heavily **Work In Progress** We would love your feedback on this
Operator so don't hesitate to let us know what is wrong and how we could improve
it, just file an [issue](https://github.com/triggermesh/knative-sources/issues/new)

