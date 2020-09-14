package logic

import (
	"github.com/zinic/forculus/config"
	"github.com/zinic/forculus/eventserver/actor"
	"github.com/zinic/forculus/eventserver/email"
	"github.com/zinic/forculus/eventserver/event"
	"github.com/zinic/forculus/log"
)

func NewEmailSender(name string, cfg config.Emailer, server config.SMTPServer) actor.Subscriber {
	emailSender := EmailSender{
		name:   name,
		cfg:    cfg,
		server: server,
	}

	return actor.NewSubscriber(emailSender.Logic)
}

type EmailSender struct {
	name   string
	cfg    config.Emailer
	server config.SMTPServer
}

func (s *EmailSender) Logic(eventC chan event.Event, exitC chan struct{}) {
	for {
		select {
		case nextEvent := <-eventC:
			emailTemplate := nextEvent.Payload.(email.Email)
			emailTemplate.Recipients = s.cfg.Recipients

			if err := email.Send(emailTemplate, s.server); err != nil {
				log.Errorf("Emailer %s failed to send email: %v", s.name, err)
			}

		case <-exitC:
			return
		}
	}
}
