package actions

import (
	"github.com/aiseeq/s2l/lib/point"
	"github.com/aiseeq/s2l/protocol/api"
)

type Actions []*api.Action

func (a *Actions) ChatSend(msg string, channel api.ActionChat_Channel) {
	*a = append(*a, &api.Action{
		ActionChat: &api.ActionChat{
			Channel: channel,
			Message: msg,
		},
	})
}

func (a *Actions) MoveCamera(p point.Pointer) {
	*a = append(*a, &api.Action{
		ActionRaw: &api.ActionRaw{
			Action: &api.ActionRaw_CameraMove{
				CameraMove: &api.ActionRawCameraMove{
					CenterWorldSpace: p.Point().To3D(),
				},
			},
		},
	})
}
