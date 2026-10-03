import { asset } from '$app/paths';
import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

/**
 * Browsers request /favicon.ico for the API's JSON pages served on this origin. It redirects to
 * the static icon because the static handlers have no MIME type for .ico, and nosniff
 * (deploy/Caddyfile) would refuse it.
 */
export const GET: RequestHandler = () => {
	redirect(302, asset('/favicon.svg'));
};
