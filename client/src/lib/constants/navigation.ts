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
	Undo2,
	ShoppingBag,
	Tag
} from 'lucide-svelte';

export type NavItem = {
	label: string;
	description?: string;
	href: string;
	icon: ComponentType;
	permission?: string;
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
			{
				label: 'Orders',
				description: 'Sales & Restock',
				href: '/orders',
				icon: ShoppingBag,
				permission: 'pos.view',
			},
		],
	},
	{
		title: 'Business',
		items: [
			{
				label: 'Reports',
				description: 'Advanced analytics & insights',
				href: '/reports',
				icon: ActivitySquare,
				permission: 'reports.view',
			},
			{
				label: 'CRM',
				description: 'Customers & loyalty',
				href: '/crm',
				icon: Users,
				permission: 'crm.view',
			},
			{
				label: 'Promotions',
				description: 'Discounts & Deals',
				href: '/promotions',
				icon: Tag,
				permission: 'products.write',
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
				permission: 'bulk.import',
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
