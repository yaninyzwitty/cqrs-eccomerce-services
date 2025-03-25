package controllers

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/internal/database"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductQueryController struct {
	pb.UnimplementedProductServiceQueryServer
	session        *gocql.Session
	memcacheClient *database.MemcachedClient
}

func NewProductQueryController(session *gocql.Session, memcacheClient *database.MemcachedClient) *ProductQueryController {
	return &ProductQueryController{session: session, memcacheClient: memcacheClient}
}

func (c *ProductQueryController) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	getCategoryQuery := `SELECT id, name, description, created_at FROM products_keyspace_v3.categories WHERE id = ?`
	var categoryId int64
	var name, description string
	var createdAt time.Time

	if err := c.session.Query(getCategoryQuery, req.Id).WithContext(ctx).Scan(&categoryId, &name, &description, &createdAt); err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "category not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get category: %v", err)
	}

	return &pb.GetCategoryResponse{
		Id:          categoryId,
		Name:        name,
		Description: description,
		CreatedAt:   timestamppb.New(createdAt),
	}, nil
}

func (c *ProductQueryController) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	if req.CategoryId == 0 || req.ProductId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "category id and product id are required")
	}

	var product pb.Product
	var createdAt, updatedAt time.Time

	getProductQuery := `SELECT id, name, description, price, stock, category_id, created_at, updated_at FROM products_keyspace_v3.products WHERE category_id = ? AND id = ?`
	err := c.session.Query(getProductQuery, req.CategoryId, req.ProductId).WithContext(ctx).Scan(
		&product.Id, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CategoryId, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch product: %v", err)
	}

	product.CreatedAt = timestamppb.New(createdAt)
	product.UpdatedAt = timestamppb.New(updatedAt)

	return &pb.GetProductResponse{Product: &product}, nil
}

func (c *ProductQueryController) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	if req.CategoryId == 0 || req.PageSize <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "category id and page size are required")
	}

	query := c.session.Query(`
		SELECT id, name, description, price, stock, created_at, updated_at 
		FROM products_keyspace_v3.products 
		WHERE category_id = ?`,
		req.CategoryId,
	).WithContext(ctx).PageSize(int(req.PageSize)).PageState(req.PagingState)

	if len(req.PagingState) > 0 {
		query = query.PageState(req.PagingState)
	}

	iter := query.Iter()
	defer iter.Close()

	var products []*pb.Product
	var (
		id          int64
		name        string
		description string
		price       float32
		stock       int32
		createdAt   time.Time
		updatedAt   time.Time
	)

	for iter.Scan(&id, &name, &description, &price, &stock, &createdAt, &updatedAt) {
		products = append(products, &pb.Product{
			Id:          id,
			Name:        name,
			Description: description,
			Price:       price,
			Stock:       stock,
			CategoryId:  req.CategoryId,
			CreatedAt:   timestamppb.New(createdAt),
			UpdatedAt:   timestamppb.New(updatedAt),
		})
	}

	// ✅ Check if Cassandra actually returns a paging state
	nextPageState := iter.PageState()

	if err := iter.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	return &pb.ListProductsResponse{
		Products:    products,
		PagingState: nextPageState,
	}, nil
}
