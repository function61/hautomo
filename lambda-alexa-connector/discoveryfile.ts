import {
	AlexaInterface,
	brightnessController,
	Category,
	colorController,
	Device,
	playbackController,
	powerController,
} from './types';
import { assertUnreachable } from './utils';

enum CapabilityCode {
	PowerController = 'PowerController',
	BrightnessController = 'BrightnessController',
	ColorController = 'ColorController',
	PlaybackController = 'PlaybackController',
}

interface DiscoveryFileDevice {
	id: string;
	friendly_name: string;
	description: string;
	display_category: string;
	capability_codes: CapabilityCode[];
}

export interface DiscoveryFile {
	user_token_hash: string; // FIXME: this is unneeded
	queue: string;
	devices: DiscoveryFileDevice[];
}

export function toAlexaStruct(file: DiscoveryFile): Device[] {
	return file.devices.map(
		(device): Device => {
			if (device.display_category !== 'LIGHT') {
				throw new Error(
					`Unexpected display_category: ${device.display_category}`,
				);
			}

			const caps: AlexaInterface[] = device.capability_codes.map(
				(code): AlexaInterface => {
					switch (code) {
						case CapabilityCode.PowerController:
							return powerController();
						case CapabilityCode.BrightnessController:
							return brightnessController();
						case CapabilityCode.ColorController:
							return colorController();
						case CapabilityCode.PlaybackController:
							return playbackController();
						default:
							return assertUnreachable(code);
					}
				},
			);

			return {
				endpointId: device.id,
				manufacturerName: 'function61.com',
				version: '1.0',
				friendlyName: device.friendly_name,
				description: device.description,
				displayCategories: [Category.LIGHT],
				capabilities: caps,
				cookie: {
					queue: file.queue,
				},
			};
		},
	);
}
