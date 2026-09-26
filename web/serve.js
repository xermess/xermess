// Serves a built app (adapter-node) with its API paths in front, the way it
// runs in a deployment. There Caddy routes the API paths to Loginer before a
// request reaches the app (deploy/Caddyfile); locally, for `scripts/start.sh
// --prod`, this does that one job. Run it from the app's directory:
//
//   PORT=5173 API_URL=http://localhost:8080 API_PATHS=/oauth2,/.well-known node ../serve.js
//
// The app's own settings (ORIGIN, BODY_SIZE_LIMIT, …) are read by its handler.

import http from 'node:http';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const { handler } = await import(pathToFileURL(resolve('build/handler.js')).href);

const { PORT = '3000', HOST = '127.0.0.1', API_URL, API_PATHS = '' } = process.env;
const api = API_URL && new URL(API_URL);
const paths = API_PATHS.split(',').filter(Boolean);

const isApi = (url) => paths.some((p) => url === p || url.startsWith(`${p}/`) || url.startsWith(`${p}?`));

function forward(req, res) {
	const headers = {
		...req.headers,
		'x-forwarded-for': req.socket.remoteAddress,
		'x-forwarded-host': req.headers.host,
		'x-forwarded-proto': 'http'
	};

	const upstream = http.request(
		{ host: api.hostname, port: api.port, method: req.method, path: req.url, headers },
		(reply) => {
			// Raw headers keep repeated ones, such as several Set-Cookie, apart.
			res.writeHead(reply.statusCode, reply.statusMessage, reply.rawHeaders);
			reply.pipe(res);
		}
	);

	upstream.on('error', () => {
		if (!res.headersSent) res.writeHead(502, { 'content-type': 'text/plain' });
		res.end('The Loginer API is not reachable.');
	});

	req.pipe(upstream);
}

http
	.createServer((req, res) => (api && isApi(req.url) ? forward(req, res) : handler(req, res)))
	.listen(Number(PORT), HOST, () => console.log(`listening on http://${HOST}:${PORT}`));
