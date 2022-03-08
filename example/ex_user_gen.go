// Code generated by generator, DO NOT EDIT.
package example

import domain "github.com/gospodinzerkalo/gokit-codegen/domain"

func (c store) CreateUser(d domain.User) (*domain.User, error) {
	if _, err := c.db.Exec("INSERT INTO users (name) VALUES ($1)", d.Name); err == nil {
		return nil, err
	}
	return &d, nil
}
func (c store) GetUserList(d domain.User) (*[]domain.User, error) {
	var res []domain.User
	rows, err := c.db.Exec("SELECT id,name FROM users")
	if err != nil {
		return nil, err
	}
	for rows.Next {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return &res, nil
}
func (c store) DeleteUser(id int64) error {
	_, err := c.db.Exec("DELETE FROM users WHERE id=$1", id)
	return err
}
