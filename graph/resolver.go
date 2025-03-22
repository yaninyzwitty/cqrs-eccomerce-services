package graph

import "github.com/yaninyzwitty/cqrs-eccomerce-service/pb"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	CommandClient pb.ProductServiceClient
	QueryClient   pb.ProductServiceClient
}
