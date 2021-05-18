package tests

import (
	"bitbucket.org/aisee/minilog"
	"github.com/chippydip/go-sc2ai/api"
	"io/ioutil"
)

type AgentInfo struct {
	observation *api.ResponseObservation
	data        *api.ResponseData
	info        *api.ResponseGameInfo
}

func (a *AgentInfo) IsRealtime() bool {
	return false
}

func (a *AgentInfo) PlayerID() api.PlayerID {
	panic("Not Implemented")
}
func (a *AgentInfo) ReplayInfo() *api.ResponseReplayInfo {
	panic("Not Implemented")
}

func (a *AgentInfo) GameInfo() *api.ResponseGameInfo {
	return a.info
}
func (a *AgentInfo) LoadInfo(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	a.info = &api.ResponseGameInfo{}
	err = a.info.Unmarshal(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *AgentInfo) Data() *api.ResponseData {
	return a.data
}
func (a *AgentInfo) LoadData(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	a.data = &api.ResponseData{}
	err = a.data.Unmarshal(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *AgentInfo) Observation() *api.ResponseObservation {
	return a.observation
}
func (a *AgentInfo) LoadObservation(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	a.observation = &api.ResponseObservation{}
	err = a.observation.Unmarshal(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *AgentInfo) Upgrades() []api.UpgradeID {
	return nil
}
func (a *AgentInfo) HasUpgrade(upgrade api.UpgradeID) bool {
	return false
}

func (a *AgentInfo) IsInGame() bool {
	return true
}
func (a *AgentInfo) Step(stepSize int) error {
	return nil
}

func (a *AgentInfo) Query(query api.RequestQuery) *api.ResponseQuery {
	panic("Not Implemented")
}
func (a *AgentInfo) SendActions(actions []*api.Action) []api.ActionResult {
	panic("Not Implemented")
}
func (a *AgentInfo) SendObserverActions(obsActions []*api.ObserverAction) {
}
func (a *AgentInfo) SendDebugCommands(commands []*api.DebugCommand) {
}
func (a *AgentInfo) ClearDebugDraw() {
}
func (a *AgentInfo) LeaveGame() {
}

func (a *AgentInfo) OnBeforeStep(func()) {
}
func (a *AgentInfo) OnObservation(func()) {
}
func (a *AgentInfo) OnAfterStep(func()) {
}
func (a *AgentInfo) SetPerfInterval(steps uint32) {
}
