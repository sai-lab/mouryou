package engine

import (
	"fmt"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"sync"
)

func DecreaseWeight(config *models.Config) {
	var mu sync.RWMutex
	mu.RLock()
	defer mu.RUnlock()

	for informations := range monitor.DataCh {
		for _, information := range informations {
			if information.Error != "" {
				err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, 5)
				if err != nil {
					fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
				}
			}
		}

	}

}
