package main

import (
	"encoding/json"
	"strings"

	sdk "github.com/opensourceways/go-gitee/gitee"
	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/opensourceways/robot-gitee-lib/client"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	ciSuccessFull   = "ci_successful"
	ciFailed        = "ci_failed"
	ciCommentPrefix = "<table><tr><th>Check Name"
)

type iClient interface {
	ListPRComments(org, repo string, number int32) ([]sdk.PullRequestComments, error)
}

type eventHandler struct {
	cli iClient
	cfg *Config
}

func NewEventHandler(cfg *Config) (*eventHandler, error) {
	c := client.NewClient(func() []byte {
		return []byte(cfg.Token)
	})

	return &eventHandler{
		cli: c,
		cfg: cfg,
	}, nil
}

func (e *eventHandler) HandlePREvent(event *sdk.PullRequestEvent) error {
	if event.GetState() != sdk.StatusOpen || event.GetActionDesc() != sdk.PRActionUpdatedLabel {
		return nil
	}

	org, repo := event.GetOrgRepo()
	if org != e.cfg.Repository.Org || repo != e.cfg.Repository.Repo {
		return nil
	}

	ciLabelSets := sets.NewString(ciFailed, ciSuccessFull)
	result := event.GetPRLabelSet().Intersection(ciLabelSets)
	if len(result) == 0 {
		return nil
	}

	msg := e.buildMsgPkgCIChecked(
		event,
		e.getCiComment(event),
		result.UnsortedList()[0],
	)

	return e.sendMsgCIChecked(msg)
}

func (e *eventHandler) getCiComment(event *sdk.PullRequestEvent) string {
	org, repo := event.GetOrgRepo()
	comments, err := e.cli.ListPRComments(org, repo, event.GetPRNumber())
	if err != nil {
		logrus.Errorf("get pr comments error: %s", err.Error())

		return ""
	}

	// Iterate from the end to get the latest ci comment
	for i := len(comments) - 1; i >= 0; i-- {
		comment := comments[i]
		if comment.User.Login != e.cfg.CIRobotName {
			continue
		}

		if strings.HasPrefix(comment.Body, ciCommentPrefix) {
			return comment.Body
		}
	}

	return ""
}

type msgPkgCIChecked struct {
	PkgId    string `json:"pkg_id"`
	Detail   string `json:"detail"`
	PRNumber int    `json:"number"`
	Success  bool   `json:"success"`
}

func (e *eventHandler) buildMsgPkgCIChecked(event *sdk.PullRequestEvent, comment, label string) msgPkgCIChecked {
	msg := msgPkgCIChecked{
		PkgId:    event.PullRequest.Body,
		PRNumber: int(event.GetPRNumber()),
	}

	switch label {
	case ciSuccessFull:
		msg.Success = true
		msg.Detail = "The CI is successful."

	case ciFailed:
		msg.Success = false
		msg.Detail = "The CI is failed."
	default:
	}

	if comment != "" {
		msg.Detail = comment
	}

	return msg
}

func (e *eventHandler) sendMsgCIChecked(msg msgPkgCIChecked) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = kafka.Publish(e.cfg.Topics.SoftwarePkgCIChecked, nil, body)
	if err == nil {
		logrus.Infoln("send success to", e.cfg.Topics.SoftwarePkgCIChecked, string(body))
	}

	return err
}
