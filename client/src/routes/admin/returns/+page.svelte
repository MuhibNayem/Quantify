<script lang="ts">
  import { onMount } from "svelte";
  import GlassCard from "$lib/components/ui/GlassCard.svelte";
  import { formatDate, formatCurrency } from "$lib/utils";
  import { fade } from "svelte/transition";

  let returns = $state([]);
  let loading = $state(true);

  // Mock Data
  const mockReturns = [
    {
      id: 1,
      orderNumber: "ORD-1733600000-1",
      user: "john_doe",
      reason: "Defective",
      amount: 100.00,
      status: "PENDING",
      date: "2025-12-08T09:00:00Z",
      items: [
        { name: "Wireless Headphones", quantity: 1, condition: "GOOD" }
      ]
    },
    {
      id: 2,
      orderNumber: "ORD-1733500000-1",
      user: "jane_smith",
      reason: "Changed Mind",
      amount: 45.50,
      status: "APPROVED",
      date: "2025-12-07T14:30:00Z",
      items: [
        { name: "USB-C Cable", quantity: 1, condition: "OPENED" }
      ]
    }
  ];

  onMount(async () => {
    // Simulate fetch
    setTimeout(() => {
      returns = mockReturns;
      loading = false;
    }, 500);
  });

  async function processReturn(id, action) {
    // Call API: await api.post(`/returns/${id}/process`, { action });
    console.log(`Processing return ${id}: ${action}`);
    
    // Optimistic update
    returns = returns.map(r => 
      r.id === id ? { ...r, status: action === 'approve' ? 'APPROVED' : 'REJECTED' } : r
    );
  }
</script>

<div class="container mx-auto p-6 space-y-8">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600">
      Return Requests
    </h1>
  </div>

  {#if loading}
    <div class="text-center text-gray-400">Loading requests...</div>
  {:else if returns.length === 0}
    <div class="text-center text-gray-400">No return requests found.</div>
  {:else}
    <div class="space-y-4">
      {#each returns as req (req.id)}
        <GlassCard class="transition-all hover:bg-white/15">
          <div class="flex flex-col md:flex-row justify-between gap-4">
            <!-- Info -->
            <div class="space-y-2 flex-1">
              <div class="flex items-center gap-3">
                <span class="font-semibold text-white text-lg">#{req.id}</span>
                <span class="text-purple-400 text-sm">{req.orderNumber}</span>
                <span class={
                  req.status === 'PENDING' ? 'text-yellow-400 bg-yellow-400/10 px-2 py-0.5 rounded text-xs' :
                  req.status === 'APPROVED' ? 'text-green-400 bg-green-400/10 px-2 py-0.5 rounded text-xs' :
                  'text-red-400 bg-red-400/10 px-2 py-0.5 rounded text-xs'
                }>
                  {req.status}
                </span>
              </div>
              
              <div class="text-sm text-gray-400">
                <span class="text-white">{req.user}</span> • {formatDate(req.date)}
              </div>
              
              <div class="bg-black/20 p-3 rounded-lg border border-white/5">
                <p class="text-sm text-gray-300 mb-2"><span class="text-gray-500">Reason:</span> {req.reason}</p>
                <div class="space-y-1">
                  {#each req.items as item}
                    <div class="flex justify-between text-xs text-gray-400">
                      <span>{item.quantity}x {item.name}</span>
                      <span class="text-gray-500">Condition: {item.condition}</span>
                    </div>
                  {/each}
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex flex-col justify-between items-end gap-4 min-w-[150px]">
              <span class="text-xl font-bold text-white">{formatCurrency(req.amount)}</span>
              
              {#if req.status === 'PENDING'}
                <div class="flex gap-2 w-full">
                  <button 
                    onclick={() => processReturn(req.id, 'reject')}
                    class="flex-1 px-3 py-2 rounded bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 transition-colors text-sm font-medium"
                  >
                    Reject
                  </button>
                  <button 
                    onclick={() => processReturn(req.id, 'approve')}
                    class="flex-1 px-3 py-2 rounded bg-green-500/10 text-green-400 hover:bg-green-500/20 border border-green-500/20 transition-colors text-sm font-medium"
                  >
                    Approve
                  </button>
                </div>
              {:else}
                <div class="text-sm text-gray-500 italic">
                  Processed
                </div>
              {/if}
            </div>
          </div>
        </GlassCard>
      {/each}
    </div>
  {/if}
</div>
