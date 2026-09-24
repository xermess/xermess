import { signIn, type LoginOptions } from '$lib/api';
import { requireUser } from '$lib/server/session';
import type { LayoutServerLoad } from './$types';

/** Every account page is for the signed-in user; anyone else signs in first,
    and comes back to the page they asked for.
 
    The login options come along because they say what a user may do with
    their own account — change the address they sign in with, today. There is
    no application here, so it is the installation's default flow that
    decides, which is the flow somebody signing in to their account followed. */
export const load: LayoutServerLoad = async ({ cookies, fetch, url, setHeaders }) => {
	setHeaders({ 'cache-control': 'private, no-store' });

	const user = await requireUser(cookies, fetch, url);

	// A page that hides a control because one request failed is worse than one
	// that offers it and is refused, so this answers null rather than throwing.
	const login: LoginOptions | null = await signIn
		.loginOptions('', fetch)
		.then((body) => body.login)
		.catch(() => null);

	return { user, login };
};
