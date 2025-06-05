package sync

import (
	"strings"
)

type SyncEntity struct {
	Key    string
	Data   string
	Store  string
	Locale string
}

func (p *SyncEntity) GetKey() string {
	return p.Key
}

func (p *SyncEntity) GetData() string {
	return p.Data
}

func (p *SyncEntity) GetStore() string {
	return p.Store
}

func (p *SyncEntity) GetLocale() string {
	return p.Locale
}

func (p *SyncEntity) IsNil() bool {
	return p == nil
}

func (p *SyncEntity) GenerateMappingKey(resourceName, source, sourceId string) string {
	keyParts := []string{resourceName}

	if p.Store != "" {
		keyParts = append(keyParts, strings.ToLower(p.Store))
	}
	if p.Locale != "" {
		keyParts = append(keyParts, strings.ToLower(p.Locale))
	}

	keyParts = append(keyParts, strings.ToLower(source), p.escapeKey(sourceId))

	return strings.Join(keyParts, ":")
}

func (p *SyncEntity) escapeKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ToLower(key)

	for _, replaceChar := range []string{"\"", "'", " ", "\000", "\n", "\r"} {
		key = strings.ReplaceAll(key, replaceChar, "-")
	}

	return key
}
