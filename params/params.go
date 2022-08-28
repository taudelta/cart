package params

type CartItem struct {
	Sku      string
	Quantity int64
}

type CartAdd struct {
	UserID string
	Items  []CartItem
}

type CartRemove struct {
	UserID    string
	Items     []CartItem
	RemoveAll bool
}

type Delete struct {
	UserID string
}
