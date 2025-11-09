<script lang="ts">
	import { onMount } from 'svelte';
	import { bulkApi } from '$lib/api/resources';
	import type { BulkImportJob } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Upload } from 'lucide-svelte';

	// --- State ---
	let file: File | null = null;
	let job = $state<BulkImportJob | null>(null);
	let jobIdQuery = $state('');
	const exportParams = $state({ format: 'csv', category: '', supplier: '' });
	let downloadingTemplate = $state(false);
	let uploadLoading = $state(false);
	let statusLoading = $state(false);

	// --- Actions ---
	const downloadTemplate = async () => {
		downloadingTemplate = true;
		try {
			const blob = await bulkApi.downloadTemplate();
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = 'product_import_template.csv';
			anchor.click();
			URL.revokeObjectURL(url);
		} finally {
			downloadingTemplate = false;
		}
	};

	const uploadFile = async () => {
		if (!file) {
			toast.warning('Choose a CSV or Excel file');
			return;
		}
		uploadLoading = true;
		try {
			const formData = new FormData();
			formData.append('file', file);
			job = await bulkApi.uploadImport(formData);
			toast.success('Import job queued');
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Upload failed';
			toast.error('Upload Failed', { description: errorMessage });
		} finally {
			uploadLoading = false;
		}
	};

	const loadJobStatus = async (id?: string) => {
		const lookupId = id ?? jobIdQuery ?? job?.jobId;
		if (!lookupId) {
			toast.warning('Enter a job ID first');
			return;
		}
		statusLoading = true;
		try {
			job = await bulkApi.status(lookupId);
		} catch (error: any) {
			const errorMessage = error.response?.data?.error || 'Job not found';
			toast.error('Failed to Load Job Status', { description: errorMessage });
		} finally {
			statusLoading = false;
		}
	};

	$effect(() => {
		let intervalId: any;
		if (job?.jobId && (job.status === 'PENDING' || job.status === 'IN_PROGRESS')) {
			intervalId = setInterval(() => {
				loadJobStatus(job.jobId);
			}, 3000);
		}
		return () => intervalId && clearInterval(intervalId);
	});

	const confirmJob = async () => {
		if (!job?.jobId) {
			toast.warning('Load a job first');
			return;
		}
		await bulkApi.confirm(job.jobId);
		toast.success('Bulk import confirmed');
		await loadJobStatus(job.jobId);
	};

	const exportCatalog = async () => {
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
		}
	};

	// --- Parallax (soft) for hero ---
	onMount(() => {
		const hero = document.querySelector('.parallax-hero') as HTMLElement | null;
		if (!hero) return;
		const handleScroll = () => {
			const y = window.scrollY / 8; // subtle
			hero.style.transform = `translateY(${y}px)`;
		};
		window.addEventListener('scroll', handleScroll, { passive: true });
		return () => window.removeEventListener('scroll', handleScroll);
	});
</script>

<!-- HERO: full-width, animated gradient, parallax content -->
<section class="relative w-full overflow-hidden bg-gradient-to-r from-amber-50 via-orange-50 to-rose-100 animate-gradientShift py-16 sm:py-20 px-6">
	<!-- soft glass veil -->
	<div class="absolute inset-0 bg-white/40 backdrop-blur-[2px]"></div>

	<!-- floating blobs -->
	<div class="pointer-events-none absolute -top-24 -right-20 h-56 w-56 rounded-full bg-gradient-to-br from-amber-300/50 to-orange-300/50 blur-3xl animate-floatSlow"></div>
	<div class="pointer-events-none absolute -bottom-24 -left-20 h-56 w-56 rounded-full bg-gradient-to-br from-rose-300/50 to-amber-300/50 blur-3xl animate-floatSlow delay-500"></div>

	<!-- content -->
	<div class="relative z-10 max-w-5xl mx-auto flex flex-col items-center text-center gap-4 parallax-hero will-change-transform">
		<div class="p-4 sm:p-5 bg-gradient-to-br from-amber-500 to-orange-600 rounded-2xl shadow-xl animate-pulseGlow">
			<Upload class="h-7 w-7 sm:h-9 sm:w-9 text-white" />
		</div>
		<h1 class="text-3xl sm:text-5xl font-extrabold tracking-tight bg-gradient-to-r from-amber-700 via-orange-700 to-rose-700 bg-clip-text text-transparent animate-fadeUp">
			Bulk Import & Export Automations
		</h1>
		<p class="max-w-2xl text-slate-700 animate-fadeUp delay-150">
			Templates, file validation, job tracking, and catalog exports — in one smooth, colorful flow.
		</p>

		<div class="mt-2 flex flex-wrap items-center justify-center gap-2 animate-fadeUp delay-200">
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-amber-200 bg-white/70 text-amber-700 shadow-sm">CSV / Excel</span>
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-orange-200 bg-white/70 text-orange-700 shadow-sm">Live job status</span>
			<span class="px-3 py-1.5 text-xs sm:text-sm rounded-full border border-rose-200 bg-white/70 text-rose-700 shadow-sm">Filterable exports</span>
		</div>
	</div>
</section>

<!-- MAIN CONTENT -->
<section class="max-w-7xl mx-auto py-12 sm:py-14 px-4 sm:px-6 bg-white space-y-8 sm:space-y-10">
	<!-- Row 1: Template / Upload -->
	<div class="grid gap-6 sm:gap-8 md:grid-cols-2">
		<!-- Template -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-amber-50 to-orange-100">
			<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-5 sm:px-6 py-5">
				<CardTitle class="text-slate-800">Template</CardTitle>
				<CardDescription class="text-slate-600">Start with the canonical CSV header map</CardDescription>
			</CardHeader>
			<CardContent class="p-5 sm:p-6">
				<Button
					class="w-full sm:w-auto bg-gradient-to-r from-amber-500 to-orange-600 hover:from-amber-600 hover:to-orange-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all"
					onclick={downloadTemplate}
					disabled={downloadingTemplate}
				>
					{downloadingTemplate ? 'Preparing…' : 'Download template'}
				</Button>
			</CardContent>
		</Card>

		<!-- Upload -->
		<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-emerald-50 to-green-100">
			<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-5 sm:px-6 py-5">
				<CardTitle class="text-slate-800">Upload File</CardTitle>
				<CardDescription class="text-slate-600">Queue validation for review & confirmation</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4 p-5 sm:p-6">
				<input
					type="file"
					accept=".csv,.xlsx"
					onchange={(e) => (file = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)}
					class="block w-full rounded-xl border border-emerald-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-400"
				/>
				<Button
					class="w-full bg-gradient-to-r from-emerald-500 to-green-600 hover:from-emerald-600 hover:to-green-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all"
					onclick={uploadFile}
					disabled={uploadLoading}
				>
					{uploadLoading ? 'Uploading…' : 'Start validation'}
				</Button>
			</CardContent>
		</Card>
	</div>

	<!-- Job Tracker -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-sky-50 to-blue-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-5 sm:px-6 py-5">
			<CardTitle class="text-slate-800">Job Tracker</CardTitle>
			<CardDescription class="text-slate-600">Monitor validation progress and confirm execution</CardDescription>
		</CardHeader>
		<CardContent class="space-y-4 p-5 sm:p-6">
			<div class="flex flex-col sm:flex-row gap-3">
				<Input class="flex-1 rounded-xl border-sky-200 bg-white/90 focus:ring-2 focus:ring-sky-400" placeholder="Job ID" bind:value={jobIdQuery} />
				<div class="flex gap-2">
					<Button
						variant="secondary"
						class="bg-white/80 border border-sky-200 text-sky-700 hover:bg-sky-50 rounded-xl"
						onclick={() => loadJobStatus()}
					>
						Load status
					</Button>
					<Button
						class="bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700 text-white rounded-xl"
						onclick={confirmJob}
						disabled={!job || job.status !== 'PENDING_CONFIRMATION'}
					>
						Confirm import
					</Button>
				</div>
			</div>

			{#if statusLoading}
				<Skeleton class="h-24 w-full bg-white/70" />
			{:else if job}
				<div class="rounded-2xl border border-sky-200 bg-white/80 backdrop-blur p-4 text-sm shadow-sm">
					<p class="text-xs uppercase text-slate-500">Job</p>
					<div class="text-2xl font-semibold text-slate-800">{job.jobId}</div>
					<p class="text-sm text-slate-600">Status: {job.status}</p>
					{#if job.message}
						<p class="mt-1 text-sm text-slate-700">{job.message}</p>
					{/if}
					<div class="mt-3 grid grid-cols-3 gap-3 text-center">
						<div class="rounded-xl bg-white/70 border border-sky-200 py-2">
							<p class="text-xs text-slate-500">Valid</p>
							<p class="text-lg font-semibold text-sky-700">{job.validRecords ?? 0}</p>
						</div>
						<div class="rounded-xl bg-white/70 border border-sky-200 py-2">
							<p class="text-xs text-slate-500">Invalid</p>
							<p class="text-lg font-semibold text-rose-600">{job.invalidRecords ?? 0}</p>
						</div>
						<div class="rounded-xl bg-white/70 border border-sky-200 py-2">
							<p class="text-xs text-slate-500">Total</p>
							<p class="text-lg font-semibold text-slate-800">{job.totalRecords ?? 0}</p>
						</div>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	<!-- Catalog Export -->
	<Card class="rounded-2xl border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:scale-[1.01] bg-gradient-to-br from-violet-50 to-purple-100">
		<CardHeader class="bg-white/80 backdrop-blur rounded-t-2xl border-b border-white/60 px-5 sm:px-6 py-5">
			<CardTitle class="text-slate-800">Catalog Export</CardTitle>
			<CardDescription class="text-slate-600">Filter by format, category or supplier</CardDescription>
		</CardHeader>
		<CardContent class="space-y-4 p-5 sm:p-6">
			<div class="grid gap-3 sm:grid-cols-3">
				<select
					class="rounded-xl border border-violet-200 bg-white/90 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-400"
					bind:value={exportParams.format}
				>
					<option value="csv">CSV</option>
					<option value="excel">Excel</option>
				</select>
				<Input class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400" placeholder="Category ID" bind:value={exportParams.category} />
				<Input class="rounded-xl border-violet-200 bg-white/90 focus:ring-2 focus:ring-violet-400" placeholder="Supplier ID" bind:value={exportParams.supplier} />
			</div>
			<Button
				class="w-full bg-gradient-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white font-semibold rounded-xl shadow-md hover:shadow-lg hover:scale-105 transition-all"
				onclick={exportCatalog}
			>
				Generate export
			</Button>
		</CardContent>
	</Card>
</section>

<style lang="postcss">
	/* Animated hero gradient & subtle motion */
	@keyframes gradientShift {
		0% { background-position: 0% 50%; }
		50% { background-position: 100% 50%; }
		100% { background-position: 0% 50%; }
	}
	.animate-gradientShift {
		background-size: 200% 200%;
		animation: gradientShift 22s ease-in-out infinite;
	}

	@keyframes pulseGlow {
		0%, 100% { transform: scale(1); box-shadow: 0 0 14px rgba(251, 146, 60, 0.35); }
		50% { transform: scale(1.08); box-shadow: 0 0 26px rgba(251, 146, 60, 0.55); }
	}
	.animate-pulseGlow { animation: pulseGlow 8s ease-in-out infinite; }

	@keyframes fadeUp {
		from { opacity: 0; transform: translateY(18px); }
		to { opacity: 1; transform: translateY(0); }
	}
	.animate-fadeUp { animation: fadeUp 900ms ease forwards; }
	.animate-fadeUp.delay-150 { animation-delay: 150ms; }
	.animate-fadeUp.delay-200 { animation-delay: 200ms; }

	@keyframes floatSlow {
		0%, 100% { transform: translateY(0px) scale(1); }
		50% { transform: translateY(-10px) scale(1.03); }
	}
	.animate-floatSlow { animation: floatSlow 10s ease-in-out infinite; }

	/* Global smooth transitions (kept light) */
	* {
		transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
		transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
		transition-duration: 300ms;
	}
</style>
