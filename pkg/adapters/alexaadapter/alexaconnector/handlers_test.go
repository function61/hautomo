package alexaconnector

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
)

func TestDiscovery(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa.Discovery",
			"name": "Discover",
			"payloadVersion": "3",
			"messageId": "b3fa7aa4-a168-4af9-80d2-d35a1f4ac5e6"
		},
		"payload": {
			"scope": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			}
		}
	}
}`

	genericAssert(t, input, `
# Output
{
  "event": {
    "header": {
      "namespace": "Alexa.Discovery",
      "name": "Discover.Response",
      "payloadVersion": "3",
      "messageId": "mockId"
    },
    "payload": {
      "endpoints": [
        {
          "endpointId": "bedroomCeilingFan",
          "manufacturerName": "function61.com",
          "version": "1.0",
          "friendlyName": "Ceiling fan",
          "description": "Bedroom ceiling fan",
          "displayCategories": [
            "SMARTPLUG"
          ],
          "capabilities": [
            {
              "type": "AlexaInterface",
              "interface": "Alexa.PowerController",
              "version": "3",
              "properties": {
                "supported": [
                  {
                    "name": "powerState"
                  }
                ],
                "proactivelyReported": false,
                "retrievable": false
              }
            }
          ],
          "cookie": {
            "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
          }
        },
        {
          "endpointId": "bathroomCeilingLight",
          "manufacturerName": "function61.com",
          "version": "1.0",
          "friendlyName": "Bathroom ceiling light",
          "description": "Bathroom ceiling light",
          "displayCategories": [
            "LIGHT"
          ],
          "capabilities": [
            {
              "type": "AlexaInterface",
              "interface": "Alexa.PowerController",
              "version": "3",
              "properties": {
                "supported": [
                  {
                    "name": "powerState"
                  }
                ],
                "proactivelyReported": false,
                "retrievable": false
              }
            }
          ],
          "cookie": {
            "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
          }
        }
      ]
    }
  }
}

# Queue msg
(None)`)
}

func TestReportState(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa",
			"name": "ReportState",
			"payloadVersion": "3",
			"messageId": "c558b9fb-2fd1-4b6a-a2f7-5affe59c2a44",
			"correlationToken": "corr1"
		},
		"endpoint": {
			"scope": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			},
			"endpointId": "038c2459-cc64-4226-8413-03c13f316797",
			"cookie": {}
		},
		"payload": {}
	}
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.ContactSensor",
        "name": "detectionState",
        "value": "NOT_DETECTED",
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      },
      {
        "namespace": "Alexa.EndpointHealth",
        "name": "connectivity",
        "value": {
          "value": "OK"
        },
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "StateReport",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr1"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "038c2459-cc64-4226-8413-03c13f316797"
    },
    "payload": {}
  }
}

# Queue msg
(None)`)
}

func TestAcceptGrant(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa.Authorization",
			"name": "AcceptGrant",
			"messageId": "9176687b-c10b-478a-8021-af0d2f413a7b",
			"payloadVersion": "3"
		},
		"payload": {
			"grant": {
				"type": "OAuth2.AuthorizationCode",
				"code": "ANLIbUFctTXOievfGlFt"
			},
			"grantee": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			}
		}
	}
}`

	genericAssert(t, input, `
# Output
{
  "event": {
    "header": {
      "namespace": "Alexa.Authorization",
      "name": "AcceptGrant.Response",
      "payloadVersion": "3",
      "messageId": "mockId"
    },
    "payload": {}
  }
}

# Queue msg
(None)`)
}

func TestPowerControllerTurnOn(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa.PowerController",
			"name": "TurnOn",
			"payloadVersion": "3",
			"messageId": "3190e501-b88d-492b-a792-7b174878f5a5",
			"correlationToken": "corr2"
		},
		"endpoint": {
			"scope": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			},
			"endpointId": "cornerLight",
			"cookie": {
				"queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
			}
		},
		"payload": {}
	}
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.PowerController",
        "name": "powerState",
        "value": "ON",
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr2"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "cornerLight",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"cornerLight","Attrs":{"on":{"value":true,"reported":"2020-07-14T14:09:00Z"}}}`)
}

func TestPowerControllerTurnOff(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa.PowerController",
			"name": "TurnOff",
			"payloadVersion": "3",
			"messageId": "23519dc3-f43b-4938-a0eb-e64f37429095",
			"correlationToken": "corr3"
		},
		"endpoint": {
			"scope": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			},
			"endpointId": "cornerLight",
			"cookie": {
				"queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
			}
		},
		"payload": {}
	}
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.PowerController",
        "name": "powerState",
        "value": "OFF",
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr3"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "cornerLight",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"cornerLight","Attrs":{"on":{"value":false,"reported":"2020-07-14T14:09:00Z"}}}`)
}

func TestSetBrightness(t *testing.T) {
	input := `{
	"directive": {
		"header": {
			"namespace": "Alexa.BrightnessController",
			"name": "SetBrightness",
			"payloadVersion": "3",
			"messageId": "6e6de8da-7e99-44dc-ae77-b57b9d0e1e46",
			"correlationToken": "corr4"
		},
		"endpoint": {
			"scope": {
				"type": "BearerToken",
				"token": "Atza|dummyAccessToken"
			},
			"endpointId": "sofaLight",
			"cookie": {
				"queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
			}
		},
		"payload": {
			"brightness": 100
		}
	}
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.BrightnessController",
        "name": "brightness",
        "value": 100,
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr4"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "sofaLight",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"sofaLight","Attrs":{"brightness":{"value":100,"reported":"2020-07-14T14:09:00Z"}}}`)
}

func TestSetColorTemp(t *testing.T) {
	input := `{
    "directive": {
        "header": {
            "namespace": "Alexa.ColorTemperatureController",
            "name": "SetColorTemperature",
            "payloadVersion": "3",
            "messageId": "df089fed-3176-4290-98e2-e985d0942aea",
            "correlationToken": "corr5"
        },
        "endpoint": {
            "scope": {
                "type": "BearerToken",
                "token": "Atza|dummyAccessToken"
            },
            "endpointId": "officeLight",
            "cookie": {
                "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
            }
        },
        "payload": {
            "colorTemperatureInKelvin": 2200
        }
    }
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.ColorTemperatureController",
        "name": "colorTemperatureInKelvin",
        "value": 2200,
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr5"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "officeLight",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"officeLight","Attrs":{"color_temp":{"value":455,"reported":"2020-07-14T14:09:00Z"}}}`)
}

func TestSetColor(t *testing.T) {
	input := `{
    "directive": {
        "header": {
            "namespace": "Alexa.ColorController",
            "name": "SetColor",
            "payloadVersion": "3",
            "messageId": "a4d87be2-189a-4a32-add4-9be6b0ebf6f9",
            "correlationToken": "corr6"
        },
        "endpoint": {
            "scope": {
                "type": "BearerToken",
                "token": "Atza|dummyAccessToken"
            },
            "endpointId": "sofaLight",
            "cookie": {
                "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
            }
        },
        "payload": {
            "color": {
                "hue": 60,
                "saturation": 1,
                "brightness": 1
            }
        }
    }
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": [
      {
        "namespace": "Alexa.ColorController",
        "name": "color",
        "value": {
          "hue": 60,
          "saturation": 1,
          "brightness": 1
        },
        "timeOfSample": "2020-07-14T14:09:00Z",
        "uncertaintyInMilliseconds": 0
      }
    ]
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr6"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "sofaLight",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"sofaLight","Attrs":{"color":{"r":255,"g":255,"b":0,"reported":"2020-07-14T14:09:00Z"}}}`)
}

func TestPlayback(t *testing.T) {
	input := `{
    "directive": {
        "header": {
            "namespace": "Alexa.PlaybackController",
            "name": "Pause",
            "payloadVersion": "3",
            "messageId": "730badbd-2eb1-40e9-9f1c-892758e24ac2",
            "correlationToken": "corr7"
        },
        "endpoint": {
            "scope": {
                "type": "BearerToken",
                "token": "Atza|dummyAccessToken"
            },
            "endpointId": "workPC",
            "cookie": {
                "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
            }
        },
        "payload": {}
    }
}`

	genericAssert(t, input, `
# Output
{
  "context": {
    "properties": null
  },
  "event": {
    "header": {
      "namespace": "Alexa",
      "name": "Response",
      "payloadVersion": "3",
      "messageId": "mockId",
      "correlationToken": "corr7"
    },
    "endpoint": {
      "scope": {
        "type": "BearerToken",
        "token": "Atza|dummyAccessToken"
      },
      "endpointId": "workPC",
      "cookie": {
        "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation"
      }
    },
    "payload": {}
  }
}

# Queue msg
{"DeviceId":"workPC","Attrs":{"playback_control":{"control":"Pause","reported":"2020-07-14T14:09:00Z"}}}`)
}

func genericAssert(t *testing.T, input string, outMsg string) {
	t.Helper()

	extSys := &dummyExternalSystems{}

	handlers := New(
		extSys,
		func() string { return "mockId" },
		func() time.Time { return time.Date(2020, 7, 14, 14, 9, 0, 0, time.UTC) },
	)

	response, err := handlers.Handle(context.Background(), []byte(input))
	assert.Ok(t, err)

	lines := []string{"", "# Output"}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		panic(err)
	}

	lines = append(lines, string(jsonBytes))

	lines = append(lines, "", "# Queue msg")

	if extSys.lastCommand != "" {
		lines = append(lines, extSys.lastCommand)
	} else {
		lines = append(lines, "(None)")
	}

	assert.EqualString(t, strings.Join(lines, "\n"), outMsg)
}

type dummyExternalSystems struct {
	lastCommand string
}

func (t *dummyExternalSystems) TokenToUserId(ctx context.Context, token string) (string, error) {
	if token != "Atza|dummyAccessToken" {
		panic("unexpected access token")
	}

	return "amzn1.account.AEEOMDZSCRTGH5TCBS23NY7G65CQ", nil
}

func (t *dummyExternalSystems) SendCommand(
	_ context.Context,
	queue string,
	command aamessages.Message,
) error {
	commandStr, err := aamessages.Marshal(command)
	if err != nil {
		return err
	}

	t.lastCommand = commandStr

	return nil
}

func (t *dummyExternalSystems) FetchDiscoveryFile(ctx context.Context, userId string) (io.ReadCloser, error) {
	dummyFile := `{
  "queue": "https://sqs.us-east-1.amazonaws.com/123456789011/JoonasHomeAutomation",
  "devices": [
    {
      "id": "bedroomCeilingFan",
      "friendly_name": "Ceiling fan",
      "description": "Bedroom ceiling fan",
      "display_category": "SMARTPLUG",
      "capability_codes": [
        "PowerController"
      ]
    },
    {
      "id": "bathroomCeilingLight",
      "friendly_name": "Bathroom ceiling light",
      "description": "Bathroom ceiling light",
      "display_category": "LIGHT",
      "capability_codes": [
        "PowerController"
      ]
    }
  ]
}`

	return ioutil.NopCloser(strings.NewReader(dummyFile)), nil
}
