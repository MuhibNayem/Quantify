<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Root,
		Content,
		Item,
		PrevButton,
		NextButton,
		Ellipsis,
		Link
	} from '$lib/components/ui/pagination';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DetailsModal from '$lib/components/DetailsModal.svelte';
	import DataTable from '$lib/components/ui/data-table/DataTable.svelte';
	import type {
		DetailBuilderContext,
		DetailExtraFetcher,
		DetailSection
	} from '$lib/components/DetailsModal.svelte';
	import {
		productsApi,
		categoriesApi,
		subCategoriesApi,
		suppliersApi,
		locationsApi
	} from '$lib/api/resources';
	import type {
		Category,
		Location,
		Product,
		StockAdjustment,
		SubCategory,
		Supplier,
		SupplierPerformance
	} from '$lib/types';
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
		Search
	} from 'lucide-svelte';
	import { auth } from '$lib/stores/auth';

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
		barCodeUPC: ''
	});

	const pagination = $state({
		currentPage: 1,
		totalPages: 1,
		totalItems: 0,
		itemsPerPage: 10
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
	let modalSectionsBuilder =
		$state<(ctx: DetailBuilderContext) => DetailSection[]>(emptySectionBuilder);
	let useLegacyModalSlot = $state(true);

	type StockSnapshot = Awaited<ReturnType<typeof productsApi.stock>>;

	const currencyFormatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });
	const dateTimeFormatter = new Intl.DateTimeFormat('en-US', {
		dateStyle: 'medium',
		timeStyle: 'short'
	});
	const percentFormatter = new Intl.NumberFormat('en-US', {
		style: 'percent',
		maximumFractionDigits: 1
	});

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
						hint: batches.length
							? `${batches.length} active batch${batches.length === 1 ? '' : 'es'}`
							: 'No active batches',
						icon: Package,
						accent: 'sky'
					},
					{
						title: 'Status',
						value: product.Status ?? 'Unknown',
						hint: lastAdjustment
							? `Updated ${formatDateTime(lastAdjustment.AdjustedAt)}`
							: 'No adjustments yet',
						icon: Tag,
						accent: 'emerald'
					},
					{
						title: 'Selling Price',
						value: formatCurrency(product.SellingPrice),
						hint: `Purchase ${formatCurrency(product.PurchasePrice)}`,
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
					{ label: 'SKU', value: product.SKU },
					{ label: 'Name', value: product.Name },
					{
						label: 'Status',
						value: product.Status ?? 'Unknown',
						badge: statusBadge(product.Status)
					},
					{ label: 'Purchase Price', value: formatCurrency(product.PurchasePrice) },
					{ label: 'Selling Price', value: formatCurrency(product.SellingPrice) },
					{ label: 'Barcode', value: product.BarcodeUPC ?? '—' }
				]
			},
			{
				type: 'description',
				title: 'Associations',
				description: 'Upstream supplier & placement details.',
				items: [
					{
						label: 'Category',
						value: product.Category?.Name ?? (product.CategoryID ? `#${product.CategoryID}` : '—')
					},
					{
						label: 'Sub-Category',
						value:
							product.SubCategory?.Name ??
							(product.SubCategoryID ? `#${product.SubCategoryID}` : '—')
					},
					{
						label: 'Supplier',
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
				rows: recentHistory,
				emptyText: 'No adjustments recorded for this product yet.'
			}
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
					{ label: 'Name', value: category.Name },
					{ label: 'ID', value: category.ID },
					{ label: 'Created', value: formatDateTime(category.CreatedAt) },
					{ label: 'Updated', value: formatDateTime(category.UpdatedAt) },
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
				rows: children,
				emptyText: 'No sub-categories associated yet.'
			}
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
					{ label: 'Name', value: subCategory.Name },
					{ label: 'Parent', value: parent?.Name ?? `Category #${subCategory.CategoryID}` },
					{ label: 'Created', value: formatDateTime(subCategory.CreatedAt) },
					{ label: 'Updated', value: formatDateTime(subCategory.UpdatedAt) }
				]
			}
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
						accent: 'emerald'
					},
					{
						title: 'Avg. lead time',
						value: performance?.averageLeadTimeDays
							? `${performance.averageLeadTimeDays} days`
							: '—',
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
				title: 'Contact Details',
				items: [
					{ label: 'Name', value: supplier.Name },
					{ label: 'Contact Person', value: supplier.ContactPerson ?? '—' },
					{ label: 'Email', value: supplier.Email ?? '—', icon: Mail },
					{ label: 'Phone', value: supplier.Phone ?? '—', icon: Phone },
					{ label: 'Address', value: supplier.Address ?? '—', icon: MapPin }
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
						value: performance?.averageLeadTimeDays
							? `${performance.averageLeadTimeDays} days`
							: '—'
					}
				],
				emptyText: 'Performance metrics unavailable.'
			}
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
				title: 'Location Profile',
				items: [
					{ label: 'Name', value: location.Name },
					{ label: 'Address', value: location.Address ?? '—' },
					{ label: 'Created', value: formatDateTime(location.CreatedAt) },
					{ label: 'Updated', value: formatDateTime(location.UpdatedAt) }
				]
			}
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
				modalSubtitle = resource.SKU ? `SKU ${resource.SKU}` : (resource.Name ?? null);
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
		loading = true;
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

	const editProduct = async (product: Product) => {
		console.log('Editing product:', product);
		editingProduct = product;
		productForm.sku = product.SKU;
		productForm.name = product.Name;
		productForm.description = product.Description ?? '';
		productForm.categoryId = product.CategoryID ? String(product.CategoryID) : '';
		productForm.subCategoryId = product.SubCategoryID ? String(product.SubCategoryID) : '';
		productForm.supplierId = product.SupplierID ? String(product.SupplierID) : '';
		productForm.locationId = product.LocationID ? String(product.LocationID) : '';
		productForm.purchasePrice = product.PurchasePrice ? String(product.PurchasePrice) : '';
		productForm.sellingPrice = product.SellingPrice ? String(product.SellingPrice) : '';
		productForm.status = product.Status ?? 'Active';
		productForm.barCodeUPC = product.BarcodeUPC ?? '';

		console.log('Form state after set:', $state.snapshot(productForm));
		console.log('Available Categories:', $state.snapshot(categories));
		console.log(
			'Category Selection Match:',
			categories.find((c) => String(c.ID) === productForm.categoryId)
		);

		if (product.CategoryID) {
			await loadSubCategories(product.CategoryID);
		}
	};

	const saveProduct = async () => {
		if (!productForm.barCodeUPC.trim()) {
			toast.error('Missing Barcode/UPC', {
				description: 'Each product must have a unique BarcodeUPC value.'
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
				description: `The BarcodeUPC "${productForm.barCodeUPC}" is already used by product "${duplicate.Name}".`
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
			BarcodeUPC: productForm.barCodeUPC.trim()
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
		if (!window.confirm(`Are you sure you want to delete ${product.Name}?`)) return;
		try {
			await productsApi.remove(product.ID);
			toast.success('Product removed');
			await loadAll();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete product';
			toast.error('Failed to Delete Product', { description: errorMessage });
		}
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
		if (!window.confirm(`Are you sure you want to delete ${category.Name}?`)) return;
		try {
			await categoriesApi.remove(category.ID);
			toast.success('Category removed');
			await loadAll();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete category';
			toast.error('Failed to Delete Category', { description: errorMessage });
		}
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
		if (!window.confirm(`Are you sure you want to delete ${subCategory.Name}?`)) return;
		try {
			await subCategoriesApi.remove(subCategory.ID);
			toast.success('Sub-category removed');
			await loadAll();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete sub-category';
			toast.error('Failed to Delete Sub-Category', { description: errorMessage });
		}
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
		if (!window.confirm(`Are you sure you want to delete ${supplier.Name}?`)) return;
		try {
			await suppliersApi.remove(supplier.ID);
			toast.success('Supplier removed');
			await loadAll();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete supplier';
			toast.error('Failed to Delete Supplier', { description: errorMessage });
		}
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
		if (!window.confirm(`Are you sure you want to delete ${location.Name}?`)) return;
		try {
			await locationsApi.remove(location.ID);
			toast.success('Location removed');
			await loadAll();
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to delete location';
			toast.error('Failed to Delete Location', { description: errorMessage });
		}
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
<section class="relative isolate w-full overflow-hidden">
	<!-- Gradient background with motion -->
	<div
		class="animate-gradientShift absolute inset-0 -z-10 bg-gradient-to-r from-sky-50 via-blue-50 to-cyan-100 bg-[length:200%_200%]"
	></div>

	<!-- Floating glow blobs -->
	<div
		class="animate-pulseGlow absolute -left-24 -top-32 h-96 w-96 rounded-full bg-sky-200/40 blur-3xl"
	></div>
	<div
		class="animate-pulseGlow absolute -bottom-28 -right-24 h-80 w-80 rounded-full bg-cyan-200/30 blur-3xl delay-700"
	></div>

	<!-- Hero container -->
	<div
		class="parallax-hero relative mx-auto max-w-7xl px-4 pb-10 pt-16 text-center sm:px-6 sm:pb-16 sm:pt-20 sm:text-left lg:px-8"
		style="transform: translateY(var(--hero-translate, 0));"
	>
		<div class="mb-3 inline-flex items-center justify-center gap-3 sm:justify-start">
			<span
				class="animate-cardFloat inline-flex rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 p-2 shadow-md"
			>
				<Zap class="h-6 w-6 text-white" />
			</span>
			<p class="text-xs font-semibold uppercase tracking-[0.18em] text-sky-700 sm:text-sm">
				Catalog Cockpit
			</p>
		</div>

		<h1
			class="mb-3 bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
		>
			Products, Categories &amp; Partners
		</h1>
		<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:mx-0 sm:text-base">
			Unified control center for your catalog data
		</p>

		<!-- Action buttons -->
		<div class="mt-6 flex flex-col justify-center gap-3 sm:flex-row sm:justify-start">
			<Button
				variant="secondary"
				onclick={loadAll}
				class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-sky-600 hover:to-blue-700 hover:shadow-xl focus:ring-2 focus:ring-sky-300 sm:w-auto"
			>
				<RefreshCcw class="mr-2 h-4 w-4" /> Sync data
			</Button>
			<Button
				href="/bulk"
				variant="outline"
				class="w-full rounded-xl border border-sky-200 bg-white/80 px-5 py-2.5 font-medium text-sky-700 shadow-md transition-all duration-300 hover:scale-105 hover:bg-sky-50 hover:shadow-lg focus:ring-2 focus:ring-sky-200 sm:w-auto"
			>
				<PlusCircle class="mr-2 h-4 w-4" /> Bulk import
			</Button>
		</div>
	</div>
</section>

<!-- ===== TABS ===== -->
<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
	<div
		class="data-animate='fade-up' mb-8 flex overflow-x-auto rounded-xl border-b border-sky-200 bg-white/60 px-2 backdrop-blur"
	>
		{#each ['products', 'categories', 'sub-categories', 'suppliers', 'locations'] as tab, i}
			<button
				class="m-1 rounded-t-xl px-5 py-2.5 text-sm font-medium transition-all duration-200
		           {activeTab === tab
					? 'border-b-2 border-sky-500 bg-gradient-to-b from-sky-100 to-blue-100 text-sky-800 shadow-sm'
					: 'text-slate-600 hover:bg-sky-50 hover:text-sky-700'}"
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
					<!-- Table Section Wrapper -->
					<div class="flex flex-col gap-6">
						<!-- Header with Search -->
						<div
							class="flex gap-4 rounded-2xl border border-sky-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
						>
							<div class="space-y-1">
								<h2 class="text-lg font-semibold text-slate-800">SKU Registry</h2>
								<p class="text-sm text-slate-500">Manage items synced with the warehouse</p>
							</div>
							<div class="flex flex-col gap-3 sm:flex-row">
								<div class="relative flex-1">
									<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
									<Input
										placeholder="Search products..."
										bind:value={search.term}
										onkeydown={(e) => e.key === 'Enter' && handleSearch()}
										class="rounded-xl border-sky-200 bg-white/90 pl-9 focus:ring-2 focus:ring-sky-400"
									/>
								</div>
								<div class="flex gap-2">
									<Button
										variant="secondary"
										onclick={handleSearch}
										class="rounded-xl border border-sky-200 bg-white text-sky-700 hover:bg-sky-50"
									>
										Search
									</Button>
									{#if auth.hasPermission('products.write')}
										<Button
											onclick={() => {
												resetProductForm();
												openModal('/products', 'Add Product');
											}}
											class="rounded-xl bg-gradient-to-r from-sky-500 to-indigo-600 text-white shadow-md transition-all hover:from-sky-600 hover:to-indigo-700 hover:shadow-lg"
										>
											<PlusCircle class="mr-2 h-4 w-4" /> Add Product
										</Button>
									{/if}
								</div>
							</div>
						</div>

						<DataTable
							data={products}
							columns={[
								{ header: 'SKU', accessorKey: 'SKU' },
								{ header: 'Name', accessorKey: 'Name' },
								{ header: 'Status', accessorKey: 'Status' },
								{ header: 'Actions', accessorKey: 'id', class: 'text-right' }
							]}
							totalItems={pagination.totalItems}
							pageSize={pagination.itemsPerPage}
							currentPage={pagination.currentPage}
							onPageChange={handlePageChange}
							{loading}
							onRowClick={(product) => viewDetails(product, 'products')}
						>
							{#snippet children(product)}
								<TableCell class="font-mono text-xs text-slate-600">{product.SKU}</TableCell>
								<TableCell class="font-medium text-slate-900">{product.Name}</TableCell>
								<TableCell>
									{@const badge = statusBadge(product.Status)}
									{#if badge}
										<span
											class="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize shadow-sm
											{badge.variant === 'success' ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : ''}
											{badge.variant === 'warning' ? 'border-amber-200 bg-amber-50 text-amber-700' : ''}
											{badge.variant === 'danger' ? 'border-rose-200 bg-rose-50 text-rose-700' : ''}
											{badge.variant === 'info' ? 'border-sky-200 bg-sky-50 text-sky-700' : ''}"
										>
											{badge.text}
										</span>
									{:else}
										<span class="text-slate-400">—</span>
									{/if}
								</TableCell>
								<TableCell class="text-right">
									<div class="flex items-center justify-end gap-1">
										<Button
											size="sm"
											variant="ghost"
											class="h-8 text-sky-600 hover:bg-sky-50 hover:text-sky-700"
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
											class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
											onclick={(event) => {
												event.stopPropagation();
												deleteProduct(product);
											}}
										>
											Delete
										</Button>
									</div>
								</TableCell>
							{/snippet}
						</DataTable>
					</div>

					<!-- Form -->
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-sky-50 to-blue-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
						data-animate="fade-up"
						style="animation-delay:240ms"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800"
								>{editingProduct ? 'Update product' : 'Create product'}</CardTitle
							>
							<CardDescription class="text-slate-600">SKU-level metadata</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input
								class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
								placeholder="SKU"
								bind:value={productForm.sku}
							/>
							<Input
								class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
								placeholder="Name"
								bind:value={productForm.name}
							/>
							<Input
								class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
								placeholder="Description"
								bind:value={productForm.description}
							/>
							<Input
								class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
								placeholder="Barcode / UPC (must be unique)"
								bind:value={productForm.barCodeUPC}
							/>

							<div class="select-wrapper">
								<select
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									bind:value={productForm.categoryId}
									onchange={(e) => loadSubCategories(Number(e?.currentTarget?.value))}
								>
									<option value="">Select category</option>
									{#each categories as category}
										<option value={String(category.ID)}>{category.Name}</option>
									{/each}
								</select>
							</div>

							<div class="flex items-center gap-3">
								<div class="select-wrapper">
									<select
										class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
										bind:value={productForm.subCategoryId}
									>
										<option value="">Select sub-category</option>
										{#each subCategories.filter((sc) => sc.CategoryID === Number(productForm.categoryId)) as subCategory}
											<option value={String(subCategory.ID)}>{subCategory.Name}</option>
										{/each}
									</select>
								</div>

								<Button
									size="sm"
									variant="outline"
									class="rounded-xl border border-sky-200 px-3 py-2 text-sky-700 hover:bg-sky-50"
									onclick={() => (activeTab = 'sub-categories')}>New</Button
								>
							</div>

							<div class="select-wrapper">
								<select
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									bind:value={productForm.supplierId}
								>
									<option value="">Select supplier</option>
									{#each suppliers as supplier}
										<option value={String(supplier.ID)}>{supplier.Name}</option>
									{/each}
								</select>
							</div>

							<div class="select-wrapper">
								<select
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									bind:value={productForm.locationId}
								>
									<option value="">Default location</option>
									{#each locations as location}
										<option value={String(location.ID)}>{location.Name}</option>
									{/each}
								</select>
							</div>

							<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
								<Input
									type="number"
									min="0"
									step="0.01"
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									placeholder="Purchase price"
									bind:value={productForm.purchasePrice}
								/>
								<Input
									type="number"
									min="0"
									step="0.01"
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									placeholder="Selling price"
									bind:value={productForm.sellingPrice}
								/>
							</div>

							<div class="select-wrapper">
								<select
									class="w-full rounded-xl border border-sky-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-sky-400"
									bind:value={productForm.status}
								>
									<option value="Active">Active</option>
									<option value="Archived">Archived</option>
									<option value="Discontinued">Discontinued</option>
								</select>
							</div>
							<div class="flex flex-col gap-3 pr-2 pt-1 sm:flex-row">
								<Button
									class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-sky-600 hover:to-blue-700 hover:shadow-xl sm:w-1/2"
									onclick={saveProduct}
								>
									{editingProduct ? 'Update' : 'Create'}
								</Button>
								<Button
									class="w-full rounded-xl border border-sky-200 py-2.5 text-sky-700 transition hover:bg-sky-50 sm:w-1/2"
									onclick={resetProductForm}
								>
									Reset
								</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'categories'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table Section Wrapper -->
					<div class="flex flex-col gap-6">
						<!-- Header with Search -->
						<div
							class="flex flex-col gap-4 rounded-2xl border border-emerald-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
						>
							<div class="space-y-1">
								<h2 class="text-lg font-semibold text-slate-800">Categories</h2>
								<p class="text-sm text-slate-500">Structure your catalog foundation</p>
							</div>
							<div class="flex flex-wrap items-center gap-2">
								<Input
									class="w-full rounded-xl border-emerald-200 bg-white/80 px-3 text-sm focus:ring-2 focus:ring-emerald-400 sm:w-48"
									placeholder="Search by name..."
									bind:value={categorySearchTerm}
								/>
								<Button
									class="rounded-xl bg-emerald-500 px-4 py-2 text-white hover:bg-emerald-600"
									onclick={handleCategorySearch}>Search</Button
								>
								<Button
									variant="ghost"
									class="rounded-xl px-4 py-2 text-emerald-600 hover:bg-emerald-50"
									onclick={clearCategorySearch}>Clear</Button
								>
							</div>
						</div>

						<DataTable
							data={categories}
							columns={[
								{ header: 'Name', accessorKey: 'Name' },
								{ header: 'Actions', accessorKey: 'id', class: 'text-right' }
							]}
							{loading}
							onRowClick={(category) => viewDetails(category, 'categories')}
						>
							{#snippet children(category)}
								<TableCell class="font-medium text-slate-800">{category.Name}</TableCell>
								<TableCell class="text-right">
									<div class="flex items-center justify-end gap-1">
										<Button
											size="sm"
											variant="ghost"
											class="h-8 text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
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
											class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
											onclick={(event) => {
												event.stopPropagation();
												deleteCategory(category);
											}}
										>
											Delete
										</Button>
									</div>
								</TableCell>
							{/snippet}
						</DataTable>
					</div>

					<!-- Form -->
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-green-50 to-emerald-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
						data-animate="fade-up"
						style="animation-delay:180ms"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800"
								>{editingCategory ? 'Update category' : 'Create category'}</CardTitle
							>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input
								class="w-full rounded-xl border border-emerald-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-emerald-400"
								placeholder="Name"
								bind:value={categoryForm.name}
							/>
							<div class="flex flex-col gap-3 pr-2 pt-1 sm:flex-row">
								<Button
									class="w-full rounded-xl bg-gradient-to-r from-emerald-500 to-green-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-emerald-600 hover:to-green-700 hover:shadow-xl sm:w-1/2"
									onclick={saveCategory}>{editingCategory ? 'Update' : 'Create'}</Button
								>
								<Button
									class="w-full rounded-xl border border-emerald-200 py-2.5 text-emerald-700 transition hover:bg-emerald-50 sm:w-1/2"
									onclick={resetCategoryForm}>Reset</Button
								>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'sub-categories'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- TABLE SECTION -->
					<div class="flex flex-col gap-6">
						<!-- Header with Dropdown -->
						<div
							class="flex flex-col gap-4 rounded-2xl border border-amber-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
						>
							<div class="space-y-1">
								<h2 class="text-lg font-semibold text-slate-800">Sub Categories</h2>
								<p class="text-sm text-slate-500">Filter by parent category</p>
							</div>
							<div class="select-wrapper w-full sm:w-1/2">
								<select
									class="w-full rounded-xl border border-amber-200 bg-white/80 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-amber-400"
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
							</div>
						</div>

						{#if !subCategoryForm.categoryId}
							<div
								class="rounded-2xl border-2 border-dashed border-amber-200/50 bg-white/40 py-12 text-center text-slate-500"
							>
								<Layers class="mx-auto mb-3 h-12 w-12 text-amber-200" />
								<p class="text-base font-medium">Select a category above</p>
								<p class="text-sm text-slate-400">Sub-categories will appear here</p>
							</div>
						{:else}
							<DataTable
								data={subCategories.filter(
									(sc) => sc.CategoryID === Number(subCategoryForm.categoryId)
								)}
								columns={[
									{ header: 'Name', accessorKey: 'Name' },
									{ header: 'Actions', accessorKey: 'id', class: 'text-right' }
								]}
								{loading}
								onRowClick={(subCategory) => viewDetails(subCategory, 'sub-categories')}
							>
								{#snippet children(subCategory)}
									<TableCell class="font-medium text-slate-800">{subCategory.Name}</TableCell>
									<TableCell class="text-right">
										<div class="flex items-center justify-end gap-1">
											<Button
												size="sm"
												variant="ghost"
												class="h-8 text-amber-700 hover:bg-amber-100 hover:text-amber-800"
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
												class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
												onclick={(event) => {
													event.stopPropagation();
													deleteSubCategory(subCategory);
												}}
											>
												Delete
											</Button>
										</div>
									</TableCell>
								{/snippet}
							</DataTable>
						{/if}
					</div>

					<!-- FORM SECTION -->
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-amber-50 to-orange-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
						data-animate="fade-up"
						style="animation-delay:180ms"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800"
								>{editingSubCategory ? 'Update sub-category' : 'Create sub-category'}</CardTitle
							>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input
								class="w-full rounded-xl border border-amber-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400"
								placeholder="Name"
								bind:value={subCategoryForm.name}
							/>
							<select
								class="w-full rounded-xl border border-amber-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-amber-400"
								bind:value={subCategoryForm.categoryId}
							>
								<option value="">Select category</option>
								{#each categories as category}
									<option value={category.ID}>{category.Name}</option>
								{/each}
							</select>
							<div class="flex flex-col gap-3 pr-2 pt-1 sm:flex-row">
								<Button
									class="w-full rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-amber-600 hover:to-orange-700 hover:shadow-xl sm:w-1/2"
									onclick={saveSubCategory}
								>
									{editingSubCategory ? 'Update' : 'Create'}
								</Button>
								<Button
									class="w-full rounded-xl border border-amber-200 py-2.5 text-amber-700 transition hover:bg-amber-50 sm:w-1/2"
									onclick={resetSubCategoryForm}
								>
									Reset
								</Button>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else if activeTab === 'suppliers'}
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table Section Wrapper -->
					<div class="flex flex-col gap-6">
						<!-- Header with Search -->
						<div
							class="flex flex-col gap-4 rounded-2xl border border-violet-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
						>
							<div class="space-y-1">
								<h2 class="text-lg font-semibold text-slate-800">Suppliers</h2>
								<p class="text-sm text-slate-500">Strategic partners powering replenishment</p>
							</div>
							<div class="flex flex-wrap items-center gap-2">
								<Input
									class="w-full rounded-xl border-violet-200 bg-white/80 px-3 text-sm focus:ring-2 focus:ring-violet-400 sm:w-48"
									placeholder="Search by name..."
									bind:value={supplierSearchTerm}
								/>
								<Button
									class="rounded-xl bg-violet-500 px-4 py-2 text-white hover:bg-violet-600"
									onclick={handleSupplierSearch}>Search</Button
								>
								<Button
									variant="ghost"
									class="rounded-xl px-4 py-2 text-violet-600 hover:bg-violet-50"
									onclick={clearSupplierSearch}>Clear</Button
								>
							</div>
						</div>

						<DataTable
							data={suppliers}
							columns={[
								{ header: 'Name', accessorKey: 'Name' },
								{ header: 'Contact', accessorKey: 'ContactPerson' },
								{ header: 'Actions', accessorKey: 'id', class: 'text-right' }
							]}
							{loading}
							onRowClick={(supplier) => viewDetails(supplier, 'suppliers')}
						>
							{#snippet children(supplier)}
								<TableCell class="font-medium text-slate-800">{supplier.Name}</TableCell>
								<TableCell>
									<p class="text-sm font-medium text-slate-700">{supplier.ContactPerson ?? '—'}</p>
									<p class="text-xs text-slate-500">{supplier.Email ?? supplier.Phone ?? ''}</p>
								</TableCell>
								<TableCell class="text-right">
									<div class="flex items-center justify-end gap-1">
										<Button
											size="sm"
											variant="ghost"
											class="h-8 text-violet-600 hover:bg-violet-50 hover:text-violet-700"
											onclick={(event) => {
												event.stopPropagation();
												editingSupplier = supplier;
												Object.assign(supplierForm, {
													name: supplier.Name,
													contactPerson: supplier.ContactPerson ?? '',
													email: supplier.Email ?? '',
													phone: supplier.Phone ?? '',
													address: supplier.Address ?? ''
												});
											}}
										>
											Edit
										</Button>
										<Button
											size="sm"
											variant="ghost"
											class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
											onclick={(event) => {
												event.stopPropagation();
												deleteSupplier(supplier);
											}}
										>
											Delete
										</Button>
									</div>
								</TableCell>
							{/snippet}
						</DataTable>
					</div>

					<!-- Form -->
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
						data-animate="fade-up"
						style="animation-delay:180ms"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800"
								>{editingSupplier ? 'Update supplier' : 'Create supplier'}</CardTitle
							>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input
								class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
								placeholder="Name"
								bind:value={supplierForm.name}
							/>
							<Input
								class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
								placeholder="Contact person"
								bind:value={supplierForm.contactPerson}
							/>
							<Input
								class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
								placeholder="Email"
								bind:value={supplierForm.email}
							/>
							<Input
								class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
								placeholder="Phone"
								bind:value={supplierForm.phone}
							/>
							<Input
								class="w-full rounded-xl border border-violet-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-400"
								placeholder="Address"
								bind:value={supplierForm.address}
							/>
							<div class="flex flex-col gap-3 pr-2 pt-1 sm:flex-row">
								<Button
									class="w-full rounded-xl bg-gradient-to-r from-violet-500 to-purple-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-violet-600 hover:to-purple-700 hover:shadow-xl sm:w-1/2"
									onclick={saveSupplier}>{editingSupplier ? 'Update' : 'Create'}</Button
								>
								<Button
									class="w-full rounded-xl border border-violet-200 py-2.5 text-violet-700 transition hover:bg-violet-50 sm:w-1/2"
									onclick={resetSupplierForm}>Reset</Button
								>
							</div>
						</CardContent>
					</Card>
				</div>
			{:else}
				<!-- LOCATIONS -->
				<div class="grid gap-8 lg:grid-cols-[2fr,1fr]">
					<!-- Table Section Wrapper -->
					<div class="flex flex-col gap-6">
						<!-- Header -->
						<div
							class="flex flex-col gap-4 rounded-2xl border border-teal-100 bg-white/50 p-4 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between"
						>
							<div class="space-y-1">
								<h2 class="text-lg font-semibold text-slate-800">Locations</h2>
								<p class="text-sm text-slate-500">Fulfilment nodes and stores</p>
							</div>
						</div>

						<DataTable
							data={locations}
							columns={[
								{ header: 'Name', accessorKey: 'Name' },
								{ header: 'Address', accessorKey: 'Address' },
								{ header: 'Actions', accessorKey: 'id', class: 'text-right' }
							]}
							{loading}
							onRowClick={(location) => viewDetails(location, 'locations')}
						>
							{#snippet children(location)}
								<TableCell class="font-medium text-slate-800">{location.Name}</TableCell>
								<TableCell class="text-slate-600">{location.Address ?? '—'}</TableCell>
								<TableCell class="text-right">
									<div class="flex items-center justify-end gap-1">
										<Button
											size="sm"
											variant="ghost"
											class="h-8 text-teal-600 hover:bg-teal-50 hover:text-teal-700"
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
											class="h-8 text-rose-600 hover:bg-rose-50 hover:text-rose-700"
											onclick={(event) => {
												event.stopPropagation();
												deleteLocation(location);
											}}
										>
											Delete
										</Button>
									</div>
								</TableCell>
							{/snippet}
						</DataTable>
					</div>

					<!-- Form -->
					<Card
						class="overflow-hidden rounded-2xl border-0 bg-gradient-to-br from-cyan-50 to-teal-100 shadow-lg transition-all duration-300 hover:scale-[1.01] hover:shadow-xl"
						data-animate="fade-up"
						style="animation-delay:180ms"
					>
						<CardHeader
							class="space-y-1 border-b border-white/60 bg-white/70 px-6 py-5 backdrop-blur"
						>
							<CardTitle class="text-slate-800"
								>{editingLocation ? 'Update location' : 'Create location'}</CardTitle
							>
						</CardHeader>
						<CardContent class="space-y-4 p-6">
							<Input
								class="w-full rounded-xl border border-teal-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400"
								placeholder="Name"
								bind:value={locationForm.name}
							/>
							<Input
								class="w-full rounded-xl border border-teal-200 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-teal-400"
								placeholder="Address"
								bind:value={locationForm.address}
							/>
							<div class="flex flex-col gap-3 pr-2 pt-1 sm:flex-row">
								<Button
									class="w-full rounded-xl bg-gradient-to-r from-cyan-500 to-teal-600 py-2.5 text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-cyan-600 hover:to-teal-700 hover:shadow-xl sm:w-1/2"
									onclick={saveLocation}>{editingLocation ? 'Update' : 'Create'}</Button
								>
								<Button
									class="w-full rounded-xl border border-teal-200 py-2.5 text-teal-700 transition hover:bg-teal-50 sm:w-1/2"
									onclick={resetLocationForm}>Reset</Button
								>
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
		transition-property:
			color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
			box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}

	/* Hero gradient animation */
	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 16s ease-in-out infinite;
	}

	/* Soft glowing blobs */
	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
			filter: blur(80px);
		}
		50% {
			transform: scale(1.08);
			opacity: 0.6;
			filter: blur(90px);
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 10s ease-in-out infinite;
	}

	/* Card float micro-motion */
	@keyframes cardFloat {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-4px);
		}
	}
	.animate-cardFloat {
		animation: cardFloat 4s ease-in-out infinite;
	}

	/* Fade-up reveal */
	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	.animate-fadeUp {
		animation: fadeUp 500ms var(--ease, cubic-bezier(0.4, 0, 0.2, 1)) forwards;
	}

	/* Pastel scrollbar */
	::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}
	::-webkit-scrollbar-track {
		background: transparent;
	}
	::-webkit-scrollbar-thumb {
		background: rgba(14, 165, 233, 0.25);
		border-radius: 9999px;
	}
	::-webkit-scrollbar-thumb:hover {
		background: rgba(14, 165, 233, 0.35);
	}

	.parallax-hero {
		transform: translateY(0);
		will-change: transform, filter;
		transition:
			transform 0.1s ease-out,
			filter 0.2s ease-out;
	}

	@keyframes gradientShift {
		0% {
			background-position: 0% 50%;
		}
		50% {
			background-position: 100% 50%;
		}
		100% {
			background-position: 0% 50%;
		}
	}
	.animate-gradientShift {
		animation: gradientShift 18s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%,
		100% {
			transform: scale(1);
			opacity: 0.45;
		}
		50% {
			transform: scale(1.08);
			opacity: 0.7;
		}
	}
	.animate-pulseGlow {
		animation: pulseGlow 12s ease-in-out infinite;
	}

	@keyframes cardFloat {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-4px);
		}
	}
	.animate-cardFloat {
		animation: cardFloat 4s ease-in-out infinite;
	}

	@keyframes fadeUp {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
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
