package content

import (
	"fmt"
	"strings"
)

type ContentStorageEntity struct {
	Key    string
	Data   string
	Locale string
}

func (p *ContentStorageEntity) GetKey() string {
	return p.Key
}

func (p *ContentStorageEntity) GetData() string {
	return p.Data
}

func (p *ContentStorageEntity) GetStore() string {
	return ""
}

func (p *ContentStorageEntity) GetLocale() string {
	return p.Locale
}

func (p *ContentStorageEntity) IsNil() bool {
	return p == nil
}

func (p *ContentStorageEntity) GenerateMappingKey(source, sourceId string) string {
	reference := fmt.Sprintf("%s:%s", source, sourceId)
	keySuffix := fmt.Sprintf("%s:%s", strings.ToLower(p.Locale), reference)

	return fmt.Sprintf("%s:%s", resourceName, keySuffix)
}

type Mapping struct {
	Source      string
	Destination string
}

func (m Mapping) GetSource() string {
	return m.Source
}

func (m Mapping) GetDestination() string {
	return m.Destination
}
