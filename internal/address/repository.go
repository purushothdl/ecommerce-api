package address

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/purushothdl/ecommerce-api/internal/domain"
    "github.com/purushothdl/ecommerce-api/internal/models"
    "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type addressRepository struct {
    db domain.DBTX
}

// NewAddressRepository creates a new AddressRepository
func NewAddressRepository(db domain.DBTX) domain.AddressRepository {
    return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, addr *models.UserAddress) error {
    query := `
        INSERT INTO user_addresses (
            user_id, name, phone, street1, street2, city, state, postal_code, country,
            is_default_shipping, is_default_billing
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, created_at, updated_at
    `
    err := r.db.QueryRowContext(ctx, query,
        addr.UserID, addr.Name, addr.Phone, addr.Street1, addr.Street2, addr.City, addr.State,
        addr.PostalCode, addr.Country, addr.IsDefaultShipping, addr.IsDefaultBilling,
    ).Scan(&addr.ID, &addr.CreatedAt, &addr.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to create address: %w", err)
    }
    return nil
}

func (r *addressRepository) GetByID(ctx context.Context, id int64) (*models.UserAddress, error) {
    query := `
        SELECT id, user_id, name, phone, street1, street2, city, state, postal_code, country,
               is_default_shipping, is_default_billing, created_at, updated_at
        FROM user_addresses WHERE id = $1
    `
    addr := &models.UserAddress{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &addr.ID, &addr.UserID, &addr.Name, &addr.Phone, &addr.Street1, &addr.Street2,
        &addr.City, &addr.State, &addr.PostalCode, &addr.Country,
        &addr.IsDefaultShipping, &addr.IsDefaultBilling, &addr.CreatedAt, &addr.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, apperrors.ErrNotFound
    } else if err != nil {
        return nil, fmt.Errorf("failed to get address by ID: %w", err)
    }
    return addr, nil
}

func (r *addressRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.UserAddress, error) {
    query := `
        SELECT id, user_id, name, phone, street1, street2, city, state, postal_code, country,
               is_default_shipping, is_default_billing, created_at, updated_at
        FROM user_addresses WHERE user_id = $1 ORDER BY created_at DESC
    `
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get addresses by user ID: %w", err)
    }
    defer rows.Close()

    var addresses []*models.UserAddress
    for rows.Next() {
        addr := &models.UserAddress{}
        if err := rows.Scan(
            &addr.ID, &addr.UserID, &addr.Name, &addr.Phone, &addr.Street1, &addr.Street2,
            &addr.City, &addr.State, &addr.PostalCode, &addr.Country,
            &addr.IsDefaultShipping, &addr.IsDefaultBilling, &addr.CreatedAt, &addr.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan address: %w", err)
        }
        addresses = append(addresses, addr)
    }
    return addresses, nil
}

func (r *addressRepository) Update(ctx context.Context, addr *models.UserAddress) error {
    query := `
        UPDATE user_addresses SET
            name = $1, phone = $2, street1 = $3, street2 = $4, city = $5, state = $6,
            postal_code = $7, country = $8, is_default_shipping = $9, is_default_billing = $10,
            updated_at = NOW()
        WHERE id = $11 AND user_id = $12
        RETURNING updated_at
    `
    err := r.db.QueryRowContext(ctx, query,
        addr.Name, addr.Phone, addr.Street1, addr.Street2, addr.City, addr.State,
        addr.PostalCode, addr.Country, addr.IsDefaultShipping, addr.IsDefaultBilling,
        addr.ID, addr.UserID,
    ).Scan(&addr.UpdatedAt)
    if err == sql.ErrNoRows {
        return apperrors.ErrNotFound
    } else if err != nil {
        return fmt.Errorf("failed to update address: %w", err)
    }
    return nil
}

func (r *addressRepository) Delete(ctx context.Context, id int64, userID int64) error {
    query := `DELETE FROM user_addresses WHERE id = $1 AND user_id = $2`
    res, err := r.db.ExecContext(ctx, query, id, userID)
    if err != nil {
        return fmt.Errorf("failed to delete address: %w", err)
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return apperrors.ErrNotFound
    }
    return nil
}

func (r *addressRepository) UnsetDefaultShipping(ctx context.Context, userID int64) error {
    query := `UPDATE user_addresses SET is_default_shipping = FALSE WHERE user_id = $1 AND is_default_shipping = TRUE`
    _, err := r.db.ExecContext(ctx, query, userID)
    return err
}

func (r *addressRepository) UnsetDefaultBilling(ctx context.Context, userID int64) error {
    query := `UPDATE user_addresses SET is_default_billing = FALSE WHERE user_id = $1 AND is_default_billing = TRUE`
    _, err := r.db.ExecContext(ctx, query, userID)
    return err
}
