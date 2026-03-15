package repository

//
//import (
//	"app-go/internal/domain"
//	"context"
//	"database/sql"
//	"errors"
//	"fmt"
//	"log"
//)
//
//const (
//	getLastInvoiceNumberQuery = `
//		SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'INV-....-(....)') AS INTEGER)), 0)
//		FROM invoices
//		WHERE tenant_id = $1 AND number LIKE $2
//	`
//
//	findInvoiceByIDQuery = `
//		SELECT id, tenant_id, customer_id, number, status, amount, issued_date, due_date, paid_at, created_at
//		FROM invoices
//		WHERE id = $1 AND tenant_id = $2
//	`
//
//	findInvoicesByTenantQuery = `
//		SELECT id, tenant_id, customer_id, number, status, amount, issued_date, due_date, paid_at, created_at
//		FROM invoices
//		WHERE tenant_id = $1
//		ORDER BY id DESC
//		LIMIT $2 OFFSET $3
//	`
//
//	findInvoicesByCustomerQuery = `
//		SELECT id, tenant_id, customer_id, number, status, amount, issued_date, due_date, paid_at, created_at
//		FROM invoices
//		WHERE tenant_id = $1 AND customer_id = $2
//		ORDER BY id DESC
//		LIMIT $3 OFFSET $4
//	`
//	findInvoicesByStatusQuery = `
//        SELECT id, tenant_id, customer_id, number, status, amount, issued_date, due_date, paid_at, created_at
//        FROM invoices
//        WHERE tenant_id = $1 AND status = $2
//        ORDER BY id DESC
//        LIMIT $3 OFFSET $4
//    `
//
//	countInvoicesByTenantQuery = `
//		SELECT COUNT(*)
//		FROM invoices
//		WHERE tenant_id = $1
//	`
//
//	countInvoicesByStatusQuery = `
//        SELECT COUNT(*)
//        FROM invoices
//        WHERE tenant_id = $1 AND status = $2
//    `
//
//	countInvoicesByCustomerQuery = `
//		SELECT COUNT(*)
//		FROM invoices
//		WHERE tenant_id = $1 AND customer_id = $2
//	`
//
//	createInvoiceQuery = `
//    INSERT INTO invoices (tenant_id, customer_id, number, status, amount, due_date)
//    VALUES ($1, $2, $3, $4, $5, $6)
//    RETURNING id, issued_date, created_at
//	`
//
//	updateInvoiceStatusQuery = `
//		UPDATE invoices
//		SET status = $1, paid_at = $2
//		WHERE id = $3 AND tenant_id = $4
//		RETURNING id
//	`
//
//	deleteInvoiceQuery = `
//		DELETE FROM invoices
//		WHERE id = $1 AND tenant_id = $2 AND status = 'draft'
//		RETURNING id
//	`
//)
//
//type InvoiceRepository struct {
//	db *sql.DB
//}
//
//func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
//	return &InvoiceRepository{db: db}
//}
//
//func (r *InvoiceRepository) GetLastNumber(ctx context.Context, tenantID, year int) (int, error) {
//	var lastNumber int
//	pattern := fmt.Sprintf("INV-%d-%%", year)
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, getLastInvoiceNumberQuery, tenantID, pattern).Scan(&lastNumber)
//	if err != nil {
//		return 0, fmt.Errorf("ошибка при получении последнего номера счета (tenantID=%d, year=%d) в БД. Причина: %w",
//			tenantID, year, err,
//		)
//	}
//
//	return lastNumber, nil
//}
//
//func (r *InvoiceRepository) FindByID(ctx context.Context, id, tenantID int64) (*domain.Invoice, error) {
//	var invoice domain.Invoice
//	var paidAt sql.NullTime
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, findInvoiceByIDQuery, id, tenantID).Scan(
//		&invoice.ID,
//		&invoice.TenantID,
//		&invoice.CustomerID,
//		&invoice.Number,
//		&invoice.Status,
//		&invoice.Amount,
//		&invoice.IssuedDate,
//		&invoice.DueDate,
//		&paidAt,
//		&invoice.CreatedAt,
//	)
//	if errors.Is(err, sql.ErrNoRows) {
//		return nil, nil
//	}
//	if err != nil {
//		return nil, fmt.Errorf("ошибка при поиске счета (id=%d, tenantID=%d) в БД. Причина: %w",
//			id, tenantID, err,
//		)
//	}
//
//	if paidAt.Valid {
//		invoice.PaidAt = &paidAt.Time
//	}
//
//	return &invoice, nil
//}
//
//func (r *InvoiceRepository) FindByTenant(ctx context.Context, tenantID int64, limit, offset int) ([]domain.Invoice, int64, error) {
//	var total int64
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, countInvoicesByTenantQuery, tenantID).Scan(&total)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при подсчете кол-ва счетов (tenantID=%d) в БД. Причина: %w",
//			tenantID, err,
//		)
//	}
//
//	if total == 0 {
//		return []domain.Invoice{}, 0, nil
//	}
//
//	rows, err := executor.QueryContext(ctx, findInvoicesByTenantQuery, tenantID, limit, offset)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при получении списка счетов (tenantID=%d, limit=%d, offset=%d) в БД. Причина: %w",
//			tenantID, limit, offset, err,
//		)
//	}
//
//	defer func() {
//		if err := rows.Close(); err != nil {
//			log.Printf("ошибка при закрытии rows после получения счетов (tenantID=%d) в БД. Причина: %v",
//				tenantID, err,
//			)
//		}
//	}()
//
//	var invoices []domain.Invoice
//
//	for rows.Next() {
//		var invoice domain.Invoice
//		var paidAt sql.NullTime
//
//		err := rows.Scan(
//			&invoice.ID,
//			&invoice.TenantID,
//			&invoice.CustomerID,
//			&invoice.Number,
//			&invoice.Status,
//			&invoice.Amount,
//			&invoice.IssuedDate,
//			&invoice.DueDate,
//			&paidAt,
//			&invoice.CreatedAt,
//		)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при сканировании счета (tenantID=%d) в БД. Причина: %w",
//				tenantID, err,
//			)
//		}
//
//		if paidAt.Valid {
//			invoice.PaidAt = &paidAt.Time
//		}
//
//		invoices = append(invoices, invoice)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, 0, fmt.Errorf("ошибка при обработке результатов счетов (tenantID=%d) в БД. Причина: %w",
//			tenantID, err,
//		)
//	}
//
//	return invoices, total, nil
//}
//
//func (r *InvoiceRepository) FindByCustomer(ctx context.Context, tenantID, customerID int64, limit, offset int) ([]domain.Invoice, int64, error) {
//	var total int64
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, countInvoicesByCustomerQuery, tenantID, customerID).Scan(&total)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при подсчете кол-ва счетов (tenantID=%d, customerID=%d) в БД. Причина: %w",
//			tenantID, customerID, err,
//		)
//	}
//
//	if total == 0 {
//		return []domain.Invoice{}, 0, nil
//	}
//
//	rows, err := executor.QueryContext(ctx, findInvoicesByCustomerQuery, tenantID, customerID, limit, offset)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при получении списка счетов (tenantID=%d, customerID=%d, limit=%d, offset=%d) в БД. Причина: %w",
//			tenantID, customerID, limit, offset, err,
//		)
//	}
//
//	defer func() {
//		if err := rows.Close(); err != nil {
//			log.Printf("ошибка при закрытии rows после получения списка счетов (customerID=%d) в БД. Причина: %v",
//				customerID, err,
//			)
//		}
//	}()
//
//	var invoices []domain.Invoice
//
//	for rows.Next() {
//		var invoice domain.Invoice
//		var paidAt sql.NullTime
//
//		err := rows.Scan(
//			&invoice.ID,
//			&invoice.TenantID,
//			&invoice.CustomerID,
//			&invoice.Number,
//			&invoice.Status,
//			&invoice.Amount,
//			&invoice.IssuedDate,
//			&invoice.DueDate,
//			&paidAt,
//			&invoice.CreatedAt,
//		)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при сканировании счета (customerID=%d, tenantID=%d) в БД. Причина: %w",
//				customerID, tenantID, err,
//			)
//		}
//
//		if paidAt.Valid {
//			invoice.PaidAt = &paidAt.Time
//		}
//
//		invoices = append(invoices, invoice)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, 0, fmt.Errorf("ошибка при обработке результатов счетов (customerID=%d) в БД. Причина: %w",
//			customerID, err,
//		)
//	}
//
//	return invoices, total, nil
//}
//
//func (r *InvoiceRepository) FindByStatus(ctx context.Context, tenantID int64, status domain.InvoiceStatus, limit, offset int) ([]domain.Invoice, int64, error) {
//	var total int64
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, countInvoicesByStatusQuery, tenantID, status).Scan(&total)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при подсчете кол-ва счетов (tenantID=%d, status=%s) в БД. Причина: %w",
//			tenantID, status, err,
//		)
//	}
//
//	if total == 0 {
//		return []domain.Invoice{}, 0, nil
//	}
//
//	rows, err := executor.QueryContext(ctx, findInvoicesByStatusQuery, tenantID, status, limit, offset)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при получении списка счетов (tenantID=%d, status=%s, limit=%d, offset=%d) в БД. Причина: %w",
//			tenantID, status, limit, offset, err,
//		)
//	}
//
//	defer func() {
//		if err := rows.Close(); err != nil {
//			log.Printf("ошибка при закрытии rows после получения списка счетов (tenantID=%d, status=%s) в БД. Причина: %v",
//				tenantID, status, err,
//			)
//		}
//	}()
//
//	var invoices []domain.Invoice
//
//	for rows.Next() {
//		var invoice domain.Invoice
//		var paidAt sql.NullTime
//
//		err := rows.Scan(
//			&invoice.ID,
//			&invoice.TenantID,
//			&invoice.CustomerID,
//			&invoice.Number,
//			&invoice.Status,
//			&invoice.Amount,
//			&invoice.IssuedDate,
//			&invoice.DueDate,
//			&paidAt,
//			&invoice.CreatedAt,
//		)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при сканировании счета (tenantID=%d, status=%s) в БД. Причина: %w",
//				tenantID, status, err,
//			)
//		}
//
//		if paidAt.Valid {
//			invoice.PaidAt = &paidAt.Time
//		}
//
//		invoices = append(invoices, invoice)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, 0, fmt.Errorf("ошибка при обработке результатов счетов (tenantID=%d, status=%s) в БД. Причина: %w",
//			tenantID, status, err)
//	}
//
//	return invoices, total, nil
//}
//
//func (r *InvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
//	executor := executor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, createInvoiceQuery,
//		invoice.TenantID,
//		invoice.CustomerID,
//		invoice.Number,
//		invoice.Status,
//		invoice.Amount,
//		invoice.DueDate,
//	).Scan(&invoice.ID, &invoice.IssuedDate, &invoice.CreatedAt)
//	if err != nil {
//		return fmt.Errorf("ошибка при создании счета (tenantID=%d, customerID=%d) в БД. Причина: %w",
//			invoice.TenantID, invoice.CustomerID, err,
//		)
//	}
//
//	return nil
//}
//
//func (r *InvoiceRepository) UpdateStatus(ctx context.Context, invoice *domain.Invoice) error {
//	executor := getExecutor(ctx, r.db)
//	result, err := executor.ExecContext(ctx, updateInvoiceStatusQuery,
//		invoice.Status,
//		invoice.PaidAt,
//		invoice.ID,
//		invoice.TenantID,
//	)
//	if err != nil {
//		return fmt.Errorf("ошибка при обновлении статуса счета (ID=%d, tenantID=%d, статус=%s) в БД. Причина: %w",
//			invoice.ID, invoice.TenantID, invoice.Status, err,
//		)
//	}
//
//	rowsAffected, _ := result.RowsAffected()
//	if rowsAffected == 0 {
//		return fmt.Errorf("счет (ID=%d, tenantID=%d) не найден",
//			invoice.ID, invoice.TenantID,
//		)
//	}
//
//	return nil
//}
//
//func (r *InvoiceRepository) Delete(ctx context.Context, id, tenantID int64) error {
//	executor := getExecutor(ctx, r.db)
//	result, err := executor.ExecContext(ctx, deleteInvoiceQuery, id, tenantID)
//	if err != nil {
//		return fmt.Errorf("ошибка при удалении счета (id=%d, tenantID=%d) в БД. Причина: %w",
//			id, tenantID, err,
//		)
//	}
//
//	rows, _ := result.RowsAffected()
//	if rows == 0 {
//		return fmt.Errorf("счет (id=%d, tenantID=%d) не найден в БД",
//			id, tenantID,
//		)
//	}
//
//	return nil
//}
