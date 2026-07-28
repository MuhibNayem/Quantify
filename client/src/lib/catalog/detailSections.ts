import { get } from 'svelte/store';
import {
	BadgeDollarSign,
	CalendarClock,
	ClipboardList,
	Layers,
	MapPin,
	Mail,
	Package,
	Phone,
	Tag,
	Users
} from 'lucide-svelte';
import type {
	DetailBuilderContext,
	DetailExtraFetcher,
	DetailSection
} from '$lib/components/DetailsModal.svelte';
import { productsApi, subCategoriesApi, suppliersApi } from '$lib/api/resources';
import type {
	Category,
	Location,
	Product,
	StockAdjustment,
	SubCategory,
	Supplier,
	SupplierPerformance
} from '$lib/types';
import { formatCurrency, formatDateTime, formatPercent } from '$lib/utils';
import { t } from '$lib/i18n';

export type CatalogTab = 'products' | 'categories' | 'sub-categories' | 'suppliers' | 'locations';

type StockSnapshot = Awaited<ReturnType<typeof productsApi.stock>>;

export const statusBadge = (status?: string) => {
	if (!status) return undefined;
	const normalized = status.toLowerCase();
	if (normalized === 'active') return { text: status, variant: 'success' as const };
	if (normalized === 'inactive') return { text: status, variant: 'warning' as const };
	if (normalized === 'archived') return { text: status, variant: 'danger' as const };
	return { text: status, variant: 'info' as const };
};

const productDetailExtraFetchers: DetailExtraFetcher[] = [
	{
		key: 'stockSnapshot',
		request: async (resourceId: string | number) => productsApi.stock(Number(resourceId))
	},
	{
		key: 'stockHistory',
		request: async (resourceId: string | number) => productsApi.stockHistory(Number(resourceId))
	}
];

const categoryDetailExtraFetchers: DetailExtraFetcher[] = [
	{
		key: 'subCategories',
		request: async (resourceId: string | number) => subCategoriesApi.list(Number(resourceId))
	}
];

const supplierDetailExtraFetchers: DetailExtraFetcher[] = [
	{
		key: 'performance',
		request: async (resourceId: string | number) => suppliersApi.performance(Number(resourceId))
	}
];

const buildProductSections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const tt = get(t);
	const product = data as unknown as Product;
	const stockSnapshot = (extras.stockSnapshot as StockSnapshot | null) ?? null;
	const stockHistory = (extras.stockHistory as StockAdjustment[] | null) ?? [];
	const batches = stockSnapshot?.batches ?? [];
	const recentHistory = stockHistory.slice(0, 10);
	const lastAdjustment = recentHistory[0];

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'Current Stock',
					value: stockSnapshot?.currentQuantity ?? '—',
					hint: batches.length
						? `${batches.length} active batch${batches.length === 1 ? '' : 'es'}`
						: 'No active batches',
					icon: Package,
					accent: 'sky'
				},
				{
					title: tt('catalog.details.status'),
					value: product.Status ?? 'Unknown',
					hint: lastAdjustment
						? `${tt('catalog.details.updated')} ${formatDateTime(lastAdjustment.AdjustedAt)}`
						: 'No adjustments yet',
					icon: Tag,
					accent: 'emerald'
				},
				{
					title: tt('catalog.details.selling_price'),
					value: formatCurrency(product.SellingPrice ?? 0),
					hint: `Purchase ${formatCurrency(product.PurchasePrice ?? 0)}`,
					icon: BadgeDollarSign,
					accent: 'amber'
				}
			]
		},
		{
			type: 'description',
			title: 'Catalog Profile',
			description: 'Key identifiers & pricing context.',
			items: [
				{ label: tt('catalog.details.sku'), value: product.SKU },
				{ label: tt('catalog.details.name'), value: product.Name },
				{
					label: tt('catalog.details.status'),
					value: product.Status ?? 'Unknown',
					badge: statusBadge(product.Status)
				},
				{
					label: tt('catalog.details.purchase_price'),
					value: formatCurrency(product.PurchasePrice ?? 0)
				},
				{
					label: tt('catalog.details.selling_price'),
					value: formatCurrency(product.SellingPrice ?? 0)
				},
				{ label: 'Barcode', value: product.BarcodeUPC ?? '—' }
			]
		},
		{
			type: 'description',
			title: 'Associations',
			description: 'Upstream supplier & placement details.',
			items: [
				{
					label: tt('catalog.products.form.select_category'),
					value: product.Category?.Name ?? (product.CategoryID ? `#${product.CategoryID}` : '—')
				},
				{
					label: tt('catalog.products.form.select_sub_category'),
					value:
						product.SubCategory?.Name ?? (product.SubCategoryID ? `#${product.SubCategoryID}` : '—')
				},
				{
					label: tt('catalog.products.form.select_supplier'),
					value: product.Supplier?.Name ?? (product.SupplierID ? `#${product.SupplierID}` : '—')
				},
				{
					label: 'Location',
					value: product.Location?.Name ?? (product.LocationID ? `#${product.LocationID}` : '—')
				},
				{ label: 'Brand', value: product.Brand ?? '—' }
			]
		},
		{
			type: 'table',
			title: 'Recent Stock Adjustments',
			description: 'Last 10 adjustments pulled from the audit log.',
			columns: [
				{
					key: 'AdjustedAt',
					label: 'Date',
					formatter: (value) => formatDateTime(value as string)
				},
				{ key: 'Type', label: 'Type' },
				{ key: 'Quantity', label: 'Qty', align: 'right' },
				{ key: 'ReasonCode', label: 'Reason' },
				{ key: 'AdjustedBy', label: 'By', align: 'right' }
			],
			rows: recentHistory as unknown as Record<string, unknown>[],
			emptyText: 'No adjustments recorded for this product yet.'
		}
	];
};

const buildCategorySections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const tt = get(t);
	const category = data as unknown as Category;
	const children = (extras.subCategories as SubCategory[] | null) ?? [];

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'Sub-categories',
					value: children.length,
					hint: children.length ? 'Active descendants' : 'No children yet',
					icon: Layers,
					accent: 'violet'
				},
				{
					title: 'Category ID',
					value: category.ID,
					hint: 'Primary identifier',
					icon: ClipboardList,
					accent: 'sky'
				},
				{
					title: 'Created',
					value: formatDateTime(category.CreatedAt),
					hint: category.UpdatedAt
						? `Updated ${formatDateTime(category.UpdatedAt)}`
						: 'No updates yet',
					icon: CalendarClock,
					accent: 'slate'
				}
			]
		},
		{
			type: 'description',
			title: 'Category Profile',
			items: [
				{ label: tt('catalog.details.name'), value: category.Name },
				{ label: tt('catalog.details.id'), value: category.ID },
				{ label: tt('catalog.details.created'), value: formatDateTime(category.CreatedAt) },
				{ label: tt('catalog.details.updated'), value: formatDateTime(category.UpdatedAt) },
				{ label: 'Sub-categories', value: children.length ? `${children.length} linked` : '—' }
			]
		},
		{
			type: 'table',
			title: 'Sub-categories',
			description: 'Direct children linked to this category.',
			columns: [
				{ key: 'Name', label: 'Name' },
				{ key: 'ID', label: 'ID', align: 'right' }
			],
			rows: children as unknown as Record<string, unknown>[],
			emptyText: 'No sub-categories associated yet.'
		}
	];
};

const buildSubCategorySections = ({ data }: DetailBuilderContext): DetailSection[] => {
	const tt = get(t);
	const subCategory = data as unknown as SubCategory;
	const parent = subCategory.Category;

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'Parent Category',
					value: parent?.Name ?? `#${subCategory.CategoryID}`,
					hint: parent ? `ID ${parent.ID}` : 'Linked parent',
					icon: ClipboardList,
					accent: 'violet'
				},
				{
					title: 'Sub-category ID',
					value: subCategory.ID,
					hint: 'Primary identifier',
					icon: Tag,
					accent: 'sky'
				},
				{
					title: 'Created',
					value: formatDateTime(subCategory.CreatedAt),
					hint: subCategory.UpdatedAt
						? `Updated ${formatDateTime(subCategory.UpdatedAt)}`
						: 'No updates yet',
					icon: CalendarClock,
					accent: 'slate'
				}
			]
		},
		{
			type: 'description',
			title: 'Sub-category Profile',
			items: [
				{ label: tt('catalog.details.name'), value: subCategory.Name },
				{
					label: tt('catalog.details.parent_category'),
					value: parent?.Name ?? `Category #${subCategory.CategoryID}`
				},
				{ label: tt('catalog.details.created'), value: formatDateTime(subCategory.CreatedAt) },
				{ label: tt('catalog.details.updated'), value: formatDateTime(subCategory.UpdatedAt) }
			]
		}
	];
};

const buildSupplierSections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const tt = get(t);
	const supplier = data as unknown as Supplier;
	const performance = (extras.performance as SupplierPerformance | null) ?? null;

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'On-time rate',
					value: formatPercent(performance?.onTimeDeliveryRate),
					hint: 'Past reporting window',
					icon: Users,
					accent: 'emerald'
				},
				{
					title: 'Avg. lead time',
					value: performance?.averageLeadTimeDays ? `${performance.averageLeadTimeDays} days` : '—',
					hint: 'Receipt to PO',
					icon: CalendarClock,
					accent: 'amber'
				},
				{
					title: 'Supplier ID',
					value: supplier.ID,
					hint: 'Primary identifier',
					icon: ClipboardList,
					accent: 'slate'
				}
			]
		},
		{
			type: 'description',
			title: tt('catalog.details.contact_details'),
			items: [
				{ label: tt('catalog.details.name'), value: supplier.Name },
				{ label: tt('catalog.details.contact_person'), value: supplier.ContactPerson ?? '—' },
				{ label: tt('catalog.details.email'), value: supplier.Email ?? '—', icon: Mail },
				{ label: tt('catalog.details.phone'), value: supplier.Phone ?? '—', icon: Phone },
				{ label: tt('catalog.details.address'), value: supplier.Address ?? '—', icon: MapPin }
			]
		},
		{
			type: 'table',
			title: 'Performance Snapshot',
			columns: [
				{ key: 'metric', label: 'Metric' },
				{ key: 'value', label: 'Value' }
			],
			rows: [
				{ metric: 'On-time delivery', value: formatPercent(performance?.onTimeDeliveryRate) },
				{
					metric: 'Average lead time',
					value: performance?.averageLeadTimeDays ? `${performance.averageLeadTimeDays} days` : '—'
				}
			],
			emptyText: 'Performance metrics unavailable.'
		}
	];
};

const buildLocationSections = ({ data }: DetailBuilderContext): DetailSection[] => {
	const tt = get(t);
	const location = data as unknown as Location;

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'Location ID',
					value: location.ID,
					hint: 'Primary identifier',
					icon: MapPin,
					accent: 'sky'
				},
				{
					title: 'Created',
					value: formatDateTime(location.CreatedAt),
					hint: location.UpdatedAt
						? `Updated ${formatDateTime(location.UpdatedAt)}`
						: 'No updates yet',
					icon: CalendarClock,
					accent: 'slate'
				}
			]
		},
		{
			type: 'description',
			title: tt('catalog.details.location_profile'),
			items: [
				{ label: tt('catalog.details.name'), value: location.Name },
				{ label: tt('catalog.details.address'), value: location.Address ?? '—' },
				{ label: tt('catalog.details.created'), value: formatDateTime(location.CreatedAt) },
				{ label: tt('catalog.details.updated'), value: formatDateTime(location.UpdatedAt) }
			]
		}
	];
};

export type DetailConfig = {
	endpoint: string;
	title: string;
	subtitle: string | null;
	extraFetchers: DetailExtraFetcher[];
	buildSections: (ctx: DetailBuilderContext) => DetailSection[];
};

const emptySectionBuilder: (ctx: DetailBuilderContext) => DetailSection[] = () => [];

export function getDetailConfig(type: CatalogTab, resource: Record<string, any>): DetailConfig {
	const baseTitle = `${type.charAt(0).toUpperCase() + type.slice(1).replace('-', ' ')} Details`;

	switch (type) {
		case 'products':
			return {
				endpoint: '/products',
				title: baseTitle,
				subtitle: resource.SKU ? `SKU ${resource.SKU}` : (resource.Name ?? null),
				extraFetchers: productDetailExtraFetchers,
				buildSections: buildProductSections
			};
		case 'categories':
			return {
				endpoint: '/categories',
				title: baseTitle,
				subtitle: resource.Name ?? null,
				extraFetchers: categoryDetailExtraFetchers,
				buildSections: buildCategorySections
			};
		case 'sub-categories':
			return {
				endpoint: '/sub-categories',
				title: baseTitle,
				subtitle: resource.Name ?? null,
				extraFetchers: [],
				buildSections: buildSubCategorySections
			};
		case 'suppliers':
			return {
				endpoint: '/suppliers',
				title: baseTitle,
				subtitle: resource.Name ?? null,
				extraFetchers: supplierDetailExtraFetchers,
				buildSections: buildSupplierSections
			};
		case 'locations':
			return {
				endpoint: '/locations',
				title: baseTitle,
				subtitle: resource.Name ?? null,
				extraFetchers: [],
				buildSections: buildLocationSections
			};
		default:
			return {
				endpoint: `/${type}`,
				title: baseTitle,
				subtitle: resource?.Name ?? resource?.SKU ?? null,
				extraFetchers: [],
				buildSections: emptySectionBuilder
			};
	}
}
