
```markdown
# CQRS E-commerce Service

A scalable e-commerce service implementing CQRS (Command Query Responsibility Segregation) pattern with separate read and write models, using GraphQL for the client interface.

## Architecture Overview

### Write Model (Commands)
- Handles all write operations (create/update/delete)
- Processes business logic and data validation
- Publishes events to message bus
- Uses separate write database optimized for writes

Commands:
- CreateProduct
- UpdateProduct 
- DeleteProduct
- CreateOrder
- UpdateOrderStatus
- AddToCart
- RemoveFromCart

### Read Model (Queries)
- Handles all read operations
- Optimized for fast querying
- Subscribes to events to update read models
- Uses separate read database optimized for reads

Queries:
- GetProduct
- GetProducts
- GetOrder
- GetOrders
- GetCart
- GetOrderHistory

### GraphQL API
- Single endpoint for all operations
- Type-safe schema
- Real-time subscriptions for updates
- Query optimization

## Tech Stack

- Node.js/TypeScript
- Apollo GraphQL Server
- PostgreSQL (Write DB)
- MongoDB (Read DB) 
- Redis (Caching)
- RabbitMQ (Event Bus)
- Docker
- Kubernetes

## Getting Started

1. Clone the repository
git clone https://github.com/your-repo/cqrs-ecommerce.gitegregation) pattern with separate read and write models, using GraphQL for the client interface.

## Architecture Overview

### Write Model (Commands)
- Handles all write operations (create/update/delete)
- Processes business logic and data validation
- Publishes events to message bus
- Uses separate write database optimized for writes

Commands:
- CreateProduct
- UpdateProduct 
- DeleteProduct
- CreateOrder
- UpdateOrderStatus
- AddToCart
- RemoveFromCart

### Read Model (Queries)
- Handles all read operations
- Optimized for fast querying
- Subscribes to events to update read models
- Uses separate read database optimized for reads

Queries:
- GetProduct
- GetProducts
- GetOrder
- GetOrders
- GetCart
- GetOrderHistory

### GraphQL API
- Single endpoint for all operations
- Type-safe schema
- Real-time subscriptions for updates
- Query optimization

## Tech Stack

- Node.js/TypeScript
- Apollo GraphQL Server
- PostgreSQL (Write DB)
- MongoDB (Read DB) 
- Redis (Caching)
- RabbitMQ (Event Bus)
- Docker
- Kubernetes


