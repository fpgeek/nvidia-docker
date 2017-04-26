package nvidia

import (
	"fmt"
	"sort"

	dockercli "github.com/fsouza/go-dockerclient"
)

const (
	endpoint = "unix:///var/run/docker.sock"
)

type (
	pair struct {
		Key   int
		Value int
	}
	pairList []pair
)

func (p pairList) Len() int           { return len(p) }
func (p pairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SelectGPU(devs []Device) (string, error) {
	devCnt := len(devs)
	usedGPUCountMap := map[int]int{}
	for i := 0; i < devCnt; i++ {
		usedGPUCountMap[i] = 0
	}

	client, err := dockercli.NewClient(endpoint)
	if err != nil {
		return "", err
	}
	apiConts, err := client.ListContainers(dockercli.ListContainersOptions{})
	if err != nil {
		return "", err
	}

	for _, apiCont := range apiConts {
		if cont, err := client.InspectContainer(apiCont.ID); err == nil {
			for _, device := range cont.HostConfig.Devices {
				for i := range devs {
					if device.PathOnHost == devs[i].Path {
						if cnt, ok := usedGPUCountMap[i]; ok {
							usedGPUCountMap[i] = cnt + 1
						} else {
							usedGPUCountMap[i] = 1
						}
					}
				}
			}
		}
	}

	if len(usedGPUCountMap) == 0 {
		return "0", nil
	}

	pairs := make(pairList, 0, len(usedGPUCountMap))
	for index, usedCnt := range usedGPUCountMap {
		pairs = append(pairs, pair{index, usedCnt})
	}
	sort.Sort(pairs)
	return fmt.Sprintf("%d", pairs[0].Key), nil
}
