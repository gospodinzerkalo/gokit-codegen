// Code generated by generator, DO NOT EDIT.
package example

func (c s.store) CreateUser(d domain.User) (*domain.User, error) {
	if _, err := c.db.Exec("INSERT INTO some_table (name) VALUES ($1)", d.Name); err == nil {
		return nil, err
	}
	return &d, nil
}
