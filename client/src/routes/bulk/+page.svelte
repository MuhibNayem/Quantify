<script lang="ts">
	import { browser } from '$app/environment';
	import { bulkApi, categoriesApi, suppliersApi } from '$lib/api/resources';
	import api from '$lib/api';
	import type {
		BulkImportJob,
		BulkImportValidationResult,
		Category,
		Supplier,
		BulkExportResult
	} from '$lib/types';
	import { toast } from 'svelte-sonner';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import {
		Upload,
		FileDown,
		CheckCircle,
		AlertTriangle,
		Loader,
		Download,
		Search,
		FileText,
		Clock,
		RefreshCw,
		X,
		FileSpreadsheet,
		Package,
		ArrowRight,
		AlertCircle,
		Check
	} from 'lucide-svelte';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { onMount } from 'svelte';

	// --- Types ---
	type JobStatusEventDetail = {
		event: string;
		jobId: number;
		status: string;
		type: string;
		result?: string;
		lastError?: string;
	};

	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	// --- State: Import Wizard ---
	let file = $state<File | null>(null);
	let importJob = $state<BulkImportJob | null>(null);
	let validationResult = $state<BulkImportValidationResult | null>(null);
	let hasShownValidation = $state(false);
	let currentStep = $state<
		'idle' | 'uploading' | 'validating' | 'importing' | 'complete' | 'failed'
	>('idle');
	let isDragging = false;

	$effect(() => {
		if (!auth.hasPermission('bulk.import')) {
			toast.error('Access Denied', {
				description: 'You do not have permission to perform bulk operations.'
			});
			goto('/');
		}
	});

	// --- State: Export ---
	const exportParams = $state({ format: 'csv', category: '', supplier: '' });
	let exporting = $state(false);
	let categories = $state<Category[]>([]);
	let suppliers = $state<Supplier[]>([]);

	// --- State: Job Lookup & History ---
	let jobIdQuery = $state('');
	let lookupJob = $state<BulkImportJob | null>(null);
	let lookupLoading = $state(false);
	let recentJobs = $state<BulkImportJob[]>([]);

	// --- Helpers ---
	const stepIndex = (s: typeof currentStep): number =>
		({
			idle: 0,
			uploading: 1,
			validating: 2,
			importing: 3,
			complete: 4,
			failed: 4
		})[s];

	const parseValidationResult = (raw: unknown): BulkImportValidationResult | null => {
		if (!raw) return null;
		if (typeof raw === 'string') {
			try {
				return JSON.parse(raw) as BulkImportValidationResult;
			} catch (error) {
				console.error('Failed to parse validation result', error);
				return null;
			}
		}
		return raw as BulkImportValidationResult;
	};

	const parseExportResult = (raw: unknown): BulkExportResult | null => {
		if (!raw) return null;
		if (typeof raw === 'string') {
			try {
				return JSON.parse(raw) as BulkExportResult;
			} catch (error) {
				console.error('Failed to parse export result', error);
				return null;
			}
		}
		return raw as BulkExportResult;
	};

	// --- Actions: Import ---
	const handleFileChange = (e: Event) => {
		const target = e.currentTarget as HTMLInputElement;
		file = target.files?.[0] ?? null;
		if (file) {
			importJob = null;
			validationResult = null;
			hasShownValidation = false;
			currentStep = 'idle';
		}
	};

	const downloadTemplate = async () => {
		try {
			const blob = await bulkApi.downloadTemplate();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'product_import_template.csv';
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			toast.success('Template downloaded.');
		} catch (error) {
			toast.error('Failed to download template.');
		}
	};

	const uploadFile = async () => {
		if (!file) {
			toast.warning('Please select a file to upload.');
			return;
		}
		currentStep = 'uploading';
		try {
			const formData = new FormData();
			formData.append('file', file);
			importJob = await bulkApi.uploadImport(formData);
			if (!importJob?.ID) {
				throw new Error('Upload succeeded but job ID missing');
			}
			currentStep = 'validating';
			toast.info('File uploaded. Validation in progress...');
			refreshJobs(); // Refresh list to show new job
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || 'Upload failed';
			toast.error('Upload Failed', { description: errorMessage });
			currentStep = 'failed';
		}
	};

	const resetImportState = () => {
		file = null;
		importJob = null;
		validationResult = null;
		hasShownValidation = false;
		currentStep = 'idle';
	};

	// --- Actions: Export ---
	const exportCatalog = async () => {
		exporting = true;
		try {
			const blob = await bulkApi.exportProducts({
				format: exportParams.format,
				category: exportParams.category || undefined,
				supplier: exportParams.supplier || undefined
			});
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = `catalog.${exportParams.format === 'excel' ? 'xlsx' : 'csv'}`;
			anchor.click();
			URL.revokeObjectURL(url);
			toast.success('Export generated');
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Unable to export';
			toast.error('Failed to Export Catalog', { description: errorMessage });
		} finally {
			exporting = false;
		}
	};

	const downloadExport = async (job: BulkImportJob) => {
		const result = parseExportResult(job.result);
		if (!result?.downloadUrl) {
			toast.error('Download URL not found');
			return;
		}

		try {
			// Backend returns full path with /api/v1/ prefix, but api client already has baseURL set.
			// We need to strip the prefix to avoid double /api/v1/api/v1/
			const cleanUrl = result.downloadUrl.replace(/^\/api\/v1/, '');
			const response = await api.get(cleanUrl, { responseType: 'blob' });
			const blob = response.data as Blob;
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			// Extract filename from URL or default
			const filename = result.downloadUrl.split('/').pop() || `export-${job.ID}.csv`;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			toast.success('Download started');
		} catch (error) {
			console.error('Download failed', error);
			toast.error('Failed to download file');
		}
	};

	// --- Actions: Job Lookup & History ---
	const loadJobStatus = async () => {
		if (!jobIdQuery) {
			toast.warning('Enter a job ID first');
			return;
		}
		lookupLoading = true;
		try {
			lookupJob = await bulkApi.status(jobIdQuery);
			toast.success('Job status loaded');
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Job not found';
			toast.error('Failed to Load Job Status', { description: errorMessage });
			lookupJob = null;
		} finally {
			lookupLoading = false;
		}
	};

	const refreshJobs = async () => {
		try {
			recentJobs = await bulkApi.listJobs();
		} catch (error) {
			console.error('Failed to fetch recent jobs', error);
		}
	};

	const viewJobDetails = (job: BulkImportJob) => {
		lookupJob = job;
	};

	// --- Websocket / Event Handling ---
	const applyJobStatusUpdate = (detail: JobStatusEventDetail) => {
		// Handle Import Wizard updates
		if (importJob && detail.jobId === importJob.ID) {
			importJob = {
				...importJob,
				status: detail.status,
				result: detail.result ?? importJob.result,
				lastError: detail.lastError ?? importJob.lastError
			};

			if (detail.result) {
				validationResult = parseValidationResult(detail.result);
			}

			if (detail.status === 'PENDING_CONFIRMATION') {
				if (validationResult?.validRecords) {
					currentStep = 'importing';
					if (!hasShownValidation) {
						toast.info('Validation complete. Importing automatically...');
						hasShownValidation = true;
					}
				}
			} else if (detail.status === 'PROCESSING') {
				currentStep = 'importing';
			} else if (detail.status === 'COMPLETED') {
				currentStep = 'complete';
				toast.success('Import completed successfully!');
				refreshJobs(); // Refresh list on completion
			} else if (detail.status === 'FAILED') {
				currentStep = 'failed';
				toast.error('Import failed', { description: detail.lastError || importJob?.lastError });
				refreshJobs(); // Refresh list on failure
			}
		}

		// Handle Lookup updates (if watching a specific job)
		if (lookupJob && detail.jobId === lookupJob.ID) {
			lookupJob = {
				...lookupJob,
				status: detail.status,
				result: detail.result ?? lookupJob.result,
				lastError: detail.lastError ?? lookupJob.lastError
			};
		}

		// Refresh list on any update to keep status current
		refreshJobs();
	};

	onMount(async () => {
		if (!browser) return;

		// Load initial data
		try {
			const [cats, sups, jobs] = await Promise.all([
				categoriesApi.list(),
				suppliersApi.list(),
				bulkApi.listJobs()
			]);
			categories = cats;
			suppliers = sups;
			recentJobs = jobs;
		} catch (error) {
			console.error('Failed to load initial data', error);
			toast.error('Failed to load some data');
		}

		const handler = (event: Event) => {
			const custom = event as CustomEvent<JobStatusEventDetail>;
			applyJobStatusUpdate(custom.detail);
		};
		window.addEventListener('bulk-job-status', handler);
		return () => window.removeEventListener('bulk-job-status', handler);
	});
</script>

<!-- Wrapper with animated gradient + hero consistency -->
<div
	class="relative isolate min-h-[100dvh] overflow-hidden bg-gradient-to-br from-sky-50 via-blue-50 to-cyan-100"
>
	<div
		class="animate-pulseGlow absolute -left-24 -top-32 h-[28rem] w-[28rem] rounded-full bg-sky-200/40 blur-3xl"
	></div>
	<div
		class="animate-pulseGlow absolute -bottom-28 -right-24 h-[24rem] w-[24rem] rounded-full bg-cyan-200/30 blur-3xl delay-700"
	></div>

	<div class="container relative mx-auto max-w-6xl px-4 pb-20 pt-16">
		<header class="mb-10 text-center sm:text-left">
			<h1
				class="mb-2 bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-4xl font-extrabold text-transparent"
			>
				Bulk Operations
			</h1>
			<p class="mx-auto max-w-2xl text-sm text-slate-600 sm:mx-0 sm:text-base">
				Manage your catalog efficiently with bulk imports, exports, and job tracking.
			</p>
		</header>

		<Tabs value="import" class="w-full space-y-8">
			<TabsList
				class="grid w-full grid-cols-3 border border-sky-100 bg-white/60 backdrop-blur-sm lg:w-[400px]"
			>
				<TabsTrigger value="import">Import</TabsTrigger>
				<TabsTrigger value="export">Export</TabsTrigger>
				<TabsTrigger value="status">Status</TabsTrigger>
			</TabsList>

			<!-- IMPORT TAB -->
			<TabsContent value="import" class="space-y-8 outline-none">
				<!-- Step Indicator -->
				<ol
					class="flex items-center justify-center overflow-x-auto rounded-2xl border border-sky-100 bg-white/70 px-4 py-3 shadow-sm backdrop-blur-sm sm:justify-start"
				>
					{#each ['Download Template', 'Upload', 'Validate', 'Import', 'Done'] as label, i}
						<li class="relative flex items-center">
							<div
								class="mr-2 grid size-8 place-items-center rounded-full border text-xs font-semibold shadow-sm transition-all duration-300"
								class:bg-gradient-to-r={i <= stepIndex(currentStep)}
								class:from-sky-500={i <= stepIndex(currentStep)}
								class:to-blue-600={i <= stepIndex(currentStep)}
								class:text-white={i <= stepIndex(currentStep)}
								class:border-sky-200={i > stepIndex(currentStep)}
								class:text-slate-500={i > stepIndex(currentStep)}
							>
								{#if i < stepIndex(currentStep)}
									<CheckCircle class="size-4" />
								{:else}
									<span>{i + 1}</span>
								{/if}
							</div>
							<span class="mr-3 whitespace-nowrap text-xs text-slate-700 sm:text-sm">{label}</span>
							{#if i < 5}<div class="h-[2px] w-8 bg-slate-200/70"></div>{/if}
						</li>
					{/each}
				</ol>

				<div class="grid gap-8 lg:grid-cols-2">
					<!-- Step 1 -->
					<Card
						class="rounded-2xl bg-gradient-to-br from-sky-50 to-blue-100 shadow-md transition-all duration-300 hover:shadow-xl"
					>
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<FileDown class="h-5 w-5 text-sky-600" /> Step 1: Download & Prepare
							</CardTitle>
							<CardDescription class="text-slate-600"
								>Use the correct CSV format for seamless import.</CardDescription
							>
						</CardHeader>
						<CardContent class="p-6">
							<Button
								onclick={downloadTemplate}
								class="w-full rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-2.5 text-white shadow-md hover:from-sky-600 hover:to-blue-700 hover:shadow-lg sm:w-auto"
								>Download Template</Button
							>
						</CardContent>
					</Card>

					<!-- Step 2 -->
					<Card
						class="rounded-2xl bg-gradient-to-br from-blue-50 to-sky-100 shadow-md transition-all duration-300 hover:shadow-xl"
					>
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<Upload class="h-5 w-5 text-sky-600" /> Step 2: Upload File
							</CardTitle>
							<CardDescription class="text-slate-600"
								>Validate your CSV before importing products.</CardDescription
							>
						</CardHeader>
						<CardContent class="flex flex-col items-center gap-4 p-6 sm:flex-row">
							<Input
								type="file"
								accept=".csv"
								onchange={handleFileChange}
								disabled={currentStep !== 'idle'}
								class="rounded-xl border border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400"
							/>
							<Button
								onclick={uploadFile}
								disabled={!file || currentStep !== 'idle'}
								class="flex w-full items-center justify-center rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-6 py-2.5 text-white shadow-md hover:shadow-lg sm:w-auto"
							>
								{#if currentStep === 'uploading'}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
								Upload & Validate
							</Button>
						</CardContent>
					</Card>
				</div>

				<!-- Step 3: Validation & Progress -->
				{#if ['validating', 'importing', 'complete', 'failed'].includes(currentStep)}
					<Card
						class="rounded-2xl bg-gradient-to-br from-white/90 to-sky-50 shadow-md transition-all duration-300 hover:shadow-xl"
					>
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<CheckCircle class="h-5 w-5 text-sky-600" /> Step 3: Review & Confirm
							</CardTitle>
							<CardDescription class="text-slate-600"
								>Ensure validation passes before final import.</CardDescription
							>
						</CardHeader>
						<CardContent class="space-y-6 p-6">
							{#if currentStep === 'validating'}
								<div class="flex items-center gap-2 text-sky-700">
									<Loader class="h-4 w-4 animate-spin" /> Validating file...
								</div>
							{/if}

							{#if validationResult}
								<div class="grid gap-4 sm:grid-cols-3">
									<div class="rounded-2xl border border-sky-100 bg-sky-50 p-4 text-center">
										<p class="text-xs text-slate-500">Valid</p>
										<p class="text-xl font-bold text-sky-700">{validationResult.validRecords}</p>
									</div>
									<div class="rounded-2xl border border-rose-100 bg-rose-50 p-4 text-center">
										<p class="text-xs text-slate-500">Invalid</p>
										<p class="text-xl font-bold text-rose-700">{validationResult.invalidRecords}</p>
									</div>
									<div class="rounded-2xl border border-amber-100 bg-amber-50 p-4 text-center">
										<p class="text-xs text-slate-500">Total</p>
										<p class="text-xl font-bold text-amber-700">{validationResult.totalRecords}</p>
									</div>
								</div>

								{#if validationResult.newEntities}
									<div class="grid grid-cols-3 gap-4">
										<div class="rounded-xl border border-indigo-100 bg-indigo-50 p-3 text-center">
											<p class="text-xs font-medium text-indigo-600">New Categories</p>
											<p class="text-lg font-bold text-indigo-800">
												{Object.keys(validationResult.newEntities.categories || {}).length}
											</p>
										</div>
										<div class="rounded-xl border border-fuchsia-100 bg-fuchsia-50 p-3 text-center">
											<p class="text-xs font-medium text-fuchsia-600">New Suppliers</p>
											<p class="text-lg font-bold text-fuchsia-800">
												{Object.keys(validationResult.newEntities.suppliers || {}).length}
											</p>
										</div>
										<div class="rounded-xl border border-emerald-100 bg-emerald-50 p-3 text-center">
											<p class="text-xs font-medium text-emerald-600">New Locations</p>
											<p class="text-lg font-bold text-emerald-800">
												{Object.keys(validationResult.newEntities.locations || {}).length}
											</p>
										</div>
									</div>
								{/if}

								{#if validationResult.errors?.length}
									<div
										class="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800"
									>
										<ul class="list-inside list-disc space-y-1">
											{#each validationResult.errors as e}<li>{e}</li>{/each}
										</ul>
									</div>
								{/if}

								{#if validationResult.validProducts?.length}
									<div class="max-h-64 overflow-y-auto rounded-xl border bg-white/70">
										<Table>
											<TableHeader
												class="bg-gradient-to-r from-sky-100/70 to-blue-100/70 backdrop-blur"
											>
												<TableRow
													><TableHead>SKU</TableHead><TableHead>Name</TableHead><TableHead
														>Category</TableHead
													><TableHead>Supplier</TableHead><TableHead class="text-right"
														>Price</TableHead
													></TableRow
												>
											</TableHeader>
											<TableBody
												>{#each validationResult.validProducts as p}<TableRow
														class="hover:bg-sky-50"
														><TableCell>{p.SKU}</TableCell><TableCell>{p.Name}</TableCell><TableCell
															>{p.CategoryName}</TableCell
														><TableCell>{p.SupplierName}</TableCell><TableCell class="text-right"
															>{p.SellingPrice}</TableCell
														></TableRow
													>{/each}</TableBody
											>
										</Table>
									</div>
								{/if}
							{/if}

							{#if currentStep === 'importing'}
								<p class="flex items-center gap-2 text-sky-700">
									<Loader class="h-4 w-4 animate-spin" /> Importing products...
								</p>
							{/if}

							{#if currentStep === 'complete'}
								<div class="rounded-2xl border border-green-200 bg-green-50 p-6 text-center">
									<CheckCircle class="mx-auto h-12 w-12 text-green-600" />
									<h3 class="mt-2 text-lg font-semibold text-green-700">Import Complete</h3>
									<p class="mt-1 text-slate-600">
										{importJob?.message || 'Products imported successfully.'}
									</p>
									<Button
										onclick={resetImportState}
										class="mt-4 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-600 px-5 py-2.5 text-white hover:brightness-110"
										>New Import</Button
									>
								</div>
							{/if}

							{#if currentStep === 'failed'}
								<div class="rounded-2xl border border-rose-200 bg-rose-50 p-6 text-center">
									<AlertTriangle class="mx-auto h-12 w-12 text-rose-600" />
									<h3 class="mt-2 text-lg font-semibold text-rose-700">Import Failed</h3>
									<p class="mt-1 text-slate-600">
										{importJob?.lastError || 'Unknown error occurred.'}
									</p>
									<Button
										onclick={resetImportState}
										class="mt-4 rounded-xl bg-gradient-to-r from-rose-500 to-orange-500 px-5 py-2.5 text-white hover:brightness-110"
										>Try Again</Button
									>
								</div>
							{/if}
						</CardContent>
					</Card>
				{/if}
			</TabsContent>

			<!-- EXPORT TAB -->
			<TabsContent value="export" class="outline-none">
				<Card
					class="rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-purple-100 shadow-lg transition-all duration-300 hover:shadow-xl"
				>
					<CardHeader
						class="rounded-t-2xl border-b border-white/60 bg-white/80 px-6 py-5 backdrop-blur"
					>
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<Download class="h-5 w-5 text-violet-600" /> Catalog Export
						</CardTitle>
						<CardDescription class="text-slate-600"
							>Filter and download your product catalog.</CardDescription
						>
					</CardHeader>
					<CardContent class="space-y-6 p-6">
						<div class="grid gap-4 sm:grid-cols-3">
							<div class="space-y-2">
								<label class="text-sm font-medium text-slate-700">Format</label>
								<select
									class="w-full rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400"
									bind:value={exportParams.format}
								>
									<option value="csv">CSV</option>
									<option value="excel">Excel</option>
								</select>
							</div>
							<div class="space-y-2">
								<label class="text-sm font-medium text-slate-700">Category (Optional)</label>
								<select
									class="w-full rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400"
									bind:value={exportParams.category}
								>
									<option value="">All Categories</option>
									{#each categories as cat}
										<option value={cat.ID.toString()}>{cat.Name}</option>
									{/each}
								</select>
							</div>
							<div class="space-y-2">
								<label class="text-sm font-medium text-slate-700">Supplier (Optional)</label>
								<select
									class="w-full rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400"
									bind:value={exportParams.supplier}
								>
									<option value="">All Suppliers</option>
									{#each suppliers as sup}
										<option value={sup.ID.toString()}>{sup.Name}</option>
									{/each}
								</select>
							</div>
						</div>
						<Button
							class="flex w-full items-center justify-center rounded-xl bg-gradient-to-r from-violet-500 to-purple-600 font-semibold text-white shadow-md transition-all hover:scale-[1.02] hover:from-violet-600 hover:to-purple-700 hover:shadow-lg"
							onclick={exportCatalog}
							disabled={exporting}
						>
							{#if exporting}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
							Generate Export
						</Button>
					</CardContent>
				</Card>
			</TabsContent>

			<!-- STATUS TAB -->
			<TabsContent value="status" class="space-y-6 outline-none">
				<!-- Header / Search -->
				<div
					class="flex flex-col items-center justify-between gap-4 rounded-2xl border border-sky-100 bg-white/60 p-4 backdrop-blur-sm sm:flex-row"
				>
					<div class="relative w-full sm:w-72">
						<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
						<Input
							placeholder="Search by Job ID..."
							bind:value={jobIdQuery}
							onkeydown={(e) => e.key === 'Enter' && loadJobStatus()}
							class="rounded-xl border-slate-200 bg-white/80 pl-10 focus:ring-sky-200"
						/>
					</div>
					<Button
						variant="outline"
						size="sm"
						onclick={refreshJobs}
						class="gap-2 rounded-xl border-slate-200 text-slate-600 hover:bg-sky-50 hover:text-sky-700"
					>
						<RefreshCw class="h-4 w-4" /> Refresh
					</Button>
				</div>

				<div class="grid gap-6 lg:grid-cols-3">
					<!-- JOB LIST -->
					<div class="space-y-4 lg:col-span-2">
						{#if lookupLoading}
							<div class="p-12 text-center">
								<Loader class="mx-auto mb-3 h-8 w-8 animate-spin text-sky-500" />
								<p class="font-medium text-slate-500">Loading jobs...</p>
							</div>
						{:else if recentJobs.length === 0}
							<div
								class="rounded-3xl border border-dashed border-slate-300 bg-white/40 p-12 text-center"
							>
								<div
									class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-50"
								>
									<Clock class="h-8 w-8 text-slate-300" />
								</div>
								<h3 class="text-lg font-medium text-slate-700">No History</h3>
								<p class="text-slate-500">Recent import and export jobs will appear here.</p>
							</div>
						{:else}
							<div class="space-y-3">
								{#each recentJobs as job}
									<!-- Job Card -->
									<div
										role="button"
										tabindex="0"
										onclick={() => viewJobDetails(job)}
										onkeydown={(e) => e.key === 'Enter' && viewJobDetails(job)}
										class="group relative cursor-pointer overflow-hidden rounded-2xl border border-slate-100 bg-white/80 p-4 shadow-sm transition-all duration-300 hover:border-sky-200 hover:bg-white hover:shadow-md"
									>
										<!-- Hover Gradient -->
										<div
											class="pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-transparent to-sky-50/50 opacity-0 transition-opacity group-hover:opacity-100"
										></div>

										<div class="relative flex items-center justify-between gap-4">
											<div class="flex items-center gap-4">
												<!-- Icon Box -->
												<div
													class="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-xl
													{job.type === 'BULK_IMPORT' ? 'bg-blue-50 text-blue-600' : 'bg-violet-50 text-violet-600'}"
												>
													{#if job.type === 'BULK_IMPORT'}
														<Package class="h-6 w-6" />
													{:else}
														<FileSpreadsheet class="h-6 w-6" />
													{/if}

													<!-- Status Indicator Dot -->
													<div
														class="absolute -right-1 -top-1 h-3 w-3 rounded-full border-2 border-white
														{job.status === 'COMPLETED'
															? 'bg-emerald-500'
															: job.status === 'FAILED'
																? 'bg-rose-500'
																: job.status === 'PROCESSING'
																	? 'animate-pulse bg-sky-500'
																	: 'bg-amber-500'}"
													></div>
												</div>

												<div>
													<div class="mb-0.5 flex items-center gap-2">
														<span class="font-bold text-slate-700">#{job.ID}</span>
														<span class="text-xs font-medium text-slate-400">•</span>
														<span class="text-sm font-medium capitalize text-slate-600"
															>{(job.type || '').toLowerCase().replace('_', ' ')}</span
														>
													</div>
													<div class="flex items-center gap-2 text-xs text-slate-500">
														<Clock class="h-3 w-3" />
														<span
															>{new Date(job.CreatedAt || '').toLocaleDateString()}
															{new Date(job.CreatedAt || '').toLocaleTimeString([], {
																hour: '2-digit',
																minute: '2-digit'
															})}</span
														>
													</div>
												</div>
											</div>

											<div class="flex items-center gap-4">
												<div class="hidden text-right sm:block">
													<div
														class="mb-1 text-xs font-semibold uppercase tracking-wider
														{job.status === 'COMPLETED'
															? 'text-emerald-600'
															: job.status === 'FAILED'
																? 'text-rose-600'
																: job.status === 'PROCESSING'
																	? 'text-sky-600'
																	: 'text-amber-600'}"
													>
														{(job.status || '').replace('_', ' ')}
													</div>
													{#if job.status === 'FAILED'}
														<p class="max-w-[150px] truncate text-[10px] text-rose-400">
															Check details
														</p>
													{:else if job.status === 'COMPLETED'}
														<p class="text-[10px] text-slate-400">Success</p>
													{/if}
												</div>
												<div class="text-slate-300 transition-colors group-hover:text-sky-400">
													<ArrowRight class="h-5 w-5" />
												</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<!-- DETAILS PANEL -->
					<div class="lg:col-span-1">
						{#if lookupJob}
							{@const summary = parseValidationResult(lookupJob.result)}
							<div
								class="animate-in slide-in-from-right-4 sticky top-24 overflow-hidden rounded-3xl border border-sky-100 bg-white/80 shadow-xl backdrop-blur-md duration-300"
							>
								<!-- Header -->
								<div
									class="flex items-start justify-between border-b border-slate-100 bg-gradient-to-r from-slate-50 to-slate-100 p-5"
								>
									<div>
										<p class="mb-1 text-xs font-bold uppercase tracking-widest text-slate-400">
											Job Details
										</p>
										<h2 class="text-2xl font-extrabold text-slate-800">#{lookupJob.ID}</h2>
									</div>
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8 rounded-full hover:bg-white hover:text-rose-500"
										onclick={() => (lookupJob = null)}
									>
										<X class="h-4 w-4" />
									</Button>
								</div>

								<div class="space-y-6 p-6">
									<!-- Status Badge Large -->
									<div
										class="flex items-center gap-3 rounded-2xl p-3
										{lookupJob.status === 'COMPLETED'
											? 'border border-emerald-100 bg-emerald-50 text-emerald-800'
											: lookupJob.status === 'FAILED'
												? 'border border-rose-100 bg-rose-50 text-rose-800'
												: 'border border-amber-100 bg-amber-50 text-amber-800'}"
									>
										{#if lookupJob.status === 'COMPLETED'}
											<CheckCircle class="h-5 w-5 shrink-0" />
										{:else if lookupJob.status === 'FAILED'}
											<AlertTriangle class="h-5 w-5 shrink-0" />
										{:else}
											<Loader class="h-5 w-5 shrink-0 animate-spin" />
										{/if}
										<div class="flex-1">
											<p class="text-sm font-bold">{(lookupJob.status || '').replace('_', ' ')}</p>
											{#if lookupJob.lastError}
												<p class="mt-0.5 line-clamp-2 text-xs opacity-90">{lookupJob.lastError}</p>
											{/if}
										</div>
									</div>

									<!-- Actions -->
									{#if lookupJob.type === 'BULK_EXPORT' && lookupJob.status === 'COMPLETED'}
										<Button
											class="w-full rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 text-white shadow-lg shadow-violet-200 hover:from-violet-700 hover:to-indigo-700"
											onclick={() => lookupJob && downloadExport(lookupJob)}
										>
											<Download class="mr-2 h-4 w-4" /> Download Export
										</Button>
									{/if}

									<!-- Stats Grid -->
									{#if summary}
										<div class="grid grid-cols-2 gap-3">
											<div
												class="rounded-2xl border border-slate-100 bg-white p-3 text-center shadow-sm"
											>
												<p class="mb-1 text-xs font-medium text-slate-400">Processed</p>
												<p class="text-2xl font-bold text-slate-700">{summary.totalRecords || 0}</p>
											</div>
											<div
												class="rounded-2xl border border-slate-100 bg-white p-3 text-center shadow-sm"
											>
												<p class="mb-1 text-xs font-medium text-slate-400">Success Rate</p>
												<p class="text-2xl font-bold text-emerald-600">
													{summary.totalRecords
														? Math.round((summary.validRecords / summary.totalRecords) * 100)
														: 0}%
												</p>
											</div>
										</div>

										<!-- Breakdown -->
										<div class="space-y-3">
											<h4 class="text-xs font-bold uppercase tracking-widest text-slate-400">
												Breakdown
											</h4>

											<div
												class="flex items-center justify-between rounded-xl border border-emerald-100 bg-emerald-50/50 p-3"
											>
												<div class="flex items-center gap-2">
													<Check class="h-4 w-4 text-emerald-600" />
													<span class="text-sm font-medium text-emerald-900">Valid Records</span>
												</div>
												<span class="font-bold text-emerald-700">{summary.validRecords}</span>
											</div>

											<div
												class="flex items-center justify-between rounded-xl border border-rose-100 bg-rose-50/50 p-3"
											>
												<div class="flex items-center gap-2">
													<AlertCircle class="h-4 w-4 text-rose-600" />
													<span class="text-sm font-medium text-rose-900">Invalid Records</span>
												</div>
												<span class="font-bold text-rose-700">{summary.invalidRecords}</span>
											</div>
										</div>

										<!-- New Entities -->
										{#if summary.newEntities && (Object.keys(summary.newEntities.categories || {}).length > 0 || Object.keys(summary.newEntities.suppliers || {}).length > 0)}
											<div class="space-y-3">
												<h4 class="text-xs font-bold uppercase tracking-widest text-slate-400">
													Changes
												</h4>
												<div class="flex flex-wrap gap-2">
													{#if Object.keys(summary.newEntities.categories || {}).length > 0}
														<span
															class="inline-flex items-center rounded-lg border border-indigo-100 bg-indigo-50 px-2.5 py-1 text-xs font-medium text-indigo-700"
														>
															+{Object.keys(summary.newEntities.categories || {}).length} Categories
														</span>
													{/if}
													{#if Object.keys(summary.newEntities.suppliers || {}).length > 0}
														<span
															class="inline-flex items-center rounded-lg border border-fuchsia-100 bg-fuchsia-50 px-2.5 py-1 text-xs font-medium text-fuchsia-700"
														>
															+{Object.keys(summary.newEntities.suppliers || {}).length} Suppliers
														</span>
													{/if}
													{#if Object.keys(summary.newEntities.locations || {}).length > 0}
														<span
															class="inline-flex items-center rounded-lg border border-cyan-100 bg-cyan-50 px-2.5 py-1 text-xs font-medium text-cyan-700"
														>
															+{Object.keys(summary.newEntities.locations || {}).length} Locations
														</span>
													{/if}
												</div>
											</div>
										{/if}
									{/if}
								</div>
							</div>
						{:else}
							<!-- Empty State for Details -->
							<div
								class="hidden h-full min-h-[400px] flex-col items-center justify-center rounded-3xl border border-dashed border-slate-200 bg-white/30 p-8 text-center lg:flex"
							>
								<div
									class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-100 text-slate-300"
								>
									<ArrowRight class="h-6 w-6" />
								</div>
								<p class="font-medium text-slate-500">Select a job to view details</p>
							</div>
						{/if}
					</div>
				</div>
			</TabsContent>
		</Tabs>
	</div>
</div>

<style lang="postcss">
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
</style>
