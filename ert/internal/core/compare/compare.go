package compare

import (
	"context"

	"github.com/yourname/ert/internal/model"
)

type SessionComparator struct{}

func New() *SessionComparator {
	return &SessionComparator{}
}

func (c *SessionComparator) Compare(ctx context.Context, s1, s2 *model.CompareResult) *model.CompareResult {
	result := &model.CompareResult{
		Session1: s1.Session1,
		Session2: s2.Session2,
	}

	result.AddedProcesses = c.findAdded(s1.AddedProcesses, s2.AddedProcesses)
	result.RemovedProcesses = c.findRemoved(s1.AddedProcesses, s2.AddedProcesses)
	result.AddedNetwork = c.findAddedNetwork(s1.AddedNetwork, s2.AddedNetwork)
	result.RemovedNetwork = c.findRemovedNetwork(s1.AddedNetwork, s2.AddedNetwork)

	return result
}

func (c *SessionComparator) findAdded(old, new []model.ProcessDTO) []model.ProcessDTO {
	added := make([]model.ProcessDTO, 0)
	newMap := make(map[uint32]bool)
	for _, p := range new {
		newMap[p.PID] = true
	}
	for _, p := range old {
		if !newMap[p.PID] {
			added = append(added, p)
		}
	}
	return added
}

func (c *SessionComparator) findRemoved(old, new []model.ProcessDTO) []model.ProcessDTO {
	return c.findAdded(new, old)
}

func (c *SessionComparator) findAddedNetwork(old, new []model.NetworkConnDTO) []model.NetworkConnDTO {
	added := make([]model.NetworkConnDTO, 0)
	newMap := make(map[string]bool)
	for _, n := range new {
		key := n.Protocol + n.LocalAddr + string(rune(n.LocalPort)) + n.RemoteAddr + string(rune(n.RemotePort))
		newMap[key] = true
	}
	for _, n := range old {
		key := n.Protocol + n.LocalAddr + string(rune(n.LocalPort)) + n.RemoteAddr + string(rune(n.RemotePort))
		if !newMap[key] {
			added = append(added, n)
		}
	}
	return added
}

func (c *SessionComparator) findRemovedNetwork(old, new []model.NetworkConnDTO) []model.NetworkConnDTO {
	return c.findAddedNetwork(new, old)
}
