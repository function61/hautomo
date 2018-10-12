import { processEvent } from './index';
import { log } from './utils';

function assertEqual(actual: any, expected: any) {
	if (actual !== expected) {
		throw new Error('assertEqual:' + actual);
	}
}

function handlesWarmupEvent(): Promise<void> {
	return new Promise((resolve, reject) => {
		processEvent({ warmup: true }).then((result) => {
			assertEqual(result.response, 'ok');
			resolve();
		}, reject);
	});
}

function handlesUnknownMessagesGracefully(): Promise<void> {
	return new Promise((resolve, reject) => {
		processEvent({ wrong_message_type: 'foobar' } as any).then((result) => {
			assertEqual(/Unknown msg/.test(result.error!.toString()), true);
			resolve();
		}, reject);
	});
}

process.on('unhandledRejection', () => {
	log('unhandledRejection');
	process.exit(1);
});

Promise.all([handlesWarmupEvent(), handlesUnknownMessagesGracefully()]).then(
	() => {
		// tslint:disable-next-line:no-console
		log('PASS');
	},
	() => {
		// tslint:disable-next-line:no-console
		log('FAIL');
		process.exit(1);
	},
);
