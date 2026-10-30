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
		Clock,
		RefreshCw,
		X,
		FileSpreadsheet,
		Package,
		ArrowRight,
		AlertCircle,
		Check
	} from 'lucide-svelte';
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
	import { t } from '$lib/i18n';
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
			toast.error($t('common.access_denied'), {
				description: $t('common.no_permission_settings')
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

	onMount(() => {
		if (!browser) return;

		// Load initial data
		(async () => {
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
		})();

		// Initial load of job status if a query ID exists
		if (jobIdQuery) {
			loadJobStatus();
		}

		// Poll for updates if there's an active job or export is in progress
		const interval = setInterval(() => {
			if (currentStep === 'importing' || currentStep === 'validating' || exporting) {
				loadJobStatus();
			}
		}, 5000);

		const handler = (event: Event) => {
			const custom = event as CustomEvent<JobStatusEventDetail>;
			applyJobStatusUpdate(custom.detail);
		};
		window.addEventListener('bulk-job-status', handler);
		return () => window.removeEventListener('bulk-job-status', handler);
	});
</script>

<div class="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
	<!-- Page header -->
	<header class="page-header" data-animate="fade-up">
		<div>
			<p class="overline">{$t('bulk.subtitle')}</p>
			<h1 class="page-title mt-1">{$t('bulk.title')}</h1>
		</div>
	</header>

	<Tabs value="import" class="mt-6 w-full space-y-4">
		<TabsList
			class="inline-flex h-11 w-full items-center justify-start gap-1 rounded-lg border border-border/70 bg-muted/60 p-1 sm:w-auto"
		>
			<TabsTrigger
				value="import"
				class="flex-1 rounded-md px-4 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
				>{$t('bulk.tabs.import')}</TabsTrigger
			>
			<TabsTrigger
				value="export"
				class="flex-1 rounded-md px-4 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
				>{$t('bulk.tabs.export')}</TabsTrigger
			>
			<TabsTrigger
				value="status"
				class="flex-1 rounded-md px-4 text-sm font-medium text-muted-foreground transition-colors data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-xs sm:flex-none"
				>{$t('bulk.tabs.status')}</TabsTrigger
			>
		</TabsList>

		<!-- IMPORT TAB -->
		<TabsContent value="import" class="space-y-4 outline-none">
			<!-- Step Indicator -->
			<ol
				class="panel custom-scrollbar flex items-center justify-center gap-1 overflow-x-auto px-4 py-3 sm:justify-start"
			>
				{#each [$t('bulk.steps.download'), $t('bulk.steps.upload'), $t('bulk.steps.validate'), $t('bulk.steps.import'), $t('bulk.steps.done')] as label, i}
					<li class="relative flex shrink-0 items-center">
						<div
							class="num mr-2 grid size-7 shrink-0 place-items-center rounded-md border text-xs font-semibold transition-colors duration-300"
							class:bg-primary={i <= stepIndex(currentStep)}
							class:text-primary-foreground={i <= stepIndex(currentStep)}
							class:border-primary={i <= stepIndex(currentStep)}
							class:border-border={i > stepIndex(currentStep)}
							class:text-muted-foreground={i > stepIndex(currentStep)}
						>
							{#if i < stepIndex(currentStep)}
								<CheckCircle class="size-3.5" />
							{:else}
								<span>{i + 1}</span>
							{/if}
						</div>
						<span
							class="mr-3 text-xs whitespace-nowrap transition-colors"
							class:text-foreground={i <= stepIndex(currentStep)}
							class:text-muted-foreground={i > stepIndex(currentStep)}>{label}</span
						>
						{#if i < 4}<div class="mr-1 h-px w-6 bg-border/80"></div>{/if}
					</li>
				{/each}
			</ol>

			<div class="grid gap-4 lg:grid-cols-2">
				<!-- Step 1 -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<FileDown class="h-4 w-4" />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('bulk.steps.step1_title')}</h2>
								<p class="text-sm text-muted-foreground">{$t('bulk.steps.step1_desc')}</p>
							</div>
						</div>
					</div>
					<div class="panel-body">
						<Button onclick={downloadTemplate} class="btn-soothe w-full sm:w-auto">
							<FileDown class="h-4 w-4" />
							{$t('bulk.buttons.download_template')}
						</Button>
					</div>
				</section>

				<!-- Step 2 -->
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<Upload class="h-4 w-4" />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('bulk.steps.step2_title')}</h2>
								<p class="text-sm text-muted-foreground">{$t('bulk.steps.step2_desc')}</p>
							</div>
						</div>
					</div>
					<div class="panel-body flex flex-col items-stretch gap-3 sm:flex-row sm:items-center">
						<Input
							type="file"
							accept=".csv"
							onchange={handleFileChange}
							disabled={currentStep !== 'idle'}
							class="h-auto w-full py-2 text-sm"
						/>
						<Button
							onclick={uploadFile}
							disabled={!file || currentStep !== 'idle'}
							class="btn-soothe shrink-0"
						>
							{#if currentStep === 'uploading'}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
							{$t('bulk.buttons.upload_validate')}
						</Button>
					</div>
				</section>
			</div>

			<!-- Step 3: Validation & Progress -->
			{#if ['validating', 'importing', 'complete', 'failed'].includes(currentStep)}
				<section class="panel overflow-hidden">
					<div class="panel-header items-start">
						<div class="flex items-start gap-3">
							<span
								class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
							>
								<CheckCircle class="h-4 w-4" />
							</span>
							<div class="space-y-0.5">
								<h2 class="section-title">{$t('bulk.steps.step3_title')}</h2>
								<p class="text-sm text-muted-foreground">{$t('bulk.steps.step3_desc')}</p>
							</div>
						</div>
					</div>
					<div class="panel-body space-y-4">
						{#if currentStep === 'validating'}
							<p class="flex items-center gap-2 text-sm font-medium text-info">
								<Loader class="h-4 w-4 animate-spin" />
								{$t('bulk.status.validating')}
							</p>
						{/if}

						{#if validationResult}
							<div class="grid gap-3 sm:grid-cols-3">
								<div class="rounded-lg bg-success/10 p-4 text-center ring-1 ring-success/20">
									<p class="overline">{$t('bulk.labels.valid')}</p>
									<p class="kpi-num mt-1 text-2xl text-success">{validationResult.validRecords}</p>
								</div>
								<div class="rounded-lg bg-danger/10 p-4 text-center ring-1 ring-danger/20">
									<p class="overline">{$t('bulk.labels.invalid')}</p>
									<p class="kpi-num mt-1 text-2xl text-danger">{validationResult.invalidRecords}</p>
								</div>
								<div class="rounded-lg bg-muted/50 p-4 text-center ring-1 ring-border">
									<p class="overline">{$t('bulk.labels.total')}</p>
									<p class="kpi-num mt-1 text-2xl text-foreground">
										{validationResult.totalRecords}
									</p>
								</div>
							</div>

							{#if validationResult.newEntities}
								<div class="grid grid-cols-3 gap-3">
									<div class="rounded-md bg-muted/40 p-3 text-center ring-1 ring-border/70">
										<p class="text-xs font-medium text-muted-foreground">
											{$t('bulk.labels.new_categories')}
										</p>
										<p class="num-mono mt-0.5 text-lg font-semibold text-foreground">
											{Object.keys(validationResult.newEntities.categories || {}).length}
										</p>
									</div>
									<div class="rounded-md bg-muted/40 p-3 text-center ring-1 ring-border/70">
										<p class="text-xs font-medium text-muted-foreground">
											{$t('bulk.labels.new_suppliers')}
										</p>
										<p class="num-mono mt-0.5 text-lg font-semibold text-foreground">
											{Object.keys(validationResult.newEntities.suppliers || {}).length}
										</p>
									</div>
									<div class="rounded-md bg-muted/40 p-3 text-center ring-1 ring-border/70">
										<p class="text-xs font-medium text-muted-foreground">
											{$t('bulk.labels.new_locations')}
										</p>
										<p class="num-mono mt-0.5 text-lg font-semibold text-foreground">
											{Object.keys(validationResult.newEntities.locations || {}).length}
										</p>
									</div>
								</div>
							{/if}

							{#if validationResult.errors?.length}
								<div class="rounded-md bg-danger/10 p-3 text-sm text-danger ring-1 ring-danger/20">
									<ul class="list-inside list-disc space-y-1">
										{#each validationResult.errors as e}<li>{e}</li>{/each}
									</ul>
								</div>
							{/if}

							{#if validationResult.validProducts?.length}
								<div
									class="custom-scrollbar max-h-64 overflow-y-auto rounded-lg ring-1 ring-border"
								>
									<table class="ledger-table min-w-[36rem]">
										<thead>
											<tr>
												<th>SKU</th>
												<th>Name</th>
												<th>Category</th>
												<th>Supplier</th>
												<th class="text-right">Price</th>
											</tr>
										</thead>
										<tbody>
											{#each validationResult.validProducts as p}
												<tr>
													<td class="num-mono text-xs text-muted-foreground">{p.SKU}</td>
													<td class="text-foreground">{p.Name}</td>
													<td>{p.CategoryName}</td>
													<td>{p.SupplierName}</td>
													<td class="col-num">{p.SellingPrice}</td>
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							{/if}
						{/if}

						{#if currentStep === 'importing'}
							<p class="flex items-center gap-2 text-sm font-medium text-info">
								<Loader class="h-4 w-4 animate-spin" />
								{$t('bulk.status.importing')}
							</p>
						{/if}

						{#if currentStep === 'complete'}
							<div class="rounded-lg bg-success/10 p-6 text-center ring-1 ring-success/20">
								<CheckCircle class="mx-auto h-10 w-10 text-success" />
								<h3 class="mt-2 text-base font-semibold text-success">
									{$t('bulk.status.import_complete')}
								</h3>
								<p class="mt-1 text-sm text-muted-foreground">
									{importJob?.message || $t('bulk.status.success')}
								</p>
								<Button onclick={resetImportState} class="btn-soothe mt-4">
									{$t('bulk.buttons.new_import')}
								</Button>
							</div>
						{/if}

						{#if currentStep === 'failed'}
							<div class="rounded-lg bg-danger/10 p-6 text-center ring-1 ring-danger/20">
								<AlertTriangle class="mx-auto h-10 w-10 text-danger" />
								<h3 class="mt-2 text-base font-semibold text-danger">
									{$t('bulk.status.import_failed')}
								</h3>
								<p class="mt-1 text-sm text-muted-foreground">
									{importJob?.lastError || 'Unknown error occurred.'}
								</p>
								<Button variant="destructive" onclick={resetImportState} class="mt-4">
									{$t('bulk.buttons.try_again')}
								</Button>
							</div>
						{/if}
					</div>
				</section>
			{/if}
		</TabsContent>

		<!-- EXPORT TAB -->
		<TabsContent value="export" class="outline-none">
			<section class="panel overflow-hidden">
				<div class="panel-header items-start">
					<div class="flex items-start gap-3">
						<span
							class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-accent-foreground ring-1 ring-accent-foreground/10"
						>
							<Download class="h-4 w-4" />
						</span>
						<div class="space-y-0.5">
							<h2 class="section-title">{$t('bulk.tabs.export')}</h2>
							<p class="text-sm text-muted-foreground">{$t('bulk.subtitle')}</p>
						</div>
					</div>
				</div>
				<div class="panel-body space-y-4">
					<div class="grid gap-4 sm:grid-cols-3">
						<div class="space-y-1.5">
							<label class="text-sm font-medium text-foreground" for="export-format"
								>{$t('bulk.labels.format')}</label
							>
							<select id="export-format" bind:value={exportParams.format} class="h-11 text-sm">
								<option value="csv">CSV</option>
								<option value="excel">Excel</option>
							</select>
						</div>
						<div class="space-y-1.5">
							<label class="text-sm font-medium text-foreground" for="export-category"
								>{$t('bulk.labels.category')}</label
							>
							<select id="export-category" bind:value={exportParams.category} class="h-11 text-sm">
								<option value="">{$t('bulk.labels.all_categories')}</option>
								{#each categories as cat}
									<option value={cat.ID.toString()}>{cat.Name}</option>
								{/each}
							</select>
						</div>
						<div class="space-y-1.5">
							<label class="text-sm font-medium text-foreground" for="export-supplier"
								>{$t('bulk.labels.supplier')}</label
							>
							<select id="export-supplier" bind:value={exportParams.supplier} class="h-11 text-sm">
								<option value="">{$t('bulk.labels.all_suppliers')}</option>
								{#each suppliers as sup}
									<option value={sup.ID.toString()}>{sup.Name}</option>
								{/each}
							</select>
						</div>
					</div>
					<Button
						class="btn-soothe flex w-full items-center justify-center"
						onclick={exportCatalog}
						disabled={exporting}
					>
						{#if exporting}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
						{$t('bulk.buttons.generate_export')}
					</Button>
				</div>
			</section>
		</TabsContent>

		<!-- STATUS TAB -->
		<TabsContent value="status" class="space-y-4 outline-none">
			<!-- Header / Search -->
			<section class="panel flex flex-col items-center justify-between gap-3 p-4 sm:flex-row">
				<div class="relative w-full sm:w-72">
					<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						placeholder={$t('bulk.labels.search_placeholder')}
						bind:value={jobIdQuery}
						onkeydown={(e) => e.key === 'Enter' && loadJobStatus()}
						class="num-mono h-10 w-full rounded-md pl-9 text-sm"
					/>
				</div>
				<Button variant="outline" size="sm" onclick={refreshJobs} class="h-9 gap-1.5">
					<RefreshCw class="h-4 w-4" />
					{$t('bulk.buttons.refresh')}
				</Button>
			</section>

			<div class="grid gap-4 lg:grid-cols-3">
				<!-- JOB LIST -->
				<div class="space-y-3 lg:col-span-2">
					{#if lookupLoading}
						<div class="panel p-12 text-center">
							<Loader class="mx-auto mb-3 h-7 w-7 animate-spin text-primary" />
							<p class="text-sm font-medium text-muted-foreground">Loading jobs...</p>
						</div>
					{:else if recentJobs.length === 0}
						<div
							class="rounded-lg border border-dashed border-border/70 bg-muted/25 p-12 text-center"
						>
							<div
								class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-lg bg-muted text-muted-foreground/70"
							>
								<Clock class="h-6 w-6" />
							</div>
							<h3 class="text-base font-semibold text-foreground">
								{$t('bulk.labels.no_history')}
							</h3>
							<p class="mt-1 text-sm text-muted-foreground">{$t('bulk.labels.no_history_desc')}</p>
						</div>
					{:else}
						<div class="space-y-2.5">
							{#each recentJobs as job}
								<!-- Job Card -->
								<div
									role="button"
									tabindex="0"
									onclick={() => viewJobDetails(job)}
									onkeydown={(e) => e.key === 'Enter' && viewJobDetails(job)}
									class="glass-hoverable group relative cursor-pointer p-4"
								>
									<div class="relative flex items-center justify-between gap-4">
										<div class="flex min-w-0 items-center gap-3.5">
											<!-- Icon Box -->
											<div
												class="relative flex h-11 w-11 shrink-0 items-center justify-center rounded-md
												{job.type === 'BULK_IMPORT' ? 'bg-primary/10 text-primary' : 'bg-info/10 text-info'}"
											>
												{#if job.type === 'BULK_IMPORT'}
													<Package class="h-5 w-5" />
												{:else}
													<FileSpreadsheet class="h-5 w-5" />
												{/if}

												<!-- Status Indicator Dot -->
												<div
													class="absolute -top-1 -right-1 h-3 w-3 rounded-full border-2 border-card
													{job.status === 'COMPLETED'
														? 'bg-success'
														: job.status === 'FAILED'
															? 'bg-danger'
															: job.status === 'PROCESSING'
																? 'animate-pulse bg-info'
																: 'bg-warning'}"
												></div>
											</div>

											<div class="min-w-0">
												<div class="mb-0.5 flex items-center gap-2">
													<span class="num-mono text-sm font-bold text-foreground">#{job.ID}</span>
													<span class="text-xs text-muted-foreground/60">•</span>
													<span class="text-xs font-medium text-muted-foreground capitalize"
														>{(job.type || '').toLowerCase().replace('_', ' ')}</span
													>
												</div>
												<div class="flex items-center gap-1.5 text-xs text-muted-foreground">
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

										<div class="flex shrink-0 items-center gap-3">
											<div class="hidden text-right sm:block">
												<div
													class="mb-0.5 overline
													{job.status === 'COMPLETED'
														? 'text-success'
														: job.status === 'FAILED'
															? 'text-danger'
															: job.status === 'PROCESSING'
																? 'text-info'
																: 'text-warning'}"
												>
													{(job.status || '').replace('_', ' ')}
												</div>
												{#if job.status === 'FAILED'}
													<p class="max-w-[9rem] truncate text-[10px] text-danger/80">
														Check details
													</p>
												{:else if job.status === 'COMPLETED'}
													<p class="text-[10px] text-muted-foreground">
														{$t('bulk.status.success')}
													</p>
												{/if}
											</div>
											<div
												class="text-muted-foreground/50 transition-colors group-hover:text-primary"
											>
												<ArrowRight class="h-4 w-4" />
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
						<div class="panel sticky top-24 overflow-hidden">
							<!-- Header -->
							<div class="panel-header items-start bg-muted/40">
								<div>
									<p class="mb-0.5 overline">Job Details</p>
									<h2 class="num-mono text-xl font-bold text-foreground">#{lookupJob.ID}</h2>
								</div>
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 rounded-md text-muted-foreground hover:text-danger"
									onclick={() => (lookupJob = null)}
									aria-label="Close job details"
								>
									<X class="h-4 w-4" />
								</Button>
							</div>

							<div class="space-y-5 p-5">
								<!-- Status Badge Large -->
								<div
									class="flex items-center gap-3 rounded-lg p-3
									{lookupJob.status === 'COMPLETED'
										? 'bg-success/10 text-success ring-1 ring-success/20'
										: lookupJob.status === 'FAILED'
											? 'bg-danger/10 text-danger ring-1 ring-danger/20'
											: 'bg-warning/10 text-warning ring-1 ring-warning/20'}"
								>
									{#if lookupJob.status === 'COMPLETED'}
										<CheckCircle class="h-4 w-4 shrink-0" />
									{:else if lookupJob.status === 'FAILED'}
										<AlertTriangle class="h-4 w-4 shrink-0" />
									{:else}
										<Loader class="h-4 w-4 shrink-0 animate-spin" />
									{/if}
									<div class="min-w-0 flex-1">
										<p class="text-sm font-bold">{(lookupJob.status || '').replace('_', ' ')}</p>
										{#if lookupJob.lastError}
											<p class="mt-0.5 line-clamp-2 text-xs opacity-90">{lookupJob.lastError}</p>
										{/if}
									</div>
								</div>

								<!-- Actions -->
								{#if lookupJob.type === 'BULK_EXPORT' && lookupJob.status === 'COMPLETED'}
									<Button
										class="btn-soothe w-full"
										onclick={() => lookupJob && downloadExport(lookupJob)}
									>
										<Download class="mr-2 h-4 w-4" />
										{$t('bulk.buttons.download_export')}
									</Button>
								{/if}

								<!-- Stats Grid -->
								{#if summary}
									<div class="grid grid-cols-2 gap-2.5">
										<div class="rounded-lg bg-muted/40 p-3 text-center ring-1 ring-border/70">
											<p class="mb-0.5 overline">{$t('bulk.labels.processed')}</p>
											<p class="kpi-num text-xl text-foreground">{summary.totalRecords || 0}</p>
										</div>
										<div class="rounded-lg bg-muted/40 p-3 text-center ring-1 ring-border/70">
											<p class="mb-0.5 overline">Success Rate</p>
											<p class="kpi-num text-xl text-success">
												{summary.totalRecords
													? Math.round((summary.validRecords / summary.totalRecords) * 100)
													: 0}%
											</p>
										</div>
									</div>

									<!-- Breakdown -->
									<div class="space-y-2">
										<h4 class="overline">{$t('bulk.labels.breakdown')}</h4>

										<div
											class="flex items-center justify-between rounded-md bg-success/10 p-3 ring-1 ring-success/20"
										>
											<div class="flex items-center gap-2">
												<Check class="h-4 w-4 text-success" />
												<span class="text-sm font-medium text-success"
													>{$t('bulk.labels.valid_records')}</span
												>
											</div>
											<span class="num-mono text-sm font-bold text-success"
												>{summary.validRecords}</span
											>
										</div>

										<div
											class="flex items-center justify-between rounded-md bg-danger/10 p-3 ring-1 ring-danger/20"
										>
											<div class="flex items-center gap-2">
												<AlertCircle class="h-4 w-4 text-danger" />
												<span class="text-sm font-medium text-danger"
													>{$t('bulk.labels.invalid_records')}</span
												>
											</div>
											<span class="num-mono text-sm font-bold text-danger"
												>{summary.invalidRecords}</span
											>
										</div>
									</div>

									<!-- New Entities -->
									{#if summary.newEntities && (Object.keys(summary.newEntities.categories || {}).length > 0 || Object.keys(summary.newEntities.suppliers || {}).length > 0)}
										<div class="space-y-2">
											<h4 class="overline">{$t('bulk.labels.changes')}</h4>
											<div class="flex flex-wrap gap-1.5">
												{#if Object.keys(summary.newEntities.categories || {}).length > 0}
													<span class="badge-neutral">
														+{Object.keys(summary.newEntities.categories || {}).length} Categories
													</span>
												{/if}
												{#if Object.keys(summary.newEntities.suppliers || {}).length > 0}
													<span class="badge-info">
														+{Object.keys(summary.newEntities.suppliers || {}).length} Suppliers
													</span>
												{/if}
												{#if Object.keys(summary.newEntities.locations || {}).length > 0}
													<span class="badge-success">
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
							class="hidden h-full min-h-[24rem] flex-col items-center justify-center rounded-lg border border-dashed border-border/70 bg-muted/20 p-8 text-center lg:flex"
						>
							<div
								class="mb-4 flex h-14 w-14 items-center justify-center rounded-lg bg-muted text-muted-foreground/70"
							>
								<ArrowRight class="h-6 w-6" />
							</div>
							<p class="text-sm font-medium text-muted-foreground">
								{$t('bulk.labels.select_job')}
							</p>
						</div>
					{/if}
				</div>
			</div>
		</TabsContent>
	</Tabs>
</div>
