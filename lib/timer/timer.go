package timer

import (
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/mutex"
)

func Set(flag *int, flagMutex *sync.RWMutex, sleep int) {
	mutex.Write(flag, flagMutex, 1)
	time.Sleep(time.Duration(sleep) * time.Second)
	mutex.Write(flag, flagMutex, 0)
}
