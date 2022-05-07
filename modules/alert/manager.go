package alert

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/Leantar/fimserver/models"
)

const (
	KindCreate = "CREATE"
	KindChange = "CHANGE"
	KindDelete = "DELETE"
)

type Manager struct {
	baseline []models.FsObject
	alerts   []models.Alert
	agentID  uint64
}

func NewManager(baseline []models.FsObject, agentID uint64) *Manager {
	if len(baseline) == 0 {
		panic("baseline must not be empty")
	}

	// Sort the slice to be able to search it faster afterwards
	sort.Slice(baseline, func(i, j int) bool {
		return baseline[i].Path < baseline[j].Path
	})

	return &Manager{
		baseline: baseline,
		alerts:   make([]models.Alert, 0),
		agentID:  agentID,
	}
}

func (m *Manager) CheckForAlert(obj models.FsObject) {
	baseLen := len(m.baseline)

	i := sort.Search(baseLen, func(i int) bool {
		return m.baseline[i].Path >= obj.Path
	})

	// Check if object exists in baseline
	if i < baseLen && m.baseline[i].Path == obj.Path {
		if !m.equal(obj, m.baseline[i]) {
			m.insert(models.Alert{
				Kind:       KindChange,
				Difference: GetDifference(m.baseline[i], obj),
				IssuedAt:   time.Now().Unix(),
				Path:       obj.Path,
				Modified:   obj.Modified,
				AgentID:    m.agentID,
			})
		}

		// Remove element from baseline to be able to check for DELETE events afterwards
		m.baseline = append(m.baseline[:i], m.baseline[i+1:]...)
	} else {
		m.insert(models.Alert{
			Kind:     KindCreate,
			IssuedAt: time.Now().Unix(),
			Path:     obj.Path,
			Modified: obj.Modified,
			AgentID:  m.agentID,
		})
	}
}

func (m *Manager) Result() []models.Alert {
	m.removeFolderContent()

	for _, obj := range m.baseline {
		m.insert(models.Alert{
			Kind:     KindDelete,
			IssuedAt: time.Now().Unix(),
			Path:     obj.Path,
			AgentID:  m.agentID,
		})
	}

	return m.alerts
}

func (m *Manager) removeFolderContent() {
	var remainingBaseline []models.FsObject

	for _, obj := range m.baseline {
		if len(remainingBaseline) == 0 || !isSuperseded(remainingBaseline, obj) {
			remainingBaseline = append(remainingBaseline, obj)
		}
	}

	m.baseline = remainingBaseline
}

func (m *Manager) equal(obj, obj2 models.FsObject) bool {
	return obj.Path == obj2.Path &&
		obj.Hash == obj2.Hash &&
		obj.Created == obj2.Created &&
		obj.Modified == obj2.Modified &&
		obj.Uid == obj2.Uid &&
		obj.Gid == obj2.Gid &&
		obj.Mode == obj2.Mode
}

func (m *Manager) insert(alert models.Alert) {
	alertLen := len(m.alerts)

	i := sort.Search(alertLen, func(i int) bool {
		return m.alerts[i].Path >= alert.Path
	})

	if i == alertLen {
		m.alerts = append(m.alerts, alert)
	} else {
		m.alerts = append(m.alerts[:i+1], m.alerts[i:]...)
		m.alerts[i] = alert
	}
}

func isSuperseded(objs []models.FsObject, obj models.FsObject) bool {
	for _, arrObj := range objs {
		path := obj.Path

		for path != "/" {
			if arrObj.Path == path {
				return true
			}

			path = filepath.Dir(path)
		}
	}

	return false
}
