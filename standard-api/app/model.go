package app

type Model struct {
	ID   int64  `json:"id"`
	Name string `json:"name" encrypt:"true"`
}
