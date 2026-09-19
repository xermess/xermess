/**
 * A failed request, as the server described it: `code` names the problem, and
 * a page says `error.<code>` from its own catalog, in the reader's language,
 * with `params` filled in (messageOf in $lib/i18n). `message` is the server's
 * English, for when a page has no sentence of its own for the code.
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
}

export type Fetch = typeof globalThis.fetch;

type Options = {
	method?: 'GET' | 'POST' | 'PATCH' | 'DELETE';
	body?: unknown;
	fetch?: Fetch;
};

/**
 * Calls the xermess API, at /api/v1 on this app's own origin. From the browser
 * the proxy routes it to the API; from a server load, pass the load's `fetch`,
 * and hooks.server.ts sends it to the API directly with the reader's cookie.
 */
export async function request<T>(path: string, options: Options = {}): Promise<T> {
	const { method = 'GET', body, fetch: fetcher = globalThis.fetch } = options;

	const headers: Record<string, string> = {};
	if (body !== undefined) headers['Content-Type'] = 'application/json';

	let response: Response;
	try {
		response = await fetcher(`/api/v1${path}`, {
			method,
			credentials: 'same-origin',
			headers,
			body: body === undefined ? undefined : JSON.stringify(body)
		});
	} catch {
		throw new ApiError(
			0,
			'Could not reach the server. Check your connection and try again.',
			'network'
		);
	}

	const payload = response.status === 204 ? {} : await response.json().catch(() => ({}));

	if (!response.ok) {
		throw new ApiError(
			response.status,
			payload.error ?? 'Something went wrong. Try again.',
			payload.code ?? 'unknown',
			payload.params ?? {}
		);
	}

	return payload as T;
}

/** The identity provider an address has to sign in through, when a password
    sign-in, registration or reset was refused for it — `sso_required`, which
    names it — and null for any other error. */
export function requiredSSO(err: unknown): { slug: string; name: string } | null {
	if (!(err instanceof ApiError) || err.code !== 'sso_required') return null;

	const { slug, name } = err.params;
	return typeof slug === 'string' && typeof name === 'string' ? { slug, name } : null;
}
