package address

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	"github.com/purushothdl/ecommerce-api/internal/shared/dto"
	"github.com/purushothdl/ecommerce-api/pkg/errors"
	"github.com/purushothdl/ecommerce-api/pkg/utils/ptr"
)

type addressService struct {
    repo   domain.AddressRepository
    store  domain.Store
    logger *slog.Logger
}

// NewAddressService creates a new AddressService
func NewAddressService(repo domain.AddressRepository, store domain.Store, logger *slog.Logger) domain.AddressService {
    return &addressService{
        repo:   repo,
        store:  store,
        logger: logger,
    }
}

func (s *addressService) Create(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*models.UserAddress, error) {
    addr := &models.UserAddress{
        UserID:            userID,
        Name:              req.Name,
        Phone:             req.Phone,
        Street1:           req.Street1,
        Street2:           req.Street2,
        City:              req.City,
        State:             req.State,
        PostalCode:        req.PostalCode,
        Country:           req.Country,
        IsDefaultShipping: req.IsDefaultShipping,
        IsDefaultBilling:  req.IsDefaultBilling,
    }

    var err error
    if req.IsDefaultShipping || req.IsDefaultBilling {
        err = s.store.ExecTx(ctx, func(q *domain.Queries) error {
            if req.IsDefaultShipping {
                if err := q.AddressRepo.UnsetDefaultShipping(ctx, userID); err != nil {
                    return fmt.Errorf("failed to unset default shipping: %w", err)
                }
            }
            if req.IsDefaultBilling {
                if err := q.AddressRepo.UnsetDefaultBilling(ctx, userID); err != nil {
                    return fmt.Errorf("failed to unset default billing: %w", err)
                }
            }
            if err := q.AddressRepo.Create(ctx, addr); err != nil {
                return fmt.Errorf("failed to create address in tx: %w", err)
            }
            return nil
        })
    } else {
        err = s.repo.Create(ctx, addr)
    }

    if err != nil {
        s.logger.Error("failed to create address", "user_id", userID, "error", err)
        return nil, fmt.Errorf("service: failed to create address: %w", err)
    }

    s.logger.Info("address created successfully", "user_id", userID, "address_id", addr.ID)
    return addr, nil
}

func (s *addressService) Update(ctx context.Context, userID int64, id int64, req *dto.UpdateAddressRequest) (*models.UserAddress, error) {
    addr, err := s.GetByID(ctx, id, userID)
    if err != nil {
        return nil, fmt.Errorf("service: failed to get address for update: %w", err)
    }

    // Apply updates using ptr utility
    ptr.UpdateStringIfProvided(&addr.Name, req.Name)
    ptr.UpdateStringIfProvided(&addr.Phone, req.Phone)
    ptr.UpdateStringIfProvided(&addr.Street1, req.Street1)
    ptr.UpdateStringIfProvided(&addr.Street2, req.Street2)
    ptr.UpdateStringIfProvided(&addr.City, req.City)
    ptr.UpdateStringIfProvided(&addr.State, req.State)
    ptr.UpdateStringIfProvided(&addr.PostalCode, req.PostalCode)
    ptr.UpdateStringIfProvided(&addr.Country, req.Country)
	
    if req.IsDefaultShipping != nil {
        addr.IsDefaultShipping = *req.IsDefaultShipping
    }
    if req.IsDefaultBilling != nil {
        addr.IsDefaultBilling = *req.IsDefaultBilling
    }

    var updateErr error
    if req.IsDefaultShipping != nil && *req.IsDefaultShipping ||
        req.IsDefaultBilling != nil && *req.IsDefaultBilling {
        updateErr = s.store.ExecTx(ctx, func(q *domain.Queries) error {
            if req.IsDefaultShipping != nil && *req.IsDefaultShipping {
                if err := q.AddressRepo.UnsetDefaultShipping(ctx, userID); err != nil {
                    return fmt.Errorf("failed to unset default shipping: %w", err)
                }
            }
            if req.IsDefaultBilling != nil && *req.IsDefaultBilling {
                if err := q.AddressRepo.UnsetDefaultBilling(ctx, userID); err != nil {
                    return fmt.Errorf("failed to unset default billing: %w", err)
                }
            }
            if err := q.AddressRepo.Update(ctx, addr); err != nil {
                return fmt.Errorf("failed to update address in tx: %w", err)
            }
            return nil
        })
    } else {
        updateErr = s.repo.Update(ctx, addr)
    }

    if updateErr != nil {
        s.logger.Error("failed to update address", "id", id, "user_id", userID, "error", updateErr)
        return nil, fmt.Errorf("service: failed to update address: %w", updateErr)
    }

    s.logger.Info("address updated successfully", "user_id", userID, "address_id", addr.ID)
    return addr, nil
}

func (s *addressService) GetByID(ctx context.Context, id int64, userID int64) (*models.UserAddress, error) {
    addr, err := s.repo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to get address by ID", "id", id, "user_id", userID, "error", err)
        return nil, err
    }
    if addr.UserID != userID {
        return nil, apperrors.ErrUnauthorized
    }
    return addr, nil
}

func (s *addressService) ListByUserID(ctx context.Context, userID int64) ([]*models.UserAddress, error) {
    addresses, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        s.logger.Error("failed to list addresses", "user_id", userID, "error", err)
        return nil, err
    }
    return addresses, nil
}


func (s *addressService) Delete(ctx context.Context, userID int64, id int64) error {
    err := s.repo.Delete(ctx, id, userID)
    if err != nil {
        s.logger.Error("failed to delete address", "id", id, "user_id", userID, "error", err)
        return err
    }
    return nil
}

func (s *addressService) SetDefault(ctx context.Context, userID int64, id int64, addressType string) error {
    addr, err := s.GetByID(ctx, id, userID)
    if err != nil {
        return err
    }

    err = s.store.ExecTx(ctx, func(q *domain.Queries) error {
        if addressType == "shipping" {
            if err := q.AddressRepo.UnsetDefaultShipping(ctx, userID); err != nil {
                return err
            }
            addr.IsDefaultShipping = true
        } else if addressType == "billing" {
            if err := q.AddressRepo.UnsetDefaultBilling(ctx, userID); err != nil {
                return err
            }
            addr.IsDefaultBilling = true
        }
        return q.AddressRepo.Update(ctx, addr)
    })

    if err != nil {
        s.logger.Error("failed to set default address", "id", id, "user_id", userID, "type", addressType, "error", err)
        return err
    }
    return nil
}
