import * as http from 'http';
import * as https from 'https';

interface AmazonUserProfile {
	user_id: string;
	name: string;
	email: string;
}

// https://developer.amazon.com/docs/login-with-amazon/obtain-customer-profile.html#use-the-login-with-amazon-sdk-for-javascript
export function resolveUser(bearerToken: string): Promise<AmazonUserProfile> {
	return new Promise((resolve, reject) => {
		const bearerTokenEscaped = encodeURIComponent(bearerToken);

		const url = `https://api.amazon.com/user/profile?access_token=${bearerTokenEscaped}`;

		const req = https.request(url, (res: http.IncomingMessage) => {
			let responseBodyRaw = '';

			res.on('error', (err: Error) => {
				// console.log('response error', err);
				reject(err);
			});
			res.setEncoding('utf8');
			res.on('data', (chunk: string) => {
				responseBodyRaw += chunk;
			});
			res.on('end', () => {
				const bodyJson: AmazonUserProfile | null = JSON.parse(
					responseBodyRaw,
				);
				if (!bodyJson) {
					reject(
						new Error(`Failed to parse JSON: ${responseBodyRaw}`),
					);
					return;
				}

				resolve(bodyJson);
			});
		});
		req.on('error', (err: Error) => {
			reject(err);
		});
		req.end();
	});
}
