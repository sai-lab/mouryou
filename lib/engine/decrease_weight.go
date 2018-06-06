package engine

import (
	"fmt"
	"sync"

	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func DecreaseWeight(config *models.Config) {
	var mu sync.RWMutex
	lowWeight := 5

	for informations := range monitor.DataCh {
		for _, information := range informations {
			if information.Error != "Connection is timeout." {
				continue
			}
			name := information.Name
			mu.RLock()
			for i, v := range monitor.States {
				if v.Name != name {
					continue
				}
				if monitor.States[i].Weight <= lowWeight {
					break
				}

				err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, lowWeight)
				if err != nil {
					fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
				}

				// サーバの重みを変更したとき、合計の重みを変更する
				mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-(monitor.States[i].Weight-lowWeight))
				mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight-(monitor.States[i].Weight-lowWeight))
				monitor.States[i].Weight = lowWeight
				break
			}
			mu.RUnlock()
		}
	}
}
