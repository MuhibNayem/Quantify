<script lang="ts">
    import { onMount } from 'svelte';
    import { Input } from '$lib/components/ui/input';
    import { Button } from '$lib/components/ui/button';
    import { Skeleton } from '$lib/components/ui/skeleton';
    import { Badge } from '$lib/components/ui/badge';
    import {
        UserPlus,
        Search,
        Star,
        Gift,
        Edit,
        Trash2,
        Users,
        Mail,
        Phone,
        Crown,
        Award,
        Zap,
        Filter,
        ChevronRight,
        ChevronLeft,

		Edit2

    } from 'lucide-svelte';
    import { toast } from 'svelte-sonner';
    import * as Dialog from '$lib/components/ui/dialog';
    import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
    import { Root, Content, Item, PrevButton, NextButton, Ellipsis, Link } from '$lib/components/ui/pagination';

    let customers: any[] = [];
    let filteredCustomers: any[] = [];
    let selectedCustomer: any = null;
    let searchTerm = '';
    let loading = true;
    let isModalOpen = false;
    let editingCustomer: any = null;
    let pointsToAdd = 0;
    let pointsToRedeem = 0;
    let showFilters = false;
    let selectedTier = 'all';

    const pagination = {
        currentPage: 1,
        totalPages: 1,
        totalItems: 0,
        itemsPerPage: 5
    };

    const customerForm = {
        name: '',
        email: '',
        phone: ''
    };

    const openModal = (customer: any = null) => {
        if (customer) {
            editingCustomer = customer;
            customerForm.name = customer.name;
            customerForm.email = customer.email;
            customerForm.phone = customer.phone;
        } else {
            editingCustomer = null;
            customerForm.name = '';
            customerForm.email = '';
            customerForm.phone = '';
        }
        isModalOpen = true;
    };

    const saveCustomer = () => {
        if (editingCustomer) {
            const index = customers.findIndex((c) => c.id === editingCustomer.id);
            if (index !== -1) {
                customers[index] = { ...customers[index], ...customerForm };
                toast.success('Customer updated');
            }
        } else {
            const newCustomer = {
                id: customers.length + 1,
                ...customerForm,
                loyalty: { points: 0, tier: 'Bronze' },
                joinedDate: new Date().toISOString().split('T')[0],
                lastOrder: new Date().toISOString().split('T')[0]
            };
            customers = [...customers, newCustomer];
            toast.success('Customer created');
        }
        handleSearch();
        isModalOpen = false;
    };

    const deleteCustomer = (customer: any) => {
		console.log({customer})
        toast.confirm(`Are you sure you want to delete ${customer.name}?`, {
            onConfirm: () => {
                customers = customers.filter((c) => c.id !== customer.id);
                handleSearch();
                if (selectedCustomer?.id === customer.id) {
                    selectedCustomer = null;
                }
                toast.success('Customer deleted');
            }
        });
    };

    const addPoints = () => {
        if (selectedCustomer && pointsToAdd > 0) {
            selectedCustomer.loyalty.points += pointsToAdd;
            toast.success(`${pointsToAdd} points added to ${selectedCustomer.name}`);
            pointsToAdd = 0;
        }
    };

    const redeemPoints = () => {
        if (selectedCustomer && pointsToRedeem > 0) {
            if (selectedCustomer.loyalty.points >= pointsToRedeem) {
                selectedCustomer.loyalty.points -= pointsToRedeem;
                toast.success(`${pointsToRedeem} points redeemed for ${selectedCustomer.name}`);
                pointsToRedeem = 0;
            } else {
                toast.error('Insufficient points');
            }
        }
    };

    const fetchCustomers = async () => {
        setTimeout(() => {
            customers = [
                {
                    id: 1,
                    name: 'John Doe',
                    email: 'john.doe@example.com',
                    phone: '123-456-7890',
                    loyalty: { points: 150, tier: 'Silver' },
                    joinedDate: '2023-01-15',
                    lastOrder: '2024-01-10'
                },
                {
                    id: 2,
                    name: 'Jane Smith',
                    email: 'jane.smith@example.com',
                    phone: '987-654-3210',
                    loyalty: { points: 500, tier: 'Gold' },
                    joinedDate: '2022-11-20',
                    lastOrder: '2024-01-12'
                },
                {
                    id: 3,
                    name: 'Peter Jones',
                    email: 'peter.jones@example.com',
                    phone: '555-555-5555',
                    loyalty: { points: 20, tier: 'Bronze' },
                    joinedDate: '2023-08-05',
                    lastOrder: '2023-12-28'
                },
                {
                    id: 4,
                    name: 'Mary Johnson',
                    email: 'mary.j@example.com',
                    phone: '111-222-3333',
                    loyalty: { points: 75, tier: 'Silver' },
                    joinedDate: '2023-03-10',
                    lastOrder: '2024-01-05'
                },
                {
                    id: 5,
                    name: 'David Williams',
                    email: 'dave.w@example.com',
                    phone: '444-555-6666',
                    loyalty: { points: 900, tier: 'Gold' },
                    joinedDate: '2021-05-25',
                    lastOrder: '2024-01-15'
                },
                {
                    id: 6,
                    name: 'Linda Brown',
                    email: 'linda.b@example.com',
                    phone: '777-888-9999',
                    loyalty: { points: 10, tier: 'Bronze' },
                    joinedDate: '2023-10-01',
                    lastOrder: '2023-11-20'
                }
            ];
            handleSearch();
            loading = false;
        }, 1000);
    };

    const handleSearch = () => {
        let filtered = customers;

        if (searchTerm) {
            filtered = filtered.filter(
                (c) =>
                    c.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                    c.email.toLowerCase().includes(searchTerm.toLowerCase())
            );
        }

        if (selectedTier !== 'all') {
            filtered = filtered.filter((c) => c.loyalty.tier === selectedTier);
        }

        filteredCustomers = filtered;
        pagination.totalItems = filteredCustomers.length;
        pagination.totalPages = Math.ceil(filteredCustomers.length / pagination.itemsPerPage);
        handlePageChange(1);
    };

    const handlePageChange = (page: number) => {
        pagination.currentPage = page;
    };

    const getTierColor = (tier: string) => {
        switch (tier) {
            case 'Bronze':
                return 'bg-orange-100 text-orange-800 border-orange-200';
            case 'Silver':
                return 'bg-indigo-100 text-indigo-800 border-indigo-200';
            case 'Gold':
                return 'bg-amber-400 text-amber-900 border-amber-500';
            default:
                return 'bg-gray-100 text-gray-800 border-gray-200';
        }
    };

    const getTierIcon = (tier: string) => {
        switch (tier) {
            case 'Bronze':
                return Award;
            case 'Silver':
                return Star;
            case 'Gold':
                return Crown;
            default:
                return Star;
        }
    };

    onMount(() => {
        fetchCustomers();
    });

    $: paginatedCustomers = filteredCustomers.slice(
        (pagination.currentPage - 1) * pagination.itemsPerPage,
        pagination.currentPage * pagination.itemsPerPage
    );
</script>

<!-- ===== HERO SECTION ===== -->
<section class="relative isolate w-full overflow-hidden">
    <!-- Gradient background with motion -->
    <div
        class="animate-gradientShift absolute inset-0 -z-10 bg-gradient-to-r from-violet-400 via-purple-400 to-indigo-500 bg-[length:200%_200%]"
    ></div>

    <!-- Floating glow blobs -->
    <div
        class="animate-pulseGlow absolute -top-32 -left-24 h-96 w-96 rounded-full bg-violet-300/50 blur-3xl"
    ></div>
    <div
        class="animate-pulseGlow absolute -right-24 -bottom-28 h-80 w-80 rounded-full bg-indigo-300/40 blur-3xl delay-700"
    ></div>

    <!-- Hero container -->
    <div class="relative mx-auto max-w-7xl px-4 pt-16 pb-10 sm:px-6 sm:pt-20 sm:pb-16 lg:px-8">
        <div class="flex flex-col items-start justify-between gap-6 lg:flex-row lg:items-center">
            <div>
                <div class="mb-3 inline-flex items-center gap-3">
                    <span
                        class="animate-cardFloat inline-flex rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 p-2 shadow-md"
                    >
                        <Users class="h-6 w-6 text-white" />
                    </span>
                    <p class="text-xs font-semibold tracking-[0.18em] text-white uppercase sm:text-sm">
                        Customer Loyalty
                    </p>
                </div>

                <h1
                    class="mb-3 bg-gradient-to-r from-white via-gray-100 to-gray-200 bg-clip-text text-3xl font-bold text-transparent sm:text-4xl lg:text-5xl"
                >
                    Customer Management
                </h1>
                <p class="max-w-2xl text-sm text-white/90 sm:text-base">
                    Manage customer relationships and loyalty programs
                </p>
            </div>

            <!-- Action buttons -->
            <div class="flex flex-col gap-3 sm:flex-row">
                <Button
                    variant="secondary"
                    onclick={fetchCustomers}
                    class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-5 py-2.5 font-medium text-white shadow-lg transition-all duration-300 hover:scale-105 hover:from-violet-700 hover:to-indigo-700 hover:shadow-xl focus:ring-2 focus:ring-violet-300"
                >
                    <Zap class="mr-2 h-4 w-4" /> Refresh
                </Button>
                <Button
                    onclick={() => openModal()}
                    class="rounded-xl border border-white/30 bg-white/20 px-5 py-2.5 font-medium text-white shadow-md transition-all duration-300 hover:scale-105 hover:bg-white/30 hover:shadow-lg focus:ring-2 focus:ring-white/50"
                >
                    <UserPlus class="mr-2 h-4 w-4" /> Add Customer
                </Button>
            </div>
        </div>
    </div>
</section>

<!-- ===== SEARCH AND FILTERS ===== -->
<div class="mx-auto mt-6 max-w-7xl px-4 sm:px-6 lg:px-8">
    <div class="rounded-2xl border border-violet-200/50 bg-white/90 p-6 shadow-lg backdrop-blur">
        <div class="flex flex-col items-start justify-between gap-4 lg:flex-row lg:items-center">
            <div class="w-full flex-1 lg:w-auto">
                <div class="relative">
                    <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <Input
                        bind:value={searchTerm}
                        on:input={handleSearch}
                        placeholder="Search customers by name or email..."
                        class="w-full rounded-xl border-violet-300 bg-white/90 px-3.5 py-2.5 pl-10 text-sm focus:ring-2 focus:ring-violet-500"
                    />
                </div>
            </div>

            <div class="flex items-center gap-3">
                <Button
                    variant="outline"
                    onclick={() => (showFilters = !showFilters)}
                    class="rounded-xl border-violet-300 px-4 py-2.5 text-violet-700 hover:bg-violet-50"
                >
                    <Filter class="mr-2 h-4 w-4" /> Filters
                    {#if selectedTier !== 'all'}
                        <Badge class="ml-2 bg-violet-100 text-violet-700">1</Badge>
                    {/if}
                </Button>
            </div>
        </div>

        {#if showFilters}
            <div class="mt-4 border-t border-violet-200/50 pt-4">
                <div class="flex items-center gap-3">
                    <span class="text-sm font-medium text-slate-700">Tier:</span>
                    <div class="flex gap-2">
                        {#each ['all', 'Bronze', 'Silver', 'Gold'] as tier}
                            <Button
                                variant={selectedTier === tier ? 'default' : 'ghost'}
                                size="sm"
                                class={selectedTier === tier
                                    ? 'bg-violet-600 text-white'
                                    : 'text-slate-600 hover:bg-violet-50 hover:text-violet-700'}
                                onclick={() => {
                                    selectedTier = tier;
                                    handleSearch();
                                }}
                            >
                                {tier === 'all' ? 'All Tiers' : tier}
                            </Button>
                        {/each}
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>

<!-- ===== MAIN CONTENT ===== -->
<div class="mx-auto mt-6 mb-12 max-w-7xl px-4 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
        <!-- Customer Table - Taking 2 columns -->
        <div class="lg:col-span-2">
            <div class="rounded-xl border border-violet-200/50 bg-white/90 p-6 shadow-lg backdrop-blur">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Customer</TableHead>
                            <TableHead>Contact</TableHead>
                            <TableHead>Loyalty Tier</TableHead>
                            <TableHead class="text-right">Points</TableHead>
                            <TableHead class="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {#if loading}
                            {#each Array(5) as _}
                                <TableRow>
                                    <TableCell colspan="5">
                                        <Skeleton class="h-8 w-full" />
                                    </TableCell>
                                </TableRow>
                            {/each}
                        {:else if paginatedCustomers.length === 0}
                            <TableRow>
                                <TableCell colspan="5" class="h-24 text-center">No customers found.</TableCell>
                            </TableRow>
                        {:else}
                            {#each paginatedCustomers as customer (customer.id)}
                                <TableRow 
									class="cursor-pointer hover:bg-violet-50/50 {selectedCustomer && selectedCustomer.id === customer.id ? 'bg-violet-100' : ''}"
									onclick={() => (selectedCustomer = customer)}
								>
                                    <TableCell>
                                        <div class="font-medium">{customer.name}</div>
                                        <div class="text-sm text-muted-foreground">{customer.email}</div>
                                    </TableCell>
                                    <TableCell>{customer.phone}</TableCell>
                                    <TableCell>
                                        <Badge class={getTierColor(customer.loyalty.tier)}>
                                            <svelte:component
                                                this={getTierIcon(customer.loyalty.tier)}
                                                class="mr-1 h-3 w-3"
                                            />
                                            {customer.loyalty.tier}
                                        </Badge>
                                    </TableCell>
                                    <TableCell class="text-right font-mono">{customer.loyalty.points}</TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            onclick={() => openModal(customer)}
                                            class="rounded-lg border-violet-300 text-violet-700 hover:bg-violet-50"
                                        >
                                            <Edit class="h-4 w-4" />
                                        </Button>
                                        <Button
											size="sm"
											variant="destructive"
											onclick={() => deleteCustomer(selectedCustomer)}
											class="bg-red-500/20 hover:bg-red-500/30 text-red-500 rounded-lg"
										>
											<Trash2 class="h-4 w-4 text-red-500" />
										</Button>
                                    </TableCell>
                                </TableRow>
                            {/each}
                        {/if}
                    </TableBody>
                </Table>
            </div>
            {#if pagination.totalPages > 1}
                <div
                    class="mt-6 flex flex-col items-center justify-center space-y-2 rounded-2xl border border-white/60 bg-white/90 py-6 shadow-md backdrop-blur"
                >
                    <Root
                        count={pagination.totalItems}
                        perPage={pagination.itemsPerPage}
                        page={pagination.currentPage}
                        onPageChange={(e) => handlePageChange(e.detail)}
                    >
                        <Content class="flex items-center gap-1">
                            <Item>
                                <PrevButton
                                    disabled={pagination.currentPage === 1}
                                    class="rounded-lg border border-violet-300 bg-violet-50 hover:bg-violet-100"
                                    onclick={() => handlePageChange(pagination.currentPage - 1)}
                                />
                            </Item>

                            {#each { length: pagination.totalPages } as _, i}
                                <Item>
                                    <Link
                                        page={i + 1}
                                        isActive={pagination.currentPage === i + 1}
                                        class="rounded-lg border border-violet-300 px-3 py-1.5 data-[active=true]:bg-violet-600 data-[active=true]:text-white data-[active=false]:bg-white/80 data-[active=false]:text-slate-700 hover:scale-105"
                                        onclick={() => handlePageChange(i + 1)}
                                    >
                                        {i + 1}
                                    </Link>
                                </Item>
                            {/each}
                            <Item>
                                <NextButton
                                    disabled={pagination.currentPage === pagination.totalPages}
                                    class="rounded-lg border border-violet-300 bg-violet-50 hover:bg-violet-100"
                                    onclick={() => handlePageChange(pagination.currentPage + 1)}
                                />
                            </Item>
                        </Content>
                    </Root>

                    <p class="text-sm text-slate-600">
                        Showing
                        {(pagination.currentPage - 1) * pagination.itemsPerPage + 1}
                        –
                        {Math.min(pagination.currentPage * pagination.itemsPerPage, pagination.totalItems)}
                        of {pagination.totalItems} customers
                    </p>
                </div>
            {/if}
        </div>

        <!-- Customer Details Sidebar -->
        <div class="lg:col-span-1">
            {#if selectedCustomer}
                <div class="sticky top-6 space-y-6">
                    <!-- Customer Profile Card -->
                    <div class="bg-gradient-to-br from-violet-500 to-indigo-600 rounded-2xl p-6 shadow-xl text-white">
                        <div class="flex items-center justify-between mb-4">
                            <h2 class="text-xl font-bold">{selectedCustomer.name}</h2>
                            <div class="flex gap-2">
                                <Button
                                    size="sm"
                                    variant="outline"
                                    onclick={() => openModal(selectedCustomer)}
                                    class="border-white/30 text-white hover:bg-white/20 rounded-lg"
                                >
                                    <Edit2 class="h-4 w-4" />
                                </Button>
                                <Button
									size="sm"
									variant="destructive"
									onclick={() => deleteCustomer(selectedCustomer)}
									class="bg-red-500/20 hover:bg-red-500/30 text-red-500 rounded-lg"
								>
									<Trash2 class="h-4 w-4 text-red-500" />
								</Button>
                            </div>
                        </div>

                        <div class="space-y-3">
                            <div class="flex items-center gap-3">
                                <Mail class="h-4 w-4 text-white/80" />
                                <span class="text-sm text-white/90">{selectedCustomer.email}</span>
                            </div>
                            <div class="flex items-center gap-3">
                                <Phone class="h-4 w-4 text-white/80" />
                                <span class="text-sm text-white/90">{selectedCustomer.phone}</span>
                            </div>
                        </div>
                    </div>

                    <!-- Loyalty Card -->
                    <div class="bg-white/90 backdrop-blur rounded-2xl p-6 shadow-lg border border-violet-200/50">
                        <h3 class="text-lg font-semibold text-slate-800 mb-4 flex items-center gap-2">
                            <Star class="text-violet-600" />
                            Loyalty Program
                        </h3>
                        
                        <div class="text-center mb-6">
                            <p class="text-3xl font-bold text-violet-600">{selectedCustomer.loyalty.points}</p>
                            <p class="text-sm text-slate-500">Available Points</p>
                            <Badge class={`mt-2 ${getTierColor(selectedCustomer.loyalty.tier)}`}>
                                <svelte:component this={getTierIcon(selectedCustomer.loyalty.tier)} class="h-3 w-3 mr-1" />
                                {selectedCustomer.loyalty.tier} Member
                            </Badge>
                        </div>

                        <div class="space-y-3">
                            <div>
                                <label class="text-sm font-medium text-slate-700">Add Points</label>
                                <div class="flex gap-2 mt-1">
                                    <Input
                                        type="number"
                                        placeholder="Amount"
                                        bind:value={pointsToAdd}
                                        class="border-violet-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-violet-500 bg-white/90"
                                    />
                                    <Button
                                        onclick={addPoints}
                                        class="bg-violet-600 hover:bg-violet-700 text-white rounded-lg px-3 py-2"
                                    >
                                        <Gift class="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                            
                            <div>
                                <label class="text-sm font-medium text-slate-700">Redeem Points</label>
                                <div class="flex gap-2 mt-1">
                                    <Input
                                        type="number"
                                        placeholder="Amount"
                                        bind:value={pointsToRedeem}
                                        class="border-violet-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-violet-500 bg-white/90"
                                    />
                                    <Button
                                        onclick={redeemPoints}
                                        class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg px-3 py-2"
                                    >
                                        Redeem
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Quick Stats -->
                    <div class="bg-white/90 backdrop-blur rounded-2xl p-6 shadow-lg border border-violet-200/50">
                        <h3 class="text-lg font-semibold text-slate-800 mb-4">Quick Stats</h3>
                        <div class="space-y-3">
                            <div class="flex justify-between">
                                <span class="text-sm text-slate-600">Member Since</span>
                                <span class="text-sm font-medium text-slate-800">{selectedCustomer.joinedDate}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-sm text-slate-600">Last Order</span>
                                <span class="text-sm font-medium text-slate-800">{selectedCustomer.lastOrder}</span>
                            </div>
                        </div>
                    </div>
                </div>
            {:else}
                <div class="sticky top-6">
                    <div class="bg-white/90 backdrop-blur rounded-2xl p-12 text-center border border-violet-200/50">
                        <Users class="h-16 w-16 text-violet-300 mx-auto mb-4" />
                        <p class="text-slate-600 text-lg font-medium">Select a Customer</p>
                        <p class="text-slate-500 text-sm">Choose a customer from list to view details</p>
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>

<Dialog.Root bind:open={isModalOpen}>
    <Dialog.Content class="rounded-2xl border-0 bg-gradient-to-br from-violet-50 to-indigo-100 shadow-xl">
        <Dialog.Header class="border-b border-violet-200/50 bg-white/80 px-6 py-5 backdrop-blur">
            <Dialog.Title class="text-slate-800"
                >{editingCustomer ? 'Edit Customer' : 'Create Customer'}</Dialog.Title
            >
            <Dialog.Description class="text-slate-600">
                {editingCustomer
                    ? 'Update details of existing customer.'
                    : 'Enter details for new customer.'}
            </Dialog.Description>
        </Dialog.Header>
        <div class="space-y-4 px-6 py-4">
            <div class="space-y-2">
                <label for="name" class="text-sm font-medium text-slate-700">Name</label>
                <Input
                    id="name"
                    bind:value={customerForm.name}
                    class="rounded-xl border-violet-300 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-500"
                />
            </div>
            <div class="space-y-2">
                <label for="email" class="text-sm font-medium text-slate-700">Email</label>
                <Input
                    id="email"
                    type="email"
                    bind:value={customerForm.email}
                    class="rounded-xl border-violet-300 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-500"
                />
            </div>
            <div class="space-y-2">
                <label for="phone" class="text-sm font-medium text-slate-700">Phone</label>
                <Input
                    id="phone"
                    bind:value={customerForm.phone}
                    class="rounded-xl border-violet-300 bg-white/90 px-3.5 py-2.5 text-sm focus:ring-2 focus:ring-violet-500"
                />
            </div>
        </div>
        <Dialog.Footer class="border-t border-violet-200/50 bg-white/80 px-6 py-4 backdrop-blur">
            <Button
                onclick={saveCustomer}
                class="rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-5 py-2.5 font-medium text-white shadow-md transition-all duration-300 hover:from-violet-700 hover:to-indigo-700 hover:shadow-lg"
                >Save</Button
            >
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
    /* Smooth transitions globally */
    * {
        transition-property:
            color, background-color, border-color, text-decoration-color, fill, stroke, opacity,
            box-shadow, transform, filter, backdrop-filter;
        transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
        transition-duration: 300ms;
    }

    /* Hero gradient animation */
    @keyframes gradientShift {
        0% {
            background-position: 0% 50%;
        }
        50% {
            background-position: 100% 50%;
        }
        100% {
            background-position: 0% 50%;
        }
    }
    .animate-gradientShift {
        background-size: 200% 200%;
        animation: gradientShift 16s ease-in-out infinite;
    }

    /* Soft glowing blobs */
    @keyframes pulseGlow {
        0%,
        100% {
            transform: scale(1);
            opacity: 0.45;
            filter: blur(80px);
        }
        50% {
            transform: scale(1.08);
            opacity: 0.6;
            filter: blur(90px);
        }
    }
    .animate-pulseGlow {
        animation: pulseGlow 10s ease-in-out infinite;
    }

    /* Card float micro-motion */
    @keyframes cardFloat {
        0%,
        100% {
            transform: translateY(0);
        }
        50% {
            transform: translateY(-4px);
        }
    }
    .animate-cardFloat {
        animation: cardFloat 4s ease-in-out infinite;
    }

    /* Fade-up reveal */
    @keyframes fadeUp {
        from {
            opacity: 0;
            transform: translateY(12px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    .animate-fadeUp {
        animation: fadeUp 500ms var(--ease, cubic-bezier(0.4, 0, 0.2, 1)) forwards;
    }

    /* Pastel scrollbar */
    ::-webkit-scrollbar {
        width: 8px;
        height: 8px;
    }
    ::-webkit-scrollbar-track {
        background: transparent;
    }
    ::-webkit-scrollbar-thumb {
        background: rgba(139, 92, 246, 0.25);
        border-radius: 9999px;
    }
    ::-webkit-scrollbar-thumb:hover {
        background: rgba(139, 92, 246, 0.35);
    }
</style>