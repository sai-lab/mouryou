package timer

import (
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/mutex"
)

func Set(flag *int, flagMutex *sync.RWMutex, sleep int) {
	mutex.Write(flag, 1, flagMutex)
	time.Sleep(time.Duration(sleep) * time.Second)
	mutex.Write(flag, 0, flagMutex)
}
