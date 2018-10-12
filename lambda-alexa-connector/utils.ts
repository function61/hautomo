import * as crypto from 'crypto';
import {
	AlexaNamespace,
	AlexaPowerOutput,
	ContextProperty,
	EndpointSpec,
} from './types';

export function assertUnreachable(x: never): never {
	throw new Error("Didn't expect to get here");
}

export function uuidv4(): string {
	return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
		// tslint:disable-next-line:no-bitwise
		const r = (Math.random() * 16) | 0;
		// tslint:disable-next-line:no-bitwise
		const v = c === 'x' ? r : (r & 0x3) | 0x8;
		return v.toString(16);
	});
}

export function log(msg: string): void {
	// tslint:disable-next-line:no-console
	console.log(msg);
}

export function generateCommonControlResponse(
	property: string | null,
	value: any,
	endpoint: EndpointSpec,
	correlationToken: string,
	namespace: AlexaNamespace,
): AlexaPowerOutput {
	const properties: ContextProperty[] = [];

	if (property !== null) {
		properties.push({
			namespace,
			name: property,
			value,
			timeOfSample: new Date().toISOString(),
			uncertaintyInMilliseconds: 500,
		});
	}

	return {
		context: {
			properties,
		},
		event: {
			header: {
				namespace: 'Alexa',
				name: 'Response',
				payloadVersion: '3',
				messageId: uuidv4(),
				correlationToken,
			},
			endpoint,
			payload: {},
		},
	};
}

export function sha1Hex(input: string): string {
	const sha1 = crypto.createHash('sha1');
	sha1.update(input);
	return sha1.digest('hex');
}
