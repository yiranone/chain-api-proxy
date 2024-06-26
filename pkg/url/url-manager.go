package url2

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	ClientRequestUrl        = "clientRequestUrl"
	ClientRequestSpecialUrl = "clientRequestSpecialUrl"
	JobRequestUrl           = "jobRequestUrl"
	JobRequestSpecialUrl    = "jobRequestSpecialUrl"
)

type Status struct {
	Url              string
	UrlType          string
	ResultErrorCount int
	ResultNullCount  int
	SendErrorCount   int
	ReadErrorCount   int
	TimeoutCount     int
	AccessCount      int
	BlockNotFound    int
	AccessTime       time.Time
	InvalidTime      time.Time
	ValidTime        time.Time
	Reason           string
}

// formatTime方法用于格式化时间，如果时间为空则返回"none"
func (s Status) formatTime(t time.Time) string {
	if t.IsZero() {
		return "none"
	}
	return t.Format(time.RFC3339)
}

// String方法用于返回Status结构体的可读字符串表示
func (s Status) String() string {
	return fmt.Sprintf("Url:%s UrlType:%s Valid:%t  AccessTime:%s  InvalidTime:%s ValidTime:%s AccessCount:%d ResultErrorCount:%d ResultNullCount:%d  SendErrorCount:%d ReadErrorCount:%d TimeoutCount:%d  BlockNotFound:%d Reason:%s",
		s.Url,
		s.UrlType,
		s.InvalidTime.IsZero(),
		s.formatTime(s.AccessTime),
		s.formatTime(s.InvalidTime),
		s.formatTime(s.ValidTime),
		s.AccessCount,
		s.ResultErrorCount,
		s.ResultNullCount,
		s.SendErrorCount,
		s.ReadErrorCount,
		s.TimeoutCount,
		s.BlockNotFound,
		s.Reason)
}

type URLManager struct {
	urlGroups    map[string][]*Status
	validIndexes map[string][]int
	mutex        sync.Mutex
}

func NewURLManager(clientRequestUrls, clientRequestSpecialUrls, jobRequestUrls, jobRequestSpecialUrls []string) *URLManager {
	urlGroups := make(map[string][]*Status)

	urlGroups[ClientRequestUrl] = make([]*Status, len(clientRequestUrls))
	for i, url := range clientRequestUrls {
		urlGroups[ClientRequestUrl][i] = &Status{
			Url:     url,
			UrlType: ClientRequestUrl,
		}
	}

	urlGroups[ClientRequestSpecialUrl] = make([]*Status, len(clientRequestSpecialUrls))
	for i, url := range clientRequestSpecialUrls {
		urlGroups[ClientRequestSpecialUrl][i] = &Status{
			Url:     url,
			UrlType: ClientRequestSpecialUrl,
		}
	}

	urlGroups[JobRequestUrl] = make([]*Status, len(jobRequestUrls))
	for i, url := range jobRequestUrls {
		urlGroups[JobRequestUrl][i] = &Status{
			Url:     url,
			UrlType: JobRequestUrl,
		}
	}

	urlGroups[JobRequestSpecialUrl] = make([]*Status, len(jobRequestSpecialUrls))
	for i, url := range jobRequestSpecialUrls {
		urlGroups[JobRequestSpecialUrl][i] = &Status{
			Url:     url,
			UrlType: JobRequestSpecialUrl,
		}
	}

	m := &URLManager{
		urlGroups:    urlGroups,
		validIndexes: make(map[string][]int),
	}
	m.refreshValidIndexes()

	return m
}

func (m *URLManager) SetAllValid() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.resetAllStatusesToValid()
}

func (m *URLManager) StartResetScheduler() {
	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			m.mutex.Lock()
			for _, statuses := range m.urlGroups {
				for _, status := range statuses {
					if !status.InvalidTime.IsZero() {
						log.Infof("24H重新设置Url=%s为可用状态", status.Url)
						status.ValidTime = time.Now()
						status.InvalidTime = time.Time{}
						status.Reason = "24h set valid"
					}
				}
			}
			m.refreshValidIndexes()
			m.mutex.Unlock()
		}
	}
}

func (m *URLManager) SetInvalid(url string, reason string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.InvalidTime = time.Now()
				status.Reason = reason
			}
		}
	}
	m.refreshValidIndexes()
}

func (m *URLManager) AddResultErrorCount(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.ResultErrorCount++
				if status.ResultErrorCount%500 == 0 {
					status.InvalidTime = time.Now()
					status.Reason = "ResultErrorCount=" + strconv.Itoa(status.ResultErrorCount)
					m.refreshValidIndexes()
				}
			}
		}
	}
}

func (m *URLManager) AddResultNullCount(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.ResultNullCount++
			}
		}
	}
}

func (m *URLManager) AddSendErrorCount(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.SendErrorCount++
				if status.SendErrorCount%100 == 0 {
					status.InvalidTime = time.Now()
					status.Reason = "SendErrorCount=" + strconv.Itoa(status.SendErrorCount)
					m.refreshValidIndexes()
				}
			}
		}
	}
}

func (m *URLManager) AddReadErrorCount(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.ReadErrorCount++
				if status.ReadErrorCount%100 == 0 {
					status.InvalidTime = time.Now()
					status.Reason = "ReadErrorCount=" + strconv.Itoa(status.ReadErrorCount)
					m.refreshValidIndexes()
				}
			}
		}
	}
}

func (m *URLManager) AddTimeoutCount(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.TimeoutCount++
				if status.TimeoutCount%100 == 0 {
					status.InvalidTime = time.Now()
					status.Reason = "TimeoutCount=" + strconv.Itoa(status.TimeoutCount)
					m.refreshValidIndexes()
				}
			}
		}
	}
}

func (m *URLManager) AddBlockNotFound(url string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			if status.Url == url {
				status.BlockNotFound++
			}
		}
	}
}

func (m *URLManager) GetRandomURL(urlType string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	actualType := m.resolveUrlType(urlType)

	indexes, ok := m.validIndexes[actualType]
	if !ok || len(indexes) == 0 {
		m.resetAllStatusesToValid()
		indexes = m.validIndexes[actualType]
	}

	if len(indexes) == 0 {
		return "", fmt.Errorf("no valid URLs available")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := indexes[rand.Intn(len(indexes))]

	rtStatus := m.urlGroups[actualType][randomIndex]
	rtStatus.AccessTime = time.Now()
	rtStatus.AccessCount++

	return rtStatus.Url, nil
}

func (m *URLManager) GetAllURLStatus() ([]Status, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var statuses []*Status
	for _, group := range m.urlGroups {
		statuses = append(statuses, group...)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Url < statuses[j].Url
	})

	resp := make([]Status, len(statuses))
	for i, status := range statuses {
		resp[i] = *status
	}

	return resp, nil
}

func (m *URLManager) resetAllStatusesToValid() {
	for _, statuses := range m.urlGroups {
		for _, status := range statuses {
			status.ValidTime = time.Now()
			status.InvalidTime = time.Time{}
			status.Reason = "reset all to valid"
		}
	}
	m.refreshValidIndexes()
}

func (m *URLManager) refreshValidIndexes() {
	for urlType, statuses := range m.urlGroups {
		var validIndexes []int
		for j, status := range statuses {
			if status.InvalidTime.IsZero() {
				validIndexes = append(validIndexes, j)
			}
		}
		m.validIndexes[urlType] = validIndexes
	}
}

func (m *URLManager) resolveUrlType(urlType string) string {
	switch urlType {
	case JobRequestSpecialUrl:
		if len(m.validIndexes[JobRequestSpecialUrl]) == 0 {
			if len(m.validIndexes[ClientRequestSpecialUrl]) > 0 {
				return ClientRequestSpecialUrl
			}
			return JobRequestUrl
		}
	case JobRequestUrl:
		if len(m.validIndexes[JobRequestUrl]) == 0 {
			return ClientRequestUrl
		}
	case ClientRequestSpecialUrl:
		if len(m.validIndexes[ClientRequestSpecialUrl]) == 0 {
			return ClientRequestUrl
		}
	}
	return urlType
}
