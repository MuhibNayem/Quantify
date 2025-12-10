<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { crmApi } from '$lib/api/resources';
	import type { ChurnRisk, UserSummary } from '$lib/types';
    import { debounce } from '$lib/utils';
	import { Loader2, Users, Search, AlertTriangle, CheckCircle2, XCircle } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	let loading = false;
	let userId: number | null = null;
    let customerSearch = '';
    let searchResults: UserSummary[] = [];
    let selectedCustomer: UserSummary | null = null;
    let searching = false;

	let churnRisk: ChurnRisk | null = null;

    const searchCustomers = debounce(async () => {
        if (customerSearch.length < 2) return;
        searching = true;
        try {
            const res = await crmApi.listCustomers({ q: customerSearch, limit: 5 });
            searchResults = res.users;
        } catch (e) {
            console.error(e);
        } finally {
            searching = false;
        }
    }, 300);

    function selectCustomer(customer: UserSummary) {
        selectedCustomer = customer;
        userId = customer.ID;
        searchResults = [];
        customerSearch = '';
        churnRisk = null;
    }

	async function analyzeChurnRisk() {
		if (!userId) {
			toast.error('Please select a customer');
			return;
		}

		loading = true;
		churnRisk = null;
		try {
			const res = await crmApi.getChurnRisk(userId);
			churnRisk = res;
			toast.success('Analysis complete');
		} catch (error: any) {
			toast.error('Failed to analyze churn risk', {
				description: error?.response?.data?.error || 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	}

    function getRiskColor(level: string) {
        switch (level?.toLowerCase()) {
            case 'high': return 'text-red-600 bg-red-50 border-red-200';
            case 'medium': return 'text-amber-600 bg-amber-50 border-amber-200';
            case 'low': return 'text-emerald-600 bg-emerald-50 border-emerald-200';
            default: return 'text-slate-600 bg-slate-50 border-slate-200';
        }
    }
</script>

<Card class="h-full border-purple-100 bg-gradient-to-br from-white to-purple-50/50 shadow-sm">
	<CardHeader>
		<CardTitle class="flex items-center gap-2 text-slate-800">
			<div class="rounded-lg bg-purple-100 p-2 text-purple-600">
				<Users class="h-5 w-5" />
			</div>
			Customer Churn Prediction
		</CardTitle>
		<CardDescription>Identify at-risk customers and retention strategies</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
        <!-- Customer Search -->
		<div class="space-y-2">
			<Label>Select Customer</Label>
            {#if selectedCustomer}
                <div class="flex items-center justify-between rounded-md border border-purple-200 bg-purple-50 p-2">
                    <div class="flex flex-col">
                        <span class="font-medium text-purple-900">{selectedCustomer.FirstName} {selectedCustomer.LastName}</span>
                        <span class="text-xs text-purple-700">{selectedCustomer.Email}</span>
                    </div>
                    <Button variant="ghost" size="sm" class="h-8 w-8 p-0 text-purple-500 hover:text-purple-700" onclick={() => { selectedCustomer = null; userId = null; churnRisk = null; }}>
                        X
                    </Button>
                </div>
            {:else}
                <div class="relative">
                    <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-slate-500" />
                    <Input 
                        placeholder="Search by name or email..." 
                        class="pl-9" 
                        bind:value={customerSearch}
                        oninput={searchCustomers}
                    />
                    {#if searchResults.length > 0}
                        <div class="absolute z-10 mt-1 w-full rounded-md border border-slate-200 bg-white shadow-lg">
                            {#each searchResults as customer}
                                <button 
                                    class="w-full px-4 py-2 text-left text-sm hover:bg-slate-50"
                                    onclick={() => selectCustomer(customer)}
                                >
                                    <div class="font-medium">{customer.FirstName} {customer.LastName}</div>
                                    <div class="text-xs text-slate-500">{customer.Email}</div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
		</div>

		<div class="flex justify-end">
			<Button onclick={analyzeChurnRisk} disabled={loading || !userId} class="bg-purple-600 hover:bg-purple-700">
				{#if loading}
					<Loader2 class="mr-2 h-4 w-4 animate-spin" />
					Analyzing...
				{:else}
					Analyze Risk
				{/if}
			</Button>
		</div>

		{#if churnRisk}
			<div class="mt-4 space-y-4 rounded-lg border border-slate-200 bg-white p-4 shadow-sm animate-in fade-in slide-in-from-bottom-4">
				<div class="flex items-center justify-between">
					<div>
						<div class="text-sm text-slate-500">Risk Level</div>
                        <div class={`mt-1 inline-flex items-center rounded-full border px-2.5 py-0.5 text-sm font-semibold ${getRiskColor(churnRisk.risk_level)}`}>
                            {churnRisk.risk_level}
                        </div>
					</div>
                    <div class="text-right">
                        <div class="text-sm text-slate-500">Risk Score</div>
                        <div class="text-2xl font-bold text-slate-900">{(churnRisk.churn_risk_score * 100).toFixed(0)}%</div>
                    </div>
				</div>
                
                {#if churnRisk.primary_factors && churnRisk.primary_factors.length > 0}
                    <div class="space-y-1">
                        <div class="text-sm font-medium text-slate-700">Primary Factors</div>
                        <ul class="list-inside list-disc text-sm text-slate-600">
                            {#each churnRisk.primary_factors as factor}
                                <li>{factor}</li>
                            {/each}
                        </ul>
                    </div>
                {/if}

                <div class="rounded-md bg-slate-50 p-3 text-sm text-slate-700">
                    <div class="mb-1 font-medium text-slate-900 flex items-center gap-2">
                        <CheckCircle2 class="h-3 w-3 text-emerald-600" /> Retention Strategy
                    </div>
                    {churnRisk.retention_strategy}
                </div>

                {#if churnRisk.suggested_discount > 0}
                    <div class="rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
                        <div class="font-medium flex items-center gap-2">
                            <AlertTriangle class="h-3 w-3" /> Suggested Action
                        </div>
                        Offer a {churnRisk.suggested_discount}% discount to retain this customer.
                    </div>
                {/if}
			</div>
		{/if}
	</CardContent>
</Card>
