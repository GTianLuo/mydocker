package common

import (
	"fmt"
	"strings"
)

func ParseVolumeParam(volume []string) ([]string, error) {
	pVolume := make([]string, len(volume)*2)
	for i, v := range volume {
		split := strings.Split(v, ":")
		if len(split) != 2 {
			return []string{}, fmt.Errorf("invalied volume param:expect 'path:path',not %v", v)
		}
		pVolume[2*i] = split[0]
		pVolume[2*i+1] = split[1]
	}
	return pVolume, nil
}
