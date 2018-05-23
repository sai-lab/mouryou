package timer

import (
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/mutex"
)

func Set(flag *int, flagMutex *sync.RWMutex, wait time.Duration) {
	mutex.Write(flag, flagMutex, *flag+1)
	time.Sleep(wait * time.Second)
	mutex.Write(flag, flagMutex, *flag-1)
}
