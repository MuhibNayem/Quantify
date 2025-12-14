export default {
    common: {
        save: 'Salva',
        saved_successfully: 'Salvato con successo',
        failed_to_save: 'Salvataggio fallito',
        access_denied: 'Accesso Negato',
        no_permission_settings: 'Non hai i permessi per visualizzare le impostazioni.',
        no_data_available: 'Nessun dato disponibile',
        configuration: 'Configurazione',
        manage_preferences: 'Gestisci preferenze di sistema, controlli di sicurezza e policy globali.',
    },
    settings: {
        tabs: {
            general: 'Generale',
            business_rules: 'Regole Aziendali',
            system_ai: 'Sistema & AI',
            security_roles: 'Sicurezza & Ruoli',
            policies: 'Policy',
            notifications: 'Notifiche'
        },
        general: {
            business_profile: 'Profilo Aziendale',
            business_profile_desc: "L'identità visibile della tua organizzazione sulla piattaforma.",
            business_name: 'Nome Azienda',
            currency: 'Valuta',
            timezone: 'Fuso Orario',
            locale: 'Lingua / Regione',
            select_currency: 'Seleziona Valuta',
            select_timezone: 'Seleziona Fuso Orario',
            select_locale: 'Seleziona Lingua'
        },
        business: {
            loyalty_program: 'Programma Fedeltà',
            loyalty_program_desc: 'Configura come i clienti guadagnano e riscattano punti, e imposta le soglie dei livelli.',
            points_configuration: 'Configurazione Punti',
            earning_rate: 'Tasso di Guadagno (Punti per $1)',
            earning_rate_hint: 'Quanti punti guadagna un cliente per ogni unità di valuta spesa.',
            redemption_value: 'Valore di Riscatto ($ per Punto)',
            redemption_value_hint: 'Il valore monetario di un singolo punto fedeltà al momento del riscatto.',
            tier_thresholds: 'Soglie Livelli',
            silver_tier: 'Livello Argento (Punti)',
            gold_tier: 'Livello Oro (Punti)',
            platinum_tier: 'Livello Platino (Punti)',
            financial_settings: 'Impostazioni Finanziarie',
            financial_settings_desc: 'Gestisci le aliquote fiscali e altri parametri finanziari.',
            default_tax_rate: 'Aliquota Fiscale Predefinita (%)',
            default_tax_rate_hint: 'Questa aliquota sarà applicata a tutte le vendite applicabili.'
        },
        policies: {
            privacy_policy: 'Informativa sulla Privacy',
            privacy_policy_placeholder: 'Inserisci la tua informativa sulla privacy (Markdown supportato)...',
            save_policy: 'Salva Policy',
            terms_of_service: 'Termini di Servizio',
            terms_of_service_placeholder: 'Inserisci i tuoi termini di servizio (Markdown supportato)...',
            save_terms: 'Salva Termini',
            return_policy: 'Politica di Reso',
            return_window: 'Finestra di Reso (Giorni)',
            return_window_hint: 'Numero di giorni dopo l\'acquisto in cui un cliente può richiedere un reso.'
        },
        system: {
            ai_system: 'AI & Sistema',
            ai_system_desc: 'Configura il comportamento degli agenti autonomi e i parametri a livello di sistema.',
            wake_up_time: 'Orario di Risveglio AI',
            wake_up_time_hint: 'L\'AI eseguirà il "Controllo Quotidiano Mattutino" a quest\'ora ogni giorno.'
        },
        notifications: {
            global_alerts_center: 'Centro Avvisi Globale',
            coming_soon: 'Il routing avanzato delle notifiche e la configurazione dei webhook arriveranno nel prossimo aggiornamento.'
        }
    },
    roles: {
        title: 'Ruoli & Accesso',
        create_new: 'Crea Nuovo Ruolo',
        create_first: 'Crea Primo Ruolo',
        system_managed: 'Gestito dal Sistema',
        custom_role: 'Ruolo Personalizzato',
        active_permissions: 'Permessi Attivi',
        delete_confirm: 'Sei sicuro di voler eliminare il ruolo "{name}"?',
        delete: 'Elimina',
        save_changes: 'Salva Modifiche',
        saving: 'Salvataggio...',
        reset: 'Ripristina',
        role_name: 'Nome Ruolo',
        description: 'Descrizione',
        capabilities: 'Capacità',
        capabilities_desc: 'Affina i controlli di accesso per questo ruolo',
        select_all: 'Seleziona Tutto',
        security_access_control: 'Sicurezza & Controllo Accessi',
        security_desc: 'Seleziona un ruolo dalla barra laterale per configurare i permessi, o crea un nuovo ruolo personalizzato per delegare specifiche capacità di accesso.',
        names: {
            admin: 'Admin',
            manager: 'Manager',
            cashier: 'Cassiere',
            staff: 'Staff',
            customer: 'Cliente',
            sales_associate: 'Addetto Vendite'
        }
    },
    permissions: {
        groups: {
            inventory: 'Magazzino',
            sales: 'Vendite',
            crm: 'CRM',
            hrm: 'HRM',
            settings: 'Impostazioni',
            reports: 'Report',
            pos: 'POS',
            dashboard: 'Dashboard',
            'product management': 'Gestione Prodotti',
            'access control': 'Controllo Accessi',
            orders: 'Ordini',
            system: 'Sistema'
        },
        names: {
            'roles_manage': 'Gestisci Ruoli',
            'roles_view': 'Visualizza Ruoli',
            'users_manage': 'Gestisci Utenti',
            'users_view': 'Visualizza Utenti',
            'crm_read': 'Leggi Dati CRM',
            'crm_view': 'Visualizza Modulo CRM',
            'crm_write': 'Modifica Dati CRM',
            'customers_read': 'Visualizza Clienti',
            'customers_write': 'Gestisci Clienti',
            'loyalty_read': 'Visualizza Info Fedeltà',
            'loyalty_write': 'Gestisci Punti Fedeltà',
            'alerts_manage': 'Risolvi Avvisi',
            'alerts_view': 'Visualizza Avvisi',
            'barcode_read': 'Cerca Codici a Barre',
            'inventory_read': 'Leggi Dati Magazzino',
            'inventory_view': 'Visualizza Modulo Magazzino',
            'inventory_write': 'Modifica Dati Magazzino',
            'locations_read': 'Visualizza Sedi',
            'locations_write': 'Gestisci Sedi',
            'replenishment_read': 'Visualizza Previsioni/Suggerimenti',
            'replenishment_write': 'Genera Previsioni & Gestisci PO',
            'suppliers_read': 'Visualizza Fornitori',
            'suppliers_write': 'Gestisci Fornitori',
            'orders_manage': 'Gestisci Ordini',
            'orders_read': 'Visualizza Storico Ordini',
            'pos_access': 'Accedi al Terminale POS',
            'pos_view': 'Visualizza Modulo POS',
            'returns_manage': 'Approva/Rifiuta Resi',
            'returns_request': 'Richiedi Reso',
            'categories_read': 'Visualizza Categorie',
            'categories_write': 'Gestisci Categorie',
            'products_delete': 'Elimina Prodotti',
            'products_read': 'Visualizza Prodotti',
            'products_write': 'Crea/Modifica Prodotti',
            'reports_financial': 'Visualizza Report Finanziari',
            'reports_inventory': 'Visualizza Report Magazzino',
            'reports_sales': 'Visualizza Report Vendite',
            'reports_view': 'Visualizza Report',
            'settings_manage': 'Modifica Impostazioni Sistema',
            'settings_view': 'Visualizza Impostazioni Sistema',
            'bulk_export': 'Esporta Dati',
            'bulk_import': 'Importa Dati',
            'dashboard_view': 'Visualizza Dashboard',
            'notifications_read': 'Visualizza Notifiche',
            'notifications_write': 'Gestisci Notifiche'
        },
        descriptions: {
            'roles_manage': 'Gestisci ruoli e permessi',
            'roles_view': 'Visualizza ruoli esistenti',
            'users_manage': 'Aggiungi o modifica account utente',
            'users_view': 'Visualizza lista utenti',
            'crm_read': 'Leggi dati relazioni clienti',
            'crm_view': 'Accedi interfaccia CRM',
            'crm_write': 'Aggiorna record CRM',
            'customers_read': 'Visualizza dettagli cliente',
            'customers_write': 'Aggiungi o modifica info cliente',
            'loyalty_read': 'Visualizza saldo punti fedeltà',
            'loyalty_write': 'Regola punti fedeltà',
            'alerts_manage': 'Agisci su avvisi di sistema',
            'alerts_view': 'Visualizza avvisi attivi',
            'barcode_read': 'Scansiona e cerca codici a barre',
            'inventory_read': 'Visualizza livelli stock',
            'inventory_view': 'Accedi gestione magazzino',
            'inventory_write': 'Aggiorna quantità stock',
            'locations_read': 'Visualizza sedi magazzini e negozi',
            'locations_write': 'Aggiungi o modifica dettagli sede',
            'replenishment_read': 'Visualizza suggerimenti riassortimento',
            'replenishment_write': 'Crea ordini di acquisto',
            'suppliers_read': 'Visualizza lista fornitori',
            'suppliers_write': 'Gestisci info fornitori',
            'orders_manage': 'Elabora ordini di vendita',
            'orders_read': 'Visualizza storico vendite',
            'pos_access': 'Usa punto vendita',
            'pos_view': 'Visualizza schermate POS',
            'returns_manage': 'Gestisci resi clienti',
            'returns_request': 'Inizia procedura di reso',
            'categories_read': 'Visualizza categorie prodotti',
            'categories_write': 'Gestisci struttura categorie',
            'products_delete': 'Rimuovi prodotti dal catalogo',
            'products_read': 'Visualizza dettagli prodotto',
            'products_write': 'Modifica o aggiungi info prodotto',
            'reports_financial': 'Visualizza report performance finanziarie',
            'reports_inventory': 'Visualizza report stato magazzino',
            'reports_sales': 'Visualizza report analisi vendite',
            'reports_view': 'Accedi reportistica generale',
            'settings_manage': 'Cambia configurazione applicazione',
            'settings_view': 'Visualizza impostazioni configurazione',
            'bulk_export': 'Scarica dati in blocco',
            'bulk_import': 'Carica dati in blocco',
            'dashboard_view': 'Visualizza panoramica dashboard principale',
            'notifications_read': 'Leggi notifiche di sistema',
            'notifications_write': 'Aggiorna stato notifiche'
        }
    },
    users: {
        title: 'Gestione Accesso Utenti',
        subtitle: 'Approva, modifica o revoca accessi allo spazio di lavoro — con filtri live e aggiornamenti sicuri.',
        badges: {
            role_control: 'Controllo basato su ruoli',
            status_filters: 'Filtri stato',
            inline_edits: 'Modifiche in linea'
        },
        filters: {
            search_placeholder: 'Cerca per username o ID',
            search_btn: 'Cerca',
            all: 'Tutti gli utenti',
            approved: 'Approvati',
            pending: 'In attesa'
        },
        table: {
            id: 'ID',
            username: 'Username',
            role: 'Ruolo',
            status: 'Stato',
            actions: 'Azioni',
            empty: 'Nessun utente corrisponde a questo filtro',
            approve: 'Approva',
            edit: 'Modifica',
            delete: 'Elimina'
        },
        form: {
            title: 'Dettagli Utente',
            subtitle: 'Aggiorna ruolo o credenziali per l\'utente selezionato',
            select_prompt: 'Seleziona un utente dalla tabella per modificare l\'accesso.',
            editing: 'Modifica in corso',
            sections: {
                credentials: 'Credenziali Account',
                personal: 'Informazioni Personali'
            },
            fields: {
                username: 'Username',
                password_placeholder: 'Reimposta password (opzionale)',
                select_role: 'Seleziona un ruolo',
                first_name: 'Nome',
                last_name: 'Cognome',
                email: 'Email',
                phone: 'Numero di Telefono',
                address: 'Indirizzo'
            },
            buttons: {
                save: 'Salva modifiche',
                approve: 'Approva',
                delete: 'Elimina'
            },
            read_only: 'Hai accesso in sola lettura alla gestione utenti.'
        },
        status: {
            approved: 'Approvato',
            pending: 'In attesa',
            active_hint: 'Accesso attivo',
            pending_hint: 'In attesa di approvazione'
        }
    },
    bulk: {
        title: 'Operazioni Massive',
        subtitle: 'Gestisci il tuo catalogo in modo efficiente con importazioni, esportazioni massive e tracciamento job.',
        tabs: {
            import: 'Importa',
            export: 'Esporta',
            status: 'Stato'
        },
        steps: {
            download: 'Scarica Template',
            upload: 'Carica',
            validate: 'Convalida',
            import: 'Importa',
            done: 'Fatto',
            step1_title: 'Step 1: Scarica & Prepara',
            step1_desc: 'Usa il formato CSV corretto per un\'importazione senza intoppi.',
            step2_title: 'Step 2: Carica File',
            step2_desc: 'Convalida il tuo CSV prima di importare i prodotti.',
            step3_title: 'Step 3: Rivedi & Conferma',
            step3_desc: 'Assicurati che la convalida sia passata prima dell\'importazione finale.'
        },
        buttons: {
            download_template: 'Scarica Template',
            upload_validate: 'Carica & Convalida',
            new_import: 'Nuova Importazione',
            try_again: 'Riprova',
            generate_export: 'Genera Esportazione',
            download_export: 'Scarica Esportazione',
            refresh: 'Aggiorna'
        },
        labels: {
            valid: 'Valido',
            invalid: 'Non valido',
            total: 'Totale',
            new_categories: 'Nuove Categorie',
            new_suppliers: 'Nuovi Fornitori',
            new_locations: 'Nuove Sedi',
            valid_records: 'Record Validi',
            invalid_records: 'Record Non Validi',
            changes: 'Modifiche',
            processed: 'Elaborati',
            success_rate: 'Tasso di Successo',
            breakdown: 'Dettaglio',
            format: 'Formato',
            category: 'Categoria (Opzionale)',
            supplier: 'Fornitore (Opzionale)',
            all_categories: 'Tutte le Categorie',
            all_suppliers: 'Tutti i Fornitori',
            search_placeholder: 'Cerca per ID Job...',
            no_history: 'Nessuna Cronologia',
            no_history_desc: 'I job recenti di importazione ed esportazione appariranno qui.',
            select_job: 'Seleziona un job per vedere i dettagli'
        },
        status: {
            import_complete: 'Importazione Completata',
            import_failed: 'Importazione Fallita',
            validating: 'Convalida file...',
            importing: 'Importazione prodotti...',
            success: 'Successo'
        }
    },
    alerts: {
        title: 'Controllo Avvisi & Notifiche',
        subtitle: 'Soglie, escalation e messaggistica utente mirata — tutto in una console rilassante e vibrante.',
        refresh: 'Aggiorna Avvisi',
        go_to_ops: 'Vai alle Operazioni',
        live_alerts: 'Avvisi Live',
        live_alerts_desc: 'Filtra per tipo o stato del ciclo di vita',
        filters: {
            placeholder_type: 'Tutti i tipi',
            placeholder_status: 'Qualsiasi stato',
            refresh_btn: 'Aggiorna',
            type_options: {
                low_stock: 'Scorta bassa',
                overstock: 'Scorta eccessiva',
                out_of_stock: 'Esaurito',
                expiry: 'Scadenza'
            },
            status_options: {
                active: 'Attivo',
                resolved: 'Risolto'
            }
        },
        table: {
            type: 'Tipo',
            product: 'Prodotto',
            message: 'Messaggio',
            status: 'Stato',
            action: 'Azione',
            empty: 'Nessun avviso trovato',
            resolve: 'Risolvi',
            product_fallback: 'Prodotto'
        },
        details: {
            title: 'Dettagli Avviso',
            type: 'Tipo Avviso',
            status: 'Stato',
            triggered: 'Attivato',
            context: 'Contesto Avviso',
            product: 'Prodotto',
            message: 'Messaggio',
            product_id: 'ID Prodotto',
            triggered_at: 'Attivato il',
            updated_at: 'Aggiornato il',
            batch_details: 'Dettagli Lotto',
            batch_id: 'ID Lotto',
            batch_number: 'Numero Lotto',
            quantity: 'Quantità',
            expiry_date: 'Data Scadenza',
            resolved_at: 'Risolto',
            awaiting_action: 'In attesa di azione',
            na: 'N/A'
        },
        thresholds: {
            title: 'Soglie Prodotto',
            subtitle: 'Configura avvisi per SKU',
            product_label: 'Prodotto',
            product_placeholder: 'Cerca prodotto...',
            low_stock: 'Scorta bassa',
            overstock: 'Scorta eccessiva',
            expiry_days: 'Giorni alla scadenza',
            save: 'Salva soglie'
        },
        notifications: {
            title: 'Notifiche Utente',
            subtitle: 'Preferenze di escalation per operatore',
            user_label: 'Utente',
            user_placeholder: 'Cerca utente...',
            email_placeholder: 'Email',
            phone_placeholder: 'Telefono',
            email_label: 'Email',
            sms_label: 'SMS',
            save: 'Salva preferenze'
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per visualizzare gli avvisi.',
            load_fail: 'Impossibile caricare gli avvisi',
            resolve_success: 'Avviso risolto',
            resolve_fail: 'Impossibile risolvere avviso',
            select_product: 'Seleziona un prodotto',
            thresholds_updated: 'Soglie aggiornate',
            thresholds_fail: 'Impossibile salvare soglie',
            provide_user: 'Fornisci un ID utente',
            prefs_saved: 'Preferenze notifica salvate',
            prefs_fail: 'Impossibile salvare preferenze'
        }
    },
    dashboard: {
        title: 'Intelligenza Magazzino Real-time',
        subtitle: 'Monitora, analizza e ottimizza il tuo ecosistema di inventario con insight potenziati dall\'AI',
        refresh: 'Aggiorna Dati',
        update_catalog: 'Aggiorna Catalogo',
        stats: {
            active_products: 'Prodotti Attivi',
            categories: 'Categorie',
            suppliers: 'Fornitori',
            active_alerts: 'Avvisi Attivi',
            forecast_hint: '📈 {value} previsti per Q4',
            supplier_hint: '🔄 Attraverso {count} fornitori',
            sla_hint: '✅ Tutti gli SLA attivi',
            escalation_hint: '🚨 Auto-escalation attive'
        },
        demand: {
            title: 'Analisi Pulsazione Domanda',
            subtitle: 'Trend di movimento magazzino in tempo reale',
            chart_hint: '📊 Basato su velocità vendite & buffer stock',
            trend_positive: '↑ Trend: Positivo',
            trend_negative: '↓ Trend: Negativo',
            trend_stable: '→ Trend: Stabile',
            growth: '📈 {value}% Crescita',
            decline: '📉 {value}% Calo',
            no_change: 'Nessun Cambiamento',
            day_label: 'Giorno {day}'
        },
        quick_actions: {
            title: 'Azioni Rapide',
            subtitle: 'Operazioni magazzino istantanee',
            balance_stock: 'Bilancia Stock',
            balance_desc: 'Ottimizza livelli inventario',
            run_forecast: 'Esegui Previsione',
            forecast_desc: 'Predizioni AI',
            export_catalog: 'Esporta Catalogo',
            export_desc: 'Operazioni massive'
        },
        fresh_inventory: {
            title: 'Inventario Fresco',
            subtitle: 'SKU aggiunti o aggiornati di recente',
            sku: 'SKU',
            product_name: 'Nome Prodotto',
            status: 'Stato',
            no_data: 'Nessun cambiamento recente inventory'
        },
        priority_alerts: {
            title: 'Avvisi Prioritari',
            subtitle: 'Richiede attenzione immediata',
            type: 'Tipo Avviso',
            product: 'Prodotto',
            status: 'Stato',
            no_data: 'Tutti i sistemi normali'
        },
        procurement: {
            title: 'Intelligenza Approvvigionamento',
            subtitle: 'Suggerimenti riordino potenziati dall\'AI',
            product: 'Prodotto',
            suggested_qty: 'Qtà Suggerita',
            supplier: 'Fornitore',
            status: 'Stato',
            no_data: 'Nessun suggerimento in sospeso',
            ready_to_order: 'Pronto per Ordinare'
        },
        toasts: {
            load_fail: 'Impossibile Caricare Dashboard',
            error_desc: 'Si è verificato un errore imprevisto'
        }
    },
    catalog: {
        hero: {
            subtitle: 'COCKPIT CATALOGO',
            title: 'Prodotti, Categorie & Partner',
            description: 'Centro di controllo unificato per i dati del tuo catalogo',
            sync_data: 'Sincronizza dati',
            bulk_import: 'Importazione massiva'
        },
        tabs: {
            products: 'Prodotti',
            categories: 'Categorie',
            sub_categories: 'Sottocategorie',
            suppliers: 'Fornitori',
            locations: 'Sedi'
        },
        common: {
            search: 'Cerca',
            clear: 'Pulisci',
            add: 'Aggiungi',
            edit: 'Modifica',
            delete: 'Elimina',
            save: 'Salva',
            create: 'Crea',
            update: 'Aggiorna',
            reset: 'Ripristina',
            actions: 'Azioni',
            new: 'Nuovo',
            loading: 'Caricamento...'
        },
        products: {
            title: 'Registro SKU',
            subtitle: 'Gestisci articoli sincronizzati con il magazzino',
            search_placeholder: 'Cerca prodotti...',
            add_button: 'Aggiungi Prodotto',
            columns: {
                sku: 'SKU',
                name: 'Nome',
                status: 'Stato'
            },
            form: {
                update_title: 'Aggiorna prodotto',
                create_title: 'Crea prodotto',
                subtitle: 'Metadati livello SKU',
                sku: 'SKU',
                name: 'Nome',
                description: 'Descrizione',
                barcode: 'Barcode / UPC (deve essere unico)',
                select_category: 'Seleziona categoria',
                select_sub_category: 'Seleziona sottocategoria',
                select_supplier: 'Seleziona fornitore',
                default_location: 'Sede predefinita',
                purchase_price: 'Prezzo acquisto',
                selling_price: 'Prezzo vendita'
            }
        },
        categories: {
            title: 'Categorie',
            subtitle: 'Struttura le fondamenta del tuo catalogo',
            search_placeholder: 'Cerca per nome...',
            form: {
                update_title: 'Aggiorna categoria',
                create_title: 'Crea categoria',
                name: 'Nome'
            }
        },
        sub_categories: {
            title: 'Sottocategorie',
            subtitle: 'Filtra per categoria genitore',
            select_category: 'Seleziona categoria per vedere le sottocategorie',
            loading_categories: 'Caricamento categorie...',
            empty_state: 'Seleziona una categoria sopra',
            empty_subtitle: 'Le sottocategorie appariranno qui',
            form: {
                update_title: 'Aggiorna sottocategoria',
                create_title: 'Crea sottocategoria',
                name: 'Nome',
                select_category: 'Seleziona categoria'
            }
        },
        suppliers: {
            title: 'Fornitori',
            subtitle: 'Partner strategici che alimentano il riassortimento',
            search_placeholder: 'Cerca per nome...',
            columns: {
                contact: 'Contatto'
            },
            form: {
                update_title: 'Aggiorna fornitore',
                create_title: 'Crea fornitore',
                name: 'Nome',
                contact_person: 'Persona di contatto',
                email: 'Email',
                phone: 'Telefono',
                address: 'Indirizzo'
            }
        },
        locations: {
            title: 'Sedi',
            subtitle: 'Nodi di logistica e negozi',
            columns: {
                address: 'Indirizzo'
            },
            form: {
                update_title: 'Aggiorna sede',
                create_title: 'Crea sede',
                name: 'Nome',
                address: 'Indirizzo'
            }
        },
        details: {
            id: 'ID',
            sku: 'SKU',
            name: 'Nome',
            status: 'Stato',
            purchase_price: 'Prezzo Acquisto',
            selling_price: 'Prezzo Vendita',
            category_id: 'ID Categoria',
            supplier_id: 'ID Fornitore',
            description: 'Descrizione',
            created: 'Creato',
            updated: 'Aggiornato',
            parent_category: 'Categoria Genitore',
            sub_category_id: 'ID Sottocategoria',
            on_time_rate: 'Tasso puntualità',
            avg_lead_time: 'Lead time medio',
            supplier_id_label: 'ID Fornitore',
            contact_details: 'Dettagli Contatto',
            contact_person: 'Persona di Contatto',
            email: 'Email',
            phone: 'Telefono',
            address: 'Indirizzo',
            performance: 'Snapshot Performance',
            location_id: 'ID Sede',
            location_profile: 'Profilo Sede'
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per visualizzare il catalogo.',
            load_fail: 'Impossibile Caricare Catalogo',
            search_fail: 'Ricerca Fallita',
            product_not_found: 'Prodotto non trovato',
            category_not_found: 'Categoria non trovata',
            supplier_not_found: 'Fornitore non trovato',
            sub_categories_fail: 'Impossibile Caricare Sottocategorie',
            missing_barcode: 'Barcode/UPC Mancante',
            missing_barcode_desc: 'Ogni prodotto deve avere un valore BarcodeUPC univoco.',
            duplicate_barcode: 'Rilevato Barcode/UPC Duplicato',
            duplicate_barcode_desc: 'Il BarcodeUPC "{barcode}" è già usato dal prodotto "{product}".',
            product_saved: 'Prodotto salvato con successo',
            product_save_fail: 'Impossibile Salvare Prodotto',
            product_removed: 'Prodotto rimosso',
            product_remove_fail: 'Impossibile Eliminare Prodotto',
            confirm_delete: 'Sei sicuro di voler eliminare {name}?',
            category_saved: 'Categoria salvata',
            category_save_fail: 'Impossibile Salvare Categoria',
            category_removed: 'Categoria rimossa',
            category_remove_fail: 'Impossibile Eliminare Categoria',
            sub_category_saved: 'Sottocategoria salvata',
            sub_category_save_fail: 'Impossibile Salvare Sottocategoria',
            sub_category_removed: 'Sottocategoria rimossa',
            sub_category_remove_fail: 'Impossibile Eliminare Sottocategoria',
            supplier_saved: 'Fornitore salvato',
            supplier_save_fail: 'Impossibile Salvare Fornitore',
            supplier_removed: 'Fornitore rimosso',
            supplier_remove_fail: 'Impossibile Eliminare Fornitore',
            location_saved: 'Sede salvata',
            location_save_fail: 'Impossibile Salvare Sede',
            location_removed: 'Sede rimossa',
            location_remove_fail: 'Impossibile Eliminare Sede'
        }
    },
    operations: {
        hero: {
            title: 'Rettifiche Stock, Trasferimenti & Intelligenza Barcode',
            subtitle: 'Controllo unificato in tempo reale per stock, movimentazione & etichettatura.'
        },
        snapshot: {
            title: 'Snapshot Inventario',
            subtitle: 'Visualizza saldo prodotto e dettagli lotto',
            search_placeholder: 'Cerca prodotto...',
            location_id: 'ID Sede (opzionale)',
            fetch_button: 'Recupera livelli stock',
            current_qty: 'Quantità attuale',
            table: {
                batch: 'Lotto',
                qty: 'Qtà',
                expiry: 'Scadenza',
                empty: 'Nessun dettaglio lotto disponibile'
            }
        },
        adjustment: {
            title: 'Rettifica Manuale',
            subtitle: 'Esegui conteggi ciclici estemporanei o ricezioni',
            select_product: 'Seleziona prodotto da rettificare...',
            stock_in: 'Stock In (+)',
            stock_out: 'Stock Out (-)',
            quantity: 'Quantità',
            reason_code: 'Codice motivo',
            notes: 'Note',
            submit_button: 'Applica rettifica'
        },
        transfer: {
            title: 'Trasferimento Stock',
            subtitle: 'Sposta inventario tra le sedi',
            select_product: 'Seleziona prodotto da trasferire...',
            source: 'Sede origine',
            dest: 'Sede destinazione',
            quantity: 'Quantità',
            submit_button: 'Crea trasferimento'
        },
        barcode: {
            title: 'Intelligenza Barcode',
            subtitle: 'Cerca e genera codici a barre per SKU',
            input_placeholder: 'Scansiona o digita barcode / SKU',
            lookup_button: 'Cerca Prodotto',
            generate_button: 'Genera Immagine',
            preview_alt: 'Anteprima Barcode'
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per accedere alle operazioni.',
            product_id_required: 'Inserisci prima un ID prodotto',
            snapshot_updated: 'Snapshot inventario aggiornato',
            fetch_stock_fail: 'Impossibile Recuperare Stock',
            select_product: 'Seleziona un prodotto',
            stock_adjusted: 'Stock rettificato',
            adjust_fail: 'Impossibile Applicare Rettifica',
            transfer_queued: 'Trasferimento in coda',
            transfer_fail: 'Impossibile Creare Trasferimento',
            barcode_required: 'Fornisci un valore barcode',
            sku_resolved: 'SKU risolto',
            lookup_fail: 'Impossibile Cercare Barcode',
            sku_or_id_required: 'Fornisci SKU o ID prodotto',
            generate_fail: 'Impossibile Generare Barcode',
            product_not_found: 'Prodotto non trovato'
        }
    },
    time_tracking: {
        hero: {
            title: 'Centro Controllo Time Tracking',
            subtitle: 'Rimani aggiornato su turni, pause e approvazioni con uno spazio di lavoro calmo progettato per sembrare invisibile. Passa tra le viste personale e manager senza perdere lo stile Apple-inspired.',
            label: 'Time Intelligence'
        },
        role_toggle: {
            staff: 'Vista Staff',
            manager: 'Vista Manager',
            label: 'Seleziona dashboard'
        },
        staff: {
            header: {
                title: 'Il Mio Time Tracker',
                subtitle: 'Flusso Personale',
                desc: 'Traccia la tua concentrazione, le pause e i progressi da un\'unica superficie calma.'
            },
            status_card: {
                title: 'Il Mio Stato',
                clocked_in: 'Entrato',
                clocked_out: 'Uscito',
                on_break: 'In Pausa',
                break_time: 'Tempo Pausa',
                today_total: "Totale Oggi",
                clock_in_btn: 'Entra',
                clock_out_btn: 'Esci',
                start_break_btn: 'Inizia Pausa',
                end_break_btn: 'Termina Pausa'
            },
            stats: {
                today_hours: "Ore Oggi",
                weekly_hours: 'Ore Settimanali',
                target_label: 'su {target} ore obiettivo'
            },
            task: {
                title: 'Task Corrente',
                label: 'Su cosa stai lavorando?',
                placeholder: 'Inserisci il tuo task corrente...',
                button: 'Aggiorna Task'
            },
            goals: {
                title: 'Obiettivi Giornalieri'
            },
            recent_shifts: {
                title: 'Turni Recenti'
            }
        },
        manager: {
            header: {
                title: 'Dashboard Team',
                subtitle: 'Leadership',
                desc: 'Monitora presenze, turni live e slancio settimanale.'
            },
            actions: {
                export: 'Esporta Report',
                filter: 'Filtra'
            },
            stats: {
                total_hours: 'Ore Totali',
                weekly_target_percent: '{percent}% dell\'obiettivo settimanale',
                active_members: 'Membri Attivi',
                working: 'Attualmente al lavoro',
                weekly_target: 'Obiettivo Settimanale',
                team_goal: 'Obiettivo Team'
            },
            team: {
                title: 'Membri del Team',
                status: {
                    working: 'Al Lavoro',
                    on_break: 'In Pausa',
                    offline: 'Offline'
                }
            },
            attendance: {
                title: 'Presenza Settimanale',
                present: 'Presente',
                late: 'Ritardo',
                absent: 'Assente'
            },
            recent_activity: {
                title: 'Attività Recente',
                clocked_in: 'è entrato',
                started_break: 'ha iniziato pausa',
                clocked_out: 'è uscito'
            },
            quick_actions: {
                title: 'Azioni Rapide',
                reports: 'Report',
                schedule: 'Orari',
                payroll: 'Paghe',
                reminders: 'Promemoria'
            }
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per accedere al time tracking.',
            clock_in_success: 'Entrata Registrata Correttamente',
            clock_in_desc: 'Il tuo turno è iniziato. Buona giornata produttiva!',
            clock_out_info: 'Uscita Registrata',
            clock_out_desc: 'Ottimo lavoro oggi! Hai completato {time} di tempo focalizzato.',
            break_start: 'Pausa Iniziata',
            break_start_desc: 'Prenditi una meritata pausa! Il tuo timer è in pausa.',
            break_end: 'Pausa Terminata',
            break_end_desc: 'Bentornato! Pronto a continuare la tua giornata produttiva?',
            task_updated: 'Task Aggiornato',
            task_updated_desc: 'Ora stai lavorando su: {task}',
            report_exported: 'Report Esportato',
            report_exported_desc: 'Il report settimanale è stato esportato con successo.',
            reminder_sent: 'Promemoria Inviato',
            reminder_sent_desc: 'Promemoria inviato a {name} per completare il foglio presenze.',
            load_fail: 'Impossibile caricare i dati',
            op_fail: 'Operazione fallita'
        }
    },
    intelligence: {
        hero: {
            title: 'Previsioni, Suggerimenti Riordino & Report Aziendali',
            subtitle: 'Pianifica in anticipo, agisci sui segnali e allinea le analisi su un unico orizzonte.'
        },
        demand_forecast: {
            title: 'Previsione Domanda AI',
            subtitle: 'Predici la domanda futura per prodotti specifici',
            select_product: 'Seleziona Prodotto',
            placeholder: 'Cerca per nome o SKU...',
            period_label: 'Periodo Previsione (Giorni)',
            generate_btn: 'Genera',
            generating_btn: 'Generazione...',
            predicted_demand: 'Domanda Prevista',
            confidence: 'Confidenza',
            reasoning: 'Ragionamento AI',
            generated_at: 'Generato il'
        },
        churn_risk: {
            title: 'Previsione Abbandono Clienti',
            subtitle: 'Identifica clienti a rischio e strategie di ritenzione',
            select_customer: 'Seleziona Cliente',
            placeholder: 'Cerca per nome o email...',
            analyze_btn: 'Analizza Rischio',
            analyzing_btn: 'Analisi...',
            risk_level: 'Livello Rischio',
            risk_score: 'Punteggio Rischio',
            primary_factors: 'Fattori Primari',
            retention_strategy: 'Strategia Ritenzione',
            suggested_action: 'Azione Suggerita',
            discount_offer: 'Offri uno sconto del {discount}% per trattenere questo cliente.'
        },
        report_range: {
            title: 'Intervallo Report',
            subtitle: 'Allinea analisi su orizzonte condiviso',
            sales_trends: 'Trend Vendite',
            turnover: 'Rotazione Inventario',
            margin: 'Margine Profitto'
        },
        reorder_suggestions: {
            title: 'Suggerimenti Riordino',
            subtitle: 'Ordini di acquisto raccomandati dall\'AI',
            refresh_btn: 'Aggiorna',
            table: {
                product: 'Prodotto',
                supplier: 'Fornitore',
                suggested_qty: 'Qtà suggerita',
                status: 'Stato',
                actions: 'Azioni',
                create_po: 'Crea PO',
                empty: 'Nessun suggerimento in sospeso'
            }
        },
        reports: {
            period: 'Periodo: {period}',
            sales: {
                title: 'Report Vendite',
                subtitle: 'Trend vendite totali vs medie',
                total_sales: 'Vendite Totali',
                avg_daily_sales: 'Vendite Giornaliere Medie'
            },
            turnover: {
                title: 'Report Rotazione',
                subtitle: 'Efficienza inventario nel tempo',
                avg_inventory_value: 'Valore Inventario Medio',
                turnover_rate: 'Tasso Rotazione'
            },
            margin: {
                title: 'Report Margine',
                subtitle: 'Visualizzazione reddittività',
                gross_profit: 'Profitto Lordo',
                total_revenue: 'Ricavi Totali'
            }
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per visualizzare i report.',
            load_suggestions_fail: 'Impossibile Caricare Suggerimenti',
            po_created: 'PO {id} creato',
            po_create_fail: 'Impossibile Creare PO',
            report_ready: 'Report pronto',
            report_fail: 'Impossibile Eseguire Report',
            suggestions_refreshed: 'Suggerimenti aggiornati',
            refresh_fail: 'Impossibile aggiornare suggerimenti',
            select_product: 'Seleziona un prodotto',
            forecast_success: 'Previsione generata con successo',
            forecast_fail: 'Impossibile generare previsione',
            select_customer: 'Seleziona un cliente',
            analysis_complete: 'Analisi completata',
            analysis_fail: 'Impossibile analizzare rischio abbandono'
        }
    },
    pos: {
        hero: {
            title: 'Console Checkout Unificata',
            subtitle: 'Scansiona, cerca e completa ordini con un flusso a basso attrito che rimane sincronizzato con il tuo catalogo.',
            label: 'Punto Vendita',
            sub_label: 'Tela checkout live per team al banco',
            new_sale_btn: 'Nuova vendita al banco',
            refresh_catalog_btn: 'Aggiorna catalogo'
        },
        header: {
            title: 'Punto Vendita',
            description: 'Tocca i prodotti per costruire il carrello, rivedi sotto, poi conferma a destra.',
            super_shop_mode: 'Modalità Super Shop',
            search_placeholder: 'Cerca per nome, barcode, o SKU...',
            search_btn: 'Cerca'
        },
        products: {
            title: 'Prodotti',
            description: 'Tocca una tile per aggiungerla al carrello attivo.',
            results_found: '{count} risultati',
            filter_status: {
                label: 'Stato Stock',
                all: 'Tutti gli Stati',
                in_stock: 'Disponibile',
                low_stock: 'Scorta Bassa',
                out_of_stock: 'Esaurito'
            },
            no_results: 'Nessun prodotto trovato. Prova a modificare la ricerca.',
            in_stock: '{count} disponibili',
            tap_to_add: 'Tocca per aggiungere'
        },
        cart: {
            title: 'Carrello',
            empty_desc: 'Nessun articolo aggiunto.',
            items_desc: '{count} articol{s} nel carrello',
            clear_btn: 'Svuota carrello',
            empty_state: 'Aggiungi prodotti dalla griglia sopra per iniziare un nuovo ordine.',
            table: {
                product: 'Prodotto',
                price: 'Prezzo',
                qty: 'Qtà',
                total: 'Totale'
            }
        },
        customer: {
            title: 'Cliente',
            description: 'Collega un cliente per ID, username, email o telefono. Opzionale per vendite al banco.',
            search_placeholder: 'Cerca per ID, username, email, telefono',
            new_btn: 'Nuovo',
            no_selected: 'Nessun cliente selezionato. Puoi comunque completare una vendita al banco.',
            loyalty_pts: '{points} pt',
            tier: '{tier}'
        },
        payment: {
            title: 'Pagamento',
            description: 'Scegli come il cliente paga per questo ordine.',
            methods: {
                cash: 'Contanti',
                card: 'Carta',
                bkash: 'bKash',
                other: 'Altro'
            },
            sub: {
                physical: 'Fisico',
                terminal: 'Terminale',
                mobile: 'Mobile',
                check_due: 'Assegno/Dovuto'
            }
        },
        loyalty: {
            redeem_label: 'Riscatta Punti Fedeltà',
            available: 'Disponibili: {points} pt (valore {value})',
            points: 'punti',
            error_exceed: 'Non può superare il saldo disponibile.'
        },
        summary: {
            title: 'Riepilogo Ordine',
            description: 'Rivedi totali e pagamento prima di confermare la vendita.',
            subtotal: 'Subtotale',
            tax: 'Tasse ({rate}%)',
            total: 'Totale',
            payment: 'Pagamento:',
            items: 'Articoli:',
            not_selected: 'Non selezionato',
            loyalty_earnings: 'Guadagni Fedeltà',
            complete_btn: 'Completa Ordine',
            processing_btn: 'Elaborazione...',
            add_items_hint: 'Aggiungi articoli al carrello per continuare',
            select_payment_hint: 'Seleziona un metodo di pagamento per completare'
        },
        new_customer_modal: {
            title: 'Nuovo Cliente',
            description: 'Aggiungi un nuovo membro alla tua base clienti.',
            name_label: 'Nome Completo',
            name_placeholder: 'Mario Rossi',
            email_label: 'Indirizzo Email',
            email_placeholder: 'mario@example.com',
            phone_label: 'Numero di Telefono',
            phone_placeholder: '+39 333 000 0000',
            cancel_btn: 'Annulla',
            create_btn: 'Crea Cliente'
        },
        toasts: {
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per accedere al POS.',
            customer_not_found: 'Cliente non trovato',
            search_error: 'Errore ricerca cliente',
            out_of_stock: 'Prodotto esaurito',
            stock_limit_reached: 'Impossibile aggiungere altro. Solo {stock} articoli disponibili.',
            processing: 'Elaborazione transazione...',
            order_success: 'Ordine completato con successo!',
            loyalty_earned: 'Il cliente ha guadagnato {points} punti fedeltà!',
            loyalty_redeemed: 'Riscattati {points} punti per uno sconto di {amount}.',
            transaction_fail: 'Transazione Fallita',
            name_required: 'Il nome è obbligatorio',
            customer_created: 'Cliente creato e selezionato!',
            create_fail: 'Impossibile creare cliente'
        }
    },
    orders: {
        title: 'Gestione Ordini',
        subtitle: 'Gestisci fatture vendita clienti e ordini acquisto & resi fornitori.',
        tabs: {
            sales: 'Vendite (Fatture)',
            purchases: 'Acquisti & Resi'
        },
        subtabs: {
            sales_orders: 'Ordini Vendita',
            customer_returns: 'Resi Clienti',
            purchase_orders: 'Ordini Acquisto',
            vendor_returns: 'Resi Fornitori'
        },
        search: {
            sales: 'Cerca vendite...',
            pos: 'Cerca PO...',
            returns: 'Cerca resi...'
        },
        status: {
            return_pending: 'RESO IN ATTESA',
            refund_value: 'Valore Rimborso',
            items_returned: 'Articoli Restituiti',
            no_items: 'Nessun dettaglio articoli',
            total_items: 'Articoli Totali',
            qty: 'Qtà'
        },
        labels: {
            customer: 'Cliente',
            items: 'Articoli',
            more: 'altro',
            supplier: 'Fornitore',
            refund: 'Rimborso',
            reason: 'Motivo',
            guest: 'Ospite',
            unknown: 'Sconosciuto'
        },
        actions: {
            view_details: 'Vedi Dettagli',
            manage_po: 'Gestisci PO',
            try_again: 'Riprova'
        },
        empty: {
            sales: 'Nessun ordine di vendita trovato.',
            customer_returns: 'Nessun reso cliente trovato.',
            purchase_orders: 'Nessun ordine di acquisto trovato.',
            vendor_returns: 'Nessun reso fornitore trovato.'
        },
        errors: {
            load_fail: 'Impossibile caricare storico ordini.',
            access_denied: 'Accesso Negato',
            access_denied_desc: 'Non hai i permessi per visualizzare gli ordini.'
        }
    },
    reports: {
        title: 'Intelligence Suite',
        subtitle: 'Approfondimenti sull\'efficienza operativa e sulle performance di vendita.',
        actions: {
            refresh: 'Aggiorna Dati'
        },
        tabs: {
            sales: 'Vendite & Staff',
            inventory: 'Salute Inventario',
            financial: 'Dati Finanziari'
        },
        heatmap: {
            title: 'Intensità Vendite Oraria',
            subtitle: 'Orari di picco transazioni',
            description: 'Visualizza le ore più impegnative della settimana in base al volume di vendite.'
        },
        staff: {
            title: 'Performance Staff',
            subtitle: 'Vendite per membro del team',
            transactions: 'Transazioni',
            description: 'Traccia i contributi alle vendite e il conteggio transazioni dei singoli dipendenti.'
        },
        customers: {
            title: 'Migliori Clienti',
            subtitle: 'Clienti con valore più alto',
            spent: 'Speso',
            orders: 'Ordini',
            last_order: 'Ultimo Ordine',
            unknown: 'Cliente Sconosciuto',
            orders_suffix: 'ordini',
            days_ago: 'giorni fa',
            lost: 'Perso?',
            lifetime_value: 'Lifetime Value',
            headers: {
                customer: 'Cliente',
                contact: 'Contatto',
                orders: 'Ordini',
                last_order: 'Ultimo Ordine',
                total_spent: 'Totale Speso'
            },
            table: {
                user: 'Utente / Email',
                name: 'Nome',
                spent: 'Totale Speso',
                orders: 'Ordini',
                last_order: 'Ultimo Ordine',
                days_ago: 'Giorni Fa'
            },
            description: 'Identifica i tuoi clienti più preziosi in base alla spesa totale.'
        },
        frequency: {
            title: 'Analisi Frequenza',
            order_count: 'Conteggio Ordini',
            orders: 'Ordini',
            unknown: 'Prodotto Sconosciuto',
            description: 'Analizza quanto frequentemente due articoli vengono acquistati insieme (Market Basket Analysis). Aiuta nelle strategie di cross-selling e posizionamento prodotti.'
        },
        stock_aging: {
            title: 'Invecchiamento Stock',
            subtitle: 'Inventario per durata giacenza',
            headers: {
                product: 'Prodotto / SKU',
                age: 'Età (Giorni)',
                quantity: 'Qtà',
                value: 'Valore'
            },
            days_suffix: 'gg',
            description: 'Categorizza l\'inventario in base a quanto tempo è rimasto in magazzino per identificare articoli stagnanti.'
        },
        dead_stock: {
            title: 'Stock Morto (180+ Giorni)',
            days_idle: 'giorni inattivo',
            description: 'Lista di prodotti che non sono stati venduti negli ultimi 180 giorni.'
        },
        supplier: {
            title: 'Affidabilità Fornitore',
            time: 'Tempo',
            rate: 'Tasso',
            days_suffix: 'gg',
            description: 'Valuta i fornitori in base ai tempi di consegna e all\'accuratezza degli ordini.'
        },

        category: {
            title: 'Ripartizione Categorie',
            items: 'Articoli',
            sales: 'Vendite',
            margin: 'Margine',
            description: 'Mostra quali categorie di prodotto generano più ricavi e profitti.'
        },
        financials: {
            revenue: 'Ricavi',
            cogs: 'COGS',
            margin: 'Margine',
            gmroi: 'GMROI',
            description: 'Metriche finanziarie chiave inclusi Ricavi, Costo del Venduto e Margine Lordo.',
            revenue_desc: 'Reddito totale generato dalle vendite prima della detrazione delle spese.',
            cogs_desc: 'Costi diretti attribuibili alla produzione dei beni venduti.',
            margin_desc: 'La percentuale di ricavi che supera il costo del venduto.',
            gmroi_desc: 'Ritorno sul Margine Lordo dell\'Investimento. Misura la redditività dell\'inventario.'
        },
        void_analysis: {
            title: 'Analisi Annullamenti',
            subtitle: 'Audit transazioni cancellate',
            risk_score: 'Punteggio Rischio',
            voids: 'Annulli',
            risk: {
                high: 'Rischio Alto',
                medium: 'Rischio Medio',
                low: 'Rischio Basso'
            },
            description: 'Ispeziona le transazioni cancellate per identificare potenziali frodi o problemi di formazione.'
        },
        tax: {
            title: 'Responsabilità Fiscale',
            collected: 'Raccolto',
            rate: 'Aliquota',
            taxable_sales: 'Vendite Imponibili',
            description: 'Riepiloga le tasse raccolte e le vendite imponibili per aliquota fiscale.'
        },
        cash_reconciliation: {
            title: 'Riconciliazione Cassa',
            discrepancy: 'Discrepanza',
            description: 'Confronta i record di sistema con i conteggi di cassa effettivi per trovare discrepanze.'
        }
    }
};
