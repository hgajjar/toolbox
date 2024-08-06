package category

import (
	"context"
	"database/sql"
	"log"
	"queue-worker/data"

	"github.com/lib/pq"
)

const categoryTreeResourceName = "category_tree"

func (r *Repository) GetCategoryTreeStorageData(ctx context.Context, filter data.Filter) (<-chan *CategoryTreeStorageEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_category_tree_storage WHERE id_category_tree_storage = ANY($1) ORDER BY id_category_tree_storage LIMIT $2 OFFSET $3;",
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, store, locale FROM spy_category_tree_storage ORDER BY id_category_tree_storage LIMIT $1 OFFSET $2;",
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	dataCh := make(chan *CategoryTreeStorageEntity)
	go func() {
		for rows.Next() {
			entity := &CategoryTreeStorageEntity{}

			err = rows.Scan(&entity.Key, &entity.Data, &entity.Store, &entity.Locale)
			if err != nil {
				log.Println(err.Error())

				close(dataCh)

				return
			}

			dataCh <- entity
		}
		close(dataCh)
	}()

	return dataCh, nil
}

func (r *Repository) GetCateogryTreeResourceName() string {
	return categoryTreeResourceName
}

func (r *Repository) GetCategoryTreeMappings() []Mapping {
	return nil
}
