package manager

import (
	"dwl"
	"dwl/progress"
	"dwl/settings"
	"encoding/json"
	"fmt"
	"sync"
)

type Manager struct {
	items []*Item
	lock  sync.Mutex
}

type Item struct {
	dwloader *dwl.DWL
	status   int
	err      error
}

type Status struct {
	Index    int
	Progress progress.Progress
	Status   int
	Error    string
}

func NewManager() *Manager {
	m := new(Manager)
	m.items = make([]*Item, 0)
	return m
}

//////////
//List
func (m *Manager) Add(js string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	itm := new(Item)
	d, err := parsejs(js)
	if err != nil {
		return err
	}
	itm.dwloader = d
	return nil
}

func (m *Manager) Rem(ind int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.items[ind].dwloader.Stop()
	m.items = append(m.items[:ind], m.items[ind+1:]...)
}

func (m *Manager) AllProgress() string {
	m.lock.Lock()
	defer m.lock.Unlock()
	sts := make([]Status, 0)

	for i, p := range m.items {
		st := getStatus(p, i)
		sts = append(sts, st)
	}
	buf, err := json.Marshal(sts)
	if err != nil {
		fmt.Println("*** Error marshal progress (AllProgress):", err)
		return ""
	}
	return string(buf)
}

//////////
//Ctrl
func (m *Manager) Load(i int) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.items[i].dwloader.Load()
}

func (m *Manager) Stop(i int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.items[i].dwloader.Stop()
}

func (m *Manager) CompleteLoad(i int) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.items[i].dwloader.Complete()
}

func (m *Manager) ProgressLoad(i int) string {
	m.lock.Lock()
	defer m.lock.Unlock()
	st := getStatus(m.items[i], i)
	buf, err := json.Marshal(&st)
	if err != nil {
		fmt.Println("*** Error marshal progress (ProgressLoad):", err)
		return ""
	}
	return string(buf)
}

//////////
//Utils
func parsejs(js string) (*dwl.DWL, error) {
	sets, err := settings.FromJson(js)
	if err != nil {
		return nil, err
	}

	return dwl.NewDWL(sets), nil
}

func getStatus(itm *Item, i int) Status {
	st := Status{}
	st.Progress = itm.dwloader.GetProgress()
	if itm.err != nil {
		st.Error = itm.err.Error()
	}
	st.Index = i
	st.Status = itm.status
	return st
}
