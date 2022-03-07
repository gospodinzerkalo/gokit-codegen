package domain

type User struct {
	ID   int	`json:"id"`
	Name string	`json:"name" col:"name"`
	Users string `json:"-" table_name:"users"'`
}
