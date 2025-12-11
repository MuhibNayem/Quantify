export default {
    common: {
        save: 'Save',
        saved_successfully: 'Saved successfully',
        failed_to_save: 'Failed to save',
        access_denied: 'Access Denied',
        no_permission_settings: 'You do not have permission to view settings.',
        no_data_available: 'No data available',
        configuration: 'Configuration',
        manage_preferences: 'Manage system preferences, security controls, and global policies.',
    },
    settings: {
        tabs: {
            general: 'General',
            business_rules: 'Business Rules',
            system_ai: 'System & AI',
            security_roles: 'Security & Roles',
            policies: 'Policies',
            notifications: 'Notifications'
        },
        general: {
            business_profile: 'Business Profile',
            business_profile_desc: "Your organization's visible identity across the platform.",
            business_name: 'Business Name',
            currency: 'Currency',
            timezone: 'Timezone',
            locale: 'Locale / Language',
            select_currency: 'Select Currency',
            select_timezone: 'Select Timezone',
            select_locale: 'Select Locale'
        },
        business: {
            loyalty_program: 'Loyalty Program',
            loyalty_program_desc: 'Configure how customers earn and redeem points, and set tier thresholds.',
            points_configuration: 'Points Configuration',
            earning_rate: 'Earning Rate (Points per $1)',
            earning_rate_hint: 'How many points a customer earns for every unit of currency spent.',
            redemption_value: 'Redemption Value ($ per Point)',
            redemption_value_hint: 'The monetary value of a single loyalty point when redeeming.',
            tier_thresholds: 'Tier Thresholds',
            silver_tier: 'Silver Tier (Points)',
            gold_tier: 'Gold Tier (Points)',
            platinum_tier: 'Platinum Tier (Points)',
            financial_settings: 'Financial Settings',
            financial_settings_desc: 'Manage tax rates and other financial parameters.',
            default_tax_rate: 'Default Tax Rate (%)',
            default_tax_rate_hint: 'This tax rate will be applied to all applicable sales.'
        },
        policies: {
            privacy_policy: 'Privacy Policy',
            privacy_policy_placeholder: 'Enter your privacy policy (Markdown supported)...',
            save_policy: 'Save Policy',
            terms_of_service: 'Terms of Service',
            terms_of_service_placeholder: 'Enter your terms of service (Markdown supported)...',
            save_terms: 'Save Terms',
            return_policy: 'Return Policy',
            return_window: 'Return Window (Days)',
            return_window_hint: 'Number of days after purchase that a customer can request a return.'
        },
        system: {
            ai_system: 'AI & System',
            ai_system_desc: 'Configure autonomous agent behavior and system-wide parameters.',
            wake_up_time: 'AI Wake Up Time',
            wake_up_time_hint: 'The AI will run the "Daily Morning Check" at this time every day.'
        },
        notifications: {
            global_alerts_center: 'Global Alerts Center',
            coming_soon: 'Advanced notification routing and webhooks configuration is coming in the next update.'
        }
    },
    roles: {
        title: 'Roles & Access',
        create_new: 'Create New Role',
        create_first: 'Create First Role',
        system_managed: 'System Managed',
        custom_role: 'Custom Role',
        active_permissions: 'Permissions Active',
        delete_confirm: 'Are you sure you want to delete role "{name}"?',
        delete: 'Delete',
        save_changes: 'Save Changes',
        saving: 'Saving...',
        reset: 'Reset',
        role_name: 'Role Name',
        description: 'Description',
        capabilities: 'Capabilities',
        capabilities_desc: 'Fine-tune access controls for this role',
        select_all: 'Select All',
        security_access_control: 'Security & Access Control',
        security_desc: 'Select a role from the sidebar to configure permissions, or create a new custom role to delegate specific access capabilities.',
        names: {
            admin: 'Admin',
            manager: 'Manager',
            cashier: 'Cashier',
            staff: 'Staff',
            customer: 'Customer',
            sales_associate: 'Sales Associate'
        }
    },
    permissions: {
        groups: {
            inventory: 'Inventory',
            sales: 'Sales',
            crm: 'CRM',
            hrm: 'HRM',
            settings: 'Settings',
            reports: 'Reports',
            pos: 'POS',
            dashboard: 'Dashboard',
            'product management': 'Product Management',
            'access control': 'Access Control',
            orders: 'Orders',
            system: 'System'
        },
        names: {
            'roles_manage': 'Manage Roles',
            'roles_view': 'View Roles',
            'users_manage': 'Manage Users',
            'users_view': 'View Users',
            'crm_read': 'Read CRM Data',
            'crm_view': 'View CRM Module',
            'crm_write': 'Modify CRM Data',
            'customers_read': 'View Customers',
            'customers_write': 'Manage Customers',
            'loyalty_read': 'View Loyalty Info',
            'loyalty_write': 'Manage Loyalty Points',
            'alerts_manage': 'Resolve Alerts',
            'alerts_view': 'View Alerts',
            'barcode_read': 'Lookup Barcodes',
            'inventory_read': 'Read Inventory Data',
            'inventory_view': 'View Inventory Module',
            'inventory_write': 'Modify Inventory Data',
            'locations_read': 'View Locations',
            'locations_write': 'Manage Locations',
            'replenishment_read': 'View Forecasts/Suggestions',
            'replenishment_write': 'Generate Forecasts & Manage POs',
            'suppliers_read': 'View Suppliers',
            'suppliers_write': 'Manage Suppliers',
            'orders_manage': 'Manage Orders',
            'orders_read': 'View Order History',
            'pos_access': 'Access POS Terminal',
            'pos_view': 'View POS Module',
            'returns_manage': 'Approve/Reject Returns',
            'returns_request': 'Request Return',
            'categories_read': 'View Categories',
            'categories_write': 'Manage Categories',
            'products_delete': 'Delete Products',
            'products_read': 'View Products',
            'products_write': 'Create/Edit Products',
            'reports_financial': 'View Financial Reports',
            'reports_inventory': 'View Inventory Reports',
            'reports_sales': 'View Sales Reports',
            'reports_view': 'View Reports',
            'settings_manage': 'Edit System Settings',
            'settings_view': 'View System Settings',
            'bulk_export': 'Export Data',
            'bulk_import': 'Import Data',
            'dashboard_view': 'View Dashboard',
            'notifications_read': 'View Notifications',
            'notifications_write': 'Manage Notifications'
        },
        descriptions: {
            'roles_manage': 'Manage roles',
            'roles_view': 'View roles',
            'users_manage': 'Manage users',
            'users_view': 'View users',
            'crm_read': 'Read CRM data',
            'crm_view': 'View CRM module',
            'crm_write': 'Modify CRM data',
            'customers_read': 'View customers',
            'customers_write': 'Manage customers',
            'loyalty_read': 'View loyalty info',
            'loyalty_write': 'Manage loyalty points',
            'alerts_manage': 'Resolve alerts',
            'alerts_view': 'View alerts',
            'barcode_read': 'Lookup barcodes',
            'inventory_read': 'Read inventory data',
            'inventory_view': 'View inventory module',
            'inventory_write': 'Modify inventory data',
            'locations_read': 'View locations',
            'locations_write': 'Manage locations',
            'replenishment_read': 'View forecasts/suggestions',
            'replenishment_write': 'Generate forecasts and manage POs',
            'suppliers_read': 'View suppliers',
            'suppliers_write': 'Manage suppliers',
            'orders_manage': 'Manage orders',
            'orders_read': 'View order history',
            'pos_access': 'Access Point of Sale terminal',
            'pos_view': 'View POS module',
            'returns_manage': 'Approve/Reject returns',
            'returns_request': 'Request a return',
            'categories_read': 'View categories',
            'categories_write': 'Manage categories',
            'products_delete': 'Delete products',
            'products_read': 'View products',
            'products_write': 'Create/Edit products',
            'reports_financial': 'View financial reports',
            'reports_inventory': 'View inventory reports',
            'reports_sales': 'View sales reports',
            'reports_view': 'View reports',
            'settings_manage': 'Edit system settings',
            'settings_view': 'View system settings',
            'bulk_export': 'Export data in bulk',
            'bulk_import': 'Import data in bulk',
            'dashboard_view': 'View dashboard overview',
            'notifications_read': 'View consolidated notifications',
            'notifications_write': 'Mark notifications as read/unread'
        }
    },
    users: {
        title: 'User Access Management',
        subtitle: 'Approve, edit, or revoke workspace access — with live filters and secure updates.',
        badges: {
            role_control: 'Role-based control',
            status_filters: 'Status filters',
            inline_edits: 'Inline edits'
        },
        filters: {
            search_placeholder: 'Search by username or ID',
            search_btn: 'Search',
            all: 'All users',
            approved: 'Approved',
            pending: 'Pending'
        },
        table: {
            id: 'ID',
            username: 'Username',
            role: 'Role',
            status: 'Status',
            actions: 'Actions',
            empty: 'No users match this filter',
            approve: 'Approve',
            edit: 'Edit',
            delete: 'Delete'
        },
        form: {
            title: 'User Details',
            subtitle: 'Update role or credentials for the selected user',
            select_prompt: 'Select a user from the table to edit access.',
            editing: 'Editing',
            sections: {
                credentials: 'Account Credentials',
                personal: 'Personal Information'
            },
            fields: {
                username: 'Username',
                password_placeholder: 'Reset password (optional)',
                select_role: 'Select a role',
                first_name: 'First Name',
                last_name: 'Last Name',
                email: 'Email',
                phone: 'Phone Number',
                address: 'Address'
            },
            buttons: {
                save: 'Save changes',
                approve: 'Approve',
                delete: 'Delete'
            },
            read_only: 'You have read-only access to user management.'
        },
        status: {
            approved: 'Approved',
            pending: 'Pending',
            active_hint: 'Active access',
            pending_hint: 'Awaiting approval'
        }
    },
    bulk: {
        title: 'Bulk Operations',
        subtitle: 'Manage your catalog efficiently with bulk imports, exports, and job tracking.',
        tabs: {
            import: 'Import',
            export: 'Export',
            status: 'Status'
        },
        steps: {
            download: 'Download Template',
            upload: 'Upload',
            validate: 'Validate',
            import: 'Import',
            done: 'Done',
            step1_title: 'Step 1: Download & Prepare',
            step1_desc: 'Use the correct CSV format for seamless import.',
            step2_title: 'Step 2: Upload File',
            step2_desc: 'Validate your CSV before importing products.',
            step3_title: 'Step 3: Review & Confirm',
            step3_desc: 'Ensure validation passes before final import.'
        },
        buttons: {
            download_template: 'Download Template',
            upload_validate: 'Upload & Validate',
            new_import: 'New Import',
            try_again: 'Try Again',
            generate_export: 'Generate Export',
            download_export: 'Download Export',
            refresh: 'Refresh'
        },
        labels: {
            valid: 'Valid',
            invalid: 'Invalid',
            total: 'Total',
            new_categories: 'New Categories',
            new_suppliers: 'New Suppliers',
            new_locations: 'New Locations',
            valid_records: 'Valid Records',
            invalid_records: 'Invalid Records',
            changes: 'Changes',
            processed: 'Processed',
            success_rate: 'Success Rate',
            breakdown: 'Breakdown',
            format: 'Format',
            category: 'Category (Optional)',
            supplier: 'Supplier (Optional)',
            all_categories: 'All Categories',
            all_suppliers: 'All Suppliers',
            search_placeholder: 'Search by Job ID...',
            no_history: 'No History',
            no_history_desc: 'Recent import and export jobs will appear here.',
            select_job: 'Select a job to view details'
        },
        status: {
            import_complete: 'Import Complete',
            import_failed: 'Import Failed',
            validating: 'Validating file...',
            importing: 'Importing products...',
            success: 'Success'
        }
    },
    alerts: {
        title: 'Alerts & Notifications Control',
        subtitle: 'Thresholds, escalations, and targeted user messaging — all in one soothing, vibrant cockpit.',
        refresh: 'Refresh Alerts',
        go_to_ops: 'Go to Operations',
        live_alerts: 'Live Alerts',
        live_alerts_desc: 'Filter by type or lifecycle state',
        filters: {
            placeholder_type: 'All types',
            placeholder_status: 'Any status',
            refresh_btn: 'Refresh',
            type_options: {
                low_stock: 'Low stock',
                overstock: 'Overstock',
                out_of_stock: 'Out of stock',
                expiry: 'Expiry'
            },
            status_options: {
                active: 'Active',
                resolved: 'Resolved'
            }
        },
        table: {
            type: 'Type',
            product: 'Product',
            message: 'Message',
            status: 'Status',
            action: 'Action',
            empty: 'No alerts found',
            resolve: 'Resolve',
            product_fallback: 'Product'
        },
        details: {
            title: 'Alert Details',
            type: 'Alert Type',
            status: 'Status',
            triggered: 'Triggered',
            context: 'Alert Context',
            product: 'Product',
            message: 'Message',
            product_id: 'Product ID',
            triggered_at: 'Triggered At',
            updated_at: 'Updated At',
            batch_details: 'Batch Details',
            batch_id: 'Batch ID',
            batch_number: 'Batch Number',
            quantity: 'Quantity',
            expiry_date: 'Expiry Date',
            resolved_at: 'Resolved',
            awaiting_action: 'Awaiting action',
            na: 'N/A'
        },
        thresholds: {
            title: 'Product Thresholds',
            subtitle: 'Configure alerting per SKU',
            product_label: 'Product',
            product_placeholder: 'Search product...',
            low_stock: 'Low stock',
            overstock: 'Overstock',
            expiry_days: 'Expiry days',
            save: 'Save thresholds'
        },
        notifications: {
            title: 'User Notifications',
            subtitle: 'Escalation preferences per operator',
            user_label: 'User',
            user_placeholder: 'Search user...',
            email_placeholder: 'Email',
            phone_placeholder: 'Phone',
            email_label: 'Email',
            sms_label: 'SMS',
            save: 'Save preferences'
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to view alerts.',
            load_fail: 'Failed to Load Alerts',
            resolve_success: 'Alert resolved',
            resolve_fail: 'Failed to Resolve Alert',
            select_product: 'Select a product',
            thresholds_updated: 'Thresholds updated',
            thresholds_fail: 'Failed to Save Thresholds',
            provide_user: 'Provide a user ID',
            prefs_saved: 'Notification preferences saved',
            prefs_fail: 'Failed to Save Preferences'
        }
    },
    dashboard: {
        title: 'Real-time Inventory Intelligence',
        subtitle: 'Monitor, analyze, and optimize your inventory ecosystem with AI-powered insights',
        refresh: 'Refresh Data',
        update_catalog: 'Update Catalog',
        stats: {
            active_products: 'Active Products',
            categories: 'Categories',
            suppliers: 'Suppliers',
            active_alerts: 'Active Alerts',
            forecast_hint: '📈 {value} forecasted for Q4',
            supplier_hint: '🔄 Across {count} suppliers',
            sla_hint: '✅ All SLAs active',
            escalation_hint: '🚨 Auto-escalations active'
        },
        demand: {
            title: 'Demand Pulse Analytics',
            subtitle: 'Real-time inventory movement trends',
            chart_hint: '📊 Based on sales velocity & stock buffers',
            trend_positive: '↑ Trend: Positive',
            trend_negative: '↓ Trend: Negative',
            trend_stable: '→ Trend: Stable',
            growth: '📈 {value}% Growth',
            decline: '📉 {value}% Decline',
            no_change: 'No Change',
            day_label: 'Day {day}'
        },
        quick_actions: {
            title: 'Quick Actions',
            subtitle: 'Instant inventory operations',
            balance_stock: 'Balance Stock',
            balance_desc: 'Optimize inventory levels',
            run_forecast: 'Run Forecast',
            forecast_desc: 'AI predictions',
            export_catalog: 'Export Catalog',
            export_desc: 'Bulk operations'
        },
        fresh_inventory: {
            title: 'Fresh Inventory',
            subtitle: 'Recently added or updated SKUs',
            sku: 'SKU',
            product_name: 'Product Name',
            status: 'Status',
            no_data: 'No recent inventory changes'
        },
        priority_alerts: {
            title: 'Priority Alerts',
            subtitle: 'Requires immediate attention',
            type: 'Alert Type',
            product: 'Product',
            status: 'Status',
            no_data: 'All systems normal'
        },
        procurement: {
            title: 'Procurement Intelligence',
            subtitle: 'AI-powered reorder suggestions',
            product: 'Product',
            suggested_qty: 'Suggested Qty',
            supplier: 'Supplier',
            status: 'Status',
            no_data: 'No pending suggestions',
            ready_to_order: 'Ready to Order'
        },
        toasts: {
            load_fail: 'Failed to Load Dashboard',
            error_desc: 'An unexpected error occurred'
        }
    },
    catalog: {
        hero: {
            subtitle: 'CATALOG COCKPIT',
            title: 'Products, Categories & Partners',
            description: 'Unified control center for your catalog data',
            sync_data: 'Sync data',
            bulk_import: 'Bulk import'
        },
        tabs: {
            products: 'Products',
            categories: 'Categories',
            sub_categories: 'Sub Categories',
            suppliers: 'Suppliers',
            locations: 'Locations'
        },
        common: {
            search: 'Search',
            clear: 'Clear',
            add: 'Add',
            edit: 'Edit',
            delete: 'Delete',
            save: 'Save',
            create: 'Create',
            update: 'Update',
            reset: 'Reset',
            actions: 'Actions',
            new: 'New',
            loading: 'Loading...'
        },
        products: {
            title: 'SKU Registry',
            subtitle: 'Manage items synced with the warehouse',
            search_placeholder: 'Search products...',
            add_button: 'Add Product',
            columns: {
                sku: 'SKU',
                name: 'Name',
                status: 'Status'
            },
            form: {
                update_title: 'Update product',
                create_title: 'Create product',
                subtitle: 'SKU-level metadata',
                sku: 'SKU',
                name: 'Name',
                description: 'Description',
                barcode: 'Barcode / UPC (must be unique)',
                select_category: 'Select category',
                select_sub_category: 'Select sub-category',
                select_supplier: 'Select supplier',
                default_location: 'Default location',
                purchase_price: 'Purchase price',
                selling_price: 'Selling price'
            }
        },
        categories: {
            title: 'Categories',
            subtitle: 'Structure your catalog foundation',
            search_placeholder: 'Search by name...',
            form: {
                update_title: 'Update category',
                create_title: 'Create category',
                name: 'Name'
            }
        },
        sub_categories: {
            title: 'Sub Categories',
            subtitle: 'Filter by parent category',
            select_category: 'Select category to view sub-categories',
            loading_categories: 'Loading categories...',
            empty_state: 'Select a category above',
            empty_subtitle: 'Sub-categories will appear here',
            form: {
                update_title: 'Update sub-category',
                create_title: 'Create sub-category',
                name: 'Name',
                select_category: 'Select category'
            }
        },
        suppliers: {
            title: 'Suppliers',
            subtitle: 'Strategic partners powering replenishment',
            search_placeholder: 'Search by name...',
            columns: {
                contact: 'Contact'
            },
            form: {
                update_title: 'Update supplier',
                create_title: 'Create supplier',
                name: 'Name',
                contact_person: 'Contact person',
                email: 'Email',
                phone: 'Phone',
                address: 'Address'
            }
        },
        locations: {
            title: 'Locations',
            subtitle: 'Fulfilment nodes and stores',
            columns: {
                address: 'Address'
            },
            form: {
                update_title: 'Update location',
                create_title: 'Create location',
                name: 'Name',
                address: 'Address'
            }
        },
        details: {
            id: 'ID',
            sku: 'SKU',
            name: 'Name',
            status: 'Status',
            purchase_price: 'Purchase Price',
            selling_price: 'Selling Price',
            category_id: 'Category ID',
            supplier_id: 'Supplier ID',
            description: 'Description',
            created: 'Created',
            updated: 'Updated',
            parent_category: 'Parent Category',
            sub_category_id: 'Sub-category ID',
            on_time_rate: 'On-time rate',
            avg_lead_time: 'Avg. lead time',
            supplier_id_label: 'Supplier ID',
            contact_details: 'Contact Details',
            contact_person: 'Contact Person',
            email: 'Email',
            phone: 'Phone',
            address: 'Address',
            performance: 'Performance Snapshot',
            location_id: 'Location ID',
            location_profile: 'Location Profile'
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to view the catalog.',
            load_fail: 'Failed to Load Catalog',
            search_fail: 'Search Failed',
            product_not_found: 'Product not found',
            category_not_found: 'Category not found',
            supplier_not_found: 'Supplier not found',
            sub_categories_fail: 'Failed to Load Sub-Categories',
            missing_barcode: 'Missing Barcode/UPC',
            missing_barcode_desc: 'Each product must have a unique BarcodeUPC value.',
            duplicate_barcode: 'Duplicate Barcode/UPC Detected',
            duplicate_barcode_desc: 'The BarcodeUPC "{barcode}" is already used by product "{product}".',
            product_saved: 'Product saved successfully',
            product_save_fail: 'Failed to Save Product',
            product_removed: 'Product removed',
            product_remove_fail: 'Failed to Delete Product',
            confirm_delete: 'Are you sure you want to delete {name}?',
            category_saved: 'Category saved',
            category_save_fail: 'Failed to Save Category',
            category_removed: 'Category removed',
            category_remove_fail: 'Failed to Delete Category',
            sub_category_saved: 'Sub-category saved',
            sub_category_save_fail: 'Failed to Save Sub-Category',
            sub_category_removed: 'Sub-category removed',
            sub_category_remove_fail: 'Failed to Delete Sub-Category',
            supplier_saved: 'Supplier saved',
            supplier_save_fail: 'Failed to Save Supplier',
            supplier_removed: 'Supplier removed',
            supplier_remove_fail: 'Failed to Delete Supplier',
            location_saved: 'Location saved',
            location_save_fail: 'Failed to Save Location',
            location_removed: 'Location removed',
            location_remove_fail: 'Failed to Delete Location'
        }
    },
    operations: {
        hero: {
            title: 'Stock Adjustments, Transfers & Barcode Intelligence',
            subtitle: 'Unified real-time control for stock, movement & labeling.'
        },
        snapshot: {
            title: 'Inventory Snapshot',
            subtitle: 'View product balance and batch details',
            search_placeholder: 'Search product...',
            location_id: 'Location ID (optional)',
            fetch_button: 'Fetch stock levels',
            current_qty: 'Current quantity',
            table: {
                batch: 'Batch',
                qty: 'Qty',
                expiry: 'Expiry',
                empty: 'No batch detail available'
            }
        },
        adjustment: {
            title: 'Manual Adjustment',
            subtitle: 'Perform adhoc cycle counts or receipts',
            select_product: 'Select product to adjust...',
            stock_in: 'Stock In (+)',
            stock_out: 'Stock Out (-)',
            quantity: 'Quantity',
            reason_code: 'Reason code',
            notes: 'Notes',
            submit_button: 'Apply adjustment'
        },
        transfer: {
            title: 'Stock Transfer',
            subtitle: 'Move inventory across locations',
            select_product: 'Select product to transfer...',
            source: 'Source location',
            dest: 'Destination location',
            quantity: 'Quantity',
            submit_button: 'Create transfer'
        },
        barcode: {
            title: 'Barcode Intelligence',
            subtitle: 'Lookup and generate barcodes for SKUs',
            input_placeholder: 'Scan or type barcode / SKU',
            lookup_button: 'Lookup Product',
            generate_button: 'Generate Image',
            preview_alt: 'Barcode preview'
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to access operations.',
            product_id_required: 'Enter a product ID first',
            snapshot_updated: 'Inventory snapshot updated',
            fetch_stock_fail: 'Failed to Fetch Stock',
            select_product: 'Select a product',
            stock_adjusted: 'Stock adjusted',
            adjust_fail: 'Failed to Apply Adjustment',
            transfer_queued: 'Transfer queued',
            transfer_fail: 'Failed to Create Transfer',
            barcode_required: 'Provide a barcode value',
            sku_resolved: 'SKU resolved',
            lookup_fail: 'Failed to Lookup Barcode',
            sku_or_id_required: 'Provide SKU or product ID',
            generate_fail: 'Failed to Generate Barcode',
            product_not_found: 'Product not found'
        }
    },
    time_tracking: {
        hero: {
            title: 'Time Tracking Control Center',
            subtitle: 'Stay on top of shifts, breaks, and approvals with a calm workspace designed to feel invisible. Switch between personal and manager views without losing the Apple-inspired polish.',
            label: 'Time Intelligence'
        },
        role_toggle: {
            staff: 'Staff View',
            manager: 'Manager View',
            label: 'Select dashboard'
        },
        staff: {
            header: {
                title: 'My Time Tracker',
                subtitle: 'Personal Flow',
                desc: 'Track your focus, breaks, and progress from one calm surface.'
            },
            status_card: {
                title: 'My Status',
                clocked_in: 'Clocked In',
                clocked_out: 'Clocked Out',
                on_break: 'On Break',
                break_time: 'Break Time',
                today_total: "Today's Total",
                clock_in_btn: 'Clock In',
                clock_out_btn: 'Clock Out',
                start_break_btn: 'Start Break',
                end_break_btn: 'End Break'
            },
            stats: {
                today_hours: "Today's Hours",
                weekly_hours: 'Weekly Hours',
                target_label: 'out of {target} hours target'
            },
            task: {
                title: 'Current Task',
                label: 'What are you working on?',
                placeholder: 'Enter your current task...',
                button: 'Update Task'
            },
            goals: {
                title: 'Daily Goals'
            },
            recent_shifts: {
                title: 'Recent Shifts'
            }
        },
        manager: {
            header: {
                title: 'Team Dashboard',
                subtitle: 'Leadership',
                desc: 'Monitor attendance, live shifts, and weekly momentum.'
            },
            actions: {
                export: 'Export Report',
                filter: 'Filter'
            },
            stats: {
                total_hours: 'Total Hours',
                weekly_target_percent: '{percent}% of weekly target',
                active_members: 'Active Members',
                working: 'Currently working',
                weekly_target: 'Weekly Target',
                team_goal: 'Team goal'
            },
            team: {
                title: 'Team Members',
                status: {
                    working: 'Working',
                    on_break: 'On Break',
                    offline: 'Offline'
                }
            },
            attendance: {
                title: 'Weekly Attendance',
                present: 'Present',
                late: 'Late',
                absent: 'Absent'
            },
            recent_activity: {
                title: 'Recent Activities',
                clocked_in: 'clocked in',
                started_break: 'started break',
                clocked_out: 'clocked out'
            },
            quick_actions: {
                title: 'Quick Actions',
                reports: 'Reports',
                schedule: 'Schedule',
                payroll: 'Payroll',
                reminders: 'Reminders'
            }
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to access time tracking.',
            clock_in_success: 'Clocked In Successfully',
            clock_in_desc: 'Your shift has started. Have a productive day!',
            clock_out_info: 'Clocked Out',
            clock_out_desc: 'Great work today! You completed {time} of focused time.',
            break_start: 'Break Started',
            break_start_desc: 'Take a well-deserved break! Your timer is paused.',
            break_end: 'Break Ended',
            break_end_desc: 'Welcome back! Ready to continue your productive day?',
            task_updated: 'Task Updated',
            task_updated_desc: 'Now working on: {task}',
            report_exported: 'Report Exported',
            report_exported_desc: 'Weekly time report has been exported successfully.',
            reminder_sent: 'Reminder Sent',
            reminder_sent_desc: 'Reminder sent to {name} to complete their timesheet.',
            load_fail: 'Failed to load data',
            op_fail: 'Operation failed'
        }
    },
    intelligence: {
        hero: {
            title: 'Forecasting, Reorder Suggestions & Business Reports',
            subtitle: 'Plan ahead, act on signals, and align analytics across one horizon.'
        },
        demand_forecast: {
            title: 'AI Demand Forecast',
            subtitle: 'Predict future demand for specific products',
            select_product: 'Select Product',
            placeholder: 'Search by name or SKU...',
            period_label: 'Forecast Period (Days)',
            generate_btn: 'Generate',
            generating_btn: 'Generating...',
            predicted_demand: 'Predicted Demand',
            confidence: 'Confidence',
            reasoning: 'AI Reasoning',
            generated_at: 'Generated at'
        },
        churn_risk: {
            title: 'Customer Churn Prediction',
            subtitle: 'Identify at-risk customers and retention strategies',
            select_customer: 'Select Customer',
            placeholder: 'Search by name or email...',
            analyze_btn: 'Analyze Risk',
            analyzing_btn: 'Analyzing...',
            risk_level: 'Risk Level',
            risk_score: 'Risk Score',
            primary_factors: 'Primary Factors',
            retention_strategy: 'Retention Strategy',
            suggested_action: 'Suggested Action',
            discount_offer: 'Offer a {discount}% discount to retain this customer.'
        },
        report_range: {
            title: 'Report Range',
            subtitle: 'Align analytics across shared horizon',
            sales_trends: 'Sales Trends',
            turnover: 'Inventory Turnover',
            margin: 'Profit Margin'
        },
        reorder_suggestions: {
            title: 'Reorder Suggestions',
            subtitle: 'AI-recommended purchase orders',
            refresh_btn: 'Refresh',
            table: {
                product: 'Product',
                supplier: 'Supplier',
                suggested_qty: 'Suggested qty',
                status: 'Status',
                actions: 'Actions',
                create_po: 'Create PO',
                empty: 'No pending suggestions'
            }
        },
        reports: {
            period: 'Period: {period}',
            sales: {
                title: 'Sales Report',
                subtitle: 'Trend of total vs average sales',
                total_sales: 'Total Sales',
                avg_daily_sales: 'Avg Daily Sales'
            },
            turnover: {
                title: 'Turnover Report',
                subtitle: 'Inventory efficiency over time',
                avg_inventory_value: 'Avg Inventory Value',
                turnover_rate: 'Turnover Rate'
            },
            margin: {
                title: 'Margin Report',
                subtitle: 'Profitability visualization',
                gross_profit: 'Gross Profit',
                total_revenue: 'Total Revenue'
            }
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to view reports.',
            load_suggestions_fail: 'Failed to Load Suggestions',
            po_created: 'PO {id} created',
            po_create_fail: 'Failed to Create PO',
            report_ready: 'Report ready',
            report_fail: 'Failed to Run Report',
            suggestions_refreshed: 'Suggestions refreshed',
            refresh_fail: 'Failed to refresh suggestions',
            select_product: 'Please select a product',
            forecast_success: 'Forecast generated successfully',
            forecast_fail: 'Failed to generate forecast',
            select_customer: 'Please select a customer',
            analysis_complete: 'Analysis complete',
            analysis_fail: 'Failed to analyze churn risk'
        }
    },
    pos: {
        hero: {
            title: 'Unified Checkout Console',
            subtitle: 'Scan, search, and complete orders with a low-friction flow that stays in sync with your catalog.',
            label: 'Point of Sale',
            sub_label: 'Live checkout canvas for counter teams',
            new_sale_btn: 'New walk-in sale',
            refresh_catalog_btn: 'Refresh catalog'
        },
        header: {
            title: 'Point of Sale',
            description: 'Tap products to build the cart, review below, then confirm on the right.',
            super_shop_mode: 'Super shop mode',
            search_placeholder: 'Search by name, barcode, or SKU...',
            search_btn: 'Search'
        },
        products: {
            title: 'Products',
            description: 'Tap a tile to add it to the active cart.',
            results_found: '{count} results',
            filter_status: {
                label: 'Stock Status',
                all: 'All Status',
                in_stock: 'In Stock',
                low_stock: 'Low Stock',
                out_of_stock: 'Out of Stock'
            },
            no_results: 'No products found. Try adjusting your search.',
            in_stock: '{count} in stock',
            tap_to_add: 'Tap to add'
        },
        cart: {
            title: 'Cart',
            empty_desc: 'No items added yet.',
            items_desc: '{count} item{s} in cart',
            clear_btn: 'Clear cart',
            empty_state: 'Add products from the grid above to start a new order.',
            table: {
                product: 'Product',
                price: 'Price',
                qty: 'Qty',
                total: 'Total'
            }
        },
        customer: {
            title: 'Customer',
            description: 'Attach a customer by ID, username, email, or phone. Optional for walk-ins.',
            search_placeholder: 'Search by ID, username, email, phone',
            new_btn: 'New',
            no_selected: 'No customer selected. You can still complete a walk-in sale.',
            loyalty_pts: '{points} pts',
            tier: '{tier}'
        },
        payment: {
            title: 'Payment',
            description: 'Choose how the customer is paying for this order.',
            methods: {
                cash: 'Cash',
                card: 'Card',
                bkash: 'bKash',
                other: 'Other'
            },
            sub: {
                physical: 'Physical',
                terminal: 'Terminal',
                mobile: 'Mobile',
                check_due: 'Check/Due'
            }
        },
        loyalty: {
            redeem_label: 'Redeem Loyalty Points',
            available: 'Available: {points} pts ({value} value)',
            points: 'points',
            error_exceed: 'Cannot exceed available balance.'
        },
        summary: {
            title: 'Order Summary',
            description: 'Review totals and payment before confirming the sale.',
            subtotal: 'Subtotal',
            tax: 'Tax ({rate}%)',
            total: 'Total',
            payment: 'Payment:',
            items: 'Items:',
            not_selected: 'Not selected',
            loyalty_earnings: 'Loyalty Earnings',
            complete_btn: 'Complete Order',
            processing_btn: 'Processing...',
            add_items_hint: 'Add items to cart to continue',
            select_payment_hint: 'Select a payment method to complete'
        },
        new_customer_modal: {
            title: 'New Customer',
            description: 'Add a new member to your customer base.',
            name_label: 'Full Name',
            name_placeholder: 'Jane Doe',
            email_label: 'Email Address',
            email_placeholder: 'jane@example.com',
            phone_label: 'Phone Number',
            phone_placeholder: '+1 (555) 000-0000',
            cancel_btn: 'Cancel',
            create_btn: 'Create Customer'
        },
        toasts: {
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to access POS.',
            customer_not_found: 'Customer not found',
            search_error: 'Error searching for customer',
            out_of_stock: 'Product is out of stock',
            stock_limit_reached: 'Cannot add more. Only {stock} items available in stock.',
            processing: 'Processing transaction...',
            order_success: 'Order completed successfully!',
            loyalty_earned: 'Customer earned {points} loyalty points!',
            loyalty_redeemed: 'Redeemed {points} points for {amount} discount.',
            transaction_fail: 'Transaction Failed',
            name_required: 'Name is required',
            customer_created: 'Customer created and selected!',
            create_fail: 'Failed to create customer'
        }
    },
    orders: {
        title: 'Orders Management',
        subtitle: 'Manage customer sales invoices and supplier purchase orders & returns.',
        tabs: {
            sales: 'Sales (Invoices)',
            purchases: 'Purchases & Returns'
        },
        subtabs: {
            sales_orders: 'Sales Orders',
            customer_returns: 'Customer Returns',
            purchase_orders: 'Purchase Orders',
            vendor_returns: 'Vendor Returns'
        },
        search: {
            sales: 'Search sales...',
            pos: 'Search POs...',
            returns: 'Search returns...'
        },
        status: {
            return_pending: 'RETURN PENDING',
            refund_value: 'Refund Value',
            items_returned: 'Items Returned',
            no_items: 'No items detail',
            total_items: 'Total Items',
            qty: 'Qty'
        },
        labels: {
            customer: 'Customer',
            items: 'Items',
            more: 'more',
            supplier: 'Supplier',
            refund: 'Refund',
            reason: 'Reason',
            guest: 'Guest',
            unknown: 'Unknown'
        },
        actions: {
            view_details: 'View Details',
            manage_po: 'Manage PO',
            try_again: 'Try Again'
        },
        empty: {
            sales: 'No sales orders found.',
            customer_returns: 'No customer returns found.',
            purchase_orders: 'No purchase orders found.',
            vendor_returns: 'No vendor returns found.'
        },
        errors: {
            load_fail: 'Failed to load order history.',
            access_denied: 'Access Denied',
            access_denied_desc: 'You do not have permission to view orders.'
        }
    }
};
