import {
	RiAppleFill,
	RiFacebookCircleFill,
	RiGoogleFill,
	RiKey2Line,
	RiShieldKeyholeLine,
	RiVkFill
} from 'svelte-remixicon';
import type { ComponentType } from 'svelte';

import type { SocialKind, SocialProvider, SocialSpec } from '$lib/api';
import type { SelectOption } from '$lib/components/ui';

/** The mark each kind is known by. The two custom kinds have no logo of their
    own, so they take a key and a shield. */
const marks: Record<SocialKind, ComponentType> = {
	google: RiGoogleFill,
	apple: RiAppleFill,
	facebook: RiFacebookCircleFill,
	yandex: RiKey2Line,
	vk: RiVkFill,
	oidc: RiShieldKeyholeLine,
	oauth2: RiKey2Line
};

export function markFor(kind: SocialKind): ComponentType {
	return marks[kind] ?? RiKey2Line;
}

/** The kinds a new provider may be, as the picker offers them. */
export function kindOptions(kinds: SocialSpec[]): SelectOption<SocialKind>[] {
	return kinds.map((kind) => ({
		value: kind.kind,
		label: kind.label,
		icon: markFor(kind.kind),
		description: kind.custom
			? 'Any provider, with its own addresses'
			: `Signed in at ${hostOf(kind.authorize_url)}`
	}));
}

/** How the secret is sent at the token endpoint. Only the custom kinds ask:
    the rest are known. */
export const tokenAuthOptions: SelectOption<'' | 'basic' | 'post'>[] = [
	{
		value: '',
		label: "The provider's usual way",
		description: 'What this kind of provider expects'
	},
	{ value: 'basic', label: 'HTTP Basic', description: 'Every OAuth 2.0 server accepts this' },
	{ value: 'post', label: 'In the request body', description: 'client_secret as a form field' }
];

/** The host part of an address, for saying where people are sent without
    showing the whole query. */
export function hostOf(address: string): string {
	try {
		return new URL(address).host;
	} catch {
		return address;
	}
}

/** What a provider still needs before anybody can sign in with it. The panel
    says so rather than letting a button appear that cannot work. */
export function missingFrom(provider: SocialProvider, spec: SocialSpec | undefined): string {
	if (spec?.signed_secret) {
		if (!provider.has_private_key) return 'no signing key';
		if (!provider.team_id || !provider.key_id) return 'no team or key id';
		return '';
	}

	return provider.has_client_secret ? '' : 'no client secret';
}
