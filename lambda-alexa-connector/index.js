'use strict';

var AWS = require('aws-sdk');
var sqs = new AWS.SQS({apiVersion: '2012-11-05'});

function createDevice(deviceId, friendlyName, friendlyDescription) {
    return     {
        // This id needs to be unique across all devices discovered for a given manufacturer
        applianceId: deviceId,
        manufacturerName: 'function61.com',
        modelName: 'home-automation-pi proxy',
        // Version number of the product
        version: '1.0',
        // The name given by the user in your application. Examples include 'Bedroom light' etc
        friendlyName: friendlyName,
        friendlyDescription: friendlyDescription,
        // at time of discovery
        isReachable: true,
        // List the actions the device can support from our API
        // The action should be the name of the actions listed here
        // https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/smart-home-skill-api-reference#discoverappliancesresponse
        actions: ['turnOn', 'turnOff']
        /*
        additionalApplianceDetails: {
            extraDetail1: 'optionalDetailForSkillAdapterToReferenceThisDevice',
            extraDetail2: 'There can be multiple entries',
            extraDetail3: 'but they should only be used for reference purposes.',
            extraDetail4: 'This is not a suitable place to maintain current device state',
        */
    };
}

const queueUrl = 'https://sqs.us-east-1.amazonaws.com/329074924855/JoonasHomeAutomation';

const USER_DEVICES = [
	createDevice('d2ff0882', 'Sofa light', 'Floor light next the sofa'),
	createDevice('98d3cb01', 'Speaker light', 'Floor light under the speaker'),
	createDevice('cfb1b27f', 'Living room', 'Device group: Living room'),
];

function log(title, msg) {
    console.log(`[${title}] ${msg}`);
}

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function generateMessageID() {
    return uuidv4();
}

function generateResponse(name, payload) {
    return {
        header: {
            messageId: generateMessageID(),
            name: name,
            namespace: 'Alexa.ConnectedHome.Control',
            payloadVersion: '2',
        },
        payload: payload,
    };
}

function getDevicesFromPartnerCloud() {
    return USER_DEVICES;
}

function isValidToken(token) { return true; }

function isDeviceOnline(applianceId) { return true; }

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

function turnOnMessage(applianceId) {
	return 'turn_on ' + JSON.stringify({ id: applianceId });
}

function turnOffMessage(applianceId) {
	return 'turn_off ' + JSON.stringify({ id: applianceId });
}

function turnOn(applianceId) {
    var msg = turnOnMessage(applianceId);
    log('DEBUG', msg);
    sendMessageToRaspberry(msg);

    return generateResponse('TurnOnConfirmation', {});
}

function turnOff(applianceId) {
    var msg = turnOffMessage(applianceId);
    log('DEBUG', msg);
    sendMessageToRaspberry(msg);

    return generateResponse('TurnOffConfirmation', {});
}

function setPercentage(applianceId, percentage) {
    log('DEBUG', `setPercentage (applianceId: ${applianceId}), percentage: ${percentage}`);

    // Call device cloud's API to set percentage

    return generateResponse('SetPercentageConfirmation', {});
}

function incrementPercentage(applianceId, delta) {
    log('DEBUG', `incrementPercentage (applianceId: ${applianceId}), delta: ${delta}`);

    // Call device cloud's API to set percentage delta

    return generateResponse('IncrementPercentageConfirmation', {});
}

function decrementPercentage(applianceId, delta) {
    log('DEBUG', `decrementPercentage (applianceId: ${applianceId}), delta: ${delta}`);

    // Call device cloud's API to set percentage delta

    return generateResponse('DecrementPercentageConfirmation', {});
}

function handleDiscovery(request, callback) {
    log('DEBUG', `Discovery Request: ${JSON.stringify(request)}`);

    // OAuth token
    const userAccessToken = request.payload.accessToken.trim();

    if (!userAccessToken || !isValidToken(userAccessToken)) {
        const errorMessage = `Discovery Request [${request.header.messageId}] failed. Invalid access token: ${userAccessToken}`;
        log('ERROR', errorMessage);
        callback(new Error(errorMessage));
    }

    const response = {
        header: {
            messageId: generateMessageID(),
            name: 'DiscoverAppliancesResponse',
            namespace: 'Alexa.ConnectedHome.Discovery',
            payloadVersion: '2',
        },
        payload: {
            discoveredAppliances: getDevicesFromPartnerCloud(userAccessToken),
        },
    };

    log('DEBUG', `Discovery Response: ${JSON.stringify(response)}`);

    callback(null, response);
}

function handleControl(request, callback) {
    log('DEBUG', `Control Request: ${JSON.stringify(request)}`);

    const userAccessToken = request.payload.accessToken.trim();

    if (!userAccessToken || !isValidToken(userAccessToken)) {
        log('ERROR', `Discovery Request [${request.header.messageId}] failed. Invalid access token: ${userAccessToken}`);
        callback(null, generateResponse('InvalidAccessTokenError', {}));
        return;
    }

    const applianceId = request.payload.appliance.applianceId;

    if (!applianceId) {
        log('ERROR', 'No applianceId provided in request');
        const payload = { faultingParameter: `applianceId: ${applianceId}` };
        callback(null, generateResponse('UnexpectedInformationReceivedError', payload));
        return;
    }

    if (!isDeviceOnline(applianceId, userAccessToken)) {
        log('ERROR', `Device offline: ${applianceId}`);
        callback(null, generateResponse('TargetOfflineError', {}));
        return;
    }

    let response;

    switch (request.header.name) {
        case 'TurnOnRequest':
            response = turnOn(applianceId, userAccessToken);
            break;

        case 'TurnOffRequest':
            response = turnOff(applianceId, userAccessToken);
            break;

        case 'SetPercentageRequest': {
            const percentage = request.payload.percentageState.value;
            if (!percentage) {
                const payload = { faultingParameter: `percentageState: ${percentage}` };
                callback(null, generateResponse('UnexpectedInformationReceivedError', payload));
                return;
            }
            response = setPercentage(applianceId, userAccessToken, percentage);
            break;
        }

        case 'IncrementPercentageRequest': {
            const delta = request.payload.deltaPercentage.value;
            if (!delta) {
                const payload = { faultingParameter: `deltaPercentage: ${delta}` };
                callback(null, generateResponse('UnexpectedInformationReceivedError', payload));
                return;
            }
            response = incrementPercentage(applianceId, userAccessToken, delta);
            break;
        }

        case 'DecrementPercentageRequest': {
            const delta = request.payload.deltaPercentage.value;
            if (!delta) {
                const payload = { faultingParameter: `deltaPercentage: ${delta}` };
                callback(null, generateResponse('UnexpectedInformationReceivedError', payload));
                return;
            }
            response = decrementPercentage(applianceId, userAccessToken, delta);
            break;
        }

        default: {
            log('ERROR', `No supported directive name: ${request.header.name}`);
            callback(null, generateResponse('UnsupportedOperationError', {}));
            return;
        }
    }

    log('DEBUG', `Control Confirmation: ${JSON.stringify(response)}`);

    callback(null, response);
}

exports.handler = (request, context, callback) => {
    switch (request.header.namespace) {
        case 'Alexa.ConnectedHome.Discovery':
            handleDiscovery(request, callback);
            break;

        case 'Alexa.ConnectedHome.Control':
            handleControl(request, callback);
            break;

        // case 'Alexa.ConnectedHome.Query':
        //     handleQuery(request, callback);
        //     break;

        default: { // unexpected message
            const errorMessage = `No supported namespace: ${request.header.namespace}`;
            log('ERROR', errorMessage);
            callback(new Error(errorMessage));
        }
    }
};
