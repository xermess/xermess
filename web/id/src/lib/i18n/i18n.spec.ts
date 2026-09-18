import { describe, expect, it } from 'vitest';

import { chooseLanguage, translator } from './index';

/** What these tests stand on: an installation offering en, ky and ru. */
const offered = ['en', 'ky', 'ru'];

describe('chooseLanguage', () => {
	it('gives somebody the language they chose', () => {
		expect(chooseLanguage('ky', 'en-GB,en;q=0.9', offered, 'en')).toBe('ky');
	});

	it("falls back to the browser's when nobody has chosen", () => {
		expect(chooseLanguage(undefined, 'ru,en;q=0.8', offered, 'en')).toBe('ru');
	});

	it('reads a region off a tag it does not have on its own', () => {
		expect(chooseLanguage(undefined, 'ru-RU,en;q=0.8', offered, 'en')).toBe('ru');
	});

	it('takes the languages in the order the browser ranked them', () => {
		expect(chooseLanguage(undefined, 'de;q=0.9,ky;q=0.8,ru;q=0.7', offered, 'en')).toBe('ky');
	});

	it('passes over a saved choice that is no longer offered', () => {
		expect(chooseLanguage('ru', null, ['en', 'ky'], 'en')).toBe('en');
	});

	it("falls back to the installation's default", () => {
		expect(chooseLanguage(undefined, 'de,fr;q=0.9', offered, 'ky')).toBe('ky');
	});

	it('falls back to the first offered when the default is not among them', () => {
		expect(chooseLanguage(undefined, null, ['ky', 'ru'], 'de')).toBe('ky');
	});

	it('falls back to the base language when nothing is offered at all', () => {
		expect(chooseLanguage(undefined, null, [], 'de')).toBe('en');
	});
});

describe('translator', () => {
	const t = (messages: Record<string, string>) => translator(() => messages);

	it('looks a key up in the text it was given', () => {
		expect(t({ 'action.sign_in': 'Войти' })('action.sign_in')).toBe('Войти');
	});

	it('falls back to the base language this app was built with', () => {
		// What happens when the server could not be asked for any text.
		expect(t({})('action.sign_in')).toBe('Sign in');
	});

	it('falls back to the key itself when nothing has it', () => {
		expect(t({})('nothing.here')).toBe('nothing.here');
	});

	it('fills in parameters', () => {
		expect(t({})('login.subtitle_app', { app: 'Shop' })).toBe('to continue to Shop');
	});

	it('leaves a parameter nobody passed as it was written', () => {
		expect(t({})('login.subtitle_app')).toBe('to continue to {app}');
	});
});
