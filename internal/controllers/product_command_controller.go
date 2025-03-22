package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/pb"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/snowflake"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductCommandController struct {
	pb.UnimplementedProductServiceServer
	session *gocql.Session
}

func NewCommandProductCommandController(session *gocql.Session) *ProductCommandController {
	return &ProductCommandController{session: session}
}

func (c *ProductCommandController) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if req.Name == "" || req.Description == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name and description are required")
	}

	categoryId, err := snowflake.GenerateID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate category id")
	}

	now := time.Now()
	createCategoryQuery := `INSERT INTO products_keyspace_v2.categories(id, name, description, created_at) VALUES(?, ?, ?, ?)`

	if err := c.session.Query(createCategoryQuery, categoryId, req.Name, req.Description, now).WithContext(ctx).Exec(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	return &pb.CreateCategoryResponse{
		Id:          int64(categoryId),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   timestamppb.New(now),
	}, nil
}

func (c *ProductCommandController) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	if req.Name == "" || req.Description == "" || req.Price == 0 || req.CategoryId == 0 || req.Stock == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "name, description, price, stock, and category id are required")
	}

	productId, err := snowflake.GenerateID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate product id")
	}

	now := time.Now()
	outboxID := gocql.TimeUUID()
	bucket := now.Format("2006-01-02")
	eventType := "create_product_event"

	product := &pb.Product{
		Id:          int64(productId),
		CategoryId:  req.CategoryId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	payload, err := json.Marshal(product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal product: %v", err)
	}

	batch := c.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	batch.Query(
		`INSERT INTO products_keyspace_v2.products 
		(id, name, description, price, stock, category_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		productId, req.Name, req.Description, req.Price, req.Stock, req.CategoryId, now, now,
	)

	batch.Query(
		`INSERT INTO products_keyspace_v2.products_outbox 
		(id, bucket, payload, event_type) 
		VALUES (?, ?, ?, ?)`,
		outboxID, bucket, payload, eventType,
	)

	if err := c.session.ExecuteBatch(batch); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &pb.CreateProductResponse{Product: product}, nil
}
