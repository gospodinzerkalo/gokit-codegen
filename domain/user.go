package domain

type User struct {
	ID   int	`json:"id" sel:"id"`
	Name string	`json:"name" col:"name" sel:"name"`
	Users string `json:"-" table_name:"users"'`
}
