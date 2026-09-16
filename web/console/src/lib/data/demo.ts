/**
 * Placeholder data for the sections that have no backend yet.
 *
 * Nothing here is real: it exists so the sidebar has somewhere to lead and so
 * the tables can be designed against realistic shapes. Delete a block as soon
 * as its section talks to the API.
 */

type Status = 'active' | 'disabled' | 'draft';

export type DemoConnection = {
	id: string;
	name: string;
	engine: string;
	users: number;
	applications: number;
	status: Status;
};

export const demoConnections: DemoConnection[] = [
	{
		id: 'db1',
		name: 'Username & password',
		engine: 'xermess',
		users: 27583,
		applications: 4,
		status: 'active'
	},
	{
		id: 'db2',
		name: 'Legacy accounts',
		engine: 'External MySQL',
		users: 4120,
		applications: 1,
		status: 'active'
	},
	{
		id: 'db3',
		name: 'Staff directory',
		engine: 'LDAP',
		users: 212,
		applications: 1,
		status: 'disabled'
	}
];

export type DemoProvider = {
	id: string;
	name: string;
	clientId: string;
	logins: number;
	status: Status;
};

export const demoProviders: DemoProvider[] = [
	{
		id: 'soc1',
		name: 'Google',
		clientId: '8417…apps.googleusercontent.com',
		logins: 12904,
		status: 'active'
	},
	{ id: 'soc2', name: 'GitHub', clientId: 'Iv1.4b2c…', logins: 3311, status: 'active' },
	{ id: 'soc3', name: 'Apple', clientId: 'dev.xermess.signin', logins: 1877, status: 'active' },
	{ id: 'soc4', name: 'Microsoft', clientId: 'f0c1…', logins: 0, status: 'draft' }
];

export type DemoFlow = {
	id: string;
	name: string;
	steps: string;
	applications: number;
	isDefault: boolean;
	status: Status;
};

export const demoFlows: DemoFlow[] = [
	{
		id: 'flow1',
		name: 'Standard login',
		steps: 'Identifier → Password → MFA',
		applications: 3,
		isDefault: true,
		status: 'active'
	},
	{
		id: 'flow2',
		name: 'Passwordless',
		steps: 'Identifier → Email code',
		applications: 1,
		isDefault: false,
		status: 'active'
	},
	{
		id: 'flow3',
		name: 'Staff login',
		steps: 'SSO → MFA',
		applications: 1,
		isDefault: false,
		status: 'active'
	},
	{
		id: 'flow4',
		name: 'Trial signup',
		steps: 'Identifier → Password',
		applications: 0,
		isDefault: false,
		status: 'draft'
	}
];

export type DemoLanguage = {
	id: string;
	name: string;
	code: string;
	translated: number;
	isDefault: boolean;
	status: Status;
};

export const demoLanguages: DemoLanguage[] = [
	{ id: 'l1', name: 'English', code: 'en', translated: 100, isDefault: true, status: 'active' },
	{ id: 'l2', name: 'Kyrgyz', code: 'ky', translated: 92, isDefault: false, status: 'active' },
	{ id: 'l3', name: 'Russian', code: 'ru', translated: 88, isDefault: false, status: 'active' },
	{ id: 'l4', name: 'Turkish', code: 'tr', translated: 41, isDefault: false, status: 'draft' }
];
