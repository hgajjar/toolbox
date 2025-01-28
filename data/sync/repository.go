package sync

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/data"

	"github.com/lib/pq"
)

type Repository struct {
	conn   *sql.DB
	config *config.SyncEntity
}

func NewRepository(conn *sql.DB, syncConfigEntity *config.SyncEntity) *Repository {
	return &Repository{
		conn:   conn,
		config: syncConfigEntity,
	}
}

func (r *Repository) GetData(ctx context.Context, filter data.Filter) (<-chan *SyncEntity, error) {
	var rows *sql.Rows
	var err error

	if len(filter.IDs) > 0 {
		rows, err = r.conn.QueryContext(
			ctx,
			fmt.Sprintf(
				"SELECT key, data %s FROM %s WHERE %s = ANY($1) ORDER BY %s LIMIT $2 OFFSET $3;",
				r.buildStoreAndLocalePlaceholder(),
				r.config.Table,
				r.config.FilterColumn,
				r.config.IdColumn,
			),
			pq.Array(filter.IDs),
			filter.Limit,
			filter.Offset,
		)
	} else {
		rows, err = r.conn.QueryContext(
			ctx,
			fmt.Sprintf(
				"SELECT key, data %s FROM %s ORDER BY %s LIMIT $1 OFFSET $2;",
				r.buildStoreAndLocalePlaceholder(),
				r.config.Table,
				r.config.IdColumn,
			),
			filter.Limit,
			filter.Offset,
		)
	}

	if err != nil {
		return nil, err
	}

	dataCh := make(chan *SyncEntity)
	go func() {
		for rows.Next() {
			entity := &SyncEntity{}

			if r.config.Store && r.config.Locale {
				err = rows.Scan(&entity.Key, &entity.Data, &entity.Store, &entity.Locale)
			} else if r.config.Store {
				err = rows.Scan(&entity.Key, &entity.Data, &entity.Store)
			} else if r.config.Locale {
				err = rows.Scan(&entity.Key, &entity.Data, &entity.Locale)
			} else {
				err = rows.Scan(&entity.Key, &entity.Data)
			}

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

func (r *Repository) buildStoreAndLocalePlaceholder() string {
	placeholders := []string{""}

	if r.config.Store {
		placeholders = append(placeholders, "store")
	}

	if r.config.Locale {
		placeholders = append(placeholders, "locale")
	}

	return strings.Join(placeholders, ", ")
}
