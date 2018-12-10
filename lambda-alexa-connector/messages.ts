export function turnOnMessage(applianceId: string): string {
	return 'turn_on ' + JSON.stringify({ id: applianceId });
}

export function turnOffMessage(applianceId: string): string {
	return 'turn_off ' + JSON.stringify({ id: applianceId });
}

export function brightnessMessage(
	applianceId: string,
	brightness: number,
): string {
	return 'brightness ' + JSON.stringify({ id: applianceId, brightness });
}

export function colorTemperature(
	applianceId: string,
	colorTemperatureInKelvin: number,
): string {
	return (
		'colorTemperature ' +
		JSON.stringify({ id: applianceId, colorTemperatureInKelvin })
	);
}

export function colorMessage(
	applianceId: string,
	color: { hue: number; saturation: number; brightness: number },
): string {
	const rgb = hsvToRgb(color.hue, color.saturation, color.brightness);

	return (
		'color ' +
		JSON.stringify({
			id: applianceId,
			red: rgb[0],
			green: rgb[1],
			blue: rgb[2],
		})
	);
}

export function playbackControlMessage(
	applianceId: string,
	action: string,
): string {
	return 'playback ' + JSON.stringify({ id: applianceId, action });
}

// helpers

function hsvToRgb(h: number, s: number, v: number): number[] {
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
		default:
			throw new Error('Should not happen');
	}
}
