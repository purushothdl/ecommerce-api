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
	logger      *slog.Logger
}

func NewCartService(cartRepo domain.CartRepository, productRepo domain.ProductRepository, logger *slog.Logger) domain.CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
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
	s.logger.Info("adding product to cart", "cart_id", cartID, "product_id", productID, "quantity", quantity)

	// Check product stock
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		s.logger.Warn("attempted to add non-existent product to cart", "product_id", productID)
		return nil, apperrors.ErrNotFound
	}
	if product.StockQuantity < quantity {
		s.logger.Warn("not enough stock to add to cart", "product_id", productID, "stock", product.StockQuantity, "requested", quantity)
		return nil, apperrors.ErrInsufficientStock 
	}

	if err := s.cartRepo.AddItem(ctx, cartID, productID, quantity); err != nil {
		s.logger.Error("failed to add item to cart repo", "cart_id", cartID, "product_id", productID, "error", err)
		return nil, err
	}

	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	return s.getCartWithItems(ctx, cart)
}

func (s *cartService) UpdateProductInCart(ctx context.Context, cartID int64, productID int64, quantity int) (*models.Cart, error) {
	s.logger.Info("updating product in cart", "cart_id", cartID, "product_id", productID, "new_quantity", quantity)

	if quantity <= 0 {
		return s.RemoveProductFromCart(ctx, cartID, productID)
	}

	if err := s.cartRepo.UpdateItemQuantity(ctx, cartID, productID, quantity); err != nil {
		s.logger.Error("failed to update item in cart repo", "cart_id", cartID, "product_id", productID, "error", err)
		return nil, err
	}
	
	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	return s.getCartWithItems(ctx, cart)
}

func (s *cartService) RemoveProductFromCart(ctx context.Context, cartID int64, productID int64) (*models.Cart, error) {
	s.logger.Info("removing product from cart", "cart_id", cartID, "product_id", productID)

	if err := s.cartRepo.RemoveItem(ctx, cartID, productID); err != nil {
		s.logger.Error("failed to remove item from cart repo", "cart_id", cartID, "product_id", productID, "error", err)
		return nil, err
	}

	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	return s.getCartWithItems(ctx, cart)
}

func (s *cartService) GetCartContents(ctx context.Context, cartID int64) (*models.Cart, error) {
	s.logger.Info("getting cart contents", "cart_id", cartID)

	cart, err := s.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	return s.getCartWithItems(ctx, cart)
}

// Add to internal/cart/service.go
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

    if userCart.ID == anonymousCartID {
        return nil // Same cart, nothing to merge
    }

    // Merge carts
    if err := q.CartRepo.MergeCarts(ctx, anonymousCartID, userCart.ID); err != nil {
        return fmt.Errorf("failed to merge carts: %w", err)
    }

    // Delete anonymous cart
    return q.CartRepo.Delete(ctx, anonymousCartID)
}
