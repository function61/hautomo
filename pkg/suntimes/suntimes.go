package suntimes

import (
	"time"

	"github.com/yaslama/astrocalc"
)

type latLng struct {
	latitude  float64
	longitude float64
}

var Tampere = latLng{
	latitude:  61.483509,
	longitude: 23.761736,
}

// between morning's and evening's golden hours? this could be defined as period
// with sufficient lighting.
//
// golden hour ~= sky is red
func IsBetweenGoldenHours(at time.Time, position latLng) bool {
	calc := astrocalc.NewSunCalc()
	sunTimes := calc.GetTimes(at, position.latitude, position.longitude)
	/*
		"2014-07-28T21:46:43.912170231Z": "nadir":         ,
		"2014-07-29T01:20:46.21797055Z": "nightEnd":      ,
		"2014-07-29T01:54:50.354016423Z": "nauticalDawn":  ,
		"2014-07-29T02:27:04.727511405Z": "dawn":          ,
		"2014-07-29T02:53:49.87563461Z": "sunrise":       ,
		"2014-07-29T02:56:32.923271656Z": "sunriseEnd":    ,
		"2014-07-29T03:28:11.204707324Z": "goldenHourEnd": ,
		"2014-07-29T09:46:43.912170231Z": "solarNoon":     ,
		"2014-07-29T16:05:16.619633138Z": "goldenHour":    ,
		"2014-07-29T16:36:54.901068806Z": "sunsetStart":   ,
		"2014-07-29T16:39:37.948705852Z": "sunset":        ,
		"2014-07-29T17:06:23.096829056Z": "dusk":          ,
		"2014-07-29T17:38:37.470324039Z": "nauticalDusk":  ,
		"2014-07-29T18:12:41.606369912Z": "night":         ,
	*/

	return at.After(sunTimes["goldenHourEnd"]) && at.Before(sunTimes["goldenHour"])
}
