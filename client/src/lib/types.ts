export interface BaseEntity {
	ID: number;
	CreatedAt?: string;
	UpdatedAt?: string;
}

export interface Category extends BaseEntity {
	Name: string;
	SubCategories?: SubCategory[];
}

export interface SubCategory extends BaseEntity {
	Name: string;
	CategoryID: number;
	Category: Category;
}

export interface Supplier extends BaseEntity {
	Name: string;
	ContactPerson?: string;
	Email?: string;
	Phone?: string;
	Address?: string;
}

export interface Location extends BaseEntity {
	Name: string;
	Address?: string;
}

// Your existing Product type
export interface Product extends BaseEntity {
	SKU: string;
	Name: string;
	Description?: string;
	CategoryID?: number;
	SubCategoryID?: number;
	SupplierID?: number;
	Brand?: string;
	PurchasePrice?: number;
	SellingPrice?: number;
	BarcodeUPC?: string;
	ImageURLs?: string;
	Status?: string;
	LocationID?: number;
	Category?: Category;
	SubCategory?: SubCategory;
	Supplier?: Supplier;
	Location?: Location;
}

// The paginated response type
export interface PaginatedProducts {
	currentPage: number;
	itemsPerPage: number;
	totalItems: number;
	totalPages: number;
	products: Product[];
}


export interface Batch extends BaseEntity {
	ProductID: number;
	BatchNumber: string;
	Quantity: number;
	ExpiryDate?: string | null;
}

export interface Alert extends BaseEntity {
	ProductID: number;
	Type: string;
	Message: string;
	TriggeredAt: string;
	Status: string;
	BatchID?: number;
	Product?: Product;
}

export interface ReorderSuggestion extends BaseEntity {
	ProductID: number;
	SupplierID: number;
	CurrentStock: number;
	PredictedDemand: number;
	SuggestedOrderQuantity: number;
	LeadTimeDays: number;
	Status: string;
	SuggestedAt: string;
	Product?: Product;
	Supplier?: Supplier;
}

export interface PurchaseOrder extends BaseEntity {
	SupplierID: number;
	Status: string;
	OrderDate: string;
	ExpectedDeliveryDate?: string | null;
	ActualDeliveryDate?: string | null;
	CreatedBy?: number;
	ApprovedBy?: number | null;
	PurchaseOrderItems?: PurchaseOrderItem[];
}

export interface PurchaseOrderItem extends BaseEntity {
	PurchaseOrderID: number;
	ProductID: number;
	OrderedQuantity: number;
	ReceivedQuantity: number;
	UnitPrice: number;
	Product?: Product;
}

export interface StockTransfer extends BaseEntity {
	ProductID: number;
	SourceLocationID: number;
	DestLocationID: number;
	Quantity: number;
	Status?: string;
}

export interface StockAdjustment extends BaseEntity {
	ProductID: number;
	LocationID: number;
	Type: string;
	Quantity: number;
	ReasonCode: string;
	Notes: string;
	AdjustedBy: number;
	AdjustedAt: string;
	PreviousQuantity: number;
	NewQuantity: number;
}

export interface DemandForecast extends BaseEntity {
	ProductID: number;
	ForecastPeriod: string;
	PredictedDemand: number;
	GeneratedAt: string;
	Product?: Product;
}

export interface BulkImportJob {
	jobId: string;
	status: string;
	message?: string;
	totalRecords?: number;
	validRecords?: number;
	invalidRecords?: number;
	errors?: Array<Record<string, unknown>>;
	preview?: Array<Record<string, unknown>>;
	filePath?: string;
}

export interface UserSummary extends BaseEntity {
	Username: string;
	Role: string;
	IsActive: boolean;
}

export interface SupplierPerformance {
	supplierId: number;
	supplierName: string;
	averageLeadTimeDays: number;
	onTimeDeliveryRate: number;
}

export interface Notification extends BaseEntity {
	UserID: number;
	Type: string;
	Title: string;
	Message: string;
	Payload?: string | null;
	IsRead: boolean;
	ReadAt?: string | null;
	TriggeredAt: string;
}
