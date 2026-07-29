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

	const inputClass =
		'liquid-input h-10 w-full rounded-md bg-card bg-none px-3 text-sm font-normal text-foreground shadow-none ring-0 backdrop-blur-none backdrop-saturate-100 from-card to-card placeholder:text-muted-foreground focus:scale-100 focus:border-primary focus:ring-0 md:text-sm';
	const btnPrimary =
		'rounded-md bg-primary bg-none px-4 text-sm font-medium text-primary-foreground shadow-none hover:bg-primary/90 hover:bg-none';
	const btnSecondary =
		'rounded-md border border-border bg-card bg-none text-foreground shadow-none backdrop-blur-none hover:bg-muted hover:bg-none';
	const tabTriggerClass =
		'rounded-md data-[state=active]:bg-accent data-[state=active]:text-accent-foreground data-[state=active]:shadow-none';

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

<div class="min-h-[100dvh]">
	<div class="container relative mx-auto max-w-6xl px-4 pt-10 pb-20 sm:pt-14">
		<header class="mb-8 space-y-2" data-animate="fade-up">
			<p class="overline">{$t('bulk.subtitle')}</p>
			<h1 class="font-display text-foreground text-3xl font-bold tracking-tight sm:text-4xl">
				{$t('bulk.title')}
			</h1>
		</header>

		<Tabs value="import" class="w-full space-y-8">
			<TabsList
				class="h-11 w-full rounded-lg border border-border bg-card p-1 shadow-sm backdrop-blur-none lg:w-[400px]"
			>
				<TabsTrigger value="import" class={tabTriggerClass}>{$t('bulk.tabs.import')}</TabsTrigger>
				<TabsTrigger value="export" class={tabTriggerClass}>{$t('bulk.tabs.export')}</TabsTrigger>
				<TabsTrigger value="status" class={tabTriggerClass}>{$t('bulk.tabs.status')}</TabsTrigger>
			</TabsList>

			<!-- IMPORT TAB -->
			<TabsContent value="import" class="space-y-8 outline-none">
				<!-- Step Indicator -->
				<ol
					class="liquid-panel custom-scrollbar flex items-center justify-center overflow-x-auto rounded-lg px-4 py-3 sm:justify-start"
				>
					{#each [$t('bulk.steps.download'), $t('bulk.steps.upload'), $t('bulk.steps.validate'), $t('bulk.steps.import'), $t('bulk.steps.done')] as label, i}
						<li class="relative flex items-center">
							<div
								class="num mr-2 grid size-8 shrink-0 place-items-center rounded-md border text-xs font-semibold transition-colors duration-300"
								class:bg-primary={i <= stepIndex(currentStep)}
								class:text-primary-foreground={i <= stepIndex(currentStep)}
								class:border-primary={i <= stepIndex(currentStep)}
								class:border-border={i > stepIndex(currentStep)}
								class:text-muted-foreground={i > stepIndex(currentStep)}
							>
								{#if i < stepIndex(currentStep)}
									<CheckCircle class="size-4" />
								{:else}
									<span>{i + 1}</span>
								{/if}
							</div>
							<span class="text-foreground mr-3 text-xs whitespace-nowrap sm:text-sm">{label}</span>
							{#if i < 4}<div class="bg-border h-[2px] w-8"></div>{/if}
						</li>
					{/each}
				</ol>

				<div class="grid gap-6 lg:grid-cols-2">
					<!-- Step 1 -->
					<div class="liquid-panel rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
								<FileDown class="text-primary h-5 w-5" /> {$t('bulk.steps.step1_title')}
							</h2>
							<p class="text-muted-foreground text-sm">{$t('bulk.steps.step1_desc')}</p>
						</div>
						<div class="p-6">
							<Button onclick={downloadTemplate} class="{btnPrimary} w-full sm:w-auto"
								>{$t('bulk.buttons.download_template')}</Button
							>
						</div>
					</div>

					<!-- Step 2 -->
					<div class="liquid-panel rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
								<Upload class="text-primary h-5 w-5" /> {$t('bulk.steps.step2_title')}
							</h2>
							<p class="text-muted-foreground text-sm">{$t('bulk.steps.step2_desc')}</p>
						</div>
						<div class="flex flex-col items-center gap-4 p-6 sm:flex-row">
							<Input
								type="file"
								accept=".csv"
								onchange={handleFileChange}
								disabled={currentStep !== 'idle'}
								class="{inputClass} h-auto py-2"
							/>
							<Button
								onclick={uploadFile}
								disabled={!file || currentStep !== 'idle'}
								class="{btnPrimary} flex w-full items-center justify-center sm:w-auto"
							>
								{#if currentStep === 'uploading'}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
								{$t('bulk.buttons.upload_validate')}
							</Button>
						</div>
					</div>
				</div>

				<!-- Step 3: Validation & Progress -->
				{#if ['validating', 'importing', 'complete', 'failed'].includes(currentStep)}
					<div class="liquid-panel rounded-lg">
						<div class="border-border space-y-1 border-b px-6 py-5">
							<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
								<CheckCircle class="text-primary h-5 w-5" /> {$t('bulk.steps.step3_title')}
							</h2>
							<p class="text-muted-foreground text-sm">{$t('bulk.steps.step3_desc')}</p>
						</div>
						<div class="space-y-6 p-6">
							{#if currentStep === 'validating'}
								<div class="text-info flex items-center gap-2">
									<Loader class="h-4 w-4 animate-spin" /> {$t('bulk.status.validating')}
								</div>
							{/if}

							{#if validationResult}
								<div class="grid gap-4 sm:grid-cols-3">
									<div class="rounded-lg border border-success/25 bg-success/10 p-4 text-center">
										<p class="overline">{$t('bulk.labels.valid')}</p>
										<p class="kpi-num text-success mt-1 text-xl">{validationResult.validRecords}</p>
									</div>
									<div class="rounded-lg border border-danger/25 bg-danger/10 p-4 text-center">
										<p class="overline">{$t('bulk.labels.invalid')}</p>
										<p class="kpi-num text-danger mt-1 text-xl">{validationResult.invalidRecords}</p>
									</div>
									<div class="border-border bg-muted/50 rounded-lg border p-4 text-center">
										<p class="overline">{$t('bulk.labels.total')}</p>
										<p class="kpi-num text-foreground mt-1 text-xl">{validationResult.totalRecords}</p>
									</div>
								</div>

								{#if validationResult.newEntities}
									<div class="grid grid-cols-3 gap-4">
										<div class="border-border bg-muted/50 rounded-md border p-3 text-center">
											<p class="text-muted-foreground text-xs font-medium">{$t('bulk.labels.new_categories')}</p>
											<p class="num-mono text-foreground text-lg font-bold">
												{Object.keys(validationResult.newEntities.categories || {}).length}
											</p>
										</div>
										<div class="border-border bg-muted/50 rounded-md border p-3 text-center">
											<p class="text-muted-foreground text-xs font-medium">{$t('bulk.labels.new_suppliers')}</p>
											<p class="num-mono text-foreground text-lg font-bold">
												{Object.keys(validationResult.newEntities.suppliers || {}).length}
											</p>
										</div>
										<div class="border-border bg-muted/50 rounded-md border p-3 text-center">
											<p class="text-muted-foreground text-xs font-medium">{$t('bulk.labels.new_locations')}</p>
											<p class="num-mono text-foreground text-lg font-bold">
												{Object.keys(validationResult.newEntities.locations || {}).length}
											</p>
										</div>
									</div>
								{/if}

								{#if validationResult.errors?.length}
									<div class="rounded-md border border-danger/25 bg-danger/10 p-3 text-danger text-sm">
										<ul class="list-inside list-disc space-y-1">
											{#each validationResult.errors as e}<li>{e}</li>{/each}
										</ul>
									</div>
								{/if}

								{#if validationResult.validProducts?.length}
									<div class="liquid-panel custom-scrollbar max-h-64 overflow-y-auto rounded-lg">
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
														<td class="num-mono text-muted-foreground text-xs">{p.SKU}</td>
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
								<p class="text-info flex items-center gap-2">
									<Loader class="h-4 w-4 animate-spin" /> {$t('bulk.status.importing')}
								</p>
							{/if}

							{#if currentStep === 'complete'}
								<div class="rounded-lg border border-success/25 bg-success/10 p-6 text-center">
									<CheckCircle class="text-success mx-auto h-12 w-12" />
									<h3 class="text-success mt-2 text-lg font-semibold">{$t('bulk.status.import_complete')}</h3>
									<p class="text-muted-foreground mt-1">
										{importJob?.message || $t('bulk.status.success')}
									</p>
									<Button onclick={resetImportState} class="{btnPrimary} mt-4"
										>{$t('bulk.buttons.new_import')}</Button
									>
								</div>
							{/if}

							{#if currentStep === 'failed'}
								<div class="rounded-lg border border-danger/25 bg-danger/10 p-6 text-center">
									<AlertTriangle class="text-danger mx-auto h-12 w-12" />
									<h3 class="text-danger mt-2 text-lg font-semibold">{$t('bulk.status.import_failed')}</h3>
									<p class="text-muted-foreground mt-1">
										{importJob?.lastError || 'Unknown error occurred.'}
									</p>
									<Button
										onclick={resetImportState}
										class="mt-4 rounded-md bg-destructive px-4 text-sm font-medium text-white shadow-none hover:bg-destructive/90"
										>{$t('bulk.buttons.try_again')}</Button
									>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</TabsContent>

			<!-- EXPORT TAB -->
			<TabsContent value="export" class="outline-none">
				<div class="liquid-panel rounded-lg">
					<div class="border-border space-y-1 border-b px-6 py-5">
						<h2 class="text-foreground flex items-center gap-2 font-display text-lg font-semibold">
							<Download class="text-primary h-5 w-5" /> {$t('bulk.tabs.export')}
						</h2>
						<p class="text-muted-foreground text-sm">{$t('bulk.subtitle')}</p>
					</div>
					<div class="space-y-6 p-6">
						<div class="grid gap-4 sm:grid-cols-3">
							<div class="space-y-2">
								<label class="text-foreground text-sm font-medium" for="export-format"
									>{$t('bulk.labels.format')}</label
								>
								<div class="select-wrapper">
									<select id="export-format" bind:value={exportParams.format}>
										<option value="csv">CSV</option>
										<option value="excel">Excel</option>
									</select>
								</div>
							</div>
							<div class="space-y-2">
								<label class="text-foreground text-sm font-medium" for="export-category"
									>{$t('bulk.labels.category')}</label
								>
								<div class="select-wrapper">
									<select id="export-category" bind:value={exportParams.category}>
										<option value="">{$t('bulk.labels.all_categories')}</option>
										{#each categories as cat}
											<option value={cat.ID.toString()}>{cat.Name}</option>
										{/each}
									</select>
								</div>
							</div>
							<div class="space-y-2">
								<label class="text-foreground text-sm font-medium" for="export-supplier"
									>{$t('bulk.labels.supplier')}</label
								>
								<div class="select-wrapper">
									<select id="export-supplier" bind:value={exportParams.supplier}>
										<option value="">{$t('bulk.labels.all_suppliers')}</option>
										{#each suppliers as sup}
											<option value={sup.ID.toString()}>{sup.Name}</option>
										{/each}
									</select>
								</div>
							</div>
						</div>
						<Button
							class="{btnPrimary} flex w-full items-center justify-center"
							onclick={exportCatalog}
							disabled={exporting}
						>
							{#if exporting}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
							{$t('bulk.buttons.generate_export')}
						</Button>
					</div>
				</div>
			</TabsContent>

			<!-- STATUS TAB -->
			<TabsContent value="status" class="space-y-6 outline-none">
				<!-- Header / Search -->
				<div
					class="liquid-panel flex flex-col items-center justify-between gap-4 rounded-lg p-4 sm:flex-row"
				>
					<div class="relative w-full sm:w-72">
						<Search class="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
						<Input
							placeholder={$t('bulk.labels.search_placeholder')}
							bind:value={jobIdQuery}
							onkeydown={(e) => e.key === 'Enter' && loadJobStatus()}
							class="{inputClass} num-mono pl-10"
						/>
					</div>
					<Button
						variant="outline"
						size="sm"
						onclick={refreshJobs}
						class="{btnSecondary} h-9 gap-2"
					>
						<RefreshCw class="h-4 w-4" /> {$t('bulk.buttons.refresh')}
					</Button>
				</div>

				<div class="grid gap-6 lg:grid-cols-3">
					<!-- JOB LIST -->
					<div class="space-y-4 lg:col-span-2">
						{#if lookupLoading}
							<div class="p-12 text-center">
								<Loader class="text-primary mx-auto mb-3 h-8 w-8 animate-spin" />
								<p class="text-muted-foreground font-medium">Loading jobs...</p>
							</div>
						{:else if recentJobs.length === 0}
							<div
								class="border-border bg-card rounded-lg border border-dashed p-12 text-center shadow-sm"
							>
								<div class="bg-muted text-muted-foreground/50 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-md">
									<Clock class="h-8 w-8" />
								</div>
								<h3 class="text-foreground text-lg font-medium">{$t('bulk.labels.no_history')}</h3>
								<p class="text-muted-foreground">{$t('bulk.labels.no_history_desc')}</p>
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
										class="liquid-hoverable group relative cursor-pointer rounded-lg p-4"
									>
										<div class="relative flex items-center justify-between gap-4">
											<div class="flex items-center gap-4">
												<!-- Icon Box -->
												<div
													class="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-md
													{job.type === 'BULK_IMPORT'
														? 'bg-primary/10 text-primary'
														: 'bg-info/10 text-info'}"
												>
													{#if job.type === 'BULK_IMPORT'}
														<Package class="h-6 w-6" />
													{:else}
														<FileSpreadsheet class="h-6 w-6" />
													{/if}

													<!-- Status Indicator Dot -->
													<div
														class="absolute -top-1 -right-1 h-3 w-3 rounded-full border-2 border-white
														{job.status === 'COMPLETED'
															? 'bg-success'
															: job.status === 'FAILED'
																? 'bg-danger'
																: job.status === 'PROCESSING'
																	? 'bg-info animate-pulse'
																	: 'bg-warning'}"
													></div>
												</div>

												<div>
													<div class="mb-0.5 flex items-center gap-2">
														<span class="num-mono text-foreground font-bold">#{job.ID}</span>
														<span class="text-muted-foreground text-xs font-medium">•</span>
														<span class="text-muted-foreground text-sm font-medium capitalize"
															>{(job.type || '').toLowerCase().replace('_', ' ')}</span
														>
													</div>
													<div class="text-muted-foreground flex items-center gap-2 text-xs">
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
														class="overline mb-1
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
														<p class="text-danger/70 max-w-[150px] truncate text-[10px]">
															Check details
														</p>
													{:else if job.status === 'COMPLETED'}
														<p class="text-muted-foreground text-[10px]">{$t('bulk.status.success')}</p>
													{/if}
												</div>
												<div class="text-muted-foreground/40 transition-colors group-hover:text-primary">
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
								class="animate-in slide-in-from-right-4 liquid-panel sticky top-24 overflow-hidden rounded-lg duration-300"
							>
								<!-- Header -->
								<div class="border-border bg-muted/50 flex items-start justify-between border-b p-5">
									<div>
										<p class="overline mb-1">Job Details</p>
										<h2 class="num-mono text-foreground text-2xl font-bold">#{lookupJob.ID}</h2>
									</div>
									<Button
										variant="ghost"
										size="icon"
										class="text-muted-foreground hover:bg-muted hover:text-danger h-8 w-8 rounded-md"
										onclick={() => (lookupJob = null)}
									>
										<X class="h-4 w-4" />
									</Button>
								</div>

								<div class="space-y-6 p-6">
									<!-- Status Badge Large -->
									<div
										class="flex items-center gap-3 rounded-lg p-3
										{lookupJob.status === 'COMPLETED'
											? 'border border-success/25 bg-success/10 text-success'
											: lookupJob.status === 'FAILED'
												? 'border border-danger/25 bg-danger/10 text-danger'
												: 'border border-warning/25 bg-warning/10 text-warning'}"
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
											class="{btnPrimary} w-full"
											onclick={() => lookupJob && downloadExport(lookupJob)}
										>
											<Download class="mr-2 h-4 w-4" /> {$t('bulk.buttons.download_export')}
										</Button>
									{/if}

									<!-- Stats Grid -->
									{#if summary}
										<div class="grid grid-cols-2 gap-3">
											<div class="border-border bg-muted/50 rounded-lg border p-3 text-center">
												<p class="overline mb-1">{$t('bulk.labels.processed')}</p>
												<p class="kpi-num text-foreground text-2xl">{summary.totalRecords || 0}</p>
											</div>
											<div class="border-border bg-muted/50 rounded-lg border p-3 text-center">
												<p class="overline mb-1">Success Rate</p>
												<p class="kpi-num text-success text-2xl">
													{summary.totalRecords
														? Math.round((summary.validRecords / summary.totalRecords) * 100)
														: 0}%
												</p>
											</div>
										</div>

										<!-- Breakdown -->
										<div class="space-y-3">
											<h4 class="overline">{$t('bulk.labels.breakdown')}</h4>

											<div
												class="flex items-center justify-between rounded-md border border-success/25 bg-success/10 p-3"
											>
												<div class="flex items-center gap-2">
													<Check class="text-success h-4 w-4" />
													<span class="text-success text-sm font-medium">{$t('bulk.labels.valid_records')}</span>
												</div>
												<span class="num-mono text-success font-bold">{summary.validRecords}</span>
											</div>

											<div
												class="flex items-center justify-between rounded-md border border-danger/25 bg-danger/10 p-3"
											>
												<div class="flex items-center gap-2">
													<AlertCircle class="text-danger h-4 w-4" />
													<span class="text-danger text-sm font-medium">{$t('bulk.labels.invalid_records')}</span>
												</div>
												<span class="num-mono text-danger font-bold">{summary.invalidRecords}</span>
											</div>
										</div>

										<!-- New Entities -->
										{#if summary.newEntities && (Object.keys(summary.newEntities.categories || {}).length > 0 || Object.keys(summary.newEntities.suppliers || {}).length > 0)}
											<div class="space-y-3">
												<h4 class="overline">{$t('bulk.labels.changes')}</h4>
												<div class="flex flex-wrap gap-2">
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
								class="border-border bg-card hidden h-full min-h-[400px] flex-col items-center justify-center rounded-lg border border-dashed p-8 text-center shadow-sm lg:flex"
							>
								<div class="bg-muted text-muted-foreground/50 mb-4 flex h-16 w-16 items-center justify-center rounded-md">
									<ArrowRight class="h-6 w-6" />
								</div>
								<p class="text-muted-foreground font-medium">{$t('bulk.labels.select_job')}</p>
							</div>
						{/if}
					</div>
				</div>
			</TabsContent>
		</Tabs>
	</div>
</div>
