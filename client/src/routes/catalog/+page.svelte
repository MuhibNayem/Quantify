<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Root, Content, Item, PrevButton, NextButton, Ellipsis, Link } from '$lib/components/ui/pagination';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
import DetailsModal from '$lib/components/DetailsModal.svelte';
import type { DetailBuilderContext, DetailExtraFetcher, DetailSection } from '$lib/components/DetailsModal.svelte';
import { productsApi, categoriesApi, subCategoriesApi, suppliersApi, locationsApi } from '$lib/api/resources';
import type { Category, Location, Product, StockAdjustment, SubCategory, Supplier, SupplierPerformance } from '$lib/types';
import {
	BadgeDollarSign,
	CalendarClock,
	ClipboardList,
	Layers,
	MapPin,
	Mail,
	Package,
	Phone,
	PlusCircle,
	RefreshCcw,
	Tag,
	Users,
	Zap,
} from 'lucide-svelte';

	type TabKey = 'products' | 'categories' | 'sub-categories' | 'suppliers' | 'locations';

	let activeTab = $state<TabKey>('products');
	let loading = $state(false);
	let search = $state({ term: '', key: 'name' });

	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let subCategories = $state<SubCategory[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let locations = $state<Location[]>([]);

	const productForm = $state({
		sku: '',
		name: '',
		description: '',
		categoryId: '',
		subCategoryId: '',
		supplierId: '',
		locationId: '',
		purchasePrice: '',
		sellingPrice: '',
		status: 'Active',
		barCodeUPC: '',
	});

	const pagination = $state({
		currentPage: 1,
		totalPages: 1,
		totalItems: 0,
		itemsPerPage: 10,
	});

	let editingProduct: Product | null = null;

	const categoryForm = $state({ name: '' });
	let editingCategory: Category | null = null;
	let categorySearchTerm = $state('');

	const subCategoryForm = $state({ name: '', categoryId: '' });
	let editingSubCategory: SubCategory | null = null;

	const supplierForm = $state({ name: '', contactPerson: '', email: '', phone: '', address: '' });
	let editingSupplier: Supplier | null = null;
	let supplierSearchTerm = $state('');

	const locationForm = $state({ name: '', address: '' });
	let editingLocation: Location | null = null;

let selectedResourceId: number | null = $state(null);
let isModalOpen = $state(false);
let modalEndpoint = $state('');
let modalTitle = $state('');
let modalSubtitle: string | null = $state(null);
let modalExtraFetchers = $state<DetailExtraFetcher[]>([]);
const emptySectionBuilder: (ctx: DetailBuilderContext) => DetailSection[] = () => [];
let modalSectionsBuilder = $state<(ctx: DetailBuilderContext) => DetailSection[]>(emptySectionBuilder);
let useLegacyModalSlot = $state(true);

type StockSnapshot = Awaited<ReturnType<typeof productsApi.stock>>;

const currencyFormatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });
const dateTimeFormatter = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' });
const percentFormatter = new Intl.NumberFormat('en-US', { style: 'percent', maximumFractionDigits: 1 });

const formatCurrency = (value?: number | null) => {
	if (value === null || value === undefined || Number.isNaN(value)) return '—';
	return currencyFormatter.format(value);
};

const formatDateTime = (value?: string | null) => {
	if (!value) return '—';
	const date = new Date(value);
	return Number.isNaN(date.getTime()) ? '—' : dateTimeFormatter.format(date);
};

const formatPercent = (value?: number | null) => {
	if (value === null || value === undefined || Number.isNaN(value)) return '—';
	const normalized = value > 1 ? value / 100 : value;
	return percentFormatter.format(normalized);
};

const statusBadge = (status?: string) => {
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
		request: async (resourceId: string | number) => productsApi.stock(Number(resourceId)),
	},
	{
		key: 'stockHistory',
		request: async (resourceId: string | number) => productsApi.stockHistory(Number(resourceId)),
	},
];

const categoryDetailExtraFetchers: DetailExtraFetcher[] = [
	{
		key: 'subCategories',
		request: async (resourceId: string | number) => subCategoriesApi.list(Number(resourceId)),
	},
];

const supplierDetailExtraFetchers: DetailExtraFetcher[] = [
	{
		key: 'performance',
		request: async (resourceId: string | number) => suppliersApi.performance(Number(resourceId)),
	},
];

const buildProductSections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const product = data as Product;
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
					hint: batches.length ? `${batches.length} active batch${batches.length === 1 ? '' : 'es'}` : 'No active batches',
					icon: Package,
					accent: 'sky',
				},
				{
					title: 'Status',
					value: product.Status ?? 'Unknown',
					hint: lastAdjustment ? `Updated ${formatDateTime(lastAdjustment.AdjustedAt)}` : 'No adjustments yet',
					icon: Tag,
					accent: 'emerald',
				},
				{
					title: 'Selling Price',
					value: formatCurrency(product.SellingPrice),
					hint: `Purchase ${formatCurrency(product.PurchasePrice)}`,
					icon: BadgeDollarSign,
					accent: 'amber',
				},
			],
		},
		{
			type: 'description',
			title: 'Catalog Profile',
			description: 'Key identifiers & pricing context.',
			items: [
				{ label: 'SKU', value: product.SKU },
				{ label: 'Name', value: product.Name },
				{ label: 'Status', value: product.Status ?? 'Unknown', badge: statusBadge(product.Status) },
				{ label: 'Purchase Price', value: formatCurrency(product.PurchasePrice) },
				{ label: 'Selling Price', value: formatCurrency(product.SellingPrice) },
				{ label: 'Barcode', value: product.BarcodeUPC ?? '—' },
			],
		},
		{
			type: 'description',
			title: 'Associations',
			description: 'Upstream supplier & placement details.',
			items: [
				{ label: 'Category', value: product.Category?.Name ?? (product.CategoryID ? `#${product.CategoryID}` : '—') },
				{ label: 'Sub-Category', value: product.SubCategory?.Name ?? (product.SubCategoryID ? `#${product.SubCategoryID}` : '—') },
				{ label: 'Supplier', value: product.Supplier?.Name ?? (product.SupplierID ? `#${product.SupplierID}` : '—') },
				{ label: 'Location', value: product.Location?.Name ?? (product.LocationID ? `#${product.LocationID}` : '—') },
				{ label: 'Brand', value: product.Brand ?? '—' },
			],
		},
		{
			type: 'table',
			title: 'Recent Stock Adjustments',
			description: 'Last 10 adjustments pulled from the audit log.',
			columns: [
				{ key: 'AdjustedAt', label: 'Date', formatter: (value) => formatDateTime(value as string) },
				{ key: 'Type', label: 'Type' },
				{ key: 'Quantity', label: 'Qty', align: 'right' },
				{ key: 'ReasonCode', label: 'Reason' },
				{ key: 'AdjustedBy', label: 'By', align: 'right' },
			],
			rows: recentHistory,
			emptyText: 'No adjustments recorded for this product yet.',
		},
	];
};

const buildCategorySections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const category = data as Category;
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
					accent: 'violet',
				},
				{
					title: 'Category ID',
					value: category.ID,
					hint: 'Primary identifier',
					icon: ClipboardList,
					accent: 'sky',
				},
				{
					title: 'Created',
					value: formatDateTime(category.CreatedAt),
					hint: category.UpdatedAt ? `Updated ${formatDateTime(category.UpdatedAt)}` : 'No updates yet',
					icon: CalendarClock,
					accent: 'slate',
				},
			],
		},
		{
			type: 'description',
			title: 'Category Profile',
			items: [
				{ label: 'Name', value: category.Name },
				{ label: 'ID', value: category.ID },
				{ label: 'Created', value: formatDateTime(category.CreatedAt) },
				{ label: 'Updated', value: formatDateTime(category.UpdatedAt) },
				{ label: 'Sub-categories', value: children.length ? `${children.length} linked` : '—' },
			],
		},
		{
			type: 'table',
			title: 'Sub-categories',
			description: 'Direct children linked to this category.',
			columns: [
				{ key: 'Name', label: 'Name' },
				{ key: 'ID', label: 'ID', align: 'right' },
			],
			rows: children,
			emptyText: 'No sub-categories associated yet.',
		},
	];
};

const buildSubCategorySections = ({ data }: DetailBuilderContext): DetailSection[] => {
	const subCategory = data as SubCategory;
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
					accent: 'violet',
				},
				{
					title: 'Sub-category ID',
					value: subCategory.ID,
					hint: 'Primary identifier',
					icon: Tag,
					accent: 'sky',
				},
				{
					title: 'Created',
					value: formatDateTime(subCategory.CreatedAt),
					hint: subCategory.UpdatedAt ? `Updated ${formatDateTime(subCategory.UpdatedAt)}` : 'No updates yet',
					icon: CalendarClock,
					accent: 'slate',
				},
			],
		},
		{
			type: 'description',
			title: 'Sub-category Profile',
			items: [
				{ label: 'Name', value: subCategory.Name },
				{ label: 'Parent', value: parent?.Name ?? `Category #${subCategory.CategoryID}` },
				{ label: 'Created', value: formatDateTime(subCategory.CreatedAt) },
				{ label: 'Updated', value: formatDateTime(subCategory.UpdatedAt) },
			],
		},
	];
};

const buildSupplierSections = ({ data, extras }: DetailBuilderContext): DetailSection[] => {
	const supplier = data as Supplier;
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
					accent: 'emerald',
				},
				{
					title: 'Avg. lead time',
					value: performance?.averageLeadTimeDays ? `${performance.averageLeadTimeDays} days` : '—',
					hint: 'Receipt to PO',
					icon: CalendarClock,
					accent: 'amber',
				},
				{
					title: 'Supplier ID',
					value: supplier.ID,
					hint: 'Primary identifier',
					icon: ClipboardList,
					accent: 'slate',
				},
			],
		},
		{
			type: 'description',
			title: 'Contact Details',
			items: [
				{ label: 'Name', value: supplier.Name },
				{ label: 'Contact Person', value: supplier.ContactPerson ?? '—' },
				{ label: 'Email', value: supplier.Email ?? '—', icon: Mail },
				{ label: 'Phone', value: supplier.Phone ?? '—', icon: Phone },
				{ label: 'Address', value: supplier.Address ?? '—', icon: MapPin },
			],
		},
		{
			type: 'table',
			title: 'Performance Snapshot',
			columns: [
				{ key: 'metric', label: 'Metric' },
				{ key: 'value', label: 'Value' },
			],
			rows: [
				{ metric: 'On-time delivery', value: formatPercent(performance?.onTimeDeliveryRate) },
				{
					metric: 'Average lead time',
					value: performance?.averageLeadTimeDays ? `${performance.averageLeadTimeDays} days` : '—',
				},
			],
			emptyText: 'Performance metrics unavailable.',
		},
	];
};

const buildLocationSections = ({ data }: DetailBuilderContext): DetailSection[] => {
	const location = data as Location;

	return [
		{
			type: 'summary',
			cards: [
				{
					title: 'Location ID',
					value: location.ID,
					hint: 'Primary identifier',
					icon: MapPin,
					accent: 'sky',
				},
				{
					title: 'Created',
					value: formatDateTime(location.CreatedAt),
					hint: location.UpdatedAt ? `Updated ${formatDateTime(location.UpdatedAt)}` : 'No updates yet',
					icon: CalendarClock,
					accent: 'slate',
				},
			],
		},
		{
			type: 'description',
			title: 'Location Profile',
			items: [
				{ label: 'Name', value: location.Name },
				{ label: 'Address', value: location.Address ?? '—' },
				{ label: 'Created', value: formatDateTime(location.CreatedAt) },
				{ label: 'Updated', value: formatDateTime(location.UpdatedAt) },
			],
		},
	];
};

const viewDetails = (resource: any, type: TabKey) => {
	selectedResourceId = resource.ID;
	modalEndpoint = `/${type}`;
	modalTitle = `${type.charAt(0).toUpperCase() + type.slice(1).replace('-', ' ')} Details`;
	modalSubtitle = resource?.Name ?? resource?.SKU ?? null;
	modalExtraFetchers = [];
	modalSectionsBuilder = emptySectionBuilder;
	useLegacyModalSlot = true;

	switch (type) {
		case 'products':
			modalSubtitle = resource.SKU ? `SKU ${resource.SKU}` : resource.Name ?? null;
			modalExtraFetchers = productDetailExtraFetchers;
			modalSectionsBuilder = buildProductSections;
			useLegacyModalSlot = false;
			break;
		case 'categories':
			modalExtraFetchers = categoryDetailExtraFetchers;
			modalSectionsBuilder = buildCategorySections;
			useLegacyModalSlot = false;
			break;
		case 'sub-categories':
			modalSectionsBuilder = buildSubCategorySections;
			useLegacyModalSlot = false;
			break;
		case 'suppliers':
			modalExtraFetchers = supplierDetailExtraFetchers;
			modalSectionsBuilder = buildSupplierSections;
			useLegacyModalSlot = false;
			break;
		case 'locations':
			modalSectionsBuilder = buildLocationSections;
			useLegacyModalSlot = false;
			break;
		default:
			useLegacyModalSlot = true;
	}

	isModalOpen = true;
};

	const loadAll = async (page = 1) => {
		loading = true;
		try {
			const categoryList = await categoriesApi.list();
			categories = Array.isArray(categoryList) ? categoryList : [categoryList];

			const supplierList = await suppliersApi.list();
			suppliers = Array.isArray(supplierList) ? supplierList : [supplierList];

			const locationList = await locationsApi.list();
			locations = Array.isArray(locationList) ? locationList : [locationList];

			const productResponse = await productsApi.list(page, 10);
			products = productResponse.products || [];

			pagination.currentPage = productResponse.currentPage;
			pagination.totalPages = productResponse.totalPages;
			pagination.totalItems = productResponse.totalItems;
			pagination.itemsPerPage = productResponse.itemsPerPage;
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load catalog';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const loadProductPerPage = async (page = 1) => {
		try {
			const productResponse = await productsApi.list(page, 10);
			products = productResponse.products || [];

			pagination.currentPage = productResponse.currentPage;
			pagination.totalPages = productResponse.totalPages;
			pagination.totalItems = productResponse.totalItems;
			pagination.itemsPerPage = productResponse.itemsPerPage;
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to load catalog';
			toast.error('Failed to Load Catalog', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const handlePageChange = (page: number) => {
		if (page !== pagination.currentPage) {
			pagination.currentPage = page;
			loadProductPerPage(page);
		}
	};

	const handleSearch = async () => {
		if (!search.term.trim()) return;
		loading = true;
		try {
			if (search.key === 'name') {
				const productResponse = await productsApi.list(1, 100, search.term.trim());
				products = productResponse.products || [];
				pagination.currentPage = productResponse.currentPage;
				pagination.totalPages = productResponse.totalPages;
				pagination.totalItems = productResponse.totalItems;
				pagination.itemsPerPage = productResponse.itemsPerPage;
			} else if (search.key === 'sku') {
				const product = await productsApi.getBySku(search.term.trim());
				if (product) {
					viewDetails(product, 'products');
				} else {
					toast.error('Product not found');
				}
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const clearSearch = () => {
		search.term = '';
		loadAll();
	};

	const handleCategorySearch = async () => {
		if (!categorySearchTerm.trim()) return;
		loading = true;
		try {
			const category = await categoriesApi.getByName(categorySearchTerm.trim());
			if (category) {
				viewDetails(category, 'categories');
			} else {
				toast.error('Category not found');
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const clearCategorySearch = () => {
		categorySearchTerm = '';
		loadAll();
	};

	const handleSupplierSearch = async () => {
		if (!supplierSearchTerm.trim()) return;
		loading = true;
		try {
			const supplier = await suppliersApi.getByName(supplierSearchTerm.trim());
			if (supplier) {
				viewDetails(supplier, 'suppliers');
			} else {
				toast.error('Supplier not found');
			}
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Search failed';
			toast.error('Search Failed', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const clearSupplierSearch = () => {
		supplierSearchTerm = '';
		loadAll();
	};

	onMount(() => {
	loadAll();

	let ticking = false;
	const hero = document.querySelector('.parallax-hero') as HTMLElement | null;

	const handleScroll = () => {
		if (!hero) return;
		if (!ticking) {
			window.requestAnimationFrame(() => {
				const scrollY = window.scrollY || 0;
				const translate = Math.min(scrollY * 0.25, 60); 
				const blur = Math.min(scrollY * 0.02, 6); 
				hero.style.transform = `translateY(${translate}px)`;
				hero.style.filter = `blur(${blur}px)`;
				ticking = false;
			});
			ticking = true;
		}
	};

	window.addEventListener('scroll', handleScroll, { passive: true });

	return () => {
		window.removeEventListener('scroll', handleScroll);
	};
});


	const loadSubCategories = async (categoryId: number) => {
		if (!categoryId) {
			subCategories = [];
			return;
		}
		loading = true;
		try {
			const subCategoryList = await subCategoriesApi.list(categoryId);
			subCategories = Array.isArray(subCategoryList) ? subCategoryList : [subCategoryList];
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to load sub-categories';
			toast.error('Failed to Load Sub-Categories', { description: errorMessage });
		} finally {
			loading = false;
		}
	};

	const resetProductForm = () => {
		editingProduct = null;
		productForm.sku = '';
		productForm.name = '';
		productForm.description = '';
		productForm.categoryId = '';
		productForm.subCategoryId = '';
		productForm.supplierId = '';
		productForm.locationId = '';
		productForm.purchasePrice = '';
		productForm.sellingPrice = '';
		productForm.status = 'Active';
		productForm.barCodeUPC = '';
	};

	const editProduct = (product: Product) => {
		editingProduct = product;
		productForm.sku = product.SKU;
		productForm.name = product.Name;
		productForm.description = product.Description ?? '';
		productForm.categoryId = String(product.CategoryID ?? '');
		productForm.subCategoryId = String(product.SubCategoryID ?? '');
		productForm.supplierId = String(product.SupplierID ?? '');
		productForm.locationId = String(product.LocationID ?? '');
		productForm.purchasePrice = product.PurchasePrice ? String(product.PurchasePrice) : '';
		productForm.sellingPrice = product.SellingPrice ? String(product.SellingPrice) : '';
		productForm.status = product.Status ?? 'Active';
	};

	const saveProduct = async () => {
		if (!productForm.barCodeUPC.trim()) {
			toast.error('Missing Barcode/UPC', {
				description: 'Each product must have a unique BarcodeUPC value.',
			});
			return;
		}

		const duplicate = products.find(
			(p) =>
				p.BarcodeUPC?.toLowerCase() === productForm.barCodeUPC.trim().toLowerCase() &&
				p.ID !== editingProduct?.ID
		);
		if (duplicate) {
			toast.error('Duplicate Barcode/UPC Detected', {
				description: `The BarcodeUPC "${productForm.barCodeUPC}" is already used by product "${duplicate.Name}".`,
			});
			return;
		}

		const payload = {
			sku: productForm.sku,
			name: productForm.name,
			description: productForm.description,
			categoryId: productForm.categoryId ? Number(productForm.categoryId) : undefined,
			subCategoryId: productForm.subCategoryId ? Number(productForm.subCategoryId) : undefined,
			supplierId: productForm.supplierId ? Number(productForm.supplierId) : undefined,
			locationId: productForm.locationId ? Number(productForm.locationId) : undefined,
			purchasePrice: productForm.purchasePrice ? Number(productForm.purchasePrice) : undefined,
			sellingPrice: productForm.sellingPrice ? Number(productForm.sellingPrice) : undefined,
			status: productForm.status,
			BarcodeUPC: productForm.barCodeUPC.trim(),
		};

		try {
			if (editingProduct) {
				await productsApi.update(editingProduct.ID, payload);
				toast.success('Product updated successfully');
			} else {
				await productsApi.create(payload);
				toast.success('Product created successfully');
			}
			await loadAll();
			resetProductForm();
		} catch (error: any) {
			const backendMessage =
				error?.response?.data?.error || error?.response?.data?.message || 'Unable to save product';
			toast.error('Failed to Save Product', { description: backendMessage });
		}
	};

	const deleteProduct = async (product: Product) => {
		toast.confirm(`Are you sure you want to delete ${product.Name}?`, {
			onConfirm: async () => {
				try {
					await productsApi.remove(product.ID);
					toast.success('Product removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete product';
					toast.error('Failed to Delete Product', { description: errorMessage });
				}
			},
		});
	};

	const resetCategoryForm = () => {
		editingCategory = null;
		categoryForm.name = '';
	};

	const saveCategory = async () => {
		try {
			if (editingCategory) {
				await categoriesApi.update(editingCategory.ID, { name: categoryForm.name });
			} else {
				await categoriesApi.create({ name: categoryForm.name });
			}
			toast.success('Category saved');
			await loadAll();
			resetCategoryForm();
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to save category';
			toast.error('Failed to Save Category', { description: errorMessage });
		}
	};

	const deleteCategory = async (category: Category) => {
		toast.confirm(`Are you sure you want to delete ${category.Name}?`, {
			onConfirm: async () => {
				try {
					await categoriesApi.remove(category.ID);
					toast.success('Category removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete category';
					toast.error('Failed to Delete Category', { description: errorMessage });
				}
			},
		});
	};

	const resetSubCategoryForm = () => {
		editingSubCategory = null;
		subCategoryForm.name = '';
		subCategoryForm.categoryId = '';
	};

	const saveSubCategory = async () => {
		try {
			const payload = { name: subCategoryForm.name };
			if (editingSubCategory) {
				await subCategoriesApi.update(editingSubCategory.ID, payload);
			} else {
				await subCategoriesApi.create(payload, Number(subCategoryForm.categoryId));
			}
			toast.success('Sub-category saved');
			await loadAll();
			resetSubCategoryForm();
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to save sub-category';
			toast.error('Failed to Save Sub-Category', { description: errorMessage });
		}
	};

	const deleteSubCategory = async (subCategory: SubCategory) => {
		toast.confirm(`Are you sure you want to delete ${subCategory.Name}?`, {
			onConfirm: async () => {
				try {
					await subCategoriesApi.remove(subCategory.ID);
					toast.success('Sub-category removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete sub-category';
					toast.error('Failed to Delete Sub-Category', { description: errorMessage });
				}
			},
		});
	};

	const resetSupplierForm = () => {
		editingSupplier = null;
		supplierForm.name = '';
		supplierForm.contactPerson = '';
		supplierForm.email = '';
		supplierForm.phone = '';
		supplierForm.address = '';
	};

	const saveSupplier = async () => {
		try {
			if (editingSupplier) {
				await suppliersApi.update(editingSupplier.ID, supplierForm);
			} else {
				await suppliersApi.create(supplierForm);
			}
			toast.success('Supplier saved');
			await loadAll();
			resetSupplierForm();
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to save supplier';
			toast.error('Failed to Save Supplier', { description: errorMessage });
		}
	};

	const deleteSupplier = async (supplier: Supplier) => {
		toast.confirm(`Are you sure you want to delete ${supplier.Name}?`, {
			onConfirm: async () => {
				try {
					await suppliersApi.remove(supplier.ID);
					toast.success('Supplier removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete supplier';
					toast.error('Failed to Delete Supplier', { description: errorMessage });
				}
			},
		});
	};

	const resetLocationForm = () => {
		editingLocation = null;
		locationForm.name = '';
		locationForm.address = '';
	};

	const saveLocation = async () => {
		try {
			if (editingLocation) {
				await locationsApi.update(editingLocation.ID, locationForm);
			} else {
				await locationsApi.create(locationForm);
			}
			toast.success('Location saved');
			await loadAll();
			resetLocationForm();
		} catch (error) {
			const errorMessage = (error as any)?.response?.data?.error || 'Unable to save location';
			toast.error('Failed to Save Location', { description: errorMessage });
		}
	};

	const deleteLocation = async (location: Location) => {
		toast.confirm(`Are you sure you want to delete ${location.Name}?`, {
			onConfirm: async () => {
				try {
					await locationsApi.remove(location.ID);
					toast.success('Location removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete location';
					toast.error('Failed to Delete Location', { description: errorMessage });
				}
			},
		});
	};
</script>

<DetailsModal
	bind:open={isModalOpen}
	resourceId={selectedResourceId}
	endpoint={modalEndpoint}
	title={modalTitle}
	subtitle={modalSubtitle}
	extraFetchers={modalExtraFetchers}
	buildSections={modalSectionsBuilder}
	let:data={resource}
>
	{#if useLegacyModalSlot && resource}
		<div class="grid gap-4">
			<div class="grid grid-cols-2 gap-2">
				<p><strong>ID:</strong> {resource.ID}</p>
				<p><strong>SKU:</strong> {resource.SKU}</p>
				<p><strong>Name:</strong> {resource.Name}</p>
				<p><strong>Status:</strong> {resource.Status}</p>
				<p><strong>Purchase Price:</strong> ${resource.PurchasePrice}</p>
				<p><strong>Selling Price:</strong> ${resource.SellingPrice}</p>
				<p><strong>Category ID:</strong> {resource.CategoryID}</p>
				<p><strong>Supplier ID:</strong> {resource.SupplierID}</p>
			</div>
			<p><strong>Description:</strong> {resource.Description}</p>
		</div>
	{/if}
</DetailsModal>

<!-- ===== FIXED HERO (responsive, clean parallax, correct layering) ===== -->
<section class="relative w-full isolate overflow-hidden">
	<!-- Gradient background with motion -->
	<div class="absolute inset-0 -z-10 animate-gradientShift bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 bg-[length:200%_200%]"></div>

	<!-- Floating glow blobs -->
	<div class="absolute -top-32 -left-24 w-96 h-96 rounded-full bg-sky-200/40 blur-3xl animate-pulseGlow"></div>
	<div class="absolute -bottom-28 -right-24 w-80 h-80 rounded-full bg-cyan-200/30 blur-3xl animate-pulseGlow delay-700"></div>

	<!-- Hero container -->
	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-16 sm:pt-20 pb-10 sm:pb-16 text-center sm:text-left"
		style="transform: translateY(var(--hero-translate, 0));"
	>
		<div class="inline-flex items-center gap-3 justify-center sm:justify-start mb-3">
			<span class="inline-flex p-2 rounded-2xl shadow-md bg-gradient-to-br from-sky-500 to-blue-600 animate-cardFloat">
				<Zap class="h-6 w-6 text-white" />
			</span>
			<p class="text-xs sm:text-sm uppercase tracking-[0.18em] text-sky-700 font-semibold">
				Catalog Cockpit
			</p>
		</div>

		<h1
			class="text-3xl sm:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-transparent mb-3"
		>
			Products, Categories &amp; Partners
		</h1>
		<p class="text-slate-600 text-sm sm:text-base max-w-2xl mx-auto sm:mx-0">
			Unified control center for your catalog data
		</p>

		<!-- Action buttons -->
		<div class="mt-6 flex flex-col sm:flex-row gap-3 justify-center sm:justify-start">
			<Button
				variant="secondary"
				onclick={loadAll}
				class="w-full sm:w-auto bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white rounded-xl px-5 py-2.5 font-medium shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105 focus:ring-2 focus:ring-sky-300"
			>
				<RefreshCcw class="h-4 w-4 mr-2" /> Sync data
			</Button>
			<Button
				href="/bulk"
				variant="outline"
				class="w-full sm:w-auto bg-white/80 border border-sky-200 text-sky-700 hover:bg-sky-50 rounded-xl px-5 py-2.5 font-medium shadow-md hover:shadow-lg transition-all duration-300 hover:scale-105 focus:ring-2 focus:ring-sky-200"
			>
				<PlusCircle class="h-4 w-4 mr-2" /> Bulk import
			</Button>
		</div>
	</div>
</section>


<!-- ===== TABS ===== -->
<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 mt-6">
	<div class="flex border-b border-sky-200 mb-8 overflow-x-auto bg-white/60 backdrop-blur rounded-xl px-2 data-animate='fade-up'">
		{#each ['products','categories','sub-categories','suppliers','locations'] as tab, i}
			<button
				class="px-5 py-2.5 text-sm font-medium transition-all duration-200 rounded-t-xl m-1
		           {activeTab === tab
		             ? 'text-sky-800 bg-gradient-to-b from-sky-100 to-blue-100 border-b-2 border-sky-500 shadow-sm'
		             : 'text-slate-600 hover:text-sky-700 hover:bg-sky-50'}"
				onclick={() => (activeTab = tab as TabKey)}
				style={`animation-delay:${100 + i * 50}ms`}
			>
				{tab.charAt(0).toUpperCase() + tab.slice(1).replace('-', ' ')}
			</button>
		{/each}
	</div>

	<!-- ===== SECTIONS ===== -->
	<section class="space-y-10">
		{#key activeTab}
			{#if activeTab === 'products'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-sky-50 to-blue-100" data-animate="fade-up" style="animation-delay:120ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">SKU Registry</CardTitle>
							<CardDescription class="text-slate-600">Manage items synced with the warehouse</CardDescription>
							<div class="flex items-center gap-2 pt-2">
								<Input class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Search..." bind:value={search.term} />
								<select class="border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={search.key}>
									<option value="name">Name</option>
									<option value="sku">SKU</option>
								</select>
								<Button class="bg-sky-500 text-white rounded-xl px-4 py-2.5" onclick={handleSearch}>Search</Button>
								<Button variant="ghost" class="text-sky-600 rounded-xl px-4 py-2.5" onclick={clearSearch}>Clear</Button>
							</div>
						</CardHeader>
						<CardContent class="pt-0 p-0 overflow-x-auto">
							<Table class="min-w-full">
								<TableHeader class="sticky top-0 bg-gradient-to-r from-sky-100/80 to-blue-100/80 backdrop-blur z-10">
									<TableRow class="border-y border-sky-200/70">
										<TableHead class="px-4 py-3 text-slate-700">SKU</TableHead>
										<TableHead class="px-4 py-3 text-slate-700">Name</TableHead>
										<TableHead class="px-4 py-3 text-slate-700">Status</TableHead>
										<TableHead class="px-4 py-3 text-right text-slate-700">Actions</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
									{#if loading}
										{#each Array(4) as _, i}
											<TableRow><TableCell colspan="4" class="px-4 py-3"><Skeleton class="h-7 w-full bg-white/60" /></TableCell></TableRow>
										{/each}
									{:else if products.length === 0}
										<TableRow>
											<TableCell colspan="4" class="text-center py-6 text-slate-600">No products found</TableCell>
										</TableRow>
									{:else}
										{#each products as product}
											<TableRow onclick={() => viewDetails(product, 'products')} class="cursor-pointer hover:bg-white/90 transition-colors">
												<TableCell class="px-4 py-3 font-mono text-xs text-slate-800">{product.SKU}</TableCell>
												<TableCell class="px-4 py-3 text-slate-900">{product.Name}</TableCell>
												<TableCell class="px-4 py-3">
													<span class="rounded-full bg-sky-100 text-sky-700 px-2.5 py-0.5 text-xs capitalize border border-sky-200 shadow-sm">
														{product.Status ?? 'active'}
													</span>
												</TableCell>
												<TableCell class="px-4 py-3 text-right space-x-2">
													<Button
														size="sm"
														variant="ghost"
														class="text-sky-700 hover:bg-sky-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															editProduct(product);
														}}
													>
														Edit
													</Button>
													<Button
														size="sm"
														variant="ghost"
														class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															deleteProduct(product);
														}}
													>
														Delete
													</Button>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>

					{#if pagination.totalPages > 1}
						<div class="flex flex-col items-center justify-center py-6 space-y-2 bg-white/70 backdrop-blur rounded-2xl shadow-md border border-white/60" data-animate="fade-up" style="animation-delay:180ms">
							<Root
								count={pagination.totalItems}
								perPage={pagination.itemsPerPage}
								page={pagination.currentPage}
								onPageChange={(e) => handlePageChange(e.detail)}
							>
								{#snippet children({ pages, currentPage })}
									<Content class="flex items-center gap-1">
										<Item>
											<PrevButton
												disabled={pagination.currentPage === 1}
												class="rounded-lg bg-sky-50 hover:bg-sky-100 border border-sky-200"
												onclick={() => handlePageChange(pagination.currentPage - 1)}
											/>
										</Item>

										{#each pages as page (page.key)}
											{#if page.type === 'ellipsis'}
												<Item><Ellipsis /></Item>
											{:else}
												<Item>
													<Link
														{page}
														isActive={currentPage === page.value}
														class="rounded-lg data-[active=true]:bg-sky-600 data-[active=true]:text-white data-[active=false]:bg-white/80 data-[active=false]:text-slate-700 hover:scale-105 border border-sky-200 px-3 py-1.5"
														onclick={() => handlePageChange(page.value)}
													>
														{page.value}
													</Link>
												</Item>
											{/if}
										{/each}
										<Item>
											<NextButton
												disabled={false}
												class="rounded-lg bg-sky-50 hover:bg-sky-100 border border-sky-200"
												onclick={() => handlePageChange(pagination.currentPage + 1)}
											/>
										</Item>
									</Content>
								{/snippet}
							</Root>

							<p class="text-sm text-slate-600">
								Showing
								{(pagination.currentPage - 1) * pagination.itemsPerPage + 1}
								–
								{Math.min(pagination.currentPage * pagination.itemsPerPage, pagination.totalItems)}
								of {pagination.totalItems} products
							</p>
						</div>
					{/if}

					<!-- Form -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-sky-50 to-blue-100" data-animate="fade-up" style="animation-delay:240ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">{editingProduct ? 'Update product' : 'Create product'}</CardTitle>
							<CardDescription class="text-slate-600">SKU-level metadata</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="SKU" bind:value={productForm.sku} />
							<Input class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Name" bind:value={productForm.name} />
							<Input class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Description" bind:value={productForm.description} />
							<Input class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Barcode / UPC (must be unique)" bind:value={productForm.barCodeUPC} />

              <div class="select-wrapper">
<select class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={productForm.categoryId} onchange={(e) => loadSubCategories(Number(e?.currentTarget?.value))}>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category.ID}>{category.Name}</option>
								{/each}
							</select>
              </div>
							

							<div class="flex items-center gap-3">
                <div class="select-wrapper">
                  <select class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={productForm.subCategoryId}>
									<option value="">Select sub-category</option>
									{#each subCategories.filter((sc) => sc.CategoryID === Number(productForm.categoryId)) as subCategory}
										<option value={subCategory.ID}>{subCategory.Name}</option>
									{/each}
								</select>
                </div>
								
								<Button size="sm" variant="outline" class="border border-sky-200 text-sky-700 hover:bg-sky-50 rounded-xl px-3 py-2" onclick={() => (activeTab = 'sub-categories')}>New</Button>
							</div>

              <div class="select-wrapper">
                <select class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={productForm.supplierId}>
								<option value="">Select supplier</option>
								{#each suppliers as supplier}
									<option value={supplier.ID}>{supplier.Name}</option>
								{/each}
							</select>
              </div>
							

              <div class="select-wrapper">
                  <select class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={productForm.locationId}>
								<option value="">Default location</option>
								{#each locations as location}
									<option value={location.ID}>{location.Name}</option>
								{/each}
							</select>
              </div>

							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								<Input type="number" min="0" step="0.01" class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Purchase price" bind:value={productForm.purchasePrice} />
								<Input type="number" min="0" step="0.01" class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" placeholder="Selling price" bind:value={productForm.sellingPrice} />
							</div>

              <div class="select-wrapper">

                <select class="w-full border border-sky-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400 bg-white/90" bind:value={productForm.status}>
								<option value="Active">Active</option>
								<option value="Archived">Archived</option>
								<option value="Discontinued">Discontinued</option>
							</select>
              </div>
							<div class="flex flex-col sm:flex-row gap-3 pt-1 pr-2">
								<Button class="w-full sm:w-1/2 bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white rounded-xl py-2.5 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105" onclick={saveProduct}>
									{editingProduct ? 'Update' : 'Create'}
								</Button>
								<Button class="w-full sm:w-1/2 border border-sky-200 text-sky-700 hover:bg-sky-50 rounded-xl py-2.5 transition" onclick={resetProductForm}>
									Reset
								</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'categories'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-green-50 to-emerald-100" data-animate="fade-up" style="animation-delay:120ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">Categories</CardTitle>
							<CardDescription class="text-slate-600">Structure your catalog foundation</CardDescription>
							<div class="flex items-center gap-2 pt-2">
								<Input class="w-full border border-emerald-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-emerald-400 bg-white/90" placeholder="Search by name..." bind:value={categorySearchTerm} />
								<Button class="bg-emerald-500 text-white rounded-xl px-4 py-2.5" onclick={handleCategorySearch}>Search</Button>
								<Button variant="ghost" class="text-emerald-600 rounded-xl px-4 py-2.5" onclick={clearCategorySearch}>Clear</Button>
							</div>
						</CardHeader>
						<CardContent class="pt-0 p-0 overflow-x-auto">
							<Table class="min-w-full">
								<TableHeader class="sticky top-0 bg-gradient-to-r from-green-100/80 to-emerald-100/80 backdrop-blur z-10">
									<TableRow class="border-y border-emerald-200/70">
										<TableHead class="px-4 py-3 text-slate-700">Name</TableHead>
										<TableHead class="px-4 py-3 text-right text-slate-700">Actions</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
									{#if loading}
										{#each Array(3) as _, i}
											<TableRow><TableCell colspan="2" class="px-4 py-3"><Skeleton class="h-7 w-full bg-white/60" /></TableCell></TableRow>
										{/each}
									{:else if categories.length === 0}
										<TableRow>
											<TableCell colspan="2" class="text-center py-6 text-slate-600">No categories found</TableCell>
										</TableRow>
									{:else}
										{#each categories as category}
											<TableRow onclick={() => viewDetails(category, 'categories')} class="cursor-pointer hover:bg-white/90 transition-colors">
												<TableCell class="px-4 py-3">{category.Name}</TableCell>
												<TableCell class="px-4 py-3 text-right space-x-2">
													<Button
														size="sm"
														variant="ghost"
														class="text-emerald-700 hover:bg-emerald-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															editingCategory = category;
															categoryForm.name = category.Name;
														}}
													>
														Edit
													</Button>
													<Button
														size="sm"
														variant="ghost"
														class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															deleteCategory(category);
														}}
													>
														Delete
													</Button>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>

					<!-- Form -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-green-50 to-emerald-100" data-animate="fade-up" style="animation-delay:180ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">{editingCategory ? 'Update category' : 'Create category'}</CardTitle>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input class="w-full border border-emerald-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-emerald-400 bg-white/90" placeholder="Name" bind:value={categoryForm.name} />
							<div class="flex flex-col sm:flex-row gap-3 pt-1 pr-2">
								<Button class="w-full sm:w-1/2 bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700 text-white rounded-xl py-2.5 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105" onclick={saveCategory}>{editingCategory ? 'Update' : 'Create'}</Button>
								<Button class="w-full sm:w-1/2 border border-emerald-200 text-emerald-700 hover:bg-emerald-50 rounded-xl py-2.5 transition" onclick={resetCategoryForm}>Reset</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'sub-categories'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- TABLE SECTION -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-amber-50 to-orange-100" data-animate="fade-up" style="animation-delay:120ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">Sub Categories</CardTitle>
							<CardDescription class="text-slate-600">Filter by parent category</CardDescription>
						</CardHeader>

						<CardContent class="space-y-4 p-6 overflow-x-auto">
							<select
								class="w-full sm:w-1/2 border border-amber-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400 bg-white/90"
								bind:value={subCategoryForm.categoryId}
								onchange={(e) => loadSubCategories(Number(e?.currentTarget?.value))}
							>
								<option value="">Select category to view sub-categories</option>
								{#if categories.length > 0}
									{#each categories as category}
										<option value={category.ID}>{category.Name}</option>
									{/each}
								{:else}
									<option disabled>Loading categories...</option>
								{/if}
							</select>

							{#if !subCategoryForm.categoryId}
								<div class="py-10 text-center text-slate-600 border border-dashed border-amber-200 rounded-xl bg-white/60">
									<p class="text-sm">Select a category to view sub-categories</p>
								</div>
							{:else}
								<Table class="min-w-full border border-amber-200 rounded-xl overflow-hidden">
									<TableHeader class="bg-gradient-to-r from-amber-100/80 to-orange-100/80 backdrop-blur">
										<TableRow class="border-y border-amber-200/70">
											<TableHead class="px-4 py-3 text-slate-700">Name</TableHead>
											<TableHead class="px-4 py-3 text-right text-slate-700">Actions</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
										{#if loading}
											{#each Array(3) as _, i}
												<TableRow><TableCell colspan="2" class="px-4 py-3"><Skeleton class="h-7 w-full bg-white/60" /></TableCell></TableRow>
											{/each}
										{:else}
											{#each subCategories.filter((sc) => sc.CategoryID === Number(subCategoryForm.categoryId)) as subCategory (subCategory.ID)}
												<TableRow onclick={() => viewDetails(subCategory, 'sub-categories')} class="cursor-pointer hover:bg-white/90 transition-colors">
													<TableCell class="px-4 py-3">{subCategory.Name}</TableCell>
													<TableCell class="px-4 py-3 text-right space-x-2">
														<Button
															size="sm"
															variant="ghost"
															class="text-amber-700 hover:bg-amber-100 rounded-lg px-3 py-1.5"
															onclick={(event) => {
																event.stopPropagation();
																editingSubCategory = subCategory;
																subCategoryForm.name = subCategory.Name;
																subCategoryForm.categoryId = String(subCategory.CategoryID);
															}}
														>
															Edit
														</Button>
														<Button
															size="sm"
															variant="ghost"
															class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
															onclick={(event) => {
																event.stopPropagation();
																deleteSubCategory(subCategory);
															}}
														>
															Delete
														</Button>
													</TableCell>
												</TableRow>
											{:else}
												<TableRow>
													<TableCell colspan="2" class="text-center py-6 text-slate-600">
														No sub-categories found for this category
													</TableCell>
												</TableRow>
											{/each}
										{/if}
									</TableBody>
								</Table>
							{/if}
						</CardContent>
					</Card>

					<!-- FORM SECTION -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-amber-50 to-orange-100" data-animate="fade-up" style="animation-delay:180ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">{editingSubCategory ? 'Update sub-category' : 'Create sub-category'}</CardTitle>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input class="w-full border border-amber-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400 bg-white/90" placeholder="Name" bind:value={subCategoryForm.name} />
							<select class="w-full border border-amber-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400 bg-white/90" bind:value={subCategoryForm.categoryId}>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category.ID}>{category.Name}</option>
								{/each}
							</select>
							<div class="flex flex-col sm:flex-row gap-3 pt-1 pr-2">
								<Button class="w-full sm:w-1/2 bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-white rounded-xl py-2.5 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105" onclick={saveSubCategory}>
									{editingSubCategory ? 'Update' : 'Create'}
								</Button>
								<Button class="w-full sm:w-1/2 border border-amber-200 text-amber-700 hover:bg-amber-50 rounded-xl py-2.5 transition" onclick={resetSubCategoryForm}>
									Reset
								</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'suppliers'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-violet-50 to-purple-100" data-animate="fade-up" style="animation-delay:120ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">Suppliers</CardTitle>
							<CardDescription class="text-slate-600">Strategic partners powering replenishment</CardDescription>
							<div class="flex items-center gap-2 pt-2">
								<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Search by name..." bind:value={supplierSearchTerm} />
								<Button class="bg-violet-500 text-white rounded-xl px-4 py-2.5" onclick={handleSupplierSearch}>Search</Button>
								<Button variant="ghost" class="text-violet-600 rounded-xl px-4 py-2.5" onclick={clearSupplierSearch}>Clear</Button>
							</div>
						</CardHeader>
						<CardContent class="pt-0 p-0 overflow-x-auto">
							<Table class="min-w-full">
								<TableHeader class="sticky top-0 bg-gradient-to-r from-violet-100/80 to-purple-100/80 backdrop-blur z-10">
									<TableRow class="border-y border-violet-200/70">
										<TableHead class="px-4 py-3 text-slate-700">Name</TableHead>
										<TableHead class="px-4 py-3 text-slate-700">Contact</TableHead>
										<TableHead class="px-4 py-3 text-right text-slate-700">Actions</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
									{#if loading}
										{#each Array(3) as _, i}
											<TableRow><TableCell colspan="3" class="px-4 py-3"><Skeleton class="h-7 w-full bg-white/60" /></TableCell></TableRow>
										{/each}
									{:else if suppliers.length === 0}
										<TableRow>
											<TableCell colspan="3" class="text-center py-6 text-slate-600">No suppliers found</TableCell>
										</TableRow>
									{:else}
										{#each suppliers as supplier}
											<TableRow onclick={() => viewDetails(supplier, 'suppliers')} class="cursor-pointer hover:bg-white/90 transition-colors">
												<TableCell class="px-4 py-3">{supplier.Name}</TableCell>
												<TableCell class="px-4 py-3">
													<p class="text-sm text-slate-800">{supplier.ContactPerson ?? '—'}</p>
													<p class="text-xs text-slate-600">{supplier.Email ?? supplier.Phone ?? ''}</p>
												</TableCell>
												<TableCell class="px-4 py-3 text-right space-x-2">
													<Button
														size="sm"
														variant="ghost"
														class="text-violet-700 hover:bg-violet-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															editingSupplier = supplier;
															Object.assign(supplierForm, {
																name: supplier.Name,
																contactPerson: supplier.ContactPerson ?? '',
																email: supplier.Email ?? '',
																phone: supplier.Phone ?? '',
																address: supplier.Address ?? '',
															});
														}}
													>
														Edit
													</Button>
													<Button
														size="sm"
														variant="ghost"
														class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															deleteSupplier(supplier);
														}}
													>
														Delete
													</Button>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>

					<!-- Form -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-violet-50 to-purple-100" data-animate="fade-up" style="animation-delay:180ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">{editingSupplier ? 'Update supplier' : 'Create supplier'}</CardTitle>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Name" bind:value={supplierForm.name} />
							<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Contact person" bind:value={supplierForm.contactPerson} />
							<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Email" bind:value={supplierForm.email} />
							<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Phone" bind:value={supplierForm.phone} />
							<Input class="w-full border border-violet-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400 bg-white/90" placeholder="Address" bind:value={supplierForm.address} />
							<div class="flex flex-col sm:flex-row gap-3 pt-1 pr-2">
								<Button class="w-full sm:w-1/2 bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white rounded-xl py-2.5 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105" onclick={saveSupplier}>{editingSupplier ? 'Update' : 'Create'}</Button>
								<Button class="w-full sm:w-1/2 border border-violet-200 text-violet-700 hover:bg-violet-50 rounded-xl py-2.5 transition" onclick={resetSupplierForm}>Reset</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else}
				<!-- LOCATIONS -->
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-cyan-50 to-teal-100" data-animate="fade-up" style="animation-delay:120ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">Locations</CardTitle>
							<CardDescription class="text-slate-600">Fulfilment nodes and stores</CardDescription>
						</CardHeader>
						<CardContent class="pt-0 p-0 overflow-x-auto">
							<Table class="min-w-full">
								<TableHeader class="sticky top-0 bg-gradient-to-r from-cyan-100/80 to-teal-100/80 backdrop-blur z-10">
									<TableRow class="border-y border-teal-200/70">
										<TableHead class="px-4 py-3 text-slate-700">Name</TableHead>
										<TableHead class="px-4 py-3 text-slate-700">Address</TableHead>
										<TableHead class="px-4 py-3 text-right text-slate-700">Actions</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody class="[&>tr:nth-child(even)]:bg-white/70 [&>tr:nth-child(odd)]:bg-white/50">
									{#if loading}
										{#each Array(3) as _, i}
											<TableRow><TableCell colspan="3" class="px-4 py-3"><Skeleton class="h-7 w-full bg-white/60" /></TableCell></TableRow>
										{/each}
									{:else if locations.length === 0}
										<TableRow>
											<TableCell colspan="3" class="text-center py-6 text-slate-600">No locations found</TableCell>
										</TableRow>
									{:else}
										{#each locations as location}
											<TableRow onclick={() => viewDetails(location, 'locations')} class="cursor-pointer hover:bg-white/90 transition-colors">
												<TableCell class="px-4 py-3">{location.Name}</TableCell>
												<TableCell class="px-4 py-3 text-slate-700">{location.Address ?? '—'}</TableCell>
												<TableCell class="px-4 py-3 text-right space-x-2">
													<Button
														size="sm"
														variant="ghost"
														class="text-cyan-700 hover:bg-cyan-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															editingLocation = location;
															locationForm.name = location.Name;
															locationForm.address = location.Address ?? '';
														}}
													>
														Edit
													</Button>
													<Button
														size="sm"
														variant="ghost"
														class="text-rose-700 hover:bg-rose-100 rounded-lg px-3 py-1.5"
														onclick={(event) => {
															event.stopPropagation();
															deleteLocation(location);
														}}
													>
														Delete
													</Button>
												</TableCell>
											</TableRow>
										{/each}
									{/if}
								</TableBody>
							</Table>
						</CardContent>
					</Card>

					<!-- Form -->
					<Card class="rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] overflow-hidden border-0 bg-gradient-to-br from-cyan-50 to-teal-100" data-animate="fade-up" style="animation-delay:180ms">
						<CardHeader class="space-y-1 bg-white/70 backdrop-blur px-6 py-5 border-b border-white/60">
							<CardTitle class="text-slate-800">{editingLocation ? 'Update location' : 'Create location'}</CardTitle>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input class="w-full border border-teal-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400 bg-white/90" placeholder="Name" bind:value={locationForm.name} />
							<Input class="w-full border border-teal-200 rounded-xl px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400 bg-white/90" placeholder="Address" bind:value={locationForm.address} />
							<div class="flex flex-col sm:flex-row gap-3 pt-1 pr-2">
								<Button class="w-full sm:w-1/2 bg-gradient-to-r from-cyan-500 to-teal-600 hover:from-cyan-600 hover:to-teal-700 text-white rounded-xl py-2.5 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-105" onclick={saveLocation}>{editingLocation ? 'Update' : 'Create'}</Button>
								<Button class="w-full sm:w-1/2 border border-teal-200 text-teal-700 hover:bg-teal-50 rounded-xl py-2.5 transition" onclick={resetLocationForm}>Reset</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{/if}
		{/key}
	</section>
</div>

<style lang="postcss">
	/* Smooth transitions globally */
	* {
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Hero gradient animation */
	@keyframes gradientShift {
		0% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
		100% { background-position: 0% 50%; }
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 16s ease-in-out infinite;
	}

	/* Soft glowing blobs */
	@keyframes pulseGlow {
		0%, 100% { transform: scale(1); opacity: 0.45; filter: blur(80px); }
		50% { transform: scale(1.08); opacity: 0.6; filter: blur(90px); }
	}
	.animate-pulseGlow { animation: pulseGlow 10s ease-in-out infinite; }

	/* Card float micro-motion */
	@keyframes cardFloat {
		0%, 100% { transform: translateY(0); }
		50% { transform: translateY(-4px); }
	}
	.animate-cardFloat { animation: cardFloat 4s ease-in-out infinite; }

	/* Fade-up reveal */
	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(12px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp {
		animation: fadeUp 500ms var(--ease, cubic-bezier(0.4, 0, 0.2, 1)) forwards;
	}

	/* Pastel scrollbar */
	::-webkit-scrollbar { width: 8px; height: 8px; }
	::-webkit-scrollbar-track { background: transparent; }
	::-webkit-scrollbar-thumb {
		background: rgba(14, 165, 233, 0.25);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover { background: rgba(14, 165, 233, 0.35); }


  .parallax-hero {
    transform: translateY(0);
    will-change: transform, filter;
    transition: transform 0.1s ease-out, filter 0.2s ease-out;
  }


@keyframes gradientShift {
	0% { background-position: 0% 50%; }
	50% { background-position: 100% 50%; }
	100% { background-position: 0% 50%; }
}
.animate-gradientShift {
	animation: gradientShift 18s ease-in-out infinite;
}

@keyframes pulseGlow {
	0%, 100% { transform: scale(1); opacity: 0.45; }
	50% { transform: scale(1.08); opacity: 0.7; }
}
.animate-pulseGlow {
	animation: pulseGlow 12s ease-in-out infinite;
}

@keyframes cardFloat {
	0%, 100% { transform: translateY(0); }
	50% { transform: translateY(-4px); }
}
.animate-cardFloat {
	animation: cardFloat 4s ease-in-out infinite;
}

@keyframes fadeUp {
	from { opacity: 0; transform: translateY(12px); }
	to { opacity: 1; transform: translateY(0); }
}
.animate-fadeUp {
	animation: fadeUp 0.6s ease-out forwards;
}

/* Responsive tweaks */
@media (max-width: 640px) {
	.parallax-hero {
		padding-top: 5rem;
		padding-bottom: 3rem;
	}
}

</style>
