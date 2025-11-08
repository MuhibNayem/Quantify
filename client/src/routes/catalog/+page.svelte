<script lang="ts">
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { productsApi, categoriesApi, subCategoriesApi, suppliersApi, locationsApi } from '$lib/api/resources';
	import type { Category, Location, Product, SubCategory, Supplier } from '$lib/types';
	import { PlusCircle, RefreshCcw } from 'lucide-svelte';

	type TabKey = 'products' | 'categories' | 'sub-categories' | 'suppliers' | 'locations';

	let activeTab = $state<TabKey>('products');
	let loading = $state(false);

	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let subCategories = $state<SubCategory[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let locations = $state<Location[]>([]);

	const productForm = $state({
		sku: '',
		name: '',
		description: '',
		categoryId: '',
		subCategoryId: '',
		supplierId: '',
		locationId: '',
		purchasePrice: '',
		sellingPrice: '',
		status: 'Active',
	});
	let editingProduct: Product | null = null;

	const categoryForm = $state({ name: '' });
	let editingCategory: Category | null = null;

	const subCategoryForm = $state({ name: '', categoryId: '' });	let editingSubCategory: SubCategory | null = null;

	const supplierForm = $state({ name: '', contactPerson: '', email: '', phone: '', address: '' });
	let editingSupplier: Supplier | null = null;

	const locationForm = $state({ name: '', address: '' });
	let editingLocation: Location | null = null;

	const loadAll = async () => {
		console.log('loadAll called');
		loading = true;
		try {
			const categoryList = await categoriesApi.list();
			categories = Array.isArray(categoryList) ? categoryList : [categoryList];

			const subCategoryList = await subCategoriesApi.list();
			subCategories = Array.isArray(subCategoryList) ? subCategoryList : [subCategoryList];

			const supplierList = await suppliersApi.list();
			suppliers = Array.isArray(supplierList) ? supplierList : [supplierList];

			const locationList = await locationsApi.list();
			locations = Array.isArray(locationList) ? locationList : [locationList];

			const productList = await productsApi.list();
			products = productList;
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to load catalog';
			toast.error('Failed to Load Catalog', {
				description: errorMessage,
			});
		} finally {
			loading = false;
		}
	};

	onMount(loadAll);

	const resetProductForm = () => {
		editingProduct = null;
		productForm.sku = '';
		productForm.name = '';
		productForm.description = '';
		productForm.categoryId = '';
		productForm.subCategoryId = '';
		productForm.supplierId = '';
		productForm.locationId = '';
		productForm.purchasePrice = '';
		productForm.sellingPrice = '';
		productForm.status = 'Active';
	};

	const editProduct = (product: Product) => {
		editingProduct = product;
		productForm.sku = product.SKU;
		productForm.name = product.Name;
		productForm.description = product.Description ?? '';
		productForm.categoryId = String(product.CategoryID ?? '');
		productForm.subCategoryId = String(product.SubCategoryID ?? '');
		productForm.supplierId = String(product.SupplierID ?? '');
		productForm.locationId = String(product.LocationID ?? '');
		productForm.purchasePrice = product.PurchasePrice ? String(product.PurchasePrice) : '';
		productForm.sellingPrice = product.SellingPrice ? String(product.SellingPrice) : '';
		productForm.status = product.Status ?? 'Active';
	};

	const saveProduct = async () => {
		const payload = {
			sku: productForm.sku,
			name: productForm.name,
			description: productForm.description,
			categoryId: productForm.categoryId ? Number(productForm.categoryId) : undefined,
			subCategoryId: productForm.subCategoryId ? Number(productForm.subCategoryId) : undefined,
			supplierId: productForm.supplierId ? Number(productForm.supplierId) : undefined,
			locationId: productForm.locationId ? Number(productForm.locationId) : undefined,
			purchasePrice: productForm.purchasePrice ? Number(productForm.purchasePrice) : undefined,
			sellingPrice: productForm.sellingPrice ? Number(productForm.sellingPrice) : undefined,
			status: productForm.status,
		};
		try {
			if (editingProduct) {
				await productsApi.update(editingProduct.ID, payload);
				toast.success('Product updated');
			} else {
				await productsApi.create(payload);
				toast.success('Product created');
			}
			await loadAll();
			resetProductForm();
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to save product';
			toast.error('Failed to Save Product', {
				description: errorMessage,
			});
		}
	};

	const deleteProduct = async (product: Product) => {
		toast.confirm(`Are you sure you want to delete ${product.Name}?`, {
			onConfirm: async () => {
				try {
					await productsApi.remove(product.ID);
					toast.success('Product removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete product';
					toast.error('Failed to Delete Product', {
						description: errorMessage,
					});
				}
			},
		});
	};

	const resetCategoryForm = () => {
		editingCategory = null;
		categoryForm.name = '';
	};

	const saveCategory = async () => {
		try {
			if (editingCategory) {
				await categoriesApi.update(editingCategory.ID, { name: categoryForm.name });
			} else {
				await categoriesApi.create({ name: categoryForm.name });
			}
			toast.success('Category saved');
			await loadAll();
			resetCategoryForm();
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to save category';
			toast.error('Failed to Save Category', {
				description: errorMessage,
			});
		}
	};

	const deleteCategory = async (category: Category) => {
		toast.confirm(`Are you sure you want to delete ${category.Name}?`, {
			onConfirm: async () => {
				try {
					await categoriesApi.remove(category.ID);
					toast.success('Category removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete category';
					toast.error('Failed to Delete Category', {
						description: errorMessage,
					});
				}
			},
		});
	};

	const resetSubCategoryForm = () => {
		editingSubCategory = null;
		subCategoryForm.name = '';
		subCategoryForm.categoryId = '';
	};

	const saveSubCategory = async () => {
		try {
			const payload = {
				name: subCategoryForm.name,
				categoryId: Number(subCategoryForm.categoryId),
			};
			if (editingSubCategory) {
				await subCategoriesApi.update(editingSubCategory.ID, payload);
			} else {
				await subCategoriesApi.create(payload);
			}
			toast.success('Sub-category saved');
			await loadAll();
			resetSubCategoryForm();
		} catch (error) {
			const errorMessage = error?.response?.data?.error || 'Unable to save sub-category';
			toast.error('Failed to Save Sub-Category', {
				description: errorMessage,
			});
		}
	};

	const deleteSubCategory = async (subCategory: SubCategory) => {
		toast.confirm(`Are you sure you want to delete ${subCategory.Name}?`, {
			onConfirm: async () => {
				try {
					await subCategoriesApi.remove(subCategory.ID);
					toast.success('Sub-category removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete sub-category';
					toast.error('Failed to Delete Sub-Category', {
						description: errorMessage,
					});
				}
			},
		});
	};

	const resetSupplierForm = () => {
		editingSupplier = null;
		supplierForm.name = '';
		supplierForm.contactPerson = '';
		supplierForm.email = '';
		supplierForm.phone = '';
		supplierForm.address = '';
	};

	const saveSupplier = async () => {
		try {
			if (editingSupplier) {
				await suppliersApi.update(editingSupplier.ID, supplierForm);
			} else {
				await suppliersApi.create(supplierForm);
			}
			toast.success('Supplier saved');
			await loadAll();
			resetSupplierForm();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to save supplier';
			toast.error('Failed to Save Supplier', {
				description: errorMessage,
			});
		}
	};

	const deleteSupplier = async (supplier: Supplier) => {
		toast.confirm(`Are you sure you want to delete ${supplier.Name}?`, {
			onConfirm: async () => {
				try {
					await suppliersApi.remove(supplier.ID);
					toast.success('Supplier removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete supplier';
					toast.error('Failed to Delete Supplier', {
						description: errorMessage,
					});
				}
			},
		});
	};

	const resetLocationForm = () => {
		editingLocation = null;
		locationForm.name = '';
		locationForm.address = '';
	};

	const saveLocation = async () => {
		try {
			if (editingLocation) {
				await locationsApi.update(editingLocation.ID, locationForm);
			} else {
				await locationsApi.create(locationForm);
			}
			toast.success('Location saved');
			await loadAll();
			resetLocationForm();
		} catch (error) {
			const errorMessage = error.response?.data?.error || 'Unable to save location';
			toast.error('Failed to Save Location', {
				description: errorMessage,
			});
		}
	};

	const deleteLocation = async (location: Location) => {
		toast.confirm(`Are you sure you want to delete ${location.Name}?`, {
			onConfirm: async () => {
				try {
					await locationsApi.remove(location.ID);
					toast.success('Location removed');
					await loadAll();
				} catch (error: any) {
					const errorMessage = error.response?.data?.error || 'Unable to delete location';
					toast.error('Failed to Delete Location', {
						description: errorMessage,
					});
				}
			},
		});
	};
	
</script>

<div class="w-full max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8 bg-white dark:bg-slate-900 shadow-xl rounded-2xl">
  <section class="space-y-8">
    <header class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="text-sm uppercase tracking-wide text-muted-foreground">Catalog cockpit</p>
        <h1 class="text-3xl font-semibold">Products, categories & partners</h1>
      </div>
      <div class="flex gap-2">
        <Button variant="secondary" onclick={loadAll}>
          <RefreshCcw class="mr-2 h-4 w-4" /> Sync data
        </Button>
        <Button variant="outline" href="/bulk" class="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200">
          <PlusCircle class="mr-2 h-4 w-4" /> Bulk import
        </Button>
      </div>
    </header>

    <div class="grid gap-4 md:grid-cols-5">
      <Button variant={activeTab === 'products' ? 'default' : 'secondary'} onclick={() => (activeTab = 'products')} class={activeTab === 'products' ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200' : ''}>
        Products
      </Button>
      <Button variant={activeTab === 'categories' ? 'default' : 'secondary'} onclick={() => (activeTab = 'categories')} class={activeTab === 'categories' ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200' : ''}>
        Categories
      </Button>
      <Button variant={activeTab === 'sub-categories' ? 'default' : 'secondary'} onclick={() => (activeTab = 'sub-categories')} class={activeTab === 'sub-categories' ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200' : ''}>
        Sub Categories
      </Button>
      <Button variant={activeTab === 'suppliers' ? 'default' : 'secondary'} onclick={() => (activeTab = 'suppliers')} class={activeTab === 'suppliers' ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200' : ''}>
        Suppliers
      </Button>
      <Button variant={activeTab === 'locations' ? 'default' : 'secondary'} onclick={() => (activeTab = 'locations')} class={activeTab === 'locations' ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200' : ''}>
        Locations
      </Button>
    </div>

    {#key activeTab}
    {#if activeTab === 'products'}
      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>SKU registry</CardTitle>
            <CardDescription>Manage items synced with the warehouse</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>SKU</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead class="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if loading}
                  {#each Array(4) as _, i}
                    <TableRow>
                      <TableCell colspan="4"><Skeleton class="h-6 w-full" /></TableCell>
                    </TableRow>
                  {/each}
                {:else}
                  {#each products as product}
                    <TableRow>
                      <TableCell class="font-mono text-xs">{product.SKU}</TableCell>
                      <TableCell>{product.Name}</TableCell>
                      <TableCell>
                        <span class="rounded-full bg-muted px-2 py-0.5 text-xs capitalize">{product.Status ?? 'active'}</span>
                      </TableCell>
                      <TableCell class="text-right space-x-1">
                        <Button size="sm" variant="ghost" onclick={() => editProduct(product)}>Edit</Button>
                        <Button size="sm" variant="ghost" class="text-destructive" onclick={() => deleteProduct(product)}>
                          Delete
                        </Button>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>{editingProduct ? 'Update product' : 'Create product'}</CardTitle>
            <CardDescription>SKU-level metadata</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3">
            <Input placeholder="SKU" bind:value={productForm.sku} />
            <Input placeholder="Name" bind:value={productForm.name} />
            <Input placeholder="Description" bind:value={productForm.description} />
            <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={productForm.categoryId}>
              <option value="">Select category</option>
              {#each categories as category}
                <option value={category.ID}>{category.Name}</option>
              {/each}
            </select>
            <div class="flex items-center gap-2">
              <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={productForm.subCategoryId}>
                <option value="">Select sub-category</option>
                {#each subCategories.filter(sc => sc.CategoryID === Number(productForm.categoryId)) as subCategory}
                  <option value={subCategory.ID}>{subCategory.Name}</option>
                {/each}
              </select>
              <Button size="sm" variant="outline" onclick={() => activeTab = 'sub-categories'}>New</Button>
            </div>
            <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={productForm.supplierId}>
              <option value="">Select supplier</option>
              {#each suppliers as supplier}
                <option value={supplier.ID}>{supplier.Name}</option>
              {/each}
            </select>
            <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={productForm.locationId}>
              <option value="">Default location</option>
              {#each locations as location}
                <option value={location.ID}>{location.Name}</option>
              {/each}
            </select>
            <div class="grid grid-cols-2 gap-2">
              <Input type="number" min="0" step="0.01" placeholder="Purchase price" bind:value={productForm.purchasePrice} />
              <Input type="number" min="0" step="0.01" placeholder="Selling price" bind:value={productForm.sellingPrice} />
            </div>
            <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={productForm.status}>
              <option value="Active">Active</option>
              <option value="Archived">Archived</option>
              <option value="Discontinued">Disoption>
            </select>
            <div class="flex gap-2">
              <Button class="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" onclick={saveProduct}>{editingProduct ? 'Update' : 'Create'}</Button>
              <Button class="w-full" variant="outline" onclick={resetProductForm}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    {:else if activeTab === 'categories'}
      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Categories</CardTitle>
            <CardDescription>Structure your catalog foundation</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead class="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if loading}
                  {#each Array(3) as _, i}
                    <TableRow>
                      <TableCell colspan="2"><Skeleton class="h-6 w-full" /></TableCell>
                    </TableRow>
                  {/each}
                {:else if categories.length === 0}
                  <TableRow>
                    <TableCell colspan='2' class="text-center text-sm text-muted-foreground">No categories found</TableCell>
                  </TableRow>
                {:else}
                  {#each categories as category}
                    <TableRow>
                      <TableCell>{category.Name}</TableCell>
                      <TableCell class="text-right space-x-1">
                        <Button size="sm" variant="ghost" onclick={() => { editingCategory = category; categoryForm.name = category.Name; }}>Edit</Button>
                        <Button size="sm" variant="ghost" class="text-destructive" onclick={() => deleteCategory(category)}>Delete</Button>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>{editingCategory ? 'Update category' : 'Create category'}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Input placeholder="Name" bind:value={categoryForm.name} />
            <div class="flex gap-2">
              <Button class="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" onclick={saveCategory}>{editingCategory ? 'Update' : 'Create'}</Button>
              <Button class="w-full" variant="outline" onclick={resetCategoryForm}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    {:else if activeTab === 'sub-categories'}
      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Sub Categories</CardTitle>
            <CardDescription>Refine your catalog structure</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Category</TableHead>
                  <TableHead class="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if loading}
                  {#each Array(3) as _, i}
                    <TableRow>
                      <TableCell colspan="3"><Skeleton class="h-6 w-full" /></TableCell>
                    </TableRow>
                  {/each}
                {:else if subCategories.length === 0}
                  <TableRow>
                    <TableCell colspan="3" class="text-center text-sm text-muted-foreground">No sub-categories found</TableCell>
                  </TableRow>
                {:else}
                  {#each subCategories as subCategory}
                    <TableRow>
                      <TableCell>{subCategory.Name}</TableCell>
                      <TableCell>{subCategory.Category.Name}</TableCell>
                      <TableCell class="text-right space-x-1">
                        <Button size="sm" variant="ghost" onclick={() => { editingSubCategory = subCategory; subCategoryForm.name = subCategory.Name; subCategoryForm.categoryId = String(subCategory.CategoryID); }}>Edit</Button>
                        <Button size="sm" variant="ghost" class="text-destructive" onclick={() => deleteSubCategory(subCategory)}>Delete</Button>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>{editingSubCategory ? 'Update sub-category' : 'Create sub-category'}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Input placeholder="Name" bind:value={subCategoryForm.name} />
            <select class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" bind:value={subCategoryForm.categoryId}>
              <option value="">Select category</option>
              {#each categories as category}
                <option value={category.ID}>{category.Name}</option>
              {/each}
            </select>
            <div class="flex gap-2">
              <Button class="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" onclick={saveSubCategory}>{editingSubCategory ? 'Update' : 'Create'}</Button>
              <Button class="w-full" variant="outline" onclick={resetSubCategoryForm}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    {:else if activeTab === 'suppliers'}
      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Suppliers</CardTitle>
            <CardDescription>Strategic partners powering replenishment</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Contact</TableHead>
                  <TableHead class="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if loading}
                  {#each Array(3) as _, i}
                    <TableRow>
                      <TableCell colspan="3"><Skeleton class="h-6 w-full" /></TableCell>
                    </TableRow>
                  {/each}
                {:else if suppliers.length === 0}
                  <TableRow>
                    <TableCell colspan="3" class="text-center text-sm text-muted-foreground">No suppliers found</TableCell>
                  </TableRow>
                {:else}
                  {#each suppliers as supplier}
                    <TableRow>
                      <TableCell>{supplier.Name}</TableCell>
                      <TableCell>
                        <p class="text-sm">{supplier.ContactPerson ?? '—'}</p>
                        <p class="text-xs text-muted-foreground">{supplier.Email ?? supplier.Phone ?? ''}</p>
                      </TableCell>
                      <TableCell class="text-right space-x-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onclick={() => {
                            editingSupplier = supplier;
                            Object.assign(supplierForm, {
                              name: supplier.Name,
                              contactPerson: supplier.ContactPerson ?? '',
                              email: supplier.Email ?? '',
                              phone: supplier.Phone ?? '',
                              address: supplier.Address ?? '',
                            });
                          }}
                        >
                          Edit
                        </Button>
                        <Button size="sm" variant="ghost" class="text-destructive" onclick={() => deleteSupplier(supplier)}>Delete</Button>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>{editingSupplier ? 'Update supplier' : 'Create supplier'}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Input placeholder="Name" bind:value={supplierForm.name} />
            <Input placeholder="Contact person" bind:value={supplierForm.contactPerson} />
            <Input placeholder="Email" bind:value={supplierForm.email} />
            <Input placeholder="Phone" bind:value={supplierForm.phone} />
            <Input placeholder="Address" bind:value={supplierForm.address} />
            <div class="flex gap-2">
              <Button class="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" onclick={saveSupplier}>{editingSupplier ? 'Update' : 'Create'}</Button>
              <Button class="w-full" variant="outline" onclick={resetSupplierForm}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    {:else}
      <div class="grid gap-6 lg:grid-cols-[2fr,1fr]">
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>Locations</CardTitle>
            <CardDescription>Fulfilment nodes and stores</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Address</TableHead>
                  <TableHead class="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if loading}
                  {#each Array(3) as _, i}
                    <TableRow>
                      <TableCell colspan="3"><Skeleton class="h-6 w-full" /></TableCell>
                    </TableRow>
                  {/each}
                {:else if locations.length === 0}
                  <TableRow>
                    <TableCell colspan="3" class="text-center text-sm text-muted-foreground">No locations found</TableCell>
                  </TableRow>
                {:else}
                  {#each locations as location}
                    <TableRow>
                      <TableCell>{location.Name}</TableCell>
                      <TableCell class="text-sm text-muted-foreground">{location.Address ?? '—'}</TableCell>
                      <TableCell class="text-right space-x-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onclick={() => {
                            editingLocation = location;
                            locationForm.name = location.Name;
                            locationForm.address = location.Address ?? '';
                          }}
                        >
                          Edit
                        </Button>
                        <Button size="sm" variant="ghost" class="text-destructive" onclick={() => deleteLocation(location)}>Delete</Button>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card class="shadow-lg rounded-xl">
          <CardHeader>
            <CardTitle>{editingLocation ? 'Update location' : 'Create location'}</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Input placeholder="Name" bind:value={locationForm.name} />
            <Input placeholder="Address" bind:value={locationForm.address} />
            <div class="flex gap-2">
              <Button class="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold shadow-lg hover:shadow-xl transition-all duration-200" onclick={saveLocation}>{editingLocation ? 'Update' : 'Create'}</Button>
              <Button class="w-full" variant="outline" onclick={resetLocationForm}>Reset</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    {/if}
    {/key}
  </section>
</div>
</div>
