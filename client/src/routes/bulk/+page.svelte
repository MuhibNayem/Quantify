<script lang="ts">
	import { bulkApi } from '$lib/api/resources';
	import type { BulkImportJob } from '$lib/types';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';

	let file: File | null = null;
	let job = $state<BulkImportJob | null>(null);
	let jobIdQuery = $state('');
	const exportParams = $state({ format: 'csv', category: '', supplier: '' });
	let downloadingTemplate = $state(false);
	let uploadLoading = $state(false);
	let statusLoading = $state(false);

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
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Upload failed';
			toast.error('Upload Failed', {
				description: errorMessage,
			});
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
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Job not found';
			toast.error('Failed to Load Job Status', {
				description: errorMessage,
			});
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
		return () => {
			if (intervalId) {
				clearInterval(intervalId);
			}
		};
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
				supplier: exportParams.supplier || undefined,
			});
			const url = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = url;
			anchor.download = `catalog.${exportParams.format === 'excel' ? 'xlsx' : 'csv'}`;
			anchor.click();
			URL.revokeObjectURL(url);
			toast.success('Export generated');
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to export';
			toast.error('Failed to Export Catalog', {
				description: errorMessage,
			});
		}
	};
</script>

<section class="space-y-8">
	<header>
		<p class="text-sm uppercase tracking-wide text-muted-foreground">Bulk automations</p>
		<h1 class="text-3xl font-semibold">Import templates, job tracking & catalog exports</h1>
	</header>

	<div class="grid gap-6 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>Template</CardTitle>
				<CardDescription>Start with the canonical CSV header map</CardDescription>
			</CardHeader>
			<CardContent>
				<Button variant="secondary" onclick={downloadTemplate} disabled={downloadingTemplate}>
					{downloadingTemplate ? 'Preparing...' : 'Download template'}
				</Button>
			</CardContent>
		</Card>
		<Card>
			<CardHeader>
				<CardTitle>Upload file</CardTitle>
				<CardDescription>Queue validation for review & confirmation</CardDescription>
			</CardHeader>
			<CardContent class="space-y-3">
				<input type="file" accept=".csv,.xlsx" onchange={(event) => (file = event.currentTarget.files?.[0] ?? null)} />
				<Button class="w-full" onclick={uploadFile} disabled={uploadLoading}>{uploadLoading ? 'Uploading...' : 'Start validation'}</Button>
			</CardContent>
		</Card>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Job tracker</CardTitle>
			<CardDescription>Monitor validation progress and confirm execution</CardDescription>
		</CardHeader>
		<CardContent class="space-y-4">
			<div class="flex flex-wrap gap-2">
				<Input class="flex-1" placeholder="Job ID" bind:value={jobIdQuery} />
				<Button variant="secondary" onclick={() => loadJobStatus()}>Load status</Button>
				<Button onclick={confirmJob} disabled={!job || job.status !== 'PENDING_CONFIRMATION'}>Confirm import</Button>
			</div>
			{#if statusLoading}
				<Skeleton class="h-24 w-full" />
			{:else if job}
				<div class="rounded-2xl border border-border/70 p-4 text-sm">
					<p class="text-xs uppercase text-muted-foreground">Job</p>
					<div class="text-2xl font-semibold">{job.jobId}</div>
					<p class="text-sm text-muted-foreground">Status: {job.status}</p>
					{#if job.message}
						<p class="mt-1 text-sm">{job.message}</p>
					{/if}
					<div class="mt-3 grid grid-cols-3 gap-3 text-center">
						<div>
							<p class="text-xs text-muted-foreground">Valid</p>
							<p class="text-lg font-semibold">{job.validRecords ?? 0}</p>
						</div>
						<div>
							<p class="text-xs text-muted-foreground">Invalid</p>
							<p class="text-lg font-semibold text-destructive">{job.invalidRecords ?? 0}</p>
						</div>
						<div>
							<p class="text-xs text-muted-foreground">Total</p>
							<p class="text-lg font-semibold">{job.totalRecords ?? 0}</p>
						</div>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>Catalog export</CardTitle>
			<CardDescription>Filter by format, category or supplier</CardDescription>
		</CardHeader>
		<CardContent class="space-y-3">
			<div class="grid gap-3 sm:grid-cols-3">
				<select class="rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={exportParams.format}>
					<option value="csv">CSV</option>
					<option value="excel">Excel</option>
				</select>
				<Input placeholder="Category ID" bind:value={exportParams.category} />
				<Input placeholder="Supplier ID" bind:value={exportParams.supplier} />
			</div>
			<Button class="w-full" onclick={exportCatalog}>Generate export</Button>
		</CardContent>
	</Card>
</section>
