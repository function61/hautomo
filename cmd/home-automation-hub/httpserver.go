package main

import (
	"context"
	"encoding/json"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

const tpl = `
<html>
<head>
	<title>home-automation-hub</title>
</head>
<body>

<table>
<thead>
<tr>
	<th></th>
	<th>name</th>
	<th>type</th>
	<th>battery</th>
	<th>link quality</th>
	<th>last online</th>
	<th>temp</th>
</tr>
</thead>
<tbody>
{{range .}}
<tr>
	<td>{{.Device.ProbablyTurnedOn}}</td>
	<td>{{.Device.Conf.DeviceId}}</td>
	<td>{{.Device.DeviceType.Manufacturer}} {{.Device.DeviceType.Model}}</td>
	<td title="type: {{.Device.DeviceType.BatteryType}} voltage: {{.Device.BatteryVoltage}} mV">{{.Device.BatteryPct}} %</td>
	<td>{{.Device.LinkQuality}} %</td>
	<td>{{.LastOnlineFormatted}}</td>
	<td>{{if .Device.LastTemperatureHumidityPressureEvent}}
		temp {{.Device.LastTemperatureHumidityPressureEvent.Temperature}}
		humidify {{.Device.LastTemperatureHumidityPressureEvent.Humidity}}
		pressure {{.Device.LastTemperatureHumidityPressureEvent.Pressure}}
	{{end}}</td>
</tr>
{{end}}
</tbody>
</table>

</body>
</html>
`

func handleHttp(app *Application, conf *hapitypes.ConfigFile, logger *log.Logger, stop *stopper.Stopper) {
	logl := logex.Levels(logger)

	defer stop.Done()
	srv := &http.Server{Addr: ":8097"}

	go func() {
		<-stop.Signal

		logl.Info.Println("stopping HTTP")

		_ = srv.Shutdown(context.TODO())
	}()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(conf)
	})

	http.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("name").Parse(tpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		devices := []*hapitypes.Device{}

		for _, dev := range app.deviceById {
			devices = append(devices, dev)
		}

		sort.Slice(devices, func(i, j int) bool {
			return devices[i].Conf.DeviceId < devices[j].Conf.DeviceId
		})

		type DeviceWithComputed struct {
			Device              *hapitypes.Device
			LastOnlineFormatted string
		}

		now := time.Now()

		devicesComputed := []DeviceWithComputed{}
		for _, device := range devices {
			lastOnlineFormatted := ""

			if device.LastOnline != nil {
				lastOnlineFormatted = now.Sub(*device.LastOnline).String()
			}

			devicesComputed = append(devicesComputed, DeviceWithComputed{
				Device:              device,
				LastOnlineFormatted: lastOnlineFormatted,
			})
		}

		if err := tmpl.Execute(w, devicesComputed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	logl.Info.Printf("Starting to listen at %s", srv.Addr)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logl.Error.Printf("ListenAndServe(): %s", err.Error())
	}
}
