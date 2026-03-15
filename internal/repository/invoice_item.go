package repository

// import (
//
//	"app-go/internal/domain"
//	"context"
//	"database/sql"
//	"errors"
//	"fmt"
//	"log"
//	"strings"
//
//	"github.com/shopspring/decimal"
//
// )
//
// const (
//
//	totalSumByInvoiceQuery = `
//		SELECT COALESCE(SUM(amount), 0)
//		FROM invoice_items
//		WHERE invoice_id = $1 AND tenant_id = $2
//	`
//
//	findItemByIDQuery = `
//		SELECT id, invoice_id, tenant_id, description, quantity, unit_price, amount, created_at
//		FROM invoice_items
//		WHERE id = $1 AND tenant_id = $2
//	`
//
//	findItemsByInvoiceIDQuery = `
//		SELECT id, invoice_id, tenant_id, description, quantity, unit_price, amount, created_at
//		FROM invoice_items
//		WHERE invoice_id = $1 AND tenant_id = $2
//		ORDER BY id
//		LIMIT $3 OFFSET $4
//	`
//
//	countItemsByInvoiceIDQuery = `
//		SELECT COUNT(*)
//		FROM invoice_items
//		WHERE invoice_id = $1 AND tenant_id = $2
//	`
//
//	// language=TEXT
//	createManyItemsBaseQuery = `
//		INSERT INTO invoice_items (invoice_id, tenant_id, description, quantity, unit_price, amount)
//		VALUES %s
//		RETURNING id, created_at
//	`
//
//	updateItemQuery = `
//		UPDATE invoice_items
//		SET description = $1, quantity = $2, unit_price = $3, amount = $4
//		WHERE id = $5 AND tenant_id = $6
//	`
//
//	deleteItemByIDQuery = `
//		DELETE FROM invoice_items
//		WHERE id = $1 AND tenant_id = $2
//	`
//
//	deleteItemsByInvoiceIDQuery = `
//		DELETE FROM invoice_items
//		WHERE invoice_id = $1 AND tenant_id = $2
//	`
//
// )
//
//	type InvoiceItemRepository struct {
//		db *sql.DB
//	}
//
//	func NewInvoiceItemRepository(db *sql.DB) *InvoiceItemRepository {
//		return &InvoiceItemRepository{
//			db: db,
//		}
//	}
//
//	func (r *InvoiceItemRepository) TotalSumByInvoice(ctx context.Context, invoiceID, tenantID int64) (decimal.Decimal, error) {
//		var total decimal.Decimal
//
//		err := getExecutor(ctx, r.db).QueryRowContext(ctx, totalSumByInvoiceQuery, invoiceID, tenantID).Scan(&total)
//		if err != nil {
//			return decimal.Zero, fmt.Errorf("ошибка при получении суммы позиций счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//				invoiceID, tenantID, err,
//			)
//		}
//
//		return total, nil
//	}
//
//	func (r *InvoiceItemRepository) FindByID(ctx context.Context, id, tenantID int64) (*domain.InvoiceItem, error) {
//		row := getExecutor(ctx, r.db).QueryRowContext(ctx, findItemByIDQuery, id, tenantID)
//
//		item, err := r.scanInvoiceItem(row)
//		if errors.Is(err, sql.ErrNoRows) {
//			return nil, nil
//		}
//		if err != nil {
//			return nil, fmt.Errorf("ошибка при поиске позиции счета (id=%d, tenantID=%d) в БД. Причина: %w",
//				id, tenantID, err,
//			)
//		}
//
//		return item, nil
//	}
//
//	func (r *InvoiceItemRepository) FindByInvoiceID(ctx context.Context, invoiceID, tenantID int64, limit, offset int) ([]domain.InvoiceItem, int64, error) {
//		var total int64
//
//		executor := getExecutor(ctx, r.db)
//		err := executor.QueryRowContext(ctx, countItemsByInvoiceIDQuery, invoiceID, tenantID).Scan(&total)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при подсчете кол-ва позиций счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//				invoiceID, tenantID, err,
//			)
//		}
//
//		if total == 0 {
//			return []domain.InvoiceItem{}, 0, nil
//		}
//
//		rows, err := executor.QueryContext(ctx, findItemsByInvoiceIDQuery, invoiceID, tenantID, limit, offset)
//		if err != nil {
//			return nil, 0, fmt.Errorf("ошибка при получении списка позиций счета (invoiceID=%d, tenantID=%d, limit=%d, offset=%d) в БД. Причина: %w",
//				invoiceID, tenantID, limit, offset, err,
//			)
//		}
//
//		defer func() {
//			if err := rows.Close(); err != nil {
//				log.Printf("ошибка при закрытии rows после получения позиций счета (invoiceID=%d, tenantID=%d, limit=%d, offset=%d) в БД. Причина: %v",
//					invoiceID, tenantID, limit, offset, err,
//				)
//			}
//		}()
//
//		var items []domain.InvoiceItem
//
//		for rows.Next() {
//			item, err := r.scanInvoiceItem(rows)
//			if err != nil {
//				return nil, 0, fmt.Errorf("ошибка при сканировании позиции счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//					invoiceID, tenantID, err,
//				)
//			}
//
//			items = append(items, *item)
//		}
//
//		if err = rows.Err(); err != nil {
//			return nil, 0, fmt.Errorf("ошибка при обработке результатов запроса позиций счета (invoiceID=%d, tenantID=%d, limit=%d, offset=%d) в БД. Причина: %w",
//				invoiceID, tenantID, limit, offset, err)
//		}
//
//		return items, total, nil
//	}
//
//	func (r *InvoiceItemRepository) Create(ctx context.Context, items []domain.InvoiceItem) error {
//		if len(items) == 0 {
//			return nil
//		}
//
//		valueStrings := make([]string, 0, len(items))
//		valueArgs := make([]interface{}, 0, len(items)*6)
//
//		for i, item := range items {
//			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
//
//			valueArgs = append(valueArgs,
//				item.InvoiceID,
//				item.TenantID,
//				item.Description,
//				item.Quantity,
//				item.UnitPrice,
//				item.Amount,
//			)
//		}
//
//		query := fmt.Sprintf(createManyItemsBaseQuery, strings.Join(valueStrings, ","))
//
//		rows, err := getExecutor(ctx, r.db).QueryContext(ctx, query, valueArgs...)
//		if err != nil {
//			return fmt.Errorf("ошибка при массовом создании позиций счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//				items[0].InvoiceID, items[0].TenantID, err,
//			)
//		}
//
//		defer func() {
//			if err := rows.Close(); err != nil && len(items) > 0 {
//				log.Printf("ошибка при закрытии rows после создания позиций счета (invoiceID=%d, tenantID=%d) в БД. Причина: %v",
//					items[0].InvoiceID, items[0].TenantID, err)
//			}
//		}()
//
//		var i int
//		for rows.Next() {
//			if err := rows.Scan(&items[i].ID, &items[i].CreatedAt); err != nil {
//				return fmt.Errorf("ошибка при сканировании ID позиции счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//					items[i].InvoiceID, items[i].TenantID, err,
//				)
//			}
//
//			i++
//		}
//
//		return nil
//	}
//
//	func (r *InvoiceItemRepository) Update(ctx context.Context, item *domain.InvoiceItem) error {
//		result, err := getExecutor(ctx, r.db).ExecContext(ctx, updateItemQuery,
//			item.Description,
//			item.Quantity,
//			item.UnitPrice,
//			item.Amount,
//			item.ID,
//			item.TenantID,
//		)
//		if err != nil {
//			return fmt.Errorf("ошибка при обновлении позиции счета (ID=%d, tenantID=%d) в БД. Причина: %w",
//				item.ID, item.TenantID, err,
//			)
//		}
//
//		rows, _ := result.RowsAffected()
//		if rows == 0 {
//			return fmt.Errorf("позиция счета (ID=%d, tenantID=%d) не найдена",
//				item.ID, item.TenantID,
//			)
//		}
//
//		return nil
//	}
//
//	func (r *InvoiceItemRepository) RemoveByID(ctx context.Context, id, tenantID int64) error {
//		result, err := getExecutor(ctx, r.db).ExecContext(ctx, deleteItemByIDQuery, id, tenantID)
//		if err != nil {
//			return fmt.Errorf("ошибка при удалении позиции счета (ID=%d, tenantID=%d) в БД. Причина: %w",
//				id, tenantID, err,
//			)
//		}
//
//		rows, _ := result.RowsAffected()
//		if rows == 0 {
//			return fmt.Errorf("позиция счета (ID=%d, tenantID=%d) не найдена",
//				id, tenantID,
//			)
//		}
//
//		return nil
//	}
//
//	func (r *InvoiceItemRepository) RemoveAllByInvoiceID(ctx context.Context, invoiceID, tenantID int64) error {
//		result, err := getExecutor(ctx, r.db).ExecContext(ctx, deleteItemsByInvoiceIDQuery, invoiceID, tenantID)
//		if err != nil {
//			return fmt.Errorf("ошибка при удалении всех позиций счета (invoiceID=%d, tenantID=%d) в БД. Причина: %w",
//				invoiceID, tenantID, err,
//			)
//		}
//
//		rows, _ := result.RowsAffected()
//		if rows == 0 {
//			log.Printf("позиции счета (invoiceID=%d, tenantID=%d) не найдены для удаления",
//				invoiceID, tenantID,
//			)
//		}
//
//		return nil
//	}
type scanner interface {
	Scan(dest ...interface{}) error
}

//
//func (r *InvoiceItemRepository) scanInvoiceItem(sc scanner) (*domain.InvoiceItem, error) {
//	var item domain.InvoiceItem
//	err := sc.Scan(
//		&item.ID,
//		&item.InvoiceID,
//		&item.TenantID,
//		&item.Description,
//		&item.Quantity,
//		&item.UnitPrice,
//		&item.Amount,
//		&item.CreatedAt,
//	)
//	if err != nil {
//		return nil, err
//	}
//	return &item, nil
//}
