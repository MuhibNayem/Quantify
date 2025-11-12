<!-- +page.svelte (or BulkImport.svelte)
Modern, colorful, light-gradient UI using Tailwind + SvelteKit 5.
This keeps your business logic intact, focusing on visuals, accessibility, and micro‑interactions.
-->
<script lang="ts">
	import { browser } from '$app/environment';
	import { bulkApi } from '$lib/api/resources';
	import type { BulkImportJob, BulkImportValidationResult } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Upload, FileDown, CheckCircle, AlertTriangle, Loader } from 'lucide-svelte';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { onMount } from 'svelte';

	// --- Component State ---
	type JobStatusEventDetail = {
		event: string;
		jobId: number;
		status: string;
		type: string;
		result?: string;
		lastError?: string;
	};

	let file = $state<File | null>(null);
	let job = $state<BulkImportJob | null>(null);
	let validationResult = $state<BulkImportValidationResult | null>(null);
	let hasShownValidation = $state(false);
	let currentStep = $state<'idle' | 'uploading' | 'validating' | 'importing' | 'complete' | 'failed'>('idle');
	let isDragging = false;

	const applyJobStatusUpdate = (detail: JobStatusEventDetail) => {
		if (!job || detail.jobId !== job.ID) return;

		job = {
			...job,
			status: detail.status,
			result: detail.result ?? job.result,
			lastError: detail.lastError ?? job.lastError,
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
		} else if (detail.status === 'FAILED') {
			currentStep = 'failed';
			toast.error('Import failed', { description: detail.lastError || job?.lastError });
		}
	};

	onMount(() => {
		if (!browser) return;
		const handler = (event: Event) => {
			const custom = event as CustomEvent<JobStatusEventDetail>;
			applyJobStatusUpdate(custom.detail);
		};
		window.addEventListener('bulk-job-status', handler);
		return () => window.removeEventListener('bulk-job-status', handler);
	});

	// --- Helpers ---
	const formatMoney = (v: string | number) => {
		const n = typeof v === 'string' ? Number(v) : v;
		if (Number.isNaN(n)) return String(v);
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD', maximumFractionDigits: 2 }).format(n);
	};

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

	// --- Actions ---
	const handleFileChange = (e: Event) => {
		const target = e.currentTarget as HTMLInputElement;
		file = target.files?.[0] ?? null;
		if (file) {
			job = null;
			validationResult = null;
			hasShownValidation = false;
			currentStep = 'idle';
		}
	};

	const onDrop = (e: DragEvent) => {
		e.preventDefault();
		isDragging = false;
		const f = e.dataTransfer?.files?.[0];
		if (f) {
			file = f;
			job = null;
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
			job = await bulkApi.uploadImport(formData);
			if (!job?.ID) {
				throw new Error('Upload succeeded but job ID missing');
			}
			currentStep = 'validating';
			toast.info('File uploaded. Validation in progress...');
		} catch (error: any) {
			const errorMessage = error?.response?.data?.error || 'Upload failed';
			toast.error('Upload Failed', { description: errorMessage });
			currentStep = 'failed';
		}
	};

	const resetState = () => {
		file = null;
		job = null;
		validationResult = null;
		hasShownValidation = false;
		currentStep = 'idle';
	};
</script>

<!-- Background wrapper -->

<!-- Wrapper with animated gradient + hero consistency -->
<div class="relative min-h-[100dvh] isolate overflow-hidden bg-gradient-to-br from-sky-50 via-blue-50 to-cyan-100">
	<div class="absolute -top-32 -left-24 w-[28rem] h-[28rem] rounded-full bg-sky-200/40 blur-3xl animate-pulseGlow"></div>
	<div class="absolute -bottom-28 -right-24 w-[24rem] h-[24rem] rounded-full bg-cyan-200/30 blur-3xl animate-pulseGlow delay-700"></div>

	<div class="relative container mx-auto max-w-6xl px-4 pb-20 pt-16">
		<header class="mb-10 text-center sm:text-left">
			<h1 class="text-4xl font-extrabold bg-gradient-to-r from-sky-700 via-blue-700 to-cyan-700 bg-clip-text text-transparent mb-2">Bulk Product Import</h1>
			<p class="text-slate-600 text-sm sm:text-base max-w-2xl mx-auto sm:mx-0">Follow the steps below to import your products safely and accurately using a CSV template.</p>
		</header>

		<!-- Step Indicator -->
			<ol class="flex items-center justify-center sm:justify-start mb-10 overflow-x-auto rounded-2xl border border-sky-100 bg-white/70 backdrop-blur-sm px-4 py-3 shadow-sm">
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

		<!-- Steps -->
		<div class="space-y-10">
			<Card class="rounded-2xl bg-gradient-to-br from-sky-50 to-blue-100 shadow-md hover:shadow-xl">
				<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<FileDown class="h-5 w-5 text-sky-600" /> Step 1: Download & Prepare
					</CardTitle>
					<CardDescription class="text-slate-600">Use the correct CSV format for seamless import.</CardDescription>
				</CardHeader>
				<CardContent class="p-6">
					<Button onclick={downloadTemplate} class="bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white rounded-xl px-5 py-2.5 shadow-md hover:shadow-lg">Download Template</Button>
				</CardContent>
			</Card>

			<Card class="rounded-2xl bg-gradient-to-br from-blue-50 to-sky-100 shadow-md hover:shadow-xl">
				<CardHeader class="border-b border-white/60 bg-white/70 backdrop-blur-sm">
					<CardTitle class="flex items-center gap-2 text-slate-800">
						<Upload class="h-5 w-5 text-sky-600" /> Step 2: Upload File
					</CardTitle>
					<CardDescription class="text-slate-600">Validate your CSV before importing products.</CardDescription>
				</CardHeader>
				<CardContent class="flex flex-col sm:flex-row items-center gap-4 p-6">
					<Input type="file" accept=".csv" onchange={handleFileChange} disabled={currentStep !== 'idle'} class="rounded-xl border border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" />
					<Button onclick={uploadFile} disabled={!file || currentStep !== 'idle'} class="bg-gradient-to-r from-sky-500 to-blue-600 text-white rounded-xl px-6 py-2.5 shadow-md hover:shadow-lg flex items-center">
						{#if currentStep === 'uploading'}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
						Upload & Validate
					</Button>
				</CardContent>
			</Card>

			{#if ['validating','importing','complete','failed'].includes(currentStep)}
				<Card class="rounded-2xl bg-gradient-to-br from-white/90 to-sky-50 shadow-md hover:shadow-xl">
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
								<p class="text-slate-600 mt-1">{job?.message || 'Products imported successfully.'}</p>
								<Button onclick={resetState} class="mt-4 bg-gradient-to-r from-emerald-500 to-teal-600 text-white rounded-xl px-5 py-2.5 hover:brightness-110">New Import</Button>
							</div>
						{/if}

						{#if currentStep === 'failed'}
							<div class="text-center p-6 bg-rose-50 rounded-2xl border border-rose-200">
								<AlertTriangle class="h-12 w-12 mx-auto text-rose-600" />
								<h3 class="text-lg font-semibold text-rose-700 mt-2">Import Failed</h3>
								<p class="text-slate-600 mt-1">{job?.lastError || 'Unknown error occurred.'}</p>
								<Button onclick={resetState} class="mt-4 bg-gradient-to-r from-rose-500 to-orange-500 text-white rounded-xl px-5 py-2.5 hover:brightness-110">Try Again</Button>
							</div>
						{/if}
					</CardContent>
				</Card>
			{/if}
		</div>
	</div>
</div>

<style lang="postcss">
@keyframes pulseGlow { 0%,100%{transform:scale(1);opacity:.45}50%{transform:scale(1.08);opacity:.7}} .animate-pulseGlow{animation:pulseGlow 12s ease-in-out infinite}
</style>




<!-- tailwind.config.ts additions (drop this snippet into your config) ----------------------------------
import type { Config } from 'tailwindcss'
export default {
  content: [
    './src/**/*.{svelte,ts,js}',
  ],
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: '#ffffff',
          50: '#f8fafc',
          100: '#f1f5f9',
        },
      },
      boxShadow: {
        'soft': '0 1px 2px rgba(0,0,0,.04), 0 8px 24px rgba(0,0,0,.06)'
      },
      borderRadius: {
        '2xl': '1.25rem'
      }
    }
  },
  plugins: []
} satisfies Config
--------------------------------------------------------------------------------------------------------->

<!-- app.postcss or app.css (global) additions -----------------------------------------------------------
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Better focus rings aligned with Material cues */
:root {
  --ring: 225 96% 60%;
}
:where(button, [role="button"], input, select, textarea):focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px hsl(var(--ring) / .25), 0 0 0 6px hsl(var(--ring) / .15);
  transition: box-shadow .2s ease;
}
--------------------------------------------------------------------------------------------------------->
