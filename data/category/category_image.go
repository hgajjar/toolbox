package category

import (
	"context"
	"database/sql"
	"queue-worker/data"

	"github.com/lib/pq"
)

const categoryImageResourceName = "category_image"

func (r *Repository) GetCategoryImageStorageData(ctx context.Context, filter data.Filter) ([]*CategoryImageStorageEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, locale FROM spy_category_image_storage WHERE fk_category = ANY($1) ORDER BY id_category_image_storage LIMIT $2 OFFSET $3;",
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, locale FROM spy_category_image_storage ORDER BY id_category_image_storage LIMIT $1 OFFSET $2;",
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	var data []*CategoryImageStorageEntity
	for rows.Next() {
		entity := &CategoryImageStorageEntity{}

		err = rows.Scan(&entity.Key, &entity.Data, &entity.Locale)
		if err != nil {
			return nil, err
		}

		data = append(data, entity)
	}

	return data, nil
}

func (r *Repository) GetCateogryImageResourceName() string {
	return categoryImageResourceName
}

func (r *Repository) GetCategoryImageMappings() []Mapping {
	return nil
}
