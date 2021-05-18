package actions

import "github.com/aiseeq/s2l/protocol/api"

type Actions []*api.Action

func (a *Actions) ChatSend(msg string) {
	*a = append(*a, &api.Action{
		ActionChat: &api.ActionChat{
			Channel: api.ActionChat_Broadcast,
			Message: msg,
		},
	})
}
