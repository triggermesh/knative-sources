package adapter

import (
	"fmt"

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

	user   string
	domain string
}

func (p *defaultProcessor) Run(stopCh <-chan struct{}) {
	rtm := p.client.NewRTM()
	defer rtm.Disconnect()
	go rtm.ManageConnection()

	for {
		select {
		case in := <-rtm.IncomingEvents:
			switch event := in.Data.(type) {

			case *slack.MessageEvent:
				ce := p.cloudEventFromMessage(event)
				p.ceCh <- *ce

			case *slack.ConnectingEvent:
				p.logger.Infof("connecting to Slack attempt %d", event.ConnectionCount)

			case *slack.ConnectedEvent:
				p.user = event.Info.User.Name
				p.domain = event.Info.Team.Domain
				p.logger.Infof("connected to Slack as %q", event.Info.User.Name)

			case *slack.ConnectionErrorEvent:
				p.logger.Error("Error connecting to Slack, retrying in %s: %e", event.ErrorObj)

			case *slack.InvalidAuthEvent:
				p.logger.Error("Bot credentials are not valid")
				return

			case *slack.DisconnectedEvent:
				if !event.Intentional {
					p.logger.Warn("Bot was disconnected from Slack", zap.Error(event.Cause))
				}
			}

		case <-stopCh:
			p.logger.Infof("received stop signal")
			return
		}
	}
}

type messageEvent struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func (p *defaultProcessor) cloudEventFromMessage(message *slack.MessageEvent) *cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(message.ClientMsgID)
	event.SetSource(fmt.Sprintf("com.slack.%s", p.domain))
	event.SetSubject(message.Channel)
	event.SetData(cloudevents.ApplicationJSON, &messageEvent{
		UserID: message.User,
		Text:   message.Text,
	})
	event.SetType("dev.knative.sources.slack/" + message.Type)
	return &event
}
