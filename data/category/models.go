package category

type CategoryImageStorageEntity struct {
	Key    string
	Data   string
	Locale string
}

type CategoryNodeStorageEntity struct {
	Key    string
	Data   string
	Store  string
	Locale string
}

type CategoryTreeStorageEntity struct {
	Key    string
	Data   string
	Store  string
	Locale string
}

func (p *CategoryImageStorageEntity) GetKey() string {
	return p.Key
}

func (p *CategoryImageStorageEntity) GetData() string {
	return p.Data
}

func (p *CategoryImageStorageEntity) GetStore() string {
	return ""
}

func (p *CategoryImageStorageEntity) GetLocale() string {
	return p.Locale
}

func (p *CategoryImageStorageEntity) IsNil() bool {
	return p == nil
}

func (p *CategoryImageStorageEntity) GenerateMappingKey(source, sourceId string) string {
	return ""
}

func (p *CategoryNodeStorageEntity) GetKey() string {
	return p.Key
}

func (p *CategoryNodeStorageEntity) GetData() string {
	return p.Data
}

func (p *CategoryNodeStorageEntity) GetStore() string {
	return p.Store
}

func (p *CategoryNodeStorageEntity) GetLocale() string {
	return p.Locale
}

func (p *CategoryNodeStorageEntity) IsNil() bool {
	return p == nil
}

func (p *CategoryNodeStorageEntity) GenerateMappingKey(source, sourceId string) string {
	return ""
}

func (p *CategoryTreeStorageEntity) GetKey() string {
	return p.Key
}

func (p *CategoryTreeStorageEntity) GetData() string {
	return p.Data
}

func (p *CategoryTreeStorageEntity) GetStore() string {
	return p.Store
}

func (p *CategoryTreeStorageEntity) GetLocale() string {
	return p.Locale
}

func (p *CategoryTreeStorageEntity) IsNil() bool {
	return p == nil
}

func (p *CategoryTreeStorageEntity) GenerateMappingKey(source, sourceId string) string {
	return ""
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
