package category

import (
	"context"
	"database/sql"
	"queue-worker/data"

	"github.com/lib/pq"
)

const categoryNodeResourceName = "category_node"

func (r *Repository) GetCategoryNodeStorageData(ctx context.Context, filter data.Filter) ([]*CategoryNodeStorageEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_category_node_storage WHERE fk_category_node = ANY($1) ORDER BY id_category_node_storage LIMIT $2 OFFSET $3;",
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_category_node_storage ORDER BY id_category_node_storage LIMIT $1 OFFSET $2;",
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	var data []*CategoryNodeStorageEntity
	for rows.Next() {
		entity := &CategoryNodeStorageEntity{}

		err = rows.Scan(&entity.Key, &entity.Data, &entity.Store, &entity.Locale)
		if err != nil {
			return nil, err
		}

		data = append(data, entity)
	}

	return data, nil
}

func (r *Repository) GetCateogryNodeResourceName() string {
	return categoryNodeResourceName
}

func (r *Repository) GetCategoryNodeMappings() []Mapping {
	return nil
}
