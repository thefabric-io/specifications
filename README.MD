# Specifications for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/thefabric-io/specifications.svg)](https://pkg.go.dev/github.com/thefabric-io/specifications)
[![Go Report Card](https://goreportcard.com/badge/github.com/thefabric-io/specifications)](https://goreportcard.com/report/github.com/thefabric-io/specifications)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Specifications** is a lightweight, extensible, and storage-agnostic way to express complex query conditions in your Go applications using the Specification pattern. It allows you to define criteria (like `Equal`, `In`, `Like`, `GreaterThan`, `Limit`, `OrderBy`) in a composable, declarative manner, keeping your domain logic clean and decoupled from database specifics.

## Key Features

- **Domain-Driven Design Friendly**: Define specifications in your domain using conceptual field names without tying them to table or column names.
- **Pluggable Visitors**: Translate specifications into actual queries (SQL, NoSQL, in-memory filters, etc.) by implementing a `SpecificationVisitor`.
- **Rich Query Language**: Includes comparisons (`Equal`, `NotEqual`, `GreaterThan`, `LowerThan`, `Like`), set membership (`In`), logical composition (`And`, `Or`), and query modifiers (`Limit`, `Offset`, `OrderBy`).
- **Extensible**: Easily add new specification types or integrate with different databases by adding custom visitors.

## Installation

```bash
go get github.com/thefabric-io/specifications
```

## Structure

- `specifications/`: Core specifications, visitor interfaces, and factories.
- `specifications/postgres`: PostgreSQL-specific visitor that converts specs into SQL queries with parameter binding.

## Basic Usage

### 1. Define Domain-Level Field Names and Specifications

In your domain layer, create a specifier that uses conceptual field names (e.g. `"Status"`, `"ID"`) and the generic specification factories:

```go
// internal/domain/product/product.go
package product

import (
    "github.com/thefabric-io/specifications"
)

const (
    FieldID     = "ID"
    FieldStatus = "Status"
    FieldPrice  = "Price"
)

type Specifier struct{}

func NewSpecifier() *Specifier {
    return &Specifier{}
}

func (s *Specifier) WithID(id string) specifications.Specification {
    return specifications.Equal(FieldID, id)
}

func (s *Specifier) WithStatuses(statuses ...string) specifications.Specification {
    vals := make([]interface{}, len(statuses))
    for i, st := range statuses {
        vals[i] = st
    }
    return specifications.In(FieldStatus, vals...)
}
```

Your domain layer only knows about logical fields, not database columns.

### 2. Build Complex Query Specifications in Application Code

Combine specifications to express complex queries cleanly:

```go
// Some application/service code
spec := specifications.And(
    product.NewSpecifier().WithID("prod_2qQIX1KrAPawnJcnCey0g0eyyKK"),
    specifications.Or(
        product.NewSpecifier().WithStatuses("archived", "deleted"),
        specifications.GreaterThan(product.FieldPrice, 100),
    ),
    specifications.OrderBy(product.FieldStatus, "ASC"),
    specifications.Limit(10),
    specifications.Offset(20),
)
```

This reads like a fluent description of what you want: find products with `ID = "id123"` and either `status IN ("archived", "deleted")` or `price > 100`, then order by status ascending, limit to 10 results, and offset by 20.

### 3. Translate Domain Specs to Database Queries

In the infrastructure layer, map domain fields to actual database columns and use a visitor to generate the final SQL:

```go
// internal/infrastructure/database/product_repository.go
package database

import (
    "context"
    "database/sql"

    "github.com/thefabric-io/specifications"
    "github.com/thefabric-io/specifications/postgres"
    "your-module/internal/domain/product"
)

type Product struct {
    ID     string
    Name   string
    Status string
    Price  float64
}

type productRepository struct {
    db       *sql.DB
    fieldMap map[string]string
}

func NewProductRepository(db *sql.DB) *productRepository {
    // Map domain fields to actual DB columns
    fieldMap := map[string]string{
        product.FieldID:     "id",
        product.FieldStatus: "status",
        product.FieldPrice:  "price",
    }
    return &productRepository{db: db, fieldMap: fieldMap}
}

func (r *productRepository) Load(ctx context.Context, spec specifications.Specification) ([]Product, error) {
    visitor := postgres.NewVisitor(r.fieldMap)
    if spec != nil {
        spec.Accept(visitor)
    }

    baseQuery := "SELECT id, name, status, price FROM products"
    query, args := visitor.BuildQuery(baseQuery)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        if err := rows.Scan(&p.ID, &p.Name, &p.Status, &p.Price); err != nil {
            return nil, err
        }
        products = append(products, p)
    }
    return products, nil
}
```

### 4. Execute the Query

```go
// In your application code:
ctx := context.Background()
packs, err := packRepository.Load(ctx, spec)
if err != nil {
    // handle error
}

// 'packs' now contains the filtered, ordered, and paginated results.
```

## Adding New Specifications or Visitors

To add a new comparison operator (e.g., `Between`), simply:

1. Add a new `Specification` struct and factory method in `specifications.go`.
2. Add a corresponding `VisitXXX` method in the `SpecificationVisitor` interface.
3. Implement that method in your `postgres.Visitor` (or any other visitor you create).

This modular approach keeps your domain logic separate from the underlying query mechanism.

## Contributing

Contributions, suggestions, and bug reports are welcome! Feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

---

Happy querying!