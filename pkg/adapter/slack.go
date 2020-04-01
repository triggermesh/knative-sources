package adapter

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// Processor parse slack events and turns them into cloud events
type processor interface {
	Run(stopCh <-chan struct{})
}

// NewProcessor for incoming slack events
func newProcessor(token string, logger *zap.SugaredLogger, ceCh chan<- cloudevents.Event) processor {
	c := slack.New(token)

	return &defaultProcessor{
		client: c,
		logger: logger,
		ceCh:   ceCh,
	}
}

type defaultProcessor struct {
	client *slack.Client
	logger *zap.SugaredLogger
	ceCh   chan<- cloudevents.Event
}

func (p *defaultProcessor) Run(stopCh <-chan struct{}) {
	rtm := p.client.NewRTM()
	defer rtm.Disconnect()
	go rtm.ManageConnection()

	for {
		select {
		case in := <-rtm.IncomingEvents:

			switch in.Type {
			case "message":
				ev, ok := in.Data.(*slack.MessageEvent)
				if !ok {
					p.logger.Errorf("Slack event cannot be converted to slack.MessageEvent: %v", in.Data)
					continue
				}
				ce := p.cloudEventFromMessage(ev)
				p.ceCh <- *ce
			case "connecting":
				ev, ok := in.Data.(*slack.ConnectingEvent)
				if !ok {
					p.logger.Errorf("Slack event cannot be converted to slack.ConnectingEvent: %v", in.Data)
					continue
				}
				p.logger.Infof("connecting to Slack attempt %d", ev.ConnectionCount)

			case "connected":
				ev, ok := in.Data.(*slack.ConnectedEvent)
				if !ok {
					p.logger.Errorf("Slack event cannot be converted to slack.ConnectedEvent: %v", in.Data)
					continue
				}
				p.logger.Infof("connected to Slack as %q", ev.Info.User.Name)
			}

		case <-stopCh:
			p.logger.Infof("received stop signal")

			return
		}
	}
}

type messageEvent struct {
	Text string `json:"text"`
}

func (p *defaultProcessor) cloudEventFromMessage(message *slack.MessageEvent) *cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(message.ClientMsgID)
	event.SetSource("com.slack.api.channel/" + message.Channel)
	event.SetSubject(message.User + " / " + message.Username)
	event.SetData(cloudevents.ApplicationJSON, &messageEvent{
		Text: message.Text,
	})
	event.SetType("dev.knative.sources.slack/new-message")
	return &event
}
