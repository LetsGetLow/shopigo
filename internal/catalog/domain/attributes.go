package domain

type AttributeMap struct {
	attributes map[string]string
}

func NewAttributeMap() *AttributeMap {
	return &AttributeMap{
		attributes: make(map[string]string),
	}
}

func (s *AttributeMap) Set(key string, value string) {
	s.attributes[key] = value
}

func (s *AttributeMap) Remove(key string) {
	delete(s.attributes, key)
}

func (s *AttributeMap) Get(key string) (value string, exists bool) {
	value, exists = s.attributes[key]
	return value, exists
}

func (s *AttributeMap) Attributes() map[string]string {
	clone := make(map[string]string, len(s.attributes))
	for key, value := range s.attributes {
		clone[key] = value
	}
	return clone
}
