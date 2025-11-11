<!-- client/src/lib/components/DetailsModal.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import * as Card from '$lib/components/ui/card';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { onMount } from 'svelte';
  import { api } from '$lib/api';

  export let open = false;
  export let resourceId: string | number | null = null;
  export let endpoint: string; // e.g., '/products'
  export let title = 'Details';

  const dispatch = createEventDispatcher();

  let data: any = null;
  let loading = false;

  async function fetchData() {
    if (!resourceId) return;
    loading = true;
    try {
      // Assuming your api utility has a get method
      const response = await api.get(`${endpoint}/${resourceId}`);
      data = response;
    } catch (error) {
      console.error('Error fetching details:', error);
      // Ideally, handle this with a user-facing error message
    } finally {
      loading = false;
    }
  }

  function handleOpenChange(newOpen: boolean) {
    open = newOpen;
    if (!open) {
      resourceId = null;
      data = null;
      dispatch('close');
    }
  }

  $: if (open && resourceId) {
    fetchData();
  }
</script>

<Dialog.Root bind:open={open} onOpenChange={handleOpenChange}>
  <Dialog.Content class="sm:max-w-[600px]">
    <Dialog.Header>
      <Dialog.Title>{title}</Dialog.Title>
    </Dialog.Header>
    <div class="p-4 max-h-[70vh] overflow-y-auto">
      {#if loading}
        <div class="space-y-4">
          <Skeleton class="h-8 w-3/4" />
          <Skeleton class="h-4 w-1/2" />
          <Skeleton class="h-4 w-1/3" />
          <div class="flex gap-4 pt-4">
            <Skeleton class="h-20 w-1/2" />
            <Skeleton class="h-20 w-1/2" />
          </div>
        </div>
      {:else if data}
        <slot {data} />
      {:else if !loading && !data}
        <p>Could not load details for the selected item.</p>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>
