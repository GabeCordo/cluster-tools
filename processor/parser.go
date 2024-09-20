package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func parse() (module, cluster string, metadata map[string]string, err error) {

	args := os.Args[1:]
	numArgs := len(args)

	if numArgs >= 1 {

		moduleCluster := args[0]
		splitModuleCluster := strings.Split(moduleCluster, ":")

		if len(splitModuleCluster) != 2 {
			return "", "", nil, errors.New("expected first parameter to be in the format module:cluster")
		}

		module = splitModuleCluster[0]
		cluster = splitModuleCluster[1]
	}

	metadata = make(map[string]string)

	if numArgs >= 2 {
		metadataStr := ""

		for i := 1; i < numArgs; i++ {
			metadataStr += args[i]
		}

		fmt.Println(metadataStr)

		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			output := fmt.Sprintf("received metadata is not a valid json: \n%s\n", metadataStr)
			return "", "", nil, errors.New(output)
		}
	}

	return module, cluster, metadata, nil
}
