package ezhub

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

func createHttpApi(stack *ezstack.Stack, nodeDatabase *nodeDb) http.Handler {
	routes := http.NewServeMux()

	routes.HandleFunc("/api/power/toggle", func(w http.ResponseWriter, r *http.Request) {
		device, found := nodeDatabase.GetDevice(zigbee.IEEEAddress(r.URL.Query().Get("addr")))
		if !found {
			httputils.Error(w, http.StatusNotFound)
			return
		}

		endpoint := ezstack.DeviceAndEndpoint{
			NetworkAddress: device.NetworkAddress,
			EndpointId:     ezstack.DefaultSingleEndpointId,
		}

		if err := stack.LocalCommand(endpoint, &cluster.GenOnOffToggleCommand{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	routes.HandleFunc("/api/clusters", func(w http.ResponseWriter, r *http.Request) {
		device, found := nodeDatabase.GetDevice(zigbee.IEEEAddress(r.URL.Query().Get("addr")))
		if !found {
			httputils.Error(w, http.StatusNotFound)
			return
		}

		lines := []string{}
		line := func(str string) { lines = append(lines, str) }

		for _, endpoint := range device.Endpoints {
			for _, clusterId := range endpoint.InClusterList {
				clusterName := func() string {
					if def := cluster.FindDefinition(clusterId); def != nil {
						return def.Name()
					} else {
						return fmt.Sprintf("unknown cluster: %d", clusterId)
					}
				}()

				line(fmt.Sprintf("%d -> %s", clusterId, clusterName))
			}
		}

		fmt.Fprint(w, strings.Join(lines, "\n")+"\n")
	})

	return routes
}
