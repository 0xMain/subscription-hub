-- ======================================================================================
-- ИНИЦИАЛИЗАЦИЯ ОСНОВНЫХ ТАБЛИЦ
-- ======================================================================================

CREATE TABLE tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_tenants_name UNIQUE (name)
);

COMMENT ON TABLE tenants IS              'Таблица компаний (тенантов)';

COMMENT ON COLUMN tenants.id IS          'ID компании';
COMMENT ON COLUMN tenants.name IS        'Название компании';
COMMENT ON COLUMN tenants.created_at IS  'Дата и время регистрации компании';


CREATE INDEX idx_tenants_name ON tenants(name);

-- ======================================================================================

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_email UNIQUE (email)
);

COMMENT ON TABLE users IS                'Таблица пользователей';

COMMENT ON COLUMN users.id IS            'ID пользователя';
COMMENT ON COLUMN users.first_name IS    'Имя пользователя';
COMMENT ON COLUMN users.last_name IS     'Фамилия пользователя';
COMMENT ON COLUMN users.email IS         'Логин / Email пользователя';
COMMENT ON COLUMN users.password_hash IS 'Пароль (хеш) пользователя';
COMMENT ON COLUMN users.created_at IS    'Дата и время регистрации пользователя';


-- ======================================================================================

CREATE TABLE user_tenants (
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL,

    CONSTRAINT pk_user_tenants PRIMARY KEY (user_id, tenant_id),
    CONSTRAINT fk_user_tenants_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_tenants_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT chk_user_tenants_role CHECK (role IN ('owner', 'admin', 'manager', 'viewer'))

);

COMMENT ON TABLE user_tenants IS            'Таблица связей пользователей с компаниями';

COMMENT ON COLUMN user_tenants.user_id IS   'ID пользователя';
COMMENT ON COLUMN user_tenants.tenant_id IS 'ID компании';
COMMENT ON COLUMN user_tenants.role IS      'Роль сотрудника в компании';

CREATE UNIQUE INDEX uq_user_tenants_one_owner_per_tenant
    ON user_tenants (tenant_id)
    WHERE (role = 'owner');

-- ======================================================================================

CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_customers_tenant_email UNIQUE(tenant_id, email),
    CONSTRAINT fk_customers_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

COMMENT ON TABLE customers IS             'Таблица клиентов компаний';

COMMENT ON COLUMN customers.id IS         'ID клиента';
COMMENT ON COLUMN customers.tenant_id IS  'ID компании, которой принадлежит клиент';
COMMENT ON COLUMN customers.first_name IS 'Имя клиента';
COMMENT ON COLUMN customers.last_name IS  'Фамилия клиента';
COMMENT ON COLUMN customers.email IS      'Email клиента';
COMMENT ON COLUMN customers.created_at IS 'Дата и время добавления клиента';

-- ======================================================================================

CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    number VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    amount DECIMAL(15,2) NOT NULL,
    issued_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_invoices_tenant_number UNIQUE(tenant_id, number),
    CONSTRAINT chk_invoices_amount_positive CHECK (amount >= 0),
    CONSTRAINT chk_invoices_dates CHECK (due_date >= issued_date),
    CONSTRAINT chk_invoices_status CHECK (status IN ('draft', 'sent', 'paid', 'cancelled')),
    CONSTRAINT fk_invoices_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
    CONSTRAINT fk_invoices_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

COMMENT ON TABLE invoices IS              'Таблица счетов';

COMMENT ON COLUMN invoices.id IS          'ID счета';
COMMENT ON COLUMN invoices.tenant_id IS   'ID компании (для изоляции данных)';
COMMENT ON COLUMN invoices.customer_id IS 'ID клиента, которому выставлен счет';
COMMENT ON COLUMN invoices.number IS      'Номер счета';
COMMENT ON COLUMN invoices.status IS      'Статус счета';
COMMENT ON COLUMN invoices.amount IS      'Сумма счета';
COMMENT ON COLUMN invoices.issued_date IS 'Дата выставления счета';
COMMENT ON COLUMN invoices.due_date IS    'Срок оплаты счета';
COMMENT ON COLUMN invoices.paid_at IS     'Дата и время фактической оплаты счета';
COMMENT ON COLUMN invoices.created_at IS  'Дата и время создания записи счета';

-- ======================================================================================

CREATE TABLE invoice_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    description VARCHAR(500) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_items_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_items_unit_price_positive CHECK (unit_price >= 0),
    CONSTRAINT chk_items_amount_positive CHECK (amount >= 0),
    CONSTRAINT chk_items_amount_correct CHECK (amount = quantity * unit_price),
    CONSTRAINT fk_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
    CONSTRAINT fk_items_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

COMMENT ON TABLE invoice_items IS              'Таблица позиции счета';

COMMENT ON COLUMN invoice_items.id IS          'ID позиции';
COMMENT ON COLUMN invoice_items.tenant_id IS   'ID компании (для изоляции данных)';
COMMENT ON COLUMN invoice_items.invoice_id IS  'ID счета, к которому относится позиция';
COMMENT ON COLUMN invoice_items.description IS 'Наименование позиции';
COMMENT ON COLUMN invoice_items.quantity IS    'Количество единиц в позиции';
COMMENT ON COLUMN invoice_items.unit_price IS  'Цена за единицу в позиции';
COMMENT ON COLUMN invoice_items.amount IS      'Сумма позиции (quantity * unit_price)';
COMMENT ON COLUMN invoice_items.created_at IS  'Дата и время создания позиции';

-- ======================================================================================