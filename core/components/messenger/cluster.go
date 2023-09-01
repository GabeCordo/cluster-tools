package messenger

import (
	"time"
)

func (cluster *Cluster) Get(supervisor uint64) ([]string, bool) {

	cluster.mutex.RLock()
	defer cluster.mutex.RUnlock()

	logs, found := cluster.supervisors[supervisor]
	return logs, found
}

func (cluster *Cluster) Add(supervisor uint64, level MessagePriority, message string) error {

	cluster.mutex.Lock()
	defer cluster.mutex.Unlock()

	logs, found := cluster.supervisors[supervisor]
	log := Log{Timestamp: time.Now(), Priority: level, Message: message}

	if !found {
		logs = make([]string, 1)
		logs[0] = log.ToString()
	} else {
		logs = append(logs, log.ToString())
	}

	cluster.supervisors[supervisor] = logs

	return nil
}
