import * as AWS from 'aws-sdk';

const sqs = new AWS.SQS({ apiVersion: '2012-11-05' });
const s3 = new AWS.S3({ apiVersion: '2006-03-01' });

import { resolveUser } from './amazonuserresolver';
import { DiscoveryFile, toAlexaStruct } from './discoveryfile';
import {
	brightnessMessage,
	colorMessage,
	playbackControlMessage,
	turnOffMessage,
	turnOnMessage,
} from './messages';
import {
	AlexaBrightnessInput,
	AlexaColorInput,
	AlexaDiscoveryInput,
	AlexaDiscoveryOutput,
	AlexaGenericMessage,
	AlexaNamespace,
	AlexaPlaybackInput,
	AlexaPowerInput,
	AlexaPowerOutput,
	LambdaCallback,
	WarmupMsg,
} from './types';
import {
	assertUnreachable,
	generateCommonControlResponse,
	log,
	sha1Hex,
	uuidv4,
} from './utils';

function handleDiscovery(
	request: AlexaDiscoveryInput,
): Promise<ProcessResult<AlexaDiscoveryOutput>> {
	if (request.header.name !== 'Discover') {
		return Promise.resolve({
			error: new Error('Unsupported directive under Discovery'),
		});
	}

	return new Promise((resolve, reject) => {
		resolveUser(request.payload.scope.token).then((profile) => {
			const userId = profile.user_id;

			log(`Discovery for ${userId}`);

			let userIdOrHack = userId;

			// FIXME: temp workaround until I can update one user's (mr. V) device
			if (
				sha1Hex(userId) === '1b3206a6fd66579cbbbf1f671e3c4a9f9417314c'
			) {
				userIdOrHack = 'a93be8a8c85febf938c6edd0b1dc5c8f32dccb3f';
			}

			s3.getObject(
				{
					Bucket: 'homeautomation.function61.com',
					Key: `discovery/${userIdOrHack}.json`,
				},
				(err: Error, data: any) => {
					if (err) {
						resolve({ error: err });
						return;
					}

					const discoveryFile: DiscoveryFile = JSON.parse(data.Body);

					resolve({
						response: {
							event: {
								header: {
									namespace: AlexaNamespace.Discovery,
									name: 'Discover.Response',
									messageId: uuidv4(),
									payloadVersion: '3',
								},
								payload: {
									endpoints: toAlexaStruct(discoveryFile),
								},
							},
						},
					});
				},
			);
		}, reject);
	});
}

function handlePowerControl(
	request: AlexaPowerInput,
): Promise<ProcessResult<AlexaPowerOutput>> {
	const newState = request.header.name === 'TurnOn' ? 'ON' : 'OFF';

	const msg =
		newState === 'ON'
			? turnOnMessage(request.endpoint.endpointId)
			: turnOffMessage(request.endpoint.endpointId);

	return Promise.resolve({
		piMsg: { msg, queue: request.endpoint.cookie.queue },
		response: generateCommonControlResponse(
			'powerState',
			newState,
			request.endpoint,
			request.header.correlationToken,
			request.header.namespace,
		),
	});
}

function handleBrightnessControl(
	request: AlexaBrightnessInput,
): Promise<ProcessResult<AlexaPowerOutput>> {
	if (request.header.name !== 'SetBrightness') {
		return Promise.resolve({ error: new Error('Unexpected directive') });
	}

	return Promise.resolve({
		piMsg: {
			msg: brightnessMessage(
				request.endpoint.endpointId,
				request.payload.brightness,
			),
			queue: request.endpoint.cookie.queue,
		},
		response: generateCommonControlResponse(
			'brightness',
			request.payload.brightness,
			request.endpoint,
			request.header.correlationToken,
			request.header.namespace,
		),
	});
}

function handleColorControl(
	request: AlexaColorInput,
): Promise<ProcessResult<AlexaPowerOutput>> {
	if (request.header.name !== 'SetColor') {
		return Promise.resolve({ error: new Error('Unexpected directive') });
	}

	return Promise.resolve({
		piMsg: {
			msg: colorMessage(
				request.endpoint.endpointId,
				request.payload.color,
			),
			queue: request.endpoint.cookie.queue,
		},
		response: generateCommonControlResponse(
			'color',
			request.payload.color,
			request.endpoint,
			request.header.correlationToken,
			request.header.namespace,
		),
	});
}

function handlePlaybackControl(
	request: AlexaPlaybackInput,
): Promise<ProcessResult<AlexaPowerOutput>> {
	return Promise.resolve({
		piMsg: {
			msg: playbackControlMessage(
				request.endpoint.endpointId,
				request.header.name,
			),
			queue: request.endpoint.cookie.queue,
		},
		response: generateCommonControlResponse(
			null,
			null,
			request.endpoint,
			request.header.correlationToken,
			request.header.namespace,
		),
	});
}

function isWarmupMsg(input: any): input is WarmupMsg {
	return 'warmup' in input;
}

function isAlexaGenericMessage(input: any): input is AlexaGenericMessage {
	return 'directive' in input;
}

interface ProcessResult<T> {
	error?: Error;
	response?: T;
	piMsg?: {
		msg: string;
		queue: string;
	};
}

export function processEvent(
	request: WarmupMsg | AlexaGenericMessage,
): Promise<ProcessResult<any>> {
	if (isWarmupMsg(request)) {
		return Promise.resolve({ response: 'ok' });
	} else if (isAlexaGenericMessage(request)) {
		const directive = request.directive;

		log(`[Alexa message] ${JSON.stringify(request)}`);

		switch (directive.header.namespace) {
			case AlexaNamespace.Discovery:
				return handleDiscovery(directive as AlexaDiscoveryInput);
			case AlexaNamespace.PowerController:
				return handlePowerControl(directive as AlexaPowerInput);
			case AlexaNamespace.BrightnessController:
				return handleBrightnessControl(
					directive as AlexaBrightnessInput,
				);
			case AlexaNamespace.ColorController:
				return handleColorControl(directive as AlexaColorInput);
			case AlexaNamespace.PlaybackController:
				return handlePlaybackControl(directive as AlexaPlaybackInput);
			default: {
				// unexpected message
				assertUnreachable(directive.header.namespace);
				const errorMessage = `No supported namespace: ${
					directive.header.namespace
				}`;
				log(`[ERROR] ${errorMessage}`);
				return Promise.resolve({ error: new Error(errorMessage) });
			}
		}
	} else {
		return Promise.resolve({
			error: new Error(`Unknown msg: ${JSON.stringify(request)}`),
		});
	}
}

export function handler(
	request: WarmupMsg | AlexaGenericMessage,
	context: undefined,
	callback: LambdaCallback,
) {
	processEvent(request).then(
		(result) => {
			if (result.error !== undefined) {
				callback(result.error);
				return;
			}

			callback(null, result.response);

			if (result.piMsg) {
				if (result.piMsg.queue !== 'drop') {
					sqs.sendMessage(
						{
							MessageBody: result.piMsg.msg,
							QueueUrl: result.piMsg.queue,
						},
						(err: Error, data: any) => {
							if (err) {
								log(err.toString());
							} else {
								log(data);
							}
						},
					);
				} else {
					log('dropping queue message');
				}
			}
		},
		(err) => {
			callback(err);
		},
	);
}
