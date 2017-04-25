package docker

import (
	"fmt"
	"sort"
	"strings"

	dockercli "github.com/fsouza/go-dockerclient"
)

const (
	endpoint = "unix:///var/run/docker.sock"
)

type (
	pair struct {
		Key   string
		Value int
	}
	pairList []pair
)

func (p pairList) Len() int           { return len(p) }
func (p pairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SelectGPU(gpuCount int) (string, error) {
	usedGPUCountMap := map[string]int{}
	for i := 0; i < gpuCount; i++ {
		usedGPUCountMap[fmt.Sprintf("%d", i)] = 0
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
			for _, env := range cont.Config.Env {
				if strings.HasPrefix(env, "NV_GPU=") {
					splitVals := strings.Split(env, "=")
					if len(splitVals) == 2 {
						gpuIndex := splitVals[1]
						if cnt, ok := usedGPUCountMap[gpuIndex]; ok {
							usedGPUCountMap[gpuIndex] = cnt + 1
						} else {
							usedGPUCountMap[gpuIndex] = 1
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
	return pairs[0].Key, nil
}
