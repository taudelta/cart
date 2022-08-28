package service

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v9"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/taudelta/cart/model"
	"github.com/taudelta/cart/service/dto"
)

func TestCart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cart Suite")
}

var redisClient *redis.Client
var cartService *Cart

var _ = BeforeSuite(func() {
	redisOptions, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(redisOptions)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	cartService = &Cart{
		Storage: redisClient,
	}
})

var _ = AfterSuite(func() {
	redisClient.FlushDB(context.Background())
})

var _ = Describe("empty cart", func() {
	When("get empty cart", func() {
		BeforeEach(func() {
			cartService.Storage.FlushDB(context.Background())
		})
		Context("empty", func() {
			It("service must return no error", func() {
				expectedCart := &model.Cart{
					Items: []model.CartItem{},
				}
				cart, err := cartService.GetCart("1")
				Expect(err).To(BeNil())
				Expect(cart).To(Equal(expectedCart))
			})
		})
	})
})

var _ = Describe("delete cart", func() {
	When("delete cart", func() {
		BeforeEach(func() {
			cartService.Storage.FlushDB(context.Background())

			err := cartService.AddItem("1", []dto.CartItem{
				{
					Sku:      "A",
					Quantity: 2,
				},
				{
					Sku:      "B",
					Quantity: 2,
				},
			})
			if err != nil {
				panic(err)
			}
		})
		Context("empty", func() {
			It("service must return no error", func() {
				expectedCart := &model.Cart{
					Items: []model.CartItem{},
				}
				err := cartService.DeleteCart("1")
				Expect(err).To(BeNil())

				cart, err := cartService.GetCart("1")
				Expect(err).To(BeNil())
				Expect(cart).To(Equal(expectedCart))
			})
		})
	})
})

var _ = Describe("cart add item", func() {
	When("add 1 item to cart", func() {
		BeforeEach(func() {
			cartService.Storage.FlushDB(context.Background())
		})
		Context("add item", func() {
			It("service must return no error", func() {
				err := cartService.AddItem("1", []dto.CartItem{
					{
						Sku:      "A",
						Quantity: 2,
					},
				})
				Expect(err).To(BeNil())

				expectedCart := &model.Cart{
					Items: []model.CartItem{
						{
							Sku:      "A",
							Quantity: 2,
						},
					},
				}
				cart, err := cartService.GetCart("1")
				Expect(err).To(BeNil())
				Expect(cart).To(Equal(expectedCart))
			})
		})
	})
})

var _ = Describe("cart remove item", func() {
	When("remove 1 item from cart", func() {
		BeforeEach(func() {
			err := cartService.Storage.FlushDB(context.Background()).Err()
			if err != nil {
				panic(err)
			}
			err = cartService.AddItem("1", []dto.CartItem{
				{Sku: "A", Quantity: 2},
				{Sku: "B", Quantity: 2},
			})
			if err != nil {
				panic(err)
			}
		})
		Context("remove item", func() {
			It("service must return no error", func() {
				err := cartService.RemoveItem("1", []dto.CartItem{
					{
						Sku:      "A",
						Quantity: 1,
					},
				}, false)
				Expect(err).To(BeNil())

				expectedCart := &model.Cart{
					Items: []model.CartItem{
						{
							Sku:      "A",
							Quantity: 1,
						},
						{
							Sku:      "B",
							Quantity: 2,
						},
					},
				}
				cart, err := cartService.GetCart("1")
				Expect(err).To(BeNil())
				Expect(cart).To(Equal(expectedCart))
			})
		})
	})
})
