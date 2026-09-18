import type { Application, Organization } from '$lib/api';

/** One agreement a page links to. The label is a message key rather than
    text: which document it is does not change with the language, and the page
    drawing it has the translator. */
export type LegalLink = { key: string; href: string };

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
		terms ? { key: 'legal.terms', href: terms } : null,
		privacy ? { key: 'legal.privacy', href: privacy } : null
	].filter((link) => link !== null);
}

/** One way to reach the organisation for help: the address or the number
    itself is the label, so it is not translated. */
export type SupportLink = { label: string; href: string };

/** How to reach the organisation for help, as the page foot lists it: an
    address to write to and a number to ring, whichever of them is published. */
export function supportLinks(organization: Organization | null | undefined): SupportLink[] {
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
