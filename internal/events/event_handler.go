package events

import (
	"encoding/json"
	"fmt"

	"github.com/yaninyzwitty/cqrs-eccomerce-service/pb"
)

func HandleCategoryCreated(payload string) ([]byte, error) {
	var category pb.Category
	if err := json.Unmarshal([]byte(payload), &category); err != nil {
		return nil, fmt.Errorf("error unmarshalling category: %w", err)
	}
	category.EventType = "category.created"

	return json.Marshal(&category)
}

func HandleProductCreated(payload string) ([]byte, error) {
	var product pb.Product
	if err := json.Unmarshal([]byte(payload), &product); err != nil {
		return nil, fmt.Errorf("error unmarshalling product: %w", err)
	}
	product.EventType = "product.created"

	return json.Marshal(&product)
}
