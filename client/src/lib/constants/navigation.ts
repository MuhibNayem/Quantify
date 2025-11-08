import type { ComponentType } from 'svelte';
import {
	LayoutDashboard,
	Boxes,
	Workflow,
	ActivitySquare,
	BellRing,
	UploadCloud,
	Users2,
} from 'lucide-svelte';

export type NavItem = {
	label: string;
	description?: string;
	href: string;
	icon: ComponentType;
};

export type NavSection = {
	title: string;
	items: NavItem[];
};

export const navSections: NavSection[] = [
	{
		title: 'Core',
		items: [
			{
				label: 'Overview',
				description: 'Live metrics and health',
				href: '/',
				icon: LayoutDashboard,
			},
		],
	},
	{
		title: 'Workspace',
		items: [
			{
				label: 'Catalog',
				description: 'Products, categories, partners',
				href: '/catalog',
				icon: Boxes,
			},
			{
				label: 'Operations',
				description: 'Stock moves, barcode, logistics',
				href: '/operations',
				icon: Workflow,
			},
			{
				label: 'Intelligence',
				description: 'Forecasts & business reports',
				href: '/intelligence',
				icon: ActivitySquare,
			},
		],
	},
	{
		title: 'Control',
		items: [
			{
				label: 'Alerts & Notifications',
				description: 'Thresholds, incidents, escalations',
				href: '/alerts',
				icon: BellRing,
			},
			{
				label: 'Bulk Ops',
				description: 'Imports, exports, automation',
				href: '/bulk',
				icon: UploadCloud,
			},
			{
				label: 'User Access',
				description: 'Approvals & roles',
				href: '/users',
				icon: Users2,
			},
		],
	},
];
