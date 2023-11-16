package main

import (
	"errors"
	"fmt"

	sdk "github.com/opensourceways/go-gitee/gitee"
)

const (
	msgHeaderUUID      = "X-Gitee-Timestamp"
	msgHeaderUserAgent = "User-Agent"
	msgHeaderEventType = "X-Gitee-Event"
)

type EventHandler interface {
	HandlePREvent(e *sdk.PullRequestEvent) error
}

type MessageServer struct {
	userAgent string
	handler   EventHandler
}

func NewMessageServer(h EventHandler, ua string) *MessageServer {
	return &MessageServer{
		userAgent: ua,
		handler:   h,
	}
}

func (m *MessageServer) handle(payload []byte, header map[string]string) error {
	eventType, err := m.parseRequest(header)
	if err != nil {
		return fmt.Errorf("invalid msg, err:%s", err.Error())
	}

	if eventType != sdk.EventTypePR {
		return nil
	}

	e, err := sdk.ConvertToPREvent(payload)
	if err != nil {
		return err
	}

	return m.handler.HandlePREvent(&e)
}

func (m *MessageServer) parseRequest(header map[string]string) (eventType string, err error) {
	if header == nil {
		err = errors.New("no header")

		return
	}

	if header[msgHeaderUserAgent] != m.userAgent {
		err = errors.New("unknown " + msgHeaderUserAgent)

		return
	}

	if eventType = header[msgHeaderEventType]; eventType == "" {
		err = errors.New("missing " + msgHeaderEventType)

		return
	}

	if header[msgHeaderUUID] == "" {
		err = errors.New("missing " + msgHeaderUUID)
	}

	return
}
