package models

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	Conn *pgxpool.Pool
}

func (m *SnippetModel) Insert(title string, content string, expires int) (num int, err error) {
	stmt := `INSERT INTO snippets (title, content, created, expires) 
VALUES($1, $2, now(), (now() + INTERVAL '1 day' * $3)) returning id`

	row := m.Conn.QueryRow(context.Background(), stmt, title, content, expires)
	var id uint64
	err = row.Scan(&id)
	if err != nil {
		return
	}
	return int(id), err
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {

	s := &Snippet{}

	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > now() AND id = $1`

	err := m.Conn.QueryRow(context.Background(), stmt, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > now() ORDER BY created DESC LIMIT 10`

	rows, err := m.Conn.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}

	// Откладываем вызов rows.Close(), чтобы быть уверенным, что набор результатов из sql.Rows
	// правильно закроется перед вызовом метода Latest(). Этот оператор откладывания
	// должен выполнится *после* проверки на наличие ошибки в методе Query().
	// В противном случае, если Query() вернет ошибку, это приведет к панике
	// так как он попытается закрыть набор результатов у которого значение: nil.
	defer rows.Close()

	var snippets []*Snippet

	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
