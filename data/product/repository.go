package product

import (
	"context"
	"database/sql"
	"queue-worker/data"

	"github.com/lib/pq"
)

const resourceName = "product_abstract"

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) GetProductAbstractStorageData(ctx context.Context, filter data.Filter) ([]*ProductAbstractStorageEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_product_abstract_storage WHERE fk_product_abstract = ANY($1) ORDER BY id_product_abstract_storage LIMIT $2 OFFSET $3;",
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_product_abstract_storage ORDER BY id_product_abstract_storage LIMIT $1 OFFSET $2;",
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	var data []*ProductAbstractStorageEntity
	for rows.Next() {
		entity := &ProductAbstractStorageEntity{}

		err = rows.Scan(&entity.Key, &entity.Data, &entity.Store, &entity.Locale)
		if err != nil {
			return nil, err
		}

		data = append(data, entity)
	}

	return data, nil
}

func (r *Repository) GetProductAbstractResourceName() string {
	return resourceName
}

func (r *Repository) GetProductAbstractMappings() []Mapping {
	return []Mapping{
		{
			Source:      "sku",
			Destination: "id_product_abstract",
		},
	}
}
