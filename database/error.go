package database

// OperationError when cannot perform a given operation on database (SET, GET or DELETE)
type OperationError struct {
	operation string
}

// OperationError when cannot perform a given operation on database (SET, GET or DELETE)
func (err *OperationError) Error() string {
	return "Could not perform the " + err.operation + " operation."
}

// DownError when its not a redis.Nil response, in this case the database is down
type DownError struct{}

// DownError when its not a redis.Nil response, in this case the database is down
func (dbe *DownError) Error() string {
	return "Database is down"
}

// CreateDatabaseError when cannot perform set on database
type CreateDatabaseError struct{}

// CreateDatabaseError when cannot perform set on database
func (err *CreateDatabaseError) Error() string {
	return "Could not create Databse"
}

// NotImplementedDatabaseError when user tries to create a not implemented database
type NotImplementedDatabaseError struct {
	database string
}

// NotImplementedDatabaseError when user tries to create a not implemented database
func (err *NotImplementedDatabaseError) Error() string {
	return err.database + " not implemented"
}
