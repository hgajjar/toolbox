package content

import (
	"context"
	"database/sql"
	"log"
	"queue-worker/data"

	"github.com/lib/pq"
)

const resourceName = "content"

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) GetContentStorageData(ctx context.Context, filter data.Filter) (<-chan *ContentStorageEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, locale FROM spy_content_storage WHERE fk_content = ANY($1) ORDER BY id_content_storage LIMIT $2 OFFSET $3;",
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			"SELECT key, data, locale FROM spy_content_storage ORDER BY id_content_storage LIMIT $1 OFFSET $2;",
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	dataCh := make(chan *ContentStorageEntity)
	go func() {
		for rows.Next() {
			entity := &ContentStorageEntity{}

			err = rows.Scan(&entity.Key, &entity.Data, &entity.Locale)
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

func (r *Repository) GetContentResourceName() string {
	return resourceName
}

func (r *Repository) GetContentMappings() []Mapping {
	return []Mapping{
		{
			Source:      "content_key",
			Destination: "idContent",
		},
	}
}
