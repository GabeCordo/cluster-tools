package messenger

import (
	"fmt"
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
	log := fmt.Sprintf("[%s][%s] %s", time.Now().Format("2006-01-02 15:04:05"), level.ToString(), message)

	if !found {
		logs = make([]string, 1)
		logs[0] = log
	} else {
		logs = append(logs, log)
	}

	cluster.supervisors[supervisor] = logs

	return nil
}
