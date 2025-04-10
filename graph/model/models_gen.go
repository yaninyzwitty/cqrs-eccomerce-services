// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	CategoryID  string  `json:"categoryId"`
}

type ListProductsResponse struct {
	Products    []*Product `json:"products"`
	PagingState *string    `json:"pagingState,omitempty"`
}

type Mutation struct {
}

type Product struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"categoryId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int32     `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Query struct {
}
