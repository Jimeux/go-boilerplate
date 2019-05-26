package app

type Model struct {
	Encryptable

	ID   int64  `json:"id"`
	Name string `json:"name" encrypt:"true"`
}
