package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	qb "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/domain/customer"
)

// CustomerRepository gives us our datastore interaction for a domain.Customer.
type CustomerRepository struct {
	db *app.DB
}

// NewCustomerRepository creates a new CustomerRepository with an encapsulated database connection.
func NewCustomerRepository(db *app.DB) *CustomerRepository {
	return &CustomerRepository{db}
}

// Add a new domain.Customer to the repository.
func (cr *CustomerRepository) Add(ctx context.Context, c *customer.Customer) error {
	query := cr.db.QB().
		Insert("customers").
		Columns("uuid", "username", "password", "created_at", "updated_at").
		Values(c.ID(), c.Username(), c.PasswordHash(), time.Now(), time.Now())

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("unable to convert customers create query to SQL: %w", err)
	}

	_, err = cr.db.Conn().Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to create customers: %w", err)
	}

	return nil
}

// FindByUsername searches the repository for a Customer with the given username.
// If none is found then nil is returned.
func (cr *CustomerRepository) FindByUsername(ctx context.Context, username string) (*customer.Customer, error) {
	query := cr.db.QB().
		Select("c.uuid as id", "c.username", "c.password").
		From("customers as c").
		Where(qb.Eq{"username": username})

	var cust struct {
		ID       string
		Username string
		Password string
	}

	if err := cr.db.Get(ctx, &cust, query); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("unable to fetch customers: %w", err)
	}

	return customer.Restore(
		uuid.MustParse(cust.ID),
		cust.Username,
		cust.Password,
	), nil
}
