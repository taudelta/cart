package model

type CartItem struct {
	Sku      string
	Quantity int
}

type Cart struct {
	Items []CartItem
}
