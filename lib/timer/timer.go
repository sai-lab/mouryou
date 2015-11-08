package timer

import (
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/mutex"
)

func Set(flag *int, flagMutex *sync.RWMutex, wait time.Duration) {
	if mutex.Read(flag, flagMutex) > 0 {
		return
	}

	mutex.Write(flag, flagMutex, 1)
	time.Sleep(wait * time.Second)
	mutex.Write(flag, flagMutex, 0)
}
