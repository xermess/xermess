/**
 * A failed request, as the server described it: `code` names the problem, and
 * a translated page says `error.<code>` in the administrator's language with
 * `params` filled in (messageOf in $lib/i18n). `message` is the server's
 * English, which the pages not yet translated show as it is.
 *
 * `status` is 0 when the server could not be reached, and the code is then
 * `network`; an answer that was not the server's is `unknown`.
 */
export class ApiError extends Error {
	constructor(
		readonly status: number,
		message: string,
		readonly code = 'unknown',
		readonly params: Record<string, string | number> = {}
	) {
		super(message);
		this.name = 'ApiError';
	}

	/** The session is missing or has expired. */
	get isUnauthorized(): boolean {
		return this.status === 401;
	}
}

/**
 * SvelteKit hands `load` functions their own fetch, which it uses to track
 * dependencies and to replay requests on the client. Passing it in is what
 * makes `invalidate` work; anything outside a load can leave it out.
 */
export type Fetch = typeof globalThis.fetch;

type RequestOptions = {
	method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
	body?: unknown;
	fetch?: Fetch;
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
	const { method = 'GET', body, fetch: fetcher = globalThis.fetch } = options;

	let response: Response;
	try {
		// The admin API is on this app's own origin: the dev server's proxy,
		// or the reverse proxy in production, routes /api/v1/admin to it.
		response = await fetcher(`/api/v1${path}`, {
			method,
			credentials: 'same-origin',
			headers: { 'Content-Type': 'application/json' },
			body: body === undefined ? undefined : JSON.stringify(body)
		});
	} catch {
		throw new ApiError(0, 'Could not reach the server. Is the admin API running?', 'network');
	}

	// A 204 has no body to read.
	const payload = response.status === 204 ? {} : await response.json().catch(() => ({}));

	if (!response.ok) {
		throw new ApiError(
			response.status,
			payload.error ?? 'Something went wrong',
			payload.code ?? 'unknown',
			payload.params ?? {}
		);
	}

	return payload as T;
}

export const api = {
	get: <T>(path: string, fetcher?: Fetch) => request<T>(path, { fetch: fetcher }),
	post: <T>(path: string, body?: unknown, fetcher?: Fetch) =>
		request<T>(path, { method: 'POST', body, fetch: fetcher }),
	put: <T>(path: string, body?: unknown, fetcher?: Fetch) =>
		request<T>(path, { method: 'PUT', body, fetch: fetcher }),
	patch: <T>(path: string, body?: unknown, fetcher?: Fetch) =>
		request<T>(path, { method: 'PATCH', body, fetch: fetcher }),
	delete: <T>(path: string, body?: unknown, fetcher?: Fetch) =>
		request<T>(path, { method: 'DELETE', body, fetch: fetcher })
};
