package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	id          uuid.UUID
	name        string
	slug        string
	description string
	createdAt   time.Time
}

func NewCategory(id uuid.UUID, name, slug, description string, createdAt time.Time) *Category {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return &Category{
		id:          id,
		name:        name,
		slug:        slug,
		description: description,
		createdAt:   createdAt,
	}
}

func (c *Category) ID() uuid.UUID {
	return c.id
}

func (c *Category) Name() string {
	return c.name
}

func (c *Category) Slug() string {
	return c.slug
}

func (c *Category) Description() string {
	return c.description
}

func (c *Category) CreatedAt() time.Time {
	return c.createdAt
}
