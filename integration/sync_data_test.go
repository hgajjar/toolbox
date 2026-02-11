//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/hgajjar/toolbox/config"
	"github.com/hgajjar/toolbox/container"
	"github.com/hgajjar/toolbox/integration/rabbitmq"
	"github.com/hgajjar/toolbox/sync"
)

func TestSyncData(t *testing.T) {
	ctx := context.Background()

	closeFn, postgresHostPort, err := setupPostgresService(ctx)
	defer closeFn()
	if err != nil {
		t.Fatalf("failed to setup postgres service: %v", err)
	}

	deferFn, rabbitmqHostPort, err := setupRabbitMqService(ctx)
	defer deferFn()
	if err != nil {
		t.Fatal(err)
	}

	dic := container.New()
	defer dic.Close()

	db := dic.DB(buildPostgresConnStr(postgresHostPort))

	rabbitmqPort, _ := strconv.Atoi(rabbitmqHostPort)
	postgresPort, _ := strconv.Atoi(postgresHostPort)
	cfg := &config.Config{
		RabbitMq: config.RabbitMQ{
			Server:   "127.0.0.1",
			Port:     rabbitmqPort,
			User:     "guest",
			Password: "guest",
		},
		Postgres: config.Postgres{
			Server:   "127.0.0.1",
			Port:     postgresPort,
			User:     "postgres",
			Password: "postgres",
			Database: "postgres",
		},
	}

	rmq, err := rabbitmq.New(rabbitmq.Config{
		URL: fmt.Sprintf("amqp://guest:guest@127.0.0.1:%s/", rabbitmqHostPort),
	})
	if err != nil {
		log.Panic(err)
	}
	defer rmq.Close()

	if err := defineQueues([]string{"sync.search.cms", "sync.storage.cms"}, rmq); err != nil {
		t.Fatal(err)
	}

	setupTestData(ctx, db)

	t.Run("It publishes resource data with additional params", func(t *testing.T) {
		cfg.SyncDataEntities = []config.SyncEntity{
			{
				Resource:     "cms_page_search",
				Table:        "spy_cms_page_search",
				IdColumn:     "id_cms_page_search",
				FilterColumn: "fk_cms_page",
				QueueGroup:   "sync.search.cms",
				Store:        true,
				Locale:       true,
				Params: map[string]string{
					"type": "page",
				},
			},
		}

		sync.RunSyncData(
			ctx,
			dic,
			cfg,
			sync.SyncDataArgs{},
		)

		messages := consumeMessagesFromQueue(rmq, "sync.search.cms")
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages in the queue, got %d", len(messages))
		}

		expectedMessages := []string{
			`{"write":{"key":"key1","value":{"test":"value1"},"resource":"cms_page_search","store":"US","params":{"type":"page"}}}`,
			`{"write":{"key":"key2","value":{"test":"value2"},"resource":"cms_page_search","store":"US","params":{"type":"page"}}}`,
			`{"write":{"key":"key3","value":{"test":"value3"},"resource":"cms_page_search","store":"US","params":{"type":"page"}}}`,
		}

		for i, msg := range messages {
			if msg != expectedMessages[i] {
				t.Errorf("expected message %d to be %s, got %s", i+1, expectedMessages[i], msg)
			}
		}
	})

	t.Run("It skips additional params when they are not configured", func(t *testing.T) {
		cfg.SyncDataEntities = []config.SyncEntity{
			{
				Resource:     "cms_page",
				Table:        "spy_cms_page_storage",
				IdColumn:     "id_cms_page_storage",
				FilterColumn: "fk_cms_page",
				QueueGroup:   "sync.storage.cms",
				Store:        true,
				Locale:       true,
			},
		}

		sync.RunSyncData(
			ctx,
			dic,
			cfg,
			sync.SyncDataArgs{},
		)

		messages := consumeMessagesFromQueue(rmq, "sync.storage.cms")
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages in the queue, got %d", len(messages))
		}

		expectedMessages := []string{
			`{"write":{"key":"key1","value":{"test":"value1"},"resource":"cms_page","store":"US"}}`,
			`{"write":{"key":"key2","value":{"test":"value2"},"resource":"cms_page","store":"US"}}`,
			`{"write":{"key":"key3","value":{"test":"value3"},"resource":"cms_page","store":"US"}}`,
		}

		for i, msg := range messages {
			if msg != expectedMessages[i] {
				t.Errorf("expected message %d to be %s, got %s", i+1, expectedMessages[i], msg)
			}
		}
	})
}

func setupTestData(ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS spy_cms_page_search (
			id_cms_page_search SERIAL PRIMARY KEY,
			fk_cms_page INTEGER,
			key VARCHAR(255) NOT NULL,
			data TEXT NOT NULL,
			store VARCHAR(2),
			locale VARCHAR(6)
		);
		INSERT INTO spy_cms_page_search (id_cms_page_search, fk_cms_page, key, data, store, locale) VALUES 
		(1, 101, 'key1', '{"test": "value1"}', 'US', 'en_US'), (2, 102, 'key2', '{"test": "value2"}', 'US', 'en_US'), (3, 103, 'key3', '{"test": "value3"}', 'US', 'en_US');

		CREATE TABLE IF NOT EXISTS spy_cms_page_storage (
			id_cms_page_storage SERIAL PRIMARY KEY,
			fk_cms_page INTEGER,
			key VARCHAR(255) NOT NULL,
			data TEXT NOT NULL,
			store VARCHAR(2),
			locale VARCHAR(6)
		);
		INSERT INTO spy_cms_page_storage (id_cms_page_storage, fk_cms_page, key, data, store, locale) VALUES 
		(1, 101, 'key1', '{"test": "value1"}', 'US', 'en_US'), (2, 102, 'key2', '{"test": "value2"}', 'US', 'en_US'), (3, 103, 'key3', '{"test": "value3"}', 'US', 'en_US');
	`)

	if err != nil {
		panic(err)
	}
}

func consumeMessagesFromQueue(rmq *rabbitmq.Rabbitmq, queueName string) []string {
	mCh, err := rmq.Consume(queueName, 100)
	if err != nil {
		log.Panic(err)
	}

	var messages []string
	for m := range mCh {
		messages = append(messages, string(m))
	}

	return messages
}
