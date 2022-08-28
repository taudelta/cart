package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/go-redis/redis/v9"
	"github.com/taudelta/cart/model"
	"github.com/taudelta/cart/service/dto"
)

type Cart struct {
	Storage *redis.Client
}

func (s *Cart) AddItem(userID string, addItems []dto.CartItem) error {
	ctx := context.Background()

	cartData := s.Storage.HGetAll(ctx, userID)
	if err := cartData.Err(); err != nil {
		return err
	}

	cartValue := cartData.Val()

	var items []*model.CartItem

	if cartValue["items"] != "" {
		if err := json.Unmarshal([]byte(cartValue["items"]), &items); err != nil {
			return err
		}
	}

	itemsMap := make(map[string]*model.CartItem)
	for _, i := range items {
		itemsMap[i.Sku] = i
	}

	for _, addItem := range addItems {
		if _, ok := itemsMap[addItem.Sku]; ok {
			itemsMap[addItem.Sku].Quantity += int(addItem.Quantity)
		} else {
			items = append(items, &model.CartItem{
				Sku:      addItem.Sku,
				Quantity: int(addItem.Quantity),
			})
		}
	}

	itemBody, err := json.Marshal(&items)
	if err != nil {
		return err
	}

	pipeline := s.Storage.Pipeline()
	pipeline.HSet(ctx, userID, "items", string(itemBody))

	if _, err := pipeline.Exec(ctx); err != nil {
		log.Println("failed to add items in cart", err)
		return err
	}

	return nil
}

func (s *Cart) RemoveItem(userID string, removeItems []dto.CartItem, removeAll bool) error {
	ctx := context.Background()

	cartData := s.Storage.HGetAll(ctx, userID)
	if err := cartData.Err(); err != nil {
		return err
	}

	cartValue := cartData.Val()

	var items []*model.CartItem

	if cartValue["items"] != "" {
		if err := json.Unmarshal([]byte(cartValue["items"]), &items); err != nil {
			return err
		}
	}

	removedItemsMap := make(map[string]dto.CartItem)
	for _, i := range removeItems {
		removedItemsMap[i.Sku] = i
	}

	var updatedItems []*model.CartItem
	for index, i := range items {
		if removeItemData, ok := removedItemsMap[i.Sku]; ok {
			if removeAll {
				items[index].Quantity = 0
			} else {
				items[index].Quantity -= int(removeItemData.Quantity)
			}
		}
		if items[index].Quantity > 0 {
			updatedItems = append(updatedItems, i)
		}
	}

	itemBody, err := json.Marshal(&updatedItems)
	if err != nil {
		return err
	}

	pipeline := s.Storage.Pipeline()
	if err := pipeline.HSet(ctx, userID, "items", string(itemBody)).Err(); err != nil {
		return err
	}

	if _, err := pipeline.Exec(ctx); err != nil {
		log.Println("failed to remove items", err)
		return err
	}

	return nil
}

func (s *Cart) GetCart(userID string) (*model.Cart, error) {
	cartData := s.Storage.HGetAll(context.Background(), userID)
	if err := cartData.Err(); err != nil {
		log.Println("get cart error", err)
		return nil, err
	}

	cartValue := cartData.Val()
	if cartValue == nil {
		return nil, errors.New("no cart is found")
	}

	items := make([]model.CartItem, 0)
	if cartValue["items"] != "" {
		if err := json.Unmarshal([]byte(cartValue["items"]), &items); err != nil {
			return nil, err
		}
	}

	var cart model.Cart
	cart.Items = items

	return &cart, nil
}

func (s *Cart) DeleteCart(userID string) error {
	ctx := context.Background()

	err := s.Storage.Del(ctx, userID).Err()
	if err != nil {
		return err
	}

	return nil
}
