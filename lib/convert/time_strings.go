package convert

import "time"

var (
	layout = "Mon Jan 2 15:04:05 MST 2006"
)

func StringsToTimeToInt(str string) (int, error) {
	t, err := time.Parse(layout, str)

	return int(t.Unix()), err
}
