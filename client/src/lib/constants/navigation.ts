import type { ComponentType } from 'svelte';
import {
	LayoutDashboard,
	Boxes,
	Workflow,
	ActivitySquare,
	BellRing,
	UploadCloud,
	Users2,
	Sparkles,
	Clock,
	Users,
	Settings
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
				label: 'Time Tracking',
				description: 'Clock in/out and manage shifts',
				href: '/time-tracking',
				icon: Clock,
			},
			{
				label: 'Intelligence',
				description: 'Forecasts & business reports',
				href: '/intelligence',
				icon: ActivitySquare,
			},
			{
				label: 'POS',
				description: 'Checkout & payments',
				href: '/pos',
				icon: Sparkles,
			},
		],
	},
	{
		title: 'Business',
		items: [
			{
				label: 'CRM',
				description: 'Customers & loyalty',
				href: '/crm',
				icon: Users,
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
			{
				label: 'Settings',
				description: 'Configuration & RBAC',
				href: '/settings',
				icon: Settings,
			},
		],
	},
];
