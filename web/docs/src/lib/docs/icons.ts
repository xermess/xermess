/**
 * Icons a page may name in front matter (`icon: key-2`), from Remix Icon. Unknown names fall
 * back to a plain page icon.
 */
import type { ComponentType } from 'svelte';
import {
	RiBookOpenLine,
	RiBracesLine,
	RiErrorWarningLine,
	RiFileList3Line,
	RiFlashlightLine,
	RiGlobalLine,
	RiHome5Line,
	RiKey2Line,
	RiLoginBoxLine,
	RiLogoutBoxRLine,
	RiRefreshLine,
	RiRobot2Line,
	RiServerLine,
	RiShieldCheckLine,
	RiShieldKeyholeLine,
	RiTerminalBoxLine,
	RiUserSettingsLine
} from 'svelte-remixicon';

const ICONS: Record<string, ComponentType> = {
	'book-open': RiBookOpenLine,
	braces: RiBracesLine,
	'error-warning': RiErrorWarningLine,
	'file-list-3': RiFileList3Line,
	flashlight: RiFlashlightLine,
	global: RiGlobalLine,
	'home-5': RiHome5Line,
	'key-2': RiKey2Line,
	'login-box': RiLoginBoxLine,
	'logout-box-r': RiLogoutBoxRLine,
	refresh: RiRefreshLine,
	'robot-2': RiRobot2Line,
	server: RiServerLine,
	'shield-check': RiShieldCheckLine,
	'shield-keyhole': RiShieldKeyholeLine,
	'terminal-box': RiTerminalBoxLine,
	'user-settings': RiUserSettingsLine
};

export function iconFor(name: string): ComponentType {
	return ICONS[name] ?? RiFileList3Line;
}
