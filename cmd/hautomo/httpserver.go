package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const tpl = `
<html>
<head>
	<title>Hautomo</title>
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
	<th>last heartbeat</th>
	<th>temp</th>
</tr>
</thead>
<tbody>
{{range .}}
<tr>
	<td>{{.Device.ProbablyTurnedOn}}</td>
	<td>{{.Device.Conf.DeviceId}}</td>
	<td>{{.Device.DeviceType.Manufacturer}} {{.Device.DeviceType.Model}}</td>
{{if .Device.DeviceType.BatteryType}}
	<td title="type: {{.Device.DeviceType.BatteryType}} voltage: {{.Device.BatteryVoltage}} mV">{{.Device.BatteryPct}} %</td>
{{else}}
	<td></td>
{{end}}
	<td>{{.Device.LinkQuality}}</td>
	<td>{{.LastOnlineFormatted}}</td>
	<td>{{if .Device.LastTemperatureHumidityPressureEvent}}
		temp {{.Device.LastTemperatureHumidityPressureEvent.Temperature}}
		humidity {{.Device.LastTemperatureHumidityPressureEvent.Humidity}}
		pressure {{.Device.LastTemperatureHumidityPressureEvent.Pressure}}
	{{end}}</td>
</tr>
{{end}}
</tbody>
</table>

</body>
</html>
`

func makeHttpServer(app *Application, conf *hapitypes.ConfigFile) *http.Server {
	srv := &http.Server{Addr: ":8097"}

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(conf)
	})

	// to easily trigger debug events ...
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		// ... so you can test your actions by subscribing to debug event
		arg := r.URL.Query().Get("arg")
		if arg == "" {
			app.publish("debug")
		} else {
			app.publish("debug:" + arg)
		}
	})

	http.Handle("/metrics", promhttp.Handler())

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

	return srv
}
