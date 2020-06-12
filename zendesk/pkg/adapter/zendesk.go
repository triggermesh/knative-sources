package adapter

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"
)

// Processor parse zendesk events and turns them into cloud events
type processor interface {
	Run(stopCh <-chan struct{})
}

// NewProcessor for incoming zendesk events
func newProcessor(token string, logger *zap.SugaredLogger, ceCh chan<- cloudevents.Event) processor {
	c := zendesk.New(token)

	return &defaultProcessor{
		client: c,
		logger: logger,
		ceCh:   ceCh,
	}
}

type defaultProcessor struct {
	client *zendesk.Client
	logger *zap.SugaredLogger
	ceCh   chan<- cloudevents.Event

	user   string
	domain string
}

func (p *defaultProcessor) Run(stopCh <-chan struct{}) {
	rtm := p.client.NewRTM()
	defer func() {
		if err := rtm.Disconnect(); err != nil {
			p.logger.Errorw("Error disconnecting", zap.Error(err))
		}
	}()

	go rtm.ManageConnection()

	for {
		select {
		case in := <-rtm.IncomingEvents:
			switch event := in.Data.(type) {

			case *zendesk.MessageEvent:
				ce := p.cloudEventFromMessage(event)
				p.ceCh <- *ce

			case *zendesk.ConnectingEvent:
				p.logger.Infof("connecting to Zendesk attempt %d", event.ConnectionCount)

			case *zendesk.ConnectedEvent:
				p.user = event.Info.User.Name
				p.domain = event.Info.Team.Domain
				p.logger.Infof("connected to Zendesk as %q", event.Info.User.Name)

			case *zendesk.ConnectionErrorEvent:
				p.logger.Error("Error connecting to Zendesk, retrying in %s: %e", event.ErrorObj)

			case *zendesk.InvalidAuthEvent:
				p.logger.Error("Bot credentials are not valid")
				return

			case *zendesk.DisconnectedEvent:
				if !event.Intentional {
					p.logger.Warn("Bot was disconnected from Zendesk", zap.Error(event.Cause))
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

func (p *defaultProcessor) cloudEventFromMessage(message *zendesk.MessageEvent) *cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(message.ClientMsgID)
	event.SetSource(fmt.Sprintf("com.zendesk.%s", p.domain))
	event.SetSubject(message.Channel)
	if err := event.SetData(cloudevents.ApplicationJSON, &messageEvent{UserID: message.User, Text: message.Text}); err != nil {
		p.logger.Errorw("error setting data at cloud event", zap.Error(err))
	}
	event.SetType("com.zendesk/message")
	return &event
}
