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
	Settings,
	Undo2
} from 'lucide-svelte';

export type NavItem = {
	label: string;
	description?: string;
	href: string;
	icon: ComponentType;
	permission?: string; // Optional: if present, requires this permission
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
				// No permission needed (Public/Auth)
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
				permission: 'products.read',
			},
			{
				label: 'Operations',
				description: 'Stock moves, barcode, logistics',
				href: '/operations',
				icon: Workflow,
				permission: 'inventory.view',
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
				permission: 'reports.sales',
			},
			{
				label: 'POS',
				description: 'Checkout & payments',
				href: '/pos',
				icon: Sparkles,
				permission: 'pos.view',
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
				permission: 'crm.view',
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
				permission: 'alerts.view',
			},
			{
				label: 'Bulk Ops',
				description: 'Imports, exports, automation',
				href: '/bulk',
				icon: UploadCloud,
				permission: 'bulk.import', // Or bulk.export, checking import as primary
			},
			{
				label: 'User Access',
				description: 'Approvals & roles',
				href: '/users',
				icon: Users2,
				permission: 'users.view',
			},
			{
				label: 'Settings',
				description: 'Configuration & RBAC',
				href: '/settings',
				icon: Settings,
				permission: 'settings.view',
			},
		],
	},
];
