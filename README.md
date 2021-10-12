**This project is now archived, please go to [triggermesh/triggermesh](https://github.com/triggermesh/triggermesh) repository for additional development.**

# TriggerMesh Sources for Knative

A collection of open Knative Sources maintained by TriggerMesh.

Other Knative Sources maintained by TriggerMesh are available in the following repositories:

- [AWS Sources][knsrc-aws]
- [GitLab Source][knsrc-gitlab] (Knative sandbox project)

## Sources

The following event sources are provided by this controller:

| Source                    | Support level |
|---------------------------|---------------|
| [HTTP][docs-http]         | DEPRECATED    |
| [Webhook][docs-webhook]   | alpha         |
| [Slack][docs-slack]       | alpha         |
| [Zendesk][docs-zd]        | alpha         |

## Contributions and support

We would love to hear your feedback on these sources. Please don't hesitate to submit bug reports and suggestions by
[filing issues][gh-issue], or contribute by [submitting pull-requests][gh-pr].

Refer to [DEVELOPMENT.md](./DEVELOPMENT.md) in order to get started.

## TriggerMesh Cloud Early Access

TriggerMesh Knative Sources can be used as is from this repo. You can also use them along with other components from our
Cloud at <https://cloud.triggermesh.io>, which has a web UI to configure and run them.

## Commercial Support

TriggerMesh Inc. supports those sources commercially. Email us at <info@triggermesh.com> to get more details.

## Code of Conduct

Although this project is not part of the [CNCF][cncf], we abide by its [code of conduct][cncf-conduct].

[knsrc-aws]: https://github.com/triggermesh/aws-event-sources
[knsrc-gitlab]: https://github.com/knative-sandbox/eventing-gitlab

[docs-http]: https://docs.triggermesh.io/sources/http/
[docs-webhook]: https://docs.triggermesh.io/sources/webhook/
[docs-slack]: https://docs.triggermesh.io/sources/slack/
[docs-zd]: https://docs.triggermesh.io/sources/zendesk/

[gh-issue]: https://github.com/triggermesh/knative-sources/issues
[gh-pr]: https://github.com/triggermesh/knative-sources/pulls

[cncf]: https://www.cncf.io/
[cncf-conduct]: https://github.com/cncf/foundation/blob/master/code-of-conduct.md
