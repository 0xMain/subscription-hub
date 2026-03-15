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
//	createCustomerQuery = `
//		INSERT INTO customers (tenant_id, first_name, last_name, email)
//		VALUES ($1, $2, $3, $4)
//		RETURNING id, created_at
//	`
//
//	findCustomerByIDQuery = `
//		SELECT id, tenant_id, first_name, last_name, email, created_at
//		FROM customers
//		WHERE id = $1 AND tenant_id = $2
//	`
//
//	findCustomerByEmailQuery = `
//		SELECT id, tenant_id, first_name, last_name, email, created_at
//		FROM customers
//		WHERE tenant_id = $1 AND email = $2
//	`
//
//	countCustomersByTenantQuery = `
//		SELECT COUNT(*)
//		FROM customers
//		WHERE tenant_id = $1
//	`
//
//	findCustomersByTenantQuery = `
//		SELECT id, tenant_id, first_name, last_name, email, created_at
//		FROM customers
//		WHERE tenant_id = $1
//		ORDER BY id DESC
//		LIMIT $2 OFFSET $3
//	`
//
//	updateCustomerQuery = `
//		UPDATE customers
//		SET first_name = $1, last_name = $2, email = $3
//		WHERE id = $4 AND tenant_id = $5
//		RETURNING id
//	`
//
//	deleteCustomerQuery = `
//		DELETE FROM customers
//		WHERE id = $1 AND tenant_id = $2
//		RETURNING id
//	`
//)
//
//type CustomerRepository struct {
//	db *sql.DB
//}
//
//func NewCustomerRepository(db *sql.DB) *CustomerRepository {
//	return &CustomerRepository{
//		db: db,
//	}
//}
//
////func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
////	executor := getExecutor(ctx, r.db)
////	err := executor.QueryRowContext(ctx, createCustomerQuery,
////		customer.TenantID,
////		customer.FirstName,
////		customer.LastName,
////		customer.Email,
////	).Scan(&customer.ID, &customer.CreatedAt)
////	if err != nil {
////		return fmt.Errorf("не удалось создать клиента (tenantID=%d, email=%s) в БД. Причина: %w",
////			customer.TenantID, customer.Email, err,
////		)
////	}
////
////	return nil
////}
//
//func (r *CustomerRepository) FindByID(ctx context.Context, id, tenantID int64) (*domain.Customer, error) {
//	var customer domain.Customer
//
//	executor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, findCustomerByIDQuery, id, tenantID).Scan(
//		&customer.ID,
//		&customer.TenantID,
//		&customer.FirstName,
//		&customer.LastName,
//		&customer.Email,
//		&customer.CreatedAt,
//	)
//	if errors.Is(err, sql.ErrNoRows) {
//		return nil, nil
//	}
//	if err != nil {
//		return nil, fmt.Errorf("ошибка при поиске клиента (ID=%d, tenantID=%d) в БД. Причина: %w",
//			id, tenantID, err,
//		)
//	}
//
//	return &customer, nil
//}
//
//func (r *CustomerRepository) FindByEmail(ctx context.Context, tenantID int64, email string) (*domain.Customer, error) {
//	var customer domain.Customer
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, findCustomerByEmailQuery, tenantID, email).Scan(
//		&customer.ID,
//		&customer.TenantID,
//		&customer.FirstName,
//		&customer.LastName,
//		&customer.Email,
//		&customer.CreatedAt,
//	)
//	if errors.Is(err, sql.ErrNoRows) {
//		return nil, nil
//	}
//	if err != nil {
//		return nil, fmt.Errorf("ошибка при поиске клиента (tenantID=%d, email=%s) в БД. Причина: %w",
//			tenantID, email, err,
//		)
//	}
//
//	return &customer, nil
//}
//
//func (r *CustomerRepository) FindAllByTenant(ctx context.Context, tenantID, limit, offset int64) ([]domain.Customer, int64, error) {
//	var total int64
//
//	executor := getExecutor(ctx, r.db)
//	err := executor.QueryRowContext(ctx, countCustomersByTenantQuery, tenantID).Scan(&total)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при подсчете клиентов (tenantID=%d) в БД. Причина: %w",
//			tenantID, err,
//		)
//	}
//
//	if total == 0 {
//		return []domain.Customer{}, 0, nil
//	}
//
//	rows, err := executor.QueryContext(ctx, findCustomersByTenantQuery, tenantID, limit, offset)
//	if err != nil {
//		return nil, 0, fmt.Errorf("ошибка при получении списка клиентов (tenantID=%d, limit=%d, offset=%d) в БД. Причина: %w",
//			tenantID, limit, offset, err,
//		)
//	}
//
//	defer func() {
//		if err := rows.Close(); err != nil {
//			log.Printf("ошибка при закрытии rows после получения клиентов (tenantID=%d) в БД. Причина: %v",
//				tenantID, err,
//			)
//		}
//	}()
//
//	var customers []domain.Customer
//
//	for rows.Next() {
//		var customer domain.Customer
//		err := rows.Scan(
//			&customer.ID,
//			&customer.TenantID,
//			&customer.FirstName,
//			&customer.LastName,
//			&customer.Email,
//			&customer.CreatedAt,
//		)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при сканировании клиента (ID=%d, tenantID=%d) в БД. Причина: %w",
//				customer.ID, tenantID, err,
//			)
//		}
//
//		customers = append(customers, customer)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, 0, fmt.Errorf("ошибка при обработке результатов клиентов (tenantID=%d) в БД. Причина: %w",
//			tenantID, err,
//		)
//	}
//
//	return customers, total, nil
//}
//
//func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
//	executor := getExecutor(ctx, r.db)
//	result, err := executor.ExecContext(ctx, updateCustomerQuery,
//		customer.FirstName,
//		customer.LastName,
//		customer.Email,
//		customer.ID,
//		customer.TenantID,
//	)
//	if err != nil {
//		return fmt.Errorf("ошибка при обновлении клиента (id=%d, tenantID=%d, email=%s) в БД. Причина: %w",
//			customer.ID, customer.TenantID, customer.Email, err,
//		)
//	}
//
//	rows, _ := result.RowsAffected()
//	if rows == 0 {
//		return fmt.Errorf("клиент (id=%d, tenantID=%d) не найден в БД",
//			customer.ID, customer.TenantID,
//		)
//	}
//
//	return nil
//}
//
//func (r *CustomerRepository) Delete(ctx context.Context, id, tenantID int64) error {
//	executor := getExecutor(ctx, r.db)
//	result, err := executor.ExecContext(ctx, deleteCustomerQuery, id, tenantID)
//	if err != nil {
//		return fmt.Errorf("ошибка при удалении клиента (id=%d, tenantID=%d) в БД. Причина: %w",
//			id, tenantID, err,
//		)
//	}
//
//	rows, _ := result.RowsAffected()
//	if rows == 0 {
//		return fmt.Errorf("клиент (id=%d, tenantID=%d) не найден в БД",
//			id, tenantID,
//		)
//	}
//
//	return nil
//}
