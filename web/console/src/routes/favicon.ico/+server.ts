import { asset } from '$app/paths';
import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

/**
 * A browser asks for /favicon.ico when the thing it is showing named no icon:
 * the provider's JSON at /.well-known/openid-configuration, /oauth2/* and the
 * account API, all of which answer on this app's origin. Pages name the icon
 * themselves in app.html and never come here.
 *
 * It answers with the icon the rest of the app uses rather than a second copy
 * in another format, so there is one file to change. A redirect, rather than
 * the bytes, because that keeps that one file in static/ — where it is served
 * with its own type, which an .ico is not: both static handlers look a type up
 * through mrmime, which has no entry for .ico, and the edge sets
 * X-Content-Type-Options: nosniff (deploy/Caddyfile), so an icon served
 * without one would be refused rather than guessed at.
 */
export const GET: RequestHandler = () => {
	redirect(302, asset('/favicon.svg'));
};
