package product

import (
	"fmt"
	"strings"
)

type ProductAbstractStorageEntity struct {
	Key    string
	Data   string
	Store  string
	Locale string
}

func (p *ProductAbstractStorageEntity) GetKey() string {
	return p.Key
}

func (p *ProductAbstractStorageEntity) GetData() string {
	return p.Data
}

func (p *ProductAbstractStorageEntity) GetStore() string {
	return p.Store
}

func (p *ProductAbstractStorageEntity) GetLocale() string {
	return p.Locale
}

func (p *ProductAbstractStorageEntity) IsNil() bool {
	return p == nil
}

func (p *ProductAbstractStorageEntity) GenerateMappingKey(source, sourceId string) string {
	reference := fmt.Sprintf("%s:%s", source, sourceId)
	keySuffix := fmt.Sprintf("%s:%s:%s", strings.ToLower(p.Store), strings.ToLower(p.Locale), reference)

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
