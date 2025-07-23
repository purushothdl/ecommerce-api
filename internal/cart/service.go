// internal/cart/service.go
package cart

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/purushothdl/ecommerce-api/internal/domain"
	"github.com/purushothdl/ecommerce-api/internal/models"
	apperrors "github.com/purushothdl/ecommerce-api/pkg/errors"
)

type cartService struct {
	cartRepo    domain.CartRepository
	productRepo domain.ProductRepository
	store       domain.Store
	logger      *slog.Logger
}

func NewCartService(cartRepo domain.CartRepository, productRepo domain.ProductRepository, store domain.Store, logger *slog.Logger) domain.CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		store:       store,
		logger:      logger,
	}
}

// getCartWithItems is a helper to assemble the full cart object.
func (s *cartService) getCartWithItems(ctx context.Context, cart *models.Cart) (*models.Cart, error) {
	items, err := s.cartRepo.GetItemsByCartID(ctx, cart.ID)
	if err != nil {
		s.logger.Error("failed to get cart items", "cart_id", cart.ID, "error", err)
		return nil, fmt.Errorf("could not retrieve cart items: %w", err)
	}
	cart.Items = items

	var total float64
	for _, item := range items {
		if item.Product != nil {
			total += item.Product.Price * float64(item.Quantity)
		}
	}
	cart.Total = total

	return cart, nil
}

func (s *cartService) GetOrCreateCart(ctx context.Context, userID *int64, anonymousCartID *int64) (*models.Cart, error) {
	if userID != nil {
		cart, err := s.cartRepo.GetByUserID(ctx, *userID)
		if err == nil {
			return cart, nil 
		}
		if !errors.Is(err, apperrors.ErrNotFound) {
			s.logger.Error("error getting cart by user id", "user_id", *userID, "error", err)
			return nil, err
		}
		// Not found, so create one for the user
		s.logger.Info("no cart found for user, creating new one", "user_id", *userID)
		return s.cartRepo.Create(ctx, userID)
	}

	if anonymousCartID != nil {
		cart, err := s.cartRepo.GetByID(ctx, *anonymousCartID)
		if err == nil {
			// Ensure this cart is actually anonymous
			if cart.UserID == nil {
				return cart, nil
			}
			// If it's not anonymous, something is wrong. Fall through to create a new one.
			s.logger.Warn("cart id from cookie belongs to a registered user", "cart_id", *anonymousCartID)
		}
	}

	// No user and no valid anonymous cart, so create a new anonymous cart
	s.logger.Info("creating new anonymous cart")
	return s.cartRepo.Create(ctx, nil)
}

func (s *cartService) AddProductToCart(ctx context.Context, cartID int64, productID int64, quantity int) (*models.Cart, error) {
	s.logger.Info("adding product to cart within transaction", "cart_id", cartID, "product_id", productID, "quantity", quantity)

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		return q.CartRepo.AddItem(ctx, cartID, productID, quantity)
	})

	if err != nil {
		// The error from the repository (e.g., ErrInsufficientStock) will be passed up
		s.logger.Warn("failed to add item to cart", "error", err)
		return nil, err
	}

	return s.GetCartContents(ctx, cartID)
}

func (s *cartService) UpdateProductInCart(ctx context.Context, cartID int64, productID int64, quantity int) (*models.Cart, error) {
	s.logger.Info("updating product in cart within transaction", "cart_id", cartID, "product_id", productID, "new_quantity", quantity)

	if quantity <= 0 {
		// Removing an item can also be done in a transaction for consistency
		return s.RemoveProductFromCart(ctx, cartID, productID)
	}

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		return q.CartRepo.UpdateItemQuantity(ctx, cartID, productID, quantity)
	})

	if err != nil {
		s.logger.Error("failed to update item in cart", "error", err)
		return nil, err
	}
	
	return s.GetCartContents(ctx, cartID)
}

func (s *cartService) RemoveProductFromCart(ctx context.Context, cartID int64, productID int64) (*models.Cart, error) {
	s.logger.Info("removing product from cart within transaction", "cart_id", cartID, "product_id", productID)

	err := s.store.ExecTx(ctx, func(q *domain.Queries) error {
		return q.CartRepo.RemoveItem(ctx, cartID, productID)
	})

	if err != nil {
		s.logger.Error("failed to remove item from cart", "error", err)
		return nil, err
	}

	return s.GetCartContents(ctx, cartID)
}

func (s *cartService) GetCartContents(ctx context.Context, cartID int64) (*models.Cart, error) {
	s.logger.Info("getting cart contents", "cart_id", cartID)

	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	return s.getCartWithItems(ctx, cart)
}

func (s *cartService) HandleLoginWithTransaction(ctx context.Context, q *domain.Queries, userID int64, anonymousCartID int64) error {
    if anonymousCartID == 0 {
        return nil 
    }

    userCart, err := q.CartRepo.GetByUserID(ctx, userID)
    if err != nil {
        if errors.Is(err, apperrors.ErrNotFound) {
            userCart, err = q.CartRepo.Create(ctx, &userID)
            if err != nil {
                return fmt.Errorf("failed to create user cart: %w", err)
            }
        } else {
            return fmt.Errorf("failed to get user cart: %w", err)
        }
    }
	// Same cart, nothing to merge
    if userCart.ID == anonymousCartID {
        return nil 
    }

    // Merge carts
    if err := q.CartRepo.MergeCarts(ctx, anonymousCartID, userCart.ID); err != nil {
        return fmt.Errorf("failed to merge carts: %w", err)
    }

    // Delete anonymous cart
    return q.CartRepo.Delete(ctx, anonymousCartID)
}
