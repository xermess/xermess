import type { Application, Organization } from '$lib/api';

/** One agreement a page links to. */
export type LegalLink = { label: string; href: string };

/**
 * The agreements shown on a sign-in page: the application's own where it has
 * them, and the organisation's where it has not.
 *
 * Each link falls back on its own, because they are two documents and an
 * application may publish one without the other. A link nobody has published
 * is not shown at all rather than pointing nowhere.
 */
export function legalLinks(
	application: Application | null | undefined,
	organization: Organization | null | undefined
): LegalLink[] {
	const terms = application?.tos_uri || organization?.terms_url || '';
	const privacy = application?.policy_uri || organization?.privacy_url || '';

	return [
		terms ? { label: 'Terms of service', href: terms } : null,
		privacy ? { label: 'Privacy policy', href: privacy } : null
	].filter((link) => link !== null);
}

/** How to reach the organisation for help, as the page foot lists it: an
    address to write to and a number to ring, whichever of them is published. */
export function supportLinks(organization: Organization | null | undefined): LegalLink[] {
	return [
		organization?.support_email
			? { label: organization.support_email, href: `mailto:${organization.support_email}` }
			: null,
		organization?.support_phone
			? {
					label: organization.support_phone,
					// A dialled number carries no spaces or brackets.
					href: `tel:${organization.support_phone.replace(/[^+0-9]/g, '')}`
				}
			: null
	].filter((link) => link !== null);
}
