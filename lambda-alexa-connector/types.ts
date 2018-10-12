export type LambdaCallback = (err: Error | null, response?: string) => void;

export interface WarmupMsg {
	warmup: boolean;
}

export enum AlexaNamespace {
	Discovery = 'Alexa.Discovery',
	PowerController = 'Alexa.PowerController',
	BrightnessController = 'Alexa.BrightnessController',
	ColorController = 'Alexa.ColorController',
	PlaybackController = 'Alexa.PlaybackController',
}

export enum Category {
	LIGHT = 'LIGHT',
	TV = 'TV',
	OTHER = 'OTHER',
	SPEAKER = 'SPEAKER',
}

export interface AlexaInterface {
	type: 'AlexaInterface';
	interface: AlexaNamespace;
	version: '3';
	properties?: any;
	supportedOperations?: string[];
}

export interface Device {
	endpointId: string;
	manufacturerName: string;
	version: string;
	friendlyName: string;
	description: string;
	displayCategories: Category[];
	capabilities: AlexaInterface[];
	cookie: { [key: string]: string };
}

export interface AlexaGenericMessage {
	directive: {
		header: {
			namespace: AlexaNamespace;
		};
	};
}

export interface AlexaScope {
	type: 'BearerToken';
	token: string;
}

export interface AlexaDiscoveryInput {
	header: {
		namespace: AlexaNamespace.Discovery;
		name: 'Discover';
	};
	payload: {
		scope: AlexaScope;
	};
}

export interface AlexaDiscoveryOutput {
	event: {
		header: {
			namespace: string; // TODO: AlexaNamespace.Discovery
			name: string; // TODO: 'Discovery.Response'
			messageId: string;
			payloadVersion: string; // TODO: '3'
		};
		payload: {
			endpoints: Device[];
		};
	};
}

export interface ContextProperty {
	namespace: AlexaNamespace;
	name: string;
	value: any;
	timeOfSample: string;
	uncertaintyInMilliseconds: number;
}

export interface EndpointSpec {
	scope: AlexaScope;
	endpointId: string;
	cookie: { [key: string]: string };
}

// ---------------- Power

export interface AlexaPowerInput {
	header: {
		namespace: AlexaNamespace.PowerController;
		name: 'TurnOn' | 'TurnOff';
		payloadVersion: string;
		messageId: string;
		correlationToken: string;
	};
	endpoint: EndpointSpec;
}

export interface AlexaPowerOutput {
	context: {
		properties: ContextProperty[];
	};
	event: {
		header: {
			namespace: string; // TODO: 'Alexa'
			name: string; // TODO: 'Response'
			payloadVersion: string; // TODO: '3'
			messageId: string;
			correlationToken: string;
		};
		endpoint: EndpointSpec;
		payload: {};
	};
}

// ---------------- Brightness

export interface AlexaBrightnessInput {
	header: {
		namespace: AlexaNamespace.BrightnessController;
		name: 'SetBrightness';
		payloadVersion: string;
		messageId: string;
		correlationToken: string;
	};
	endpoint: EndpointSpec;
	payload: {
		brightness: number;
	};
}

export interface AlexaColorInput {
	header: {
		namespace: AlexaNamespace.ColorController;
		name: 'SetColor';
		payloadVersion: string;
		messageId: string;
		correlationToken: string;
	};
	endpoint: EndpointSpec;
	payload: {
		color: {
			hue: number;
			saturation: number;
			brightness: number;
		};
	};
}

export interface AlexaPlaybackInput {
	header: {
		namespace: AlexaNamespace.PlaybackController;
		name: 'Play' | 'Stop' | 'Pause';
		payloadVersion: string;
		messageId: string;
		correlationToken: string;
	};
	endpoint: EndpointSpec;
	payload: {};
}

// factories for various AlexaInterface structs

export function powerController(): AlexaInterface {
	return {
		type: 'AlexaInterface',
		interface: AlexaNamespace.PowerController,
		version: '3',
		properties: {
			supported: [{ name: 'powerState' }],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

export function brightnessController(): AlexaInterface {
	return {
		type: 'AlexaInterface',
		interface: AlexaNamespace.BrightnessController,
		version: '3',
		properties: {
			supported: [{ name: 'brightness' }],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

export function colorController(): AlexaInterface {
	return {
		type: 'AlexaInterface',
		interface: AlexaNamespace.ColorController,
		version: '3',
		properties: {
			supported: [{ name: 'color' }],
			proactivelyReported: false,
			retrievable: false,
		},
	};
}

export function playbackController(): AlexaInterface {
	return {
		type: 'AlexaInterface',
		interface: AlexaNamespace.PlaybackController,
		version: '3',
		supportedOperations: ['Play', 'Pause', 'Stop'],
	};
}

/*
	BrightnessController = 'Alexa.BrightnessController',
	ColorController = 'Alexa.ColorController',
	PlaybackController = 'Alexa.PlaybackController',
*/
