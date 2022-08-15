package util

type ListMeta struct {
	Count  int `json:"count,omitempty"`
	Total  int `json:"total,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type ListResponse[T interface{}] struct {
	Data []T      `json:"data"`
	Meta ListMeta `json:"meta"`
}
