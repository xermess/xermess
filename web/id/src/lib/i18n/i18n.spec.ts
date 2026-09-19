import { describe, expect, it } from 'vitest';

import { ApiError } from '$lib/api/client';
import { chooseLanguage, messageOf, translator } from './index';

/** What these tests stand on: an installation offering en, de and ru. */
const offered = ['en', 'de', 'ru'];

describe('chooseLanguage', () => {
	it('gives somebody the language they chose', () => {
		expect(chooseLanguage('de', 'en-GB,en;q=0.9', offered, 'en')).toBe('de');
	});

	it("falls back to the browser's when nobody has chosen", () => {
		expect(chooseLanguage(undefined, 'ru,en;q=0.8', offered, 'en')).toBe('ru');
	});

	it('reads a region off a tag it does not have on its own', () => {
		expect(chooseLanguage(undefined, 'ru-RU,en;q=0.8', offered, 'en')).toBe('ru');
	});

	it('takes the languages in the order the browser ranked them', () => {
		expect(chooseLanguage(undefined, 'fr;q=0.9,de;q=0.8,ru;q=0.7', offered, 'en')).toBe('de');
	});

	it('passes over a saved choice that is no longer offered', () => {
		expect(chooseLanguage('ru', null, ['en', 'de'], 'en')).toBe('en');
	});

	it("falls back to the installation's default", () => {
		expect(chooseLanguage(undefined, 'fr,it;q=0.9', offered, 'de')).toBe('de');
	});

	it('falls back to the first offered when the default is not among them', () => {
		expect(chooseLanguage(undefined, null, ['de', 'ru'], 'fr')).toBe('de');
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

describe('messageOf', () => {
	/** Russian, as far as these tests need it. */
	const t = translator(() => ({
		'error.invalid_credentials': 'Неверный адрес почты или пароль.',
		'error.validation.required': 'Поле «{field}» обязательно.',
		'error.rate_limited': 'Слишком много попыток. Попробуйте снова через {seconds} с.',
		'field.email': 'Электронная почта'
	}));

	it("says a problem in the reader's language", () => {
		const err = new ApiError(401, 'Wrong email or password.', 'invalid_credentials');
		expect(messageOf(err, t)).toBe('Неверный адрес почты или пароль.');
	});

	it('fills in what the server sent', () => {
		const err = new ApiError(429, 'Too many attempts.', 'rate_limited', { seconds: 12 });
		expect(messageOf(err, t)).toBe('Слишком много попыток. Попробуйте снова через 12 с.');
	});

	it('names a field the way the form labels it', () => {
		const err = new ApiError(400, 'email is required.', 'validation.required', { field: 'email' });
		expect(messageOf(err, t)).toBe('Поле «Электронная почта» обязательно.');
	});

	it('falls back to the English the page was built with for a sentence the language lacks', () => {
		const err = new ApiError(409, 'An account with this email already exists.', 'email_taken');
		expect(messageOf(err, t)).toBe('An account with this email already exists.');
	});

	it("falls back to the server's English for a code no catalog has", () => {
		const err = new ApiError(400, 'Something only the server knows.', 'brand_new_problem');
		expect(messageOf(err, t)).toBe('Something only the server knows.');
	});

	it('has a sentence for a server that could not be reached', () => {
		const err = new ApiError(0, 'Could not reach the server.', 'network');
		expect(
			messageOf(
				err,
				translator(() => ({}))
			)
		).toBe('Could not reach the server. Check your connection and try again.');
	});

	it('says something for what is not an answer at all', () => {
		expect(messageOf(new Error('boom'), t)).toBe('Something went wrong. Try again.');
	});
});

describe('translator.has', () => {
	const t = translator(() => ({ 'login.title': 'Вход' }));

	it('knows a key the language has, and one the base language has', () => {
		expect(t.has('login.title')).toBe(true);
		expect(t.has('action.sign_in')).toBe(true);
	});

	it('does not know a key nobody has', () => {
		expect(t.has('nothing.here')).toBe(false);
	});
});
