export default {
    common: {
        save: 'Guardar',
        saved_successfully: 'Guardado con éxito',
        failed_to_save: 'Error al guardar',
        access_denied: 'Acceso Denegado',
        no_permission_settings: 'No tienes permiso para ver la configuración.',
        no_data_available: 'No hay datos disponibles',
        configuration: 'Configuración',
        manage_preferences: 'Gestiona preferencias del sistema, controles de seguridad y políticas globales.',
    },
    settings: {
        tabs: {
            general: 'General',
            business_rules: 'Reglas de Negocio',
            system_ai: 'Sistema e IA',
            security_roles: 'Seguridad y Roles',
            policies: 'Políticas',
            notifications: 'Notificaciones'
        },
        general: {
            business_profile: 'Perfil de Negocio',
            business_profile_desc: "La identidad visible de tu organización en la plataforma.",
            business_name: 'Nombre del Negocio',
            currency: 'Moneda',
            timezone: 'Zona Horaria',
            locale: 'Idioma / Región',
            select_currency: 'Seleccionar Moneda',
            select_timezone: 'Seleccionar Zona Horaria',
            select_locale: 'Seleccionar Idioma'
        },
        business: {
            loyalty_program: 'Programa de Lealtad',
            loyalty_program_desc: 'Configura cómo los clientes ganan y canjean puntos, y establece los umbrales de nivel.',
            points_configuration: 'Configuración de Puntos',
            earning_rate: 'Tasa de Ganancia (Puntos por $1)',
            earning_rate_hint: 'Cuántos puntos gana un cliente por cada unidad de moneda gastada.',
            redemption_value: 'Valor de Canje ($ por Punto)',
            redemption_value_hint: 'El valor monetario de un solo punto de lealtad al canjear.',
            tier_thresholds: 'Umbrales de Nivel',
            silver_tier: 'Nivel Plata (Puntos)',
            gold_tier: 'Nivel Oro (Puntos)',
            platinum_tier: 'Nivel Platino (Puntos)',
            financial_settings: 'Ajustes Financieros',
            financial_settings_desc: 'Gestiona tasas impositivas y otros parámetros financieros.',
            default_tax_rate: 'Tasa Impositiva Predeterminada (%)',
            default_tax_rate_hint: 'Esta tasa se aplicará a todas las ventas correspondientes.'
        },
        policies: {
            privacy_policy: 'Política de Privacidad',
            privacy_policy_placeholder: 'Introduce tu política de privacidad (soporta Markdown)...',
            save_policy: 'Guardar Política',
            terms_of_service: 'Términos de Servicio',
            terms_of_service_placeholder: 'Introduce tus términos de servicio (soporta Markdown)...',
            save_terms: 'Guardar Términos',
            return_policy: 'Política de Devolución',
            return_window: 'Ventana de Devolución (Días)',
            return_window_hint: 'Número de días después de la compra en que un cliente puede solicitar una devolución.'
        },
        system: {
            ai_system: 'IA y Sistema',
            ai_system_desc: 'Configura el comportamiento de los agentes autónomos y parámetros del sistema.',
            wake_up_time: 'Hora de Despertar de la IA',
            wake_up_time_hint: 'La IA ejecutará el "Chequeo Matutino Diario" a esta hora todos los días.'
        },
        notifications: {
            global_alerts_center: 'Centro de Alertas Global',
            coming_soon: 'El enrutamiento avanzado de notificaciones y la configuración de webhooks llegarán en la próxima actualización.'
        }
    },
    roles: {
        title: 'Roles y Acceso',
        create_new: 'Crear Nuevo Rol',
        create_first: 'Crear Primer Rol',
        system_managed: 'Gestionado por el Sistema',
        custom_role: 'Rol Personalizado',
        active_permissions: 'Permisos Activos',
        delete_confirm: '¿Estás seguro de que quieres eliminar el rol "{name}"?',
        delete: 'Eliminar',
        save_changes: 'Guardar Cambios',
        saving: 'Guardando...',
        reset: 'Restablecer',
        role_name: 'Nombre del Rol',
        description: 'Descripción',
        capabilities: 'Capacidades',
        capabilities_desc: 'Ajusta los controles de acceso para este rol',
        select_all: 'Seleccionar Todo',
        security_access_control: 'Seguridad y Control de Acceso',
        security_desc: 'Selecciona un rol de la barra lateral para configurar permisos, o crea un nuevo rol personalizado para delegar capacidades de acceso específicas.',
        names: {
            admin: 'Administrador',
            manager: 'Gerente',
            cashier: 'Cajero',
            staff: 'Personal',
            customer: 'Cliente',
            sales_associate: 'Asociado de Ventas'
        }
    },
    permissions: {
        groups: {
            inventory: 'Inventario',
            sales: 'Ventas',
            crm: 'CRM',
            hrm: 'RRHH',
            settings: 'Configuración',
            reports: 'Informes',
            pos: 'TPV',
            dashboard: 'Tablero',
            'product management': 'Gestión de Productos',
            'access control': 'Control de Acceso',
            orders: 'Pedidos',
            system: 'Sistema'
        },
        names: {
            'roles_manage': 'Gestionar Roles',
            'roles_view': 'Ver Roles',
            'users_manage': 'Gestionar Usuarios',
            'users_view': 'Ver Usuarios',
            'crm_read': 'Leer Datos CRM',
            'crm_view': 'Ver Módulo CRM',
            'crm_write': 'Modificar Datos CRM',
            'customers_read': 'Ver Clientes',
            'customers_write': 'Gestionar Clientes',
            'loyalty_read': 'Ver Info Lealtad',
            'loyalty_write': 'Gestionar Puntos Lealtad',
            'alerts_manage': 'Resolver Alertas',
            'alerts_view': 'Ver Alertas',
            'barcode_read': 'Buscar Códigos de Barras',
            'inventory_read': 'Leer Datos Inventario',
            'inventory_view': 'Ver Módulo Inventario',
            'inventory_write': 'Modificar Datos Inventario',
            'locations_read': 'Ver Ubicaciones',
            'locations_write': 'Gestionar Ubicaciones',
            'replenishment_read': 'Ver Previsiones/Sugerencias',
            'replenishment_write': 'Generar Previsiones y Gestionar PO',
            'suppliers_read': 'Ver Proveedores',
            'suppliers_write': 'Gestionar Proveedores',
            'orders_manage': 'Gestionar Pedidos',
            'orders_read': 'Ver Historial de Pedidos',
            'pos_access': 'Acceder a Terminal TPV',
            'pos_view': 'Ver Módulo TPV',
            'returns_manage': 'Aprobar/Rechazar Devoluciones',
            'returns_request': 'Solicitar Devolución',
            'categories_read': 'Ver Categorías',
            'categories_write': 'Gestionar Categorías',
            'products_delete': 'Eliminar Productos',
            'products_read': 'Ver Productos',
            'products_write': 'Crear/Editar Productos',
            'reports_financial': 'Ver Informes Financieros',
            'reports_inventory': 'Ver Informes de Inventario',
            'reports_sales': 'Ver Informes de Ventas',
            'reports_view': 'Ver Informes',
            'settings_manage': 'Editar Configuración del Sistema',
            'settings_view': 'Ver Configuración del Sistema',
            'bulk_export': 'Exportar Datos',
            'bulk_import': 'Importar Datos',
            'dashboard_view': 'Ver Tablero',
            'notifications_read': 'Ver Notificaciones',
            'notifications_write': 'Gestionar Notificaciones'
        },
        descriptions: {
            'roles_manage': 'Gestionar roles y permisos',
            'roles_view': 'Ver roles existentes',
            'users_manage': 'Añadir o editar cuentas de usuario',
            'users_view': 'Ver lista de usuarios',
            'crm_read': 'Leer datos de relaciones con clientes',
            'crm_view': 'Acceder a la interfaz CRM',
            'crm_write': 'Actualizar registros CRM',
            'customers_read': 'Ver detalles del cliente',
            'customers_write': 'Añadir o editar información del cliente',
            'loyalty_read': 'Ver saldo de puntos de lealtad',
            'loyalty_write': 'Ajustar puntos de lealtad',
            'alerts_manage': 'Actuar sobre alertas del sistema',
            'alerts_view': 'Ver alertas activas',
            'barcode_read': 'Escanear y buscar códigos de barras',
            'inventory_read': 'Ver niveles de stock',
            'inventory_view': 'Acceder a gestión de inventario',
            'inventory_write': 'Actualizar cantidades de stock',
            'locations_read': 'Ver almacenes y ubicaciones de tiendas',
            'locations_write': 'Añadir o editar detalles de ubicación',
            'replenishment_read': 'Ver sugerencias de reabastecimiento',
            'replenishment_write': 'Crear órdenes de compra',
            'suppliers_read': 'Ver lista de proveedores',
            'suppliers_write': 'Gestionar información de proveedores',
            'orders_manage': 'Procesar pedidos de venta',
            'orders_read': 'Ver historial de ventas',
            'pos_access': 'Usar punto de venta',
            'pos_view': 'Ver pantallas del TPV',
            'returns_manage': 'Manejar devoluciones de clientes',
            'returns_request': 'Iniciar proceso de devolución',
            'categories_read': 'Ver categorías de productos',
            'categories_write': 'Gestionar estructura de categorías',
            'products_delete': 'Eliminar productos del catálogo',
            'products_read': 'Ver detalles del producto',
            'products_write': 'Añadir o actualizar información del producto',
            'reports_financial': 'Ver informes de rendimiento financiero',
            'reports_inventory': 'Ver informes de estado del inventario',
            'reports_sales': 'Ver informes de análisis de ventas',
            'reports_view': 'Acceder a informes generales',
            'settings_manage': 'Cambiar configuración de la aplicación',
            'settings_view': 'Ver configuración',
            'bulk_export': 'Descargar datos en masa',
            'bulk_import': 'Cargar datos en masa',
            'dashboard_view': 'Ver resumen del tablero principal',
            'notifications_read': 'Leer notificaciones del sistema',
            'notifications_write': 'Actualizar estado de notificaciones'
        }
    },
    users: {
        title: 'Gestión de Acceso de Usuarios',
        subtitle: 'Aprueba, edita o revoca el acceso al espacio de trabajo - con filtros en vivo y actualizaciones seguras.',
        badges: {
            role_control: 'Control basado en roles',
            status_filters: 'Filtros de estado',
            inline_edits: 'Ediciones en línea'
        },
        filters: {
            search_placeholder: 'Buscar por usuario o ID',
            search_btn: 'Buscar',
            all: 'Todos los usuarios',
            approved: 'Aprobados',
            pending: 'Pendientes'
        },
        table: {
            id: 'ID',
            username: 'Nombre de usuario',
            role: 'Rol',
            status: 'Estado',
            actions: 'Acciones',
            empty: 'Ningún usuario coincide con este filtro',
            approve: 'Aprobar',
            edit: 'Editar',
            delete: 'Eliminar'
        },
        form: {
            title: 'Detalles del Usuario',
            subtitle: 'Actualizar rol o credenciales para el usuario seleccionado',
            select_prompt: 'Selecciona un usuario de la tabla para editar el acceso.',
            editing: 'Editando',
            sections: {
                credentials: 'Credenciales de Cuenta',
                personal: 'Información Personal'
            },
            fields: {
                username: 'Nombre de usuario',
                password_placeholder: 'Restablecer contraseña (opcional)',
                select_role: 'Seleccionar un rol',
                first_name: 'Nombre',
                last_name: 'Apellido',
                email: 'Correo electrónico',
                phone: 'Número de teléfono',
                address: 'Dirección'
            },
            buttons: {
                save: 'Guardar cambios',
                approve: 'Aprobar',
                delete: 'Eliminar'
            },
            read_only: 'Tienes acceso de solo lectura a la gestión de usuarios.'
        },
        status: {
            approved: 'Aprobado',
            pending: 'Pendiente',
            active_hint: 'Acceso activo',
            pending_hint: 'Esperando aprobación'
        }
    },
    bulk: {
        title: 'Operaciones Masivas',
        subtitle: 'Gestiona tu catálogo eficientemente con importaciones, exportaciones y seguimiento de trabajos.',
        tabs: {
            import: 'Importar',
            export: 'Exportar',
            status: 'Estado'
        },
        steps: {
            download: 'Descargar Plantilla',
            upload: 'Subir',
            validate: 'Validar',
            import: 'Importar',
            done: 'Hecho',
            step1_title: 'Paso 1: Descargar y Preparar',
            step1_desc: 'Usa el formato CSV correcto para una importación sin problemas.',
            step2_title: 'Paso 2: Subir Archivo',
            step2_desc: 'Valida tu CSV antes de importar productos.',
            step3_title: 'Paso 3: Revisar y Confirmar',
            step3_desc: 'Asegúrate de que la validación pase antes de la importación final.'
        },
        buttons: {
            download_template: 'Descargar Plantilla',
            upload_validate: 'Subir y Validar',
            new_import: 'Nueva Importación',
            try_again: 'Intentar De Nuevo',
            generate_export: 'Generar Exportación',
            download_export: 'Descargar Exportación',
            refresh: 'Refrescar'
        },
        labels: {
            valid: 'Válido',
            invalid: 'Inválido',
            total: 'Total',
            new_categories: 'Nuevas Categorías',
            new_suppliers: 'Nuevos Proveedores',
            new_locations: 'Nuevas Ubicaciones',
            valid_records: 'Registros Válidos',
            invalid_records: 'Registros Inválidos',
            changes: 'Cambios',
            processed: 'Procesados',
            success_rate: 'Tasa de Éxito',
            breakdown: 'Desglose',
            format: 'Formato',
            category: 'Categoría (Opcional)',
            supplier: 'Proveedor (Opcional)',
            all_categories: 'Todas las Categorías',
            all_suppliers: 'Todos los Proveedores',
            search_placeholder: 'Buscar por ID de trabajo...',
            no_history: 'Sin Historial',
            no_history_desc: 'Los trabajos recientes de importación y exportación aparecerán aquí.',
            select_job: 'Selecciona un trabajo para ver detalles'
        },
        status: {
            import_complete: 'Importación Completa',
            import_failed: 'Error en Importación',
            validating: 'Validando archivo...',
            importing: 'Importando productos...',
            success: 'Éxito'
        }
    },
    alerts: {
        title: 'Control de Alertas y Notificaciones',
        subtitle: 'Umbrales, escaladas y mensajes dirigidos a usuarios — todo en una cabina relajante y vibrante.',
        refresh: 'Refrescar Alertas',
        go_to_ops: 'Ir a Operaciones',
        live_alerts: 'Alertas en Vivo',
        live_alerts_desc: 'Filtrar por tipo o estado del ciclo de vida',
        filters: {
            placeholder_type: 'Todos los tipos',
            placeholder_status: 'Cualquier estado',
            refresh_btn: 'Refrescar',
            type_options: {
                low_stock: 'Stock bajo',
                overstock: 'Exceso de stock',
                out_of_stock: 'Agotado',
                expiry: 'Caducidad'
            },
            status_options: {
                active: 'Activo',
                resolved: 'Resuelto'
            }
        },
        table: {
            type: 'Tipo',
            product: 'Producto',
            message: 'Mensaje',
            status: 'Estado',
            action: 'Acción',
            empty: 'No se encontraron alertas',
            resolve: 'Resolver',
            product_fallback: 'Producto'
        },
        details: {
            title: 'Detalles de Alerta',
            type: 'Tipo de Alerta',
            status: 'Estado',
            triggered: 'Activada',
            context: 'Contexto de Alerta',
            product: 'Producto',
            message: 'Mensaje',
            product_id: 'ID de Producto',
            triggered_at: 'Activada en',
            updated_at: 'Actualizada en',
            batch_details: 'Detalles del Lote',
            batch_id: 'ID del Lote',
            batch_number: 'Número de Lote',
            quantity: 'Cantidad',
            expiry_date: 'Fecha de Caducidad',
            resolved_at: 'Resuelta',
            awaiting_action: 'Esperando acción',
            na: 'N/A'
        },
        thresholds: {
            title: 'Umbrales de Producto',
            subtitle: 'Configurar alertas por SKU',
            product_label: 'Producto',
            product_placeholder: 'Buscar producto...',
            low_stock: 'Stock bajo',
            overstock: 'Exceso de stock',
            expiry_days: 'Días de caducidad',
            save: 'Guardar umbrales'
        },
        notifications: {
            title: 'Notificaciones de Usuario',
            subtitle: 'Preferencias de escalada por operador',
            user_label: 'Usuario',
            user_placeholder: 'Buscar usuario...',
            email_placeholder: 'Correo electrónico',
            phone_placeholder: 'Teléfono',
            email_label: 'Correo electrónico',
            sms_label: 'SMS',
            save: 'Guardar preferencias'
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para ver alertas.',
            load_fail: 'Error al cargar alertas',
            resolve_success: 'Alerta resuelta',
            resolve_fail: 'Error al resolver alerta',
            select_product: 'Selecciona un producto',
            thresholds_updated: 'Umbrales actualizados',
            thresholds_fail: 'Error al guardar umbrales',
            provide_user: 'Proporciona un ID de usuario',
            prefs_saved: 'Preferencias de notificación guardadas',
            prefs_fail: 'Error al guardar preferencias'
        }
    },
    dashboard: {
        title: 'Inteligencia de Inventario en Tiempo Real',
        subtitle: 'Monitorea, analiza y optimiza tu ecosistema de inventario con conocimientos impulsados por IA',
        refresh: 'Refrescar Datos',
        update_catalog: 'Actualizar Catálogo',
        stats: {
            active_products: 'Productos Activos',
            categories: 'Categorías',
            suppliers: 'Proveedores',
            active_alerts: 'Alertas Activas',
            forecast_hint: '📈 {value} pronosticados para Q4',
            supplier_hint: '🔄 A través de {count} proveedores',
            sla_hint: '✅ Todos los SLA activos',
            escalation_hint: '🚨 Auto-escaladas activas'
        },
        demand: {
            title: 'Análisis de Pulso de Demanda',
            subtitle: 'Tendencias de movimiento de inventario en tiempo real',
            chart_hint: '📊 Basado en velocidad de ventas y stock de seguridad',
            trend_positive: '↑ Tendencia: Positiva',
            trend_negative: '↓ Tendencia: Negativa',
            trend_stable: '→ Tendencia: Estable',
            growth: '📈 {value}% Crecimiento',
            decline: '📉 {value}% Declive',
            no_change: 'Sin Cambios',
            day_label: 'Día {day}'
        },
        quick_actions: {
            title: 'Acciones Rápidas',
            subtitle: 'Operaciones de inventario instantáneas',
            balance_stock: 'Equilibrar Stock',
            balance_desc: 'Optimizar niveles de inventario',
            run_forecast: 'Ejecutar Pronóstico',
            forecast_desc: 'Predicciones de IA',
            export_catalog: 'Exportar Catálogo',
            export_desc: 'Operaciones masivas'
        },
        fresh_inventory: {
            title: 'Inventario Fresco',
            subtitle: 'SKU añadidos o actualizados recientemente',
            sku: 'SKU',
            product_name: 'Nombre del Producto',
            status: 'Estado',
            no_data: 'Sin cambios de inventario recientes'
        },
        priority_alerts: {
            title: 'Alertas Prioritarias',
            subtitle: 'Requiere atención inmediata',
            type: 'Tipo de Alerta',
            product: 'Producto',
            status: 'Estado',
            no_data: 'Todos los sistemas normales'
        },
        procurement: {
            title: 'Inteligencia de Adquisiciones',
            subtitle: 'Sugerencias de reabastecimiento recomendadas por IA',
            product: 'Producto',
            suggested_qty: 'Cant. Sugerida',
            supplier: 'Proveedor',
            status: 'Estado',
            no_data: 'Sin sugerencias pendientes',
            ready_to_order: 'Listo para Ordenar'
        },
        toasts: {
            load_fail: 'Error al Cargar Tablero',
            error_desc: 'Ocurrió un error inesperado'
        }
    },
    catalog: {
        hero: {
            subtitle: 'CABINA DE CATÁLOGO',
            title: 'Productos, Categorías y Socios',
            description: 'Centro de control unificado para tus datos de catálogo',
            sync_data: 'Sincronizar datos',
            bulk_import: 'Importación masiva'
        },
        tabs: {
            products: 'Productos',
            categories: 'Categorías',
            sub_categories: 'Sub Categorías',
            suppliers: 'Proveedores',
            locations: 'Ubicaciones'
        },
        common: {
            search: 'Buscar',
            clear: 'Limpiar',
            add: 'Añadir',
            edit: 'Editar',
            delete: 'Eliminar',
            save: 'Guardar',
            create: 'Crear',
            update: 'Actualizar',
            reset: 'Restablecer',
            actions: 'Acciones',
            new: 'Nuevo',
            loading: 'Cargando...'
        },
        products: {
            title: 'Registro SKU',
            subtitle: 'Gestiona artículos sincronizados con el almacén',
            search_placeholder: 'Buscar productos...',
            add_button: 'Añadir Producto',
            columns: {
                sku: 'SKU',
                name: 'Nombre',
                status: 'Estado'
            },
            form: {
                update_title: 'Actualizar producto',
                create_title: 'Crear producto',
                subtitle: 'Metadatos a nivel de SKU',
                sku: 'SKU',
                name: 'Nombre',
                description: 'Descripción',
                barcode: 'Código de barras / UPC (debe ser único)',
                select_category: 'Seleccionar categoría',
                select_sub_category: 'Seleccionar sub-categoría',
                select_supplier: 'Seleccionar proveedor',
                default_location: 'Ubicación predeterminada',
                purchase_price: 'Precio de compra',
                selling_price: 'Precio de venta'
            }
        },
        categories: {
            title: 'Categorías',
            subtitle: 'Estructura la base de tu catálogo',
            search_placeholder: 'Buscar por nombre...',
            form: {
                update_title: 'Actualizar categoría',
                create_title: 'Crear categoría',
                name: 'Nombre'
            }
        },
        sub_categories: {
            title: 'Sub Categorías',
            subtitle: 'Filtrar por categoría padre',
            select_category: 'Selecciona categoría para ver sub-categorías',
            loading_categories: 'Cargando categorías...',
            empty_state: 'Selecciona una categoría arriba',
            empty_subtitle: 'Las sub-categorías aparecerán aquí',
            form: {
                update_title: 'Actualizar sub-categoría',
                create_title: 'Crear sub-categoría',
                name: 'Nombre',
                select_category: 'Seleccionar categoría'
            }
        },
        suppliers: {
            title: 'Proveedores',
            subtitle: 'Socios estratégicos impulsando el reabastecimiento',
            search_placeholder: 'Buscar por nombre...',
            columns: {
                contact: 'Contacto'
            },
            form: {
                update_title: 'Actualizar proveedor',
                create_title: 'Crear proveedor',
                name: 'Nombre',
                contact_person: 'Persona de contacto',
                email: 'Correo electrónico',
                phone: 'Teléfono',
                address: 'Dirección'
            }
        },
        locations: {
            title: 'Ubicaciones',
            subtitle: 'Nodos de cumplimiento y tiendas',
            columns: {
                address: 'Dirección'
            },
            form: {
                update_title: 'Actualizar ubicación',
                create_title: 'Crear ubicación',
                name: 'Nombre',
                address: 'Dirección'
            }
        },
        details: {
            id: 'ID',
            sku: 'SKU',
            name: 'Nombre',
            status: 'Estado',
            purchase_price: 'Precio Compra',
            selling_price: 'Precio Venta',
            category_id: 'ID Categoría',
            supplier_id: 'ID Proveedor',
            description: 'Descripción',
            created: 'Creado',
            updated: 'Actualizado',
            parent_category: 'Categoría Padre',
            sub_category_id: 'ID Sub-categoría',
            on_time_rate: 'Tasa a tiempo',
            avg_lead_time: 'Tiempo entrega prom.',
            supplier_id_label: 'ID Proveedor',
            contact_details: 'Detalles de Contacto',
            contact_person: 'Persona de Contacto',
            email: 'Correo electrónico',
            phone: 'Teléfono',
            address: 'Dirección',
            performance: 'Instantánea de Rendimiento',
            location_id: 'ID Ubicación',
            location_profile: 'Perfil de Ubicación'
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para ver el catálogo.',
            load_fail: 'Error al Cargar Catálogo',
            search_fail: 'Error en Búsqueda',
            product_not_found: 'Producto no encontrado',
            category_not_found: 'Categoría no encontrada',
            supplier_not_found: 'Proveedor no encontrado',
            sub_categories_fail: 'Error al Cargar Sub-Categorías',
            missing_barcode: 'Falta Código de Barras/UPC',
            missing_barcode_desc: 'Cada producto debe tener un valor BarcodeUPC único.',
            duplicate_barcode: 'Código de Barras/UPC Duplicado Detectado',
            duplicate_barcode_desc: 'El BarcodeUPC "{barcode}" ya está en uso por el producto "{product}".',
            product_saved: 'Producto guardado con éxito',
            product_save_fail: 'Error al Guardar Producto',
            product_removed: 'Producto eliminado',
            product_remove_fail: 'Error al Eliminar Producto',
            confirm_delete: '¿Estás seguro de que quieres eliminar {name}?',
            category_saved: 'Categoría guardada',
            category_save_fail: 'Error al Guardar Categoría',
            category_removed: 'Categoría eliminada',
            category_remove_fail: 'Error al Eliminar Categoría',
            sub_category_saved: 'Sub-categoría guardada',
            sub_category_save_fail: 'Error al Guardar Sub-Categoría',
            sub_category_removed: 'Sub-categoría eliminada',
            sub_category_remove_fail: 'Error al Eliminar Sub-Categoría',
            supplier_saved: 'Proveedor guardado',
            supplier_save_fail: 'Error al Guardar Proveedor',
            supplier_removed: 'Proveedor eliminado',
            supplier_remove_fail: 'Error al Eliminar Proveedor',
            location_saved: 'Ubicación guardada',
            location_save_fail: 'Error al Guardar Ubicación',
            location_removed: 'Ubicación eliminada',
            location_remove_fail: 'Error al Eliminar Ubicación'
        }
    },
    operations: {
        hero: {
            title: 'Ajustes de Stock, Transferencias e Inteligencia de Código de Barras',
            subtitle: 'Control unificado en tiempo real para stock, movimiento y etiquetado.'
        },
        snapshot: {
            title: 'Instantánea de Inventario',
            subtitle: 'Ver saldo de producto y detalles de lote',
            search_placeholder: 'Buscar producto...',
            location_id: 'ID Ubicación (opcional)',
            fetch_button: 'Obtener niveles de stock',
            current_qty: 'Cantidad actual',
            table: {
                batch: 'Lote',
                qty: 'Cant.',
                expiry: 'Caducidad',
                empty: 'No hay detalles de lote disponible'
            }
        },
        adjustment: {
            title: 'Ajuste Manual',
            subtitle: 'Realizar conteos cíclicos ad-hoc o recepciones',
            select_product: 'Selecciona producto para ajustar...',
            stock_in: 'Entrada Stock (+)',
            stock_out: 'Salida Stock (-)',
            quantity: 'Cantidad',
            reason_code: 'Código de motivo',
            notes: 'Notas',
            submit_button: 'Aplicar ajuste'
        },
        transfer: {
            title: 'Transferencia de Stock',
            subtitle: 'Mover inventario entre ubicaciones',
            select_product: 'Selecciona producto para transferir...',
            source: 'Ubicación origen',
            dest: 'Ubicación destino',
            quantity: 'Cantidad',
            submit_button: 'Crear transferencia'
        },
        barcode: {
            title: 'Inteligencia de Código de Barras',
            subtitle: 'Buscar y generar códigos de barras para SKU',
            input_placeholder: 'Escanear o escribir código de barras / SKU',
            lookup_button: 'Buscar Producto',
            generate_button: 'Generar Imagen',
            preview_alt: 'Vista previa de código de barras'
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para acceder a operaciones.',
            product_id_required: 'Introduce un ID de producto primero',
            snapshot_updated: 'Instantánea de inventario actualizada',
            fetch_stock_fail: 'Error al Obtener Stock',
            select_product: 'Selecciona un producto',
            stock_adjusted: 'Stock ajustado',
            adjust_fail: 'Error al Aplicar Ajuste',
            transfer_queued: 'Transferencia en cola',
            transfer_fail: 'Error al Crear Transferencia',
            barcode_required: 'Proporciona un valor de código de barras',
            sku_resolved: 'SKU resuelto',
            lookup_fail: 'Error al Buscar Código de Barras',
            sku_or_id_required: 'Proporciona SKU o ID de producto',
            generate_fail: 'Error al Generar Código de Barras',
            product_not_found: 'Producto no encontrado'
        }
    },
    time_tracking: {
        hero: {
            title: 'Centro de Control de Seguimiento de Tiempo',
            subtitle: 'Mantente al tanto de los turnos, descansos y aprobaciones con un espacio de trabajo tranquilo diseñado para sentirse invisible. Cambia entre vistas personales y de gerente sin perder el estilo inspirado en Apple.',
            label: 'Inteligencia de Tiempo'
        },
        role_toggle: {
            staff: 'Vista Personal',
            manager: 'Vista Gerente',
            label: 'Seleccionar tablero'
        },
        staff: {
            header: {
                title: 'Mi Rastreador de Tiempo',
                subtitle: 'Flujo Personal',
                desc: 'Rastrea tu enfoque, descansos y progreso desde una superficie tranquila.'
            },
            status_card: {
                title: 'Mi Estado',
                clocked_in: 'Registrado Entrada',
                clocked_out: 'Registrado Salida',
                on_break: 'En Descanso',
                break_time: 'Tiempo de Descanso',
                today_total: "Total Hoy",
                clock_in_btn: 'Registrar Entrada',
                clock_out_btn: 'Registrar Salida',
                start_break_btn: 'Iniciar Descanso',
                end_break_btn: 'Terminar Descanso'
            },
            stats: {
                today_hours: "Horas Hoy",
                weekly_hours: 'Horas Semanales',
                target_label: 'de {target} horas objetivo'
            },
            task: {
                title: 'Tarea Actual',
                label: '¿En qué estás trabajando?',
                placeholder: 'Introduce tu tarea actual...',
                button: 'Actualizar Tarea'
            },
            goals: {
                title: 'Objetivos Diarios'
            },
            recent_shifts: {
                title: 'Turnos Recientes'
            }
        },
        manager: {
            header: {
                title: 'Tablero de Equipo',
                subtitle: 'Liderazgo',
                desc: 'Monitorea asistencia, turnos en vivo y momentum semanal.'
            },
            actions: {
                export: 'Exportar Informe',
                filter: 'Filtrar'
            },
            stats: {
                total_hours: 'Horas Totales',
                weekly_target_percent: '{percent}% del objetivo semanal',
                active_members: 'Miembros Activos',
                working: 'Trabajando actualmente',
                weekly_target: 'Objetivo Semanal',
                team_goal: 'Objetivo de equipo'
            },
            team: {
                title: 'Miembros del Equipo',
                status: {
                    working: 'Trabajando',
                    on_break: 'En Descanso',
                    offline: 'Desconectado'
                }
            },
            attendance: {
                title: 'Asistencia Semanal',
                present: 'Presente',
                late: 'Tarde',
                absent: 'Ausente'
            },
            recent_activity: {
                title: 'Actividad Reciente',
                clocked_in: 'registró entrada',
                started_break: 'inició descanso',
                clocked_out: 'registró salida'
            },
            quick_actions: {
                title: 'Acciones Rápidas',
                reports: 'Informes',
                schedule: 'Horario',
                payroll: 'Nómina',
                reminders: 'Recordatorios'
            }
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para acceder al seguimiento de tiempo.',
            clock_in_success: 'Registrada Entrada Con Éxito',
            clock_in_desc: 'Tu turno ha comenzado. ¡Que tengas un día productivo!',
            clock_out_info: 'Registrada Salida',
            clock_out_desc: '¡Gran trabajo hoy! Completaste {time} de tiempo enfocado.',
            break_start: 'Descanso Iniciado',
            break_start_desc: '¡Tómate un descanso bien merecido! Tu temporizador está pausado.',
            break_end: 'Descanso Terminado',
            break_end_desc: '¡Bienvenido de nuevo! ¿Listo para continuar tu día productivo?',
            task_updated: 'Tarea Actualizada',
            task_updated_desc: 'Ahora trabajando en: {task}',
            report_exported: 'Informe Exportado',
            report_exported_desc: 'El informe semanal de tiempo ha sido exportado con éxito.',
            reminder_sent: 'Recordatorio Enviado',
            reminder_sent_desc: 'Recordatorio enviado a {name} para completar su hoja de tiempo.',
            load_fail: 'Error al cargar datos',
            op_fail: 'Operación fallida'
        }
    },
    intelligence: {
        hero: {
            title: 'Pronóstico, Sugerencias de Reabastecimiento e Informes Comerciales',
            subtitle: 'Planifica con antelación, actúa sobre señales y alinea los análisis en un horizonte.'
        },
        demand_forecast: {
            title: 'Pronóstico de Demanda IA',
            subtitle: 'Predice la demanda futura para productos específicos',
            select_product: 'Seleccionar Producto',
            placeholder: 'Buscar por nombre o SKU...',
            period_label: 'Período de Pronóstico (Días)',
            generate_btn: 'Generar',
            generating_btn: 'Generando...',
            predicted_demand: 'Demanda Predicha',
            confidence: 'Confianza',
            reasoning: 'Razonamiento IA',
            generated_at: 'Generado en'
        },
        churn_risk: {
            title: 'Predicción de Abandono de Clientes',
            subtitle: 'Identifica clientes en riesgo y estrategias de retención',
            select_customer: 'Seleccionar Cliente',
            placeholder: 'Buscar por nombre o correo...',
            analyze_btn: 'Analizar Riesgo',
            analyzing_btn: 'Analizando...',
            risk_level: 'Nivel de Riesgo',
            risk_score: 'Puntuación de Riesgo',
            primary_factors: 'Factores Primarios',
            retention_strategy: 'Estrategia de Retención',
            suggested_action: 'Acción Sugerida',
            discount_offer: 'Ofrece un descuento del {discount}% para retener a este cliente.'
        },
        report_range: {
            title: 'Rango de Informe',
            subtitle: 'Alinear análisis en horizonte compartido',
            sales_trends: 'Tendencias de Ventas',
            turnover: 'Rotación de Inventario',
            margin: 'Margen de Beneficio'
        },
        reorder_suggestions: {
            title: 'Sugerencias de Reabastecimiento',
            subtitle: 'Órdenes de compra recomendadas por IA',
            refresh_btn: 'Refrescar',
            table: {
                product: 'Producto',
                supplier: 'Proveedor',
                suggested_qty: 'Cant. sugerida',
                status: 'Estado',
                actions: 'Acciones',
                create_po: 'Crear PO',
                empty: 'Sin sugerencias pendientes'
            }
        },
        reports: {
            period: 'Período: {period}',
            sales: {
                title: 'Informe de Ventas',
                subtitle: 'Tendencia de ventas totales vs promedio',
                total_sales: 'Ventas Totales',
                avg_daily_sales: 'Ventas Diarias Promedio'
            },
            turnover: {
                title: 'Informe de Rotación',
                subtitle: 'Eficiencia de inventario a lo largo del tiempo',
                avg_inventory_value: 'Valor Promedio Inventario',
                turnover_rate: 'Tasa de Rotación'
            },
            margin: {
                title: 'Informe de Margen',
                subtitle: 'Visualización de rentabilidad',
                gross_profit: 'Beneficio Bruto',
                total_revenue: 'Ingresos Totales'
            }
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para ver informes.',
            load_suggestions_fail: 'Error al Cargar Sugerencias',
            po_created: 'PO {id} creado',
            po_create_fail: 'Error al Crear PO',
            report_ready: 'Informe listo',
            report_fail: 'Error al Ejecutar Informe',
            suggestions_refreshed: 'Sugerencias refrescadas',
            refresh_fail: 'Error al refrescar sugerencias',
            select_product: 'Por favor selecciona un producto',
            forecast_success: 'Pronóstico generado con éxito',
            forecast_fail: 'Error al generar pronóstico',
            select_customer: 'Por favor selecciona un cliente',
            analysis_complete: 'Análisis completo',
            analysis_fail: 'Error al analizar riesgo de abandono'
        }
    },
    pos: {
        hero: {
            title: 'Consola de Pago Unificada',
            subtitle: 'Escanea, busca y completa pedidos con un flujo de baja fricción que se mantiene sincronizado con tu catálogo.',
            label: 'Punto de Venta',
            sub_label: 'Lienzo de pago en vivo para equipos de mostrador',
            new_sale_btn: 'Nueva venta presencial',
            refresh_catalog_btn: 'Refrescar catálogo'
        },
        header: {
            title: 'Punto de Venta',
            description: 'Toca productos para construir el carrito, revisa abajo, luego confirma a la derecha.',
            super_shop_mode: 'Modo Super Shop',
            search_placeholder: 'Buscar por nombre, código de barras o SKU...',
            search_btn: 'Buscar'
        },
        products: {
            title: 'Productos',
            description: 'Toca una ficha para añadirla al carrito activo.',
            results_found: '{count} resultados',
            filter_status: {
                label: 'Estado de Stock',
                all: 'Todos los Estados',
                in_stock: 'En Stock',
                low_stock: 'Stock Bajo',
                out_of_stock: 'Agotado'
            },
            no_results: 'No se encontraron productos. Intenta ajustar tu búsqueda.',
            in_stock: '{count} en stock',
            tap_to_add: 'Toca para añadir'
        },
        cart: {
            title: 'Carrito',
            empty_desc: 'No hay artículos añadidos aún.',
            items_desc: '{count} artículo{s} en carrito',
            clear_btn: 'Limpiar carrito',
            empty_state: 'Añade productos de la cuadrícula superior para iniciar un nuevo pedido.',
            table: {
                product: 'Producto',
                price: 'Precio',
                qty: 'Cant.',
                total: 'Total'
            }
        },
        customer: {
            title: 'Cliente',
            description: 'Adjuntar un cliente por ID, usuario, correo o teléfono. Opcional para ventas presenciales.',
            search_placeholder: 'Buscar por ID, usuario, correo, teléfono',
            new_btn: 'Nuevo',
            no_selected: 'Ningún cliente seleccionado. Aún puedes completar una venta presencial.',
            loyalty_pts: '{points} pts',
            tier: '{tier}'
        },
        payment: {
            title: 'Pago',
            description: 'Elige cómo paga el cliente este pedido.',
            methods: {
                cash: 'Efectivo',
                card: 'Tarjeta',
                bkash: 'bKash',
                other: 'Otro'
            },
            sub: {
                physical: 'Físico',
                terminal: 'Terminal',
                mobile: 'Móvil',
                check_due: 'Cheque/Deuda'
            }
        },
        loyalty: {
            redeem_label: 'Canjear Puntos de Lealtad',
            available: 'Disponible: {points} pts (valor {value})',
            points: 'puntos',
            error_exceed: 'No puede exceder el saldo disponible.'
        },
        summary: {
            title: 'Resumen del Pedido',
            description: 'Revisa totales y pago antes de confirmar la venta.',
            subtotal: 'Subtotal',
            tax: 'Impuesto ({rate}%)',
            total: 'Total',
            payment: 'Pago:',
            items: 'Artículos:',
            not_selected: 'No seleccionado',
            loyalty_earnings: 'Ganancias de Lealtad',
            complete_btn: 'Completar Pedido',
            processing_btn: 'Procesando...',
            add_items_hint: 'Añade artículos al carrito para continuar',
            select_payment_hint: 'Selecciona un método de pago para completar'
        },
        new_customer_modal: {
            title: 'Nuevo Cliente',
            description: 'Añadir un nuevo miembro a tu base de clientes.',
            name_label: 'Nombre Completo',
            name_placeholder: 'Ana García',
            email_label: 'Dirección de Correo',
            email_placeholder: 'ana@example.com',
            phone_label: 'Número de Teléfono',
            phone_placeholder: '+34 600 000 000',
            cancel_btn: 'Cancelar',
            create_btn: 'Crear Cliente'
        },
        toasts: {
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para acceder al TPV.',
            customer_not_found: 'Cliente no encontrado',
            search_error: 'Error al buscar cliente',
            out_of_stock: 'El producto está agotado',
            stock_limit_reached: 'No se puede añadir más. Solo {stock} artículos disponibles en stock.',
            processing: 'Procesando transacción...',
            order_success: '¡Pedido completado con éxito!',
            loyalty_earned: '¡El cliente ganó {points} puntos de lealtad!',
            loyalty_redeemed: 'Canjeados {points} puntos por {amount} de descuento.',
            transaction_fail: 'Transacción Fallida',
            name_required: 'El nombre es obligatorio',
            customer_created: '¡Cliente creado y seleccionado!',
            create_fail: 'Error al crear cliente'
        }
    },
    orders: {
        title: 'Gestión de Pedidos',
        subtitle: 'Gestiona facturas de ventas de clientes y órdenes de compra y devoluciones de proveedores.',
        tabs: {
            sales: 'Ventas (Facturas)',
            purchases: 'Compras y Devoluciones'
        },
        subtabs: {
            sales_orders: 'Pedidos de Venta',
            customer_returns: 'Devoluciones de Clientes',
            purchase_orders: 'Órdenes de Compra',
            vendor_returns: 'Devoluciones a Proveedores'
        },
        search: {
            sales: 'Buscar ventas...',
            pos: 'Buscar PO...',
            returns: 'Buscar devoluciones...'
        },
        status: {
            return_pending: 'DEVOLUCIÓN PENDIENTE',
            refund_value: 'Valor de Reembolso',
            items_returned: 'Artículos Devueltos',
            no_items: 'Sin detalle de artículos',
            total_items: 'Total de Artículos',
            qty: 'Cant.'
        },
        labels: {
            customer: 'Cliente',
            items: 'Artículos',
            more: 'más',
            supplier: 'Proveedor',
            refund: 'Reembolso',
            reason: 'Razón',
            guest: 'Invitado',
            unknown: 'Desconocido'
        },
        actions: {
            view_details: 'Ver Detalles',
            manage_po: 'Gestionar PO',
            try_again: 'Intentar De Nuevo'
        },
        empty: {
            sales: 'No se encontraron pedidos de venta.',
            customer_returns: 'No se encontraron devoluciones de clientes.',
            purchase_orders: 'No se encontraron órdenes de compra.',
            vendor_returns: 'No se encontraron devoluciones a proveedores.'
        },
        errors: {
            load_fail: 'Error al cargar historial de pedidos.',
            access_denied: 'Acceso Denegado',
            access_denied_desc: 'No tienes permiso para ver pedidos.'
        }
    },
    reports: {
        title: 'Suite de Inteligencia',
        subtitle: 'Perspectivas profundas sobre la eficiencia operativa y el rendimiento de ventas.',
        actions: {
            refresh: 'Refrescar Datos'
        },
        tabs: {
            sales: 'Ventas y Personal',
            inventory: 'Salud del Inventario',
            financial: 'Finanzas'
        },
        heatmap: {
            title: 'Intensidad de Ventas por Hora',
            subtitle: 'Horas pico de transacciones',
            description: 'Visualiza las horas más ocupadas de la semana según el volumen de ventas.'
        },
        staff: {
            title: 'Rendimiento del Personal',
            subtitle: 'Ventas por miembro del equipo',
            transactions: 'Transacciones',
            description: 'Rastrea las contribuciones de ventas individuales de los empleados y el conteo de transacciones.'
        },
        customers: {
            title: 'Mejores Clientes',
            subtitle: 'Clientes de mayor valor',
            spent: 'Gastado',
            orders: 'Pedidos',
            last_order: 'Último Pedido',
            unknown: 'Cliente Desconocido',
            orders_suffix: 'pedidos',
            days_ago: 'días atrás',
            lost: '¿Perdido?',
            lifetime_value: 'Valor de Vida',
            headers: {
                customer: 'Cliente',
                contact: 'Contacto',
                orders: 'Pedidos',
                last_order: 'Último Pedido',
                total_spent: 'Total Gastado'
            },
            table: {
                user: 'Usuario / Correo',
                name: 'Nombre',
                spent: 'Total Gastado',
                orders: 'Pedidos',
                last_order: 'Último Pedido',
                days_ago: 'Días Atrás'
            },
            description: 'Identifica a tus clientes más valiosos según el gasto total.'
        },
        frequency: {
            title: 'Análisis de Frecuencia',
            order_count: 'Conteo de Pedidos',
            orders: 'Pedidos',
            unknown: 'Producto Desconocido',
            description: 'Analiza con qué frecuencia se compran dos artículos juntos (Análisis de Cesta de Mercado). Ayuda en estrategias de venta cruzada y optimización de colocación de productos.'
        },
        stock_aging: {
            title: 'Envejecimiento de Stock',
            subtitle: 'Inventario por duración de antigüedad',
            headers: {
                product: 'Producto / SKU',
                age: 'Edad (Días)',
                quantity: 'Cant.',
                value: 'Valor'
            },
            days_suffix: 'd',
            description: 'Categoriza el inventario por cuánto tiempo ha estado en stock para identificar artículos estancados.'
        },
        dead_stock: {
            title: 'Stock Muerto (180+ Días)',
            days_idle: 'días inactivo',
            description: 'Lista de productos que no se han vendido en los últimos 180 días.'
        },
        supplier: {
            title: 'Fiabilidad del Proveedor',
            time: 'Tiempo',
            rate: 'Tasa',
            days_suffix: 'd',
            description: 'Evalúa a los proveedores según los tiempos de entrega y la precisión de los pedidos.'
        },
        category: {
            title: 'Desglose por Categoría',
            items: 'Artículos',
            sales: 'Ventas',
            margin: 'Margen',
            description: 'Muestra qué categorías de productos generan más ingresos y beneficios.'
        },
        financials: {
            revenue: 'Ingresos',
            cogs: 'COGS',
            margin: 'Margen',
            gmroi: 'GMROI',
            description: 'Métricas financieras clave que incluyen Ingresos, Costo de los Bienes Vendidos y Margen Bruto.',
            revenue_desc: 'Ingresos totales generados por ventas antes de deducir gastos.',
            cogs_desc: 'Costos directos atribuibles a la producción de los bienes vendidos.',
            margin_desc: 'El porcentaje de ingresos que excede el costo de los bienes vendidos.',
            gmroi_desc: 'Retorno de Margen Bruto sobre la Inversión. Mide la rentabilidad del inventario.'
        },
        void_analysis: {
            title: 'Análisis de Anulaciones',
            subtitle: 'Auditoría de transacciones canceladas',
            risk_score: 'Puntuación de Riesgo',
            voids: 'Anulaciones',
            risk: {
                high: 'Alto Riesgo',
                medium: 'Riesgo Medio',
                low: 'Bajo Riesgo'
            },
            description: 'Audita transacciones canceladas para identificar posibles fraudes o problemas de capacitación.'
        },
        tax: {
            title: 'Responsabilidad Tributaria',
            collected: 'Recaudado',
            rate: 'Tasa',
            taxable_sales: 'Ventas Gravables',
            description: 'Resume el impuesto recaudado y las ventas gravables por tasa impositiva.'
        },
        cash_reconciliation: {
            title: 'Conciliación de Efectivo',
            discrepancy: 'Discrepancia',
            description: 'Compara los registros del sistema con los recuentos de efectivo reales para encontrar discrepancias.'
        }
    }
};
