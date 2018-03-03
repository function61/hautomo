'use strict';

var AWS = require('aws-sdk');
var sqs = new AWS.SQS({apiVersion: '2012-11-05'});

// FIXME:
// - https://developer.amazon.com/docs/smarthome/build-smart-home-skills-for-lights.html#choose-capabilities
// - https://developer.amazon.com/docs/device-apis/alexa-discovery.html#capability-versions
// - https://developer.amazon.com/docs/smarthome/smart-home-skill-api-message-reference.html

function powerController() {
	return {
		type: "AlexaInterface",
		interface: "Alexa.PowerController",
		version: "3",
		properties: {
			supported: [
				{ name: "powerState" },
			],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

function brightnessController() {
	return {
		type: "AlexaInterface",
		interface: "Alexa.BrightnessController",
		version: "3",
		properties: {
			supported: [
				{ name: "brightness" },
			],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

function colorController() {
	return {
		type: "AlexaInterface",
		interface: "Alexa.ColorController",
		version: "3",
		properties: {
			supported: [
				{ name: "color" },
			],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

function createDevice(deviceId, friendlyName, friendlyDescription) {
	return     {
		endpointId: deviceId,
		manufacturerName: 'function61.com',
		version: '1.0',
		friendlyName: friendlyName,
		description: friendlyDescription,
		displayCategories: [ 'LIGHT' ], // where are these listed?
		capabilities: [
			powerController(),
			brightnessController(),
			colorController(),
		],
		/*
		cookie: {
			extraDetail1: 'optionalDetailForSkillAdapterToReferenceThisDevice',
		*/
	};
}

const queueUrl = 'https://sqs.us-east-1.amazonaws.com/329074924855/JoonasHomeAutomation';

function getDevicesFromPartnerCloud() {
	return [
		createDevice('d2ff0882', 'Sofa light', 'Floor light next the sofa'),
		createDevice('98d3cb01', 'Speaker light', 'Floor light under the speaker'),
		createDevice('e97d7d4c', 'Cat light', 'Light above the cat painting'),
		createDevice('23e06f45', 'Nightstand light', 'Light on the nightstand'),
		createDevice('52fe368b', 'Kitchen light', 'Under-cabinet lighting'),
		createDevice('39664b86', 'Bar light', 'Under-cabinet lighting'),
		createDevice('c0730bb2', 'Amplifier', 'Onkyo TX-NR515'),
		createDevice('7e7453da', 'TV', 'Philips 55'' 4K 55PUS7909'),
		createDevice('cfb1b27f', 'Living room lights', 'Device group: Living room lights'),
	];
}

function log(title, msg) {
	console.log(`[${title}] ${msg}`);
}

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
	var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
	return v.toString(16);
  });
}

function isValidToken(token) { return true; }

function sendMessageToRaspberry(message) {
	var params = {
	  MessageBody: message,
	  QueueUrl: queueUrl
	};

	sqs.sendMessage(params, function(err, data) {
		if (err) console.log(err, err.stack); // an error occurred
		else     console.log(data);           // successful response
	});
}

/*
	green => 
		"hue": 120,
		"saturation": 1,
		"brightness": 1

*/
function hsvToRgb (h, s, v) {
	h /= 360;
	v = Math.round(v * 255);

	const i = Math.floor(h * 6);
	const f = h * 6 - i;
	const p = Math.round(v * (1 - s));
	const q = Math.round(v * (1 - f * s));
	const t = Math.round(v * (1 - (1 - f) * s));

	switch (i % 6) {
		case 0:
			return [v, t, p];
		case 1:
			return [q, v, p];
		case 2:
			return [p, v, t];
		case 3:
			return [p, q, v];
		case 4:
			return [t, p, v];
		case 5:
			return [v, p, q];
	}
}

function turnOnMessage(applianceId) {
	return 'turn_on ' + JSON.stringify({ id: applianceId });
}

function turnOffMessage(applianceId) {
	return 'turn_off ' + JSON.stringify({ id: applianceId });
}

function brightnessMessage(applianceId, brightness) {
	return 'brightness ' + JSON.stringify({ id: applianceId, brightness: brightness });
}

function colorMessage(applianceId, color) {
	const rgb = hsvToRgb(color.hue, color.saturation, color.brightness);

	return 'color ' + JSON.stringify({
		id: applianceId,
		red: rgb[0],
		green: rgb[1],
		blue: rgb[2],
	});
}

function handleDiscovery(request, callback) {
	log('DEBUG', `Discovery Request: ${JSON.stringify(request)}`);

	// OAuth token
	const userAccessToken = request.directive.payload.scope.token;

	if (!userAccessToken || !isValidToken(userAccessToken)) {
		const errorMessage = `Discovery Request [${request.directive.header.messageId}] failed. Invalid access token: ${userAccessToken}`;
		log('ERROR', errorMessage);
		callback(new Error(errorMessage));
	}

	const response = {
		event: {
			header: {
				namespace: 'Alexa.Discovery',
				name: 'Discover.Response',
				messageId: uuidv4(),
				payloadVersion: '3',
			},
			payload: {
				endpoints: getDevicesFromPartnerCloud(userAccessToken),
			},
		},
	};

	log('DEBUG', `Discovery Response: ${JSON.stringify(response)}`);

	callback(null, response);
}

function handleCommonTasks(directive, callback) {
	log('DEBUG', `Control Request: ${JSON.stringify(directive)}`);

	const userAccessToken = directive.endpoint.scope.token;

	if (!userAccessToken || !isValidToken(userAccessToken)) {
		const errMsg = `Invalid access token: ${userAccessToken}`;
		log('ERROR', errMsg);
		callback(new Error(errMsg));
		return false;
	}

	if (!directive.endpoint.endpointId) {
		const errMsg = 'No endpointId provided in request';
		log('ERROR', errMsg);
		callback(new Error(errMsg));
		return false;
	}

	// if (!isDeviceOnline(endpointId, userAccessToken)) { callback(new Error('TargetOfflineError'), null); return; }

	return true;

}

function handlePowerControl(directive, callback) {
	if (!handleCommonTasks(directive, callback)) {
		return;
	}

	switch (directive.header.name) {
		case 'TurnOff':
			sendMessageToRaspberry(turnOffMessage(directive.endpoint.endpointId));

			callback(null, generateCommonControlResponse(
				directive,
				'Alexa.PowerController',
				'powerState',
				'ON'));
			return;
		case 'TurnOn':
			sendMessageToRaspberry(turnOnMessage(directive.endpoint.endpointId));

			callback(null, generateCommonControlResponse(
				directive,
				'Alexa.PowerController',
				'powerState',
				'OFF'));
			return;
		default: {
			const errMsg = `No supported directive name: ${directive.header.name}`;
			log('ERROR', errMsg);
			callback(new Error(errMsg));
			return;
		}
	}

	// log('DEBUG', `Control Confirmation: ${JSON.stringify(response)}`);
}

function handleBrightnessControl(directive, callback) {
	if (!handleCommonTasks(directive, callback)) {
		return;
	}

	switch (directive.header.name) {
		case 'SetBrightness':
			sendMessageToRaspberry(brightnessMessage(directive.endpoint.endpointId, directive.payload.brightness));

			callback(null, generateCommonControlResponse(
				directive,
				'Alexa.BrightnessController',
				'brightness',
				directive.payload.brightness));
			return;
		default: {
			const errMsg = `No supported directive name: ${directive.header.name}`;
			log('ERROR', errMsg);
			callback(new Error(errMsg));
			return;
		}
	}
}

function handleColorControl(directive, callback) {
	if (!handleCommonTasks(directive, callback)) {
		return;
	}

	switch (directive.header.name) {
		case 'SetColor':
			sendMessageToRaspberry(colorMessage(directive.endpoint.endpointId, directive.payload.color));

			callback(null, generateCommonControlResponse(
				directive,
				'Alexa.ColorController',
				'color',
				directive.payload.color));
			return;
		default: {
			const errMsg = `No supported directive name: ${directive.header.name}`;
			log('ERROR', errMsg);
			callback(new Error(errMsg));
			return;
		}
	}
}

function generateCommonControlResponse(directive, namespace, property, value) {
	return {
		context: {
			properties: [ {
				namespace: namespace,
				name: property,
				value: value,
				timeOfSample: new Date().toISOString(),
				uncertaintyInMilliseconds: 500,
			} ]
		},
		event: {
			header: {
				namespace: "Alexa",
				name: "Response",
				payloadVersion: "3",
				messageId: uuidv4(),
				correlationToken: directive.header.correlationToken,
			},
			endpoint: directive.endpoint,
			payload: {},
		},
	};
}

exports.handler = (request, context, callback) => {
	log('DEBUG', JSON.stringify(request));

	const directive = request.directive;

	switch (directive.header.namespace) {
		case 'Alexa.Discovery':
			if (directive.header.name !== 'Discover') {
				callback(new Error('Unsupported directive under Discovery'));
				return;
			}

			handleDiscovery(request, callback);
			break;
		case 'Alexa.PowerController':
			handlePowerControl(directive, callback);
			break;
		case 'Alexa.BrightnessController':
			handleBrightnessControl(directive, callback);
			break;
		case 'Alexa.ColorController':
			handleColorControl(directive, callback);
			break;
		default: { // unexpected message
			const errorMessage = `No supported namespace: ${directive.header.namespace}`;
			log('ERROR', errorMessage);
			callback(new Error(errorMessage));
		}
	}
};
