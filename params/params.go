package params

type CartItem struct {
	Sku      string
	Quantity int64
}

type CartAdd struct {
	Version string
	UserID  string
	Items   []CartItem
}
