package customers

import (
	"context"
	"errors"
	"log"
	"time"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4"
)

// ErrNotFound возвращается, когда покупатель не найден.
var ErrNotFound = errors.New("item not found")

// ErrInternal возвращается, когда произошла внутренняя ошибка.
var ErrInternal = errors.New("internal error")

// Service описывает сервис работы с покупателями
type Service struct {
	pool *pgxpool.Pool
}

// NewService создаёт сервис.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Customer представляет информацию о покупателе
type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

// All возвращает список всех менеджеров
func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers 
	`)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Print("No rows")
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// AllActive возвращает список всех клиентов только с активными статусами
func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	items := make([]*Customer, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE active
	`)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Print("No rows")
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}

// ByID возвращает покупателя по идентификатору
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

// Save сохраняет/обновляет данные клиента
func (s *Service) Save(ctx context.Context, item *Customer) (*Customer, error) {
	items := &Customer{}
	if item.ID == 0 {
		err := s.pool.QueryRow(ctx, `
		INSERT INTO customers (name, phone) VALUES ($1, $2) ON CONFLICT (phone) DO UPDATE SET name = excluded.name RETURNING id, name, phone, active, created; 
	`, item.Name, item.Phone).Scan(&items.ID, &items.Name, &items.Phone, &items.Active, &items.Created)
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
		return items, nil
	}

	_, err := s.ByID(ctx, item.ID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
	}

	err = s.pool.QueryRow(ctx, `
		UPDATE customers SET name = $2, phone = $3 WHERE id = $1 RETURNING id, name, phone, active, created;
	`, item.ID, item.Name, item.Phone).Scan(&items.ID, &items.Name, &items.Phone, &items.Active, &items.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return items, nil
}

// RemoveByID удаляет клиента из бд, находя по id
func (s *Service) RemoveByID(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		DELETE FROM customers WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return cust, nil
}

// BlockUser блочит плохих клиентов)))
func (s *Service) BlockUser(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE customers SET active = false WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return cust, nil
}

// UnblockUser вытаскивает клиента из ЧС
func (s *Service) UnblockUser(ctx context.Context, id int64) (*Customer, error) {
	cust, err := s.ByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			log.Print(err)
			return nil, ErrNotFound
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE customers SET active = true WHERE id = $1;
	`, id)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return cust, nil
}
