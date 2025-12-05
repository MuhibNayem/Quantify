<script lang="ts">
	import { browser } from '$app/environment';
	import { bulkApi, categoriesApi, suppliersApi } from '$lib/api/resources';
	import api from '$lib/api';
	import type { BulkImportJob, BulkImportValidationResult, Category, Supplier, BulkExportResult } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Upload, FileDown, CheckCircle, AlertTriangle, Loader, Download, Search, FileText } from 'lucide-svelte';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
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

	// --- State: Import Wizard ---
	let file = $state<File | null>(null);
	let importJob = $state<BulkImportJob | null>(null);
	let validationResult = $state<BulkImportValidationResult | null>(null);
	let hasShownValidation = $state(false);
	let currentStep = $state<'idle' | 'uploading' | 'validating' | 'importing' | 'complete' | 'failed'>('idle');
	let isDragging = false;

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
	const stepIndex = (s: typeof currentStep): number => ({
		idle: 0,
		uploading: 1,
		validating: 2,
		importing: 3,
		complete: 4,
		failed: 4
	}[s]);

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
		const result = parseExportResult(job.Result);
		if (!result?.downloadUrl) {
			toast.error('Download URL not found');
			return;
		}

		try {
			const response = await api.get(result.downloadUrl, { responseType: 'blob' });
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
				lastError: detail.lastError ?? importJob.lastError,
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
				lastError: detail.lastError ?? lookupJob.lastError,
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
<div class="relative min-h-[100dvh] isolate overflow-hidden bg-gradient-to-br from-sky-50 via-blue-50 to-cyan-100">
	<div class="absolute -top-32 -left-24 w-[28rem] h-[28rem] rounded-full bg-sky-200/40 blur-3xl animate-pulseGlow"></div>
	<div class="absolute -bottom-28 -right-24 w-[24rem] h-[24rem] rounded-full bg-cyan-200/30 blur-3xl animate-pulseGlow delay-700"></div>

	<div class="relative container mx-auto max-w-6xl px-4 pb-20 pt-16">
		<header class="mb-10 text-center sm:text-left">
			<h1 class="text-4xl font-extrabold bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-transparent mb-2">Bulk Operations</h1>
			<p class="text-slate-600 text-sm sm:text-base max-w-2xl mx-auto sm:mx-0">Manage your catalog efficiently with bulk imports, exports, and job tracking.</p>
		</header>

		<Tabs value="import" class="w-full space-y-8">
			<TabsList class="grid w-full grid-cols-3 lg:w-[400px] bg-white/60 backdrop-blur-sm border border-sky-100">
				<TabsTrigger value="import">Import</TabsTrigger>
				<TabsTrigger value="export">Export</TabsTrigger>
				<TabsTrigger value="status">Status</TabsTrigger>
			</TabsList>

			<!-- IMPORT TAB -->
			<TabsContent value="import" class="space-y-8 outline-none">
				<!-- Step Indicator -->
				<ol class="flex items-center justify-center sm:justify-start overflow-x-auto rounded-2xl border border-sky-100 bg-white/70 backdrop-blur-sm px-4 py-3 shadow-sm">
					{#each ['Download Template','Upload','Validate','Import','Done'] as label, i}
						<li class="relative flex items-center">
							<div class="grid place-items-center rounded-full size-8 font-semibold text-xs border transition-all duration-300 shadow-sm mr-2"
								class:bg-gradient-to-r={i <= stepIndex(currentStep)}
								class:from-sky-500={i <= stepIndex(currentStep)}
								class:to-blue-600={i <= stepIndex(currentStep)}
								class:text-white={i <= stepIndex(currentStep)}
								class:border-sky-200={i > stepIndex(currentStep)}
								class:text-slate-500={i > stepIndex(currentStep)}>
								{#if i < stepIndex(currentStep)}
									<CheckCircle class="size-4" />
								{:else}
									<span>{i + 1}</span>
								{/if}
							</div>
							<span class="text-xs sm:text-sm text-slate-700 mr-3 whitespace-nowrap">{label}</span>
							{#if i < 5}<div class="h-[2px] w-8 bg-slate-200/70"></div>{/if}
						</li>
					{/each}
				</ol>

				<div class="grid gap-8 lg:grid-cols-2">
					<!-- Step 1 -->
					<Card class="rounded-2xl bg-gradient-to-br from-sky-50 to-blue-100 shadow-md hover:shadow-xl transition-all duration-300">
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<FileDown class="h-5 w-5 text-sky-600" /> Step 1: Download & Prepare
							</CardTitle>
							<CardDescription class="text-slate-600">Use the correct CSV format for seamless import.</CardDescription>
						</CardHeader>
						<CardContent class="p-6">
							<Button onclick={downloadTemplate} class="w-full sm:w-auto bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white rounded-xl px-5 py-2.5 shadow-md hover:shadow-lg">Download Template</Button>
						</CardContent>
					</Card>

					<!-- Step 2 -->
					<Card class="rounded-2xl bg-gradient-to-br from-blue-50 to-sky-100 shadow-md hover:shadow-xl transition-all duration-300">
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<Upload class="h-5 w-5 text-sky-600" /> Step 2: Upload File
							</CardTitle>
							<CardDescription class="text-slate-600">Validate your CSV before importing products.</CardDescription>
						</CardHeader>
						<CardContent class="flex flex-col sm:flex-row items-center gap-4 p-6">
							<Input type="file" accept=".csv" onchange={handleFileChange} disabled={currentStep !== 'idle'} class="rounded-xl border border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" />
							<Button onclick={uploadFile} disabled={!file || currentStep !== 'idle'} class="w-full sm:w-auto bg-gradient-to-r from-sky-500 to-blue-600 text-white rounded-xl px-6 py-2.5 shadow-md hover:shadow-lg flex items-center justify-center">
								{#if currentStep === 'uploading'}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
								Upload & Validate
							</Button>
						</CardContent>
					</Card>
				</div>

				<!-- Step 3: Validation & Progress -->
				{#if ['validating','importing','complete','failed'].includes(currentStep)}
					<Card class="rounded-2xl bg-gradient-to-br from-white/90 to-sky-50 shadow-md hover:shadow-xl transition-all duration-300">
						<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
							<CardTitle class="flex items-center gap-2 text-slate-800">
								<CheckCircle class="h-5 w-5 text-sky-600" /> Step 3: Review & Confirm
							</CardTitle>
							<CardDescription class="text-slate-600">Ensure validation passes before final import.</CardDescription>
						</CardHeader>
						<CardContent class="p-6 space-y-6">
							{#if currentStep === 'validating'}
								<div class="flex items-center gap-2 text-sky-700"><Loader class="h-4 w-4 animate-spin" /> Validating file...</div>
							{/if}

							{#if validationResult}
								<div class="grid sm:grid-cols-3 gap-4">
									<div class="rounded-2xl border border-sky-100 bg-sky-50 p-4 text-center"><p class="text-xs text-slate-500">Valid</p><p class="text-xl font-bold text-sky-700">{validationResult.validRecords}</p></div>
									<div class="rounded-2xl border border-rose-100 bg-rose-50 p-4 text-center"><p class="text-xs text-slate-500">Invalid</p><p class="text-xl font-bold text-rose-700">{validationResult.invalidRecords}</p></div>
									<div class="rounded-2xl border border-amber-100 bg-amber-50 p-4 text-center"><p class="text-xs text-slate-500">Total</p><p class="text-xl font-bold text-amber-700">{validationResult.totalRecords}</p></div>
								</div>

								{#if validationResult.newEntities}
									<div class="grid grid-cols-3 gap-4">
										<div class="rounded-xl border border-indigo-100 bg-indigo-50 p-3 text-center">
											<p class="text-xs text-indigo-600 font-medium">New Categories</p>
											<p class="text-lg font-bold text-indigo-800">{Object.keys(validationResult.newEntities.categories || {}).length}</p>
										</div>
										<div class="rounded-xl border border-fuchsia-100 bg-fuchsia-50 p-3 text-center">
											<p class="text-xs text-fuchsia-600 font-medium">New Suppliers</p>
											<p class="text-lg font-bold text-fuchsia-800">{Object.keys(validationResult.newEntities.suppliers || {}).length}</p>
										</div>
										<div class="rounded-xl border border-emerald-100 bg-emerald-50 p-3 text-center">
											<p class="text-xs text-emerald-600 font-medium">New Locations</p>
											<p class="text-lg font-bold text-emerald-800">{Object.keys(validationResult.newEntities.locations || {}).length}</p>
										</div>
									</div>
								{/if}

								{#if validationResult.errors?.length}
									<div class="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
										<ul class="list-disc list-inside space-y-1">{#each validationResult.errors as e}<li>{e}</li>{/each}</ul>
									</div>
								{/if}

								{#if validationResult.validProducts?.length}
									<div class="max-h-64 overflow-y-auto border rounded-xl bg-white/70">
										<Table>
											<TableHeader class="bg-gradient-to-r from-sky-100/70 to-blue-100/70 backdrop-blur">
												<TableRow><TableHead>SKU</TableHead><TableHead>Name</TableHead><TableHead>Category</TableHead><TableHead>Supplier</TableHead><TableHead class="text-right">Price</TableHead></TableRow>
											</TableHeader>
											<TableBody>{#each validationResult.validProducts as p}<TableRow class="hover:bg-sky-50"><TableCell>{p.SKU}</TableCell><TableCell>{p.Name}</TableCell><TableCell>{p.CategoryName}</TableCell><TableCell>{p.SupplierName}</TableCell><TableCell class="text-right">{p.SellingPrice}</TableCell></TableRow>{/each}</TableBody>
										</Table>
									</div>
								{/if}
							{/if}

							{#if currentStep === 'importing'}
								<p class="text-sky-700 flex items-center gap-2"><Loader class="h-4 w-4 animate-spin" /> Importing products...</p>
							{/if}

							{#if currentStep === 'complete'}
								<div class="text-center p-6 bg-green-50 rounded-2xl border border-green-200">
									<CheckCircle class="h-12 w-12 mx-auto text-green-600" />
									<h3 class="text-lg font-semibold text-green-700 mt-2">Import Complete</h3>
									<p class="text-slate-600 mt-1">{importJob?.message || 'Products imported successfully.'}</p>
									<Button onclick={resetImportState} class="mt-4 bg-gradient-to-r from-emerald-500 to-teal-600 text-white rounded-xl px-5 py-2.5 hover:brightness-110">New Import</Button>
								</div>
							{/if}

							{#if currentStep === 'failed'}
								<div class="text-center p-6 bg-rose-50 rounded-2xl border border-rose-200">
									<AlertTriangle class="h-12 w-12 mx-auto text-rose-600" />
									<h3 class="text-lg font-semibold text-rose-700 mt-2">Import Failed</h3>
									<p class="text-slate-600 mt-1">{importJob?.lastError || 'Unknown error occurred.'}</p>
									<Button onclick={resetImportState} class="mt-4 bg-gradient-to-r from-rose-500 to-orange-500 text-white rounded-xl px-5 py-2.5 hover:brightness-110">Try Again</Button>
								</div>
							{/if}
						</CardContent>
					</Card>
				{/if}
			</TabsContent>

			<!-- EXPORT TAB -->
			<TabsContent value="export" class="outline-none">
				<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 bg-gradient-to-br from-violet-50 to-purple-100">
					<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<Download class="h-5 w-5 text-violet-600" /> Catalog Export
						</CardTitle>
						<CardDescription class="text-slate-600">Filter and download your product catalog.</CardDescription>
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
								<select class="w-full rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400" bind:value={exportParams.category}>
									<option value="">All Categories</option>
									{#each categories as cat}
										<option value={cat.ID.toString()}>{cat.Name}</option>
									{/each}
								</select>
							</div>
							<div class="space-y-2">
								<label class="text-sm font-medium text-slate-700">Supplier (Optional)</label>
								<select class="w-full rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400" bind:value={exportParams.supplier}>
									<option value="">All Suppliers</option>
									{#each suppliers as sup}
										<option value={sup.ID.toString()}>{sup.Name}</option>
									{/each}
								</select>
							</div>
						</div>
						<Button
							class="w-full bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-[1.02] transition-all flex items-center justify-center"
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
			<TabsContent value="status" class="outline-none">
				<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 bg-gradient-to-br from-amber-50 to-orange-100">
					<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-6 py-5">
						<CardTitle class="flex items-center gap-2 text-slate-800">
							<Search class="h-5 w-5 text-amber-600" /> Recent Jobs
						</CardTitle>
						<CardDescription class="text-slate-600">View the status and history of your bulk operations.</CardDescription>
					</CardHeader>
					<CardContent class="p-0">
						{#if lookupLoading}
							<div class="p-8 text-center text-slate-500"><Loader class="h-6 w-6 animate-spin mx-auto mb-2" /> Loading jobs...</div>
						{:else if recentJobs.length === 0}
							<div class="p-8 text-center text-slate-500">No recent jobs found.</div>
						{:else}
							<div class="overflow-x-auto">
								<Table>
									<TableHeader class="bg-amber-100/50">
										<TableRow>
											<TableHead>ID</TableHead>
											<TableHead>Type</TableHead>
											<TableHead>Status</TableHead>
											<TableHead>Date</TableHead>
											<TableHead class="text-right">Action</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{#each recentJobs as job}
											<TableRow class="hover:bg-amber-50/50 cursor-pointer" onclick={() => viewJobDetails(job)}>
												<TableCell class="font-medium">#{job.ID}</TableCell>
												<TableCell class="capitalize">{job.Type?.toLowerCase() ?? 'unknown'}</TableCell>
												<TableCell>
													<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
														{job.Status === 'COMPLETED' ? 'bg-green-100 text-green-800' :
														 job.Status === 'FAILED' ? 'bg-rose-100 text-rose-800' :
														 job.Status === 'PENDING_CONFIRMATION' ? 'bg-blue-100 text-blue-800' :
														 'bg-amber-100 text-amber-800'}">
														{job.Status}
													</span>
												</TableCell>
												<TableCell>{new Date(job.CreatedAt || '').toLocaleDateString()} {new Date(job.CreatedAt || '').toLocaleTimeString()}</TableCell>
												<TableCell class="text-right flex items-center justify-end gap-2">
													{#if job.Type === 'BULK_EXPORT' && job.Status === 'COMPLETED'}
														<Button variant="ghost" size="sm" class="h-8 w-8 p-0" onclick={(e) => { e.stopPropagation(); downloadExport(job); }}>
															<Download class="h-4 w-4 text-violet-600" />
														</Button>
													{/if}
													<Button variant="ghost" size="sm" class="h-8 w-8 p-0">
														<FileText class="h-4 w-4 text-slate-500" />
													</Button>
												</TableCell>
											</TableRow>
										{/each}
									</TableBody>
								</Table>
							</div>
						{/if}

						{#if lookupJob}
							{@const summary = parseValidationResult(lookupJob.result)}
							<div class="border-t border-amber-200 bg-white/60 p-6 animate-in slide-in-from-top-2">
								<div class="flex items-center justify-between mb-4">
									<h3 class="font-semibold text-lg text-slate-800">Job #{lookupJob.ID} Details</h3>
									<Button variant="ghost" size="sm" onclick={() => lookupJob = null}>Close</Button>
								</div>

								<div class="grid grid-cols-3 gap-4 text-center mb-4">
									<div class="rounded-xl bg-white border border-amber-100 py-3">
										<p class="text-xs text-slate-500 mb-1">Valid</p>
										<p class="text-2xl font-bold text-green-600">{summary?.validRecords ?? 0}</p>
									</div>
									<div class="rounded-xl bg-white border border-amber-100 py-3">
										<p class="text-xs text-slate-500 mb-1">Invalid</p>
										<p class="text-2xl font-bold text-rose-600">{summary?.invalidRecords ?? 0}</p>
									</div>
									<div class="rounded-xl bg-white border border-amber-100 py-3">
										<p class="text-xs text-slate-500 mb-1">Total</p>
										<p class="text-2xl font-bold text-slate-800">{summary?.totalRecords ?? 0}</p>
									</div>
								</div>

								{#if summary?.newEntities}
									<div class="grid grid-cols-3 gap-4 mb-4">
										<div class="rounded-xl border border-indigo-100 bg-indigo-50/50 p-2 text-center">
											<p class="text-[10px] uppercase text-indigo-600 font-bold">New Categories</p>
											<p class="text-lg font-bold text-indigo-800">{Object.keys(summary.newEntities.categories || {}).length}</p>
										</div>
										<div class="rounded-xl border border-fuchsia-100 bg-fuchsia-50/50 p-2 text-center">
											<p class="text-[10px] uppercase text-fuchsia-600 font-bold">New Suppliers</p>
											<p class="text-lg font-bold text-fuchsia-800">{Object.keys(summary.newEntities.suppliers || {}).length}</p>
										</div>
										<div class="rounded-xl border border-emerald-100 bg-emerald-50/50 p-2 text-center">
											<p class="text-[10px] uppercase text-emerald-600 font-bold">New Locations</p>
											<p class="text-lg font-bold text-emerald-800">{Object.keys(summary.newEntities.locations || {}).length}</p>
										</div>
									</div>
								{/if}

								{#if lookupJob.lastError}
									<div class="p-3 rounded-xl bg-rose-50 border border-rose-100 text-rose-700 text-sm">
										<span class="font-semibold">Error:</span> {lookupJob.lastError}
									</div>
								{/if}
							</div>
						{/if}
					</CardContent>
				</Card>
			</TabsContent>
		</Tabs>
	</div>
</div>

<style lang="postcss">
@keyframes pulseGlow { 0%,100%{transform:scale(1);opacity:.45}50%{transform:scale(1.08);opacity:.7}} .animate-pulseGlow{animation:pulseGlow 12s ease-in-out infinite}
</style>
