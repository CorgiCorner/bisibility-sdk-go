package bisibility

import "context"

// Pager iterates cursor-paginated resources on Go versions before iter.Seq2.
// Call Next, read Item, and check Err after iteration stops.
type Pager[T any] struct {
	ctx    context.Context
	cursor string
	done   bool
	err    error
	fetch  func(context.Context, string) ([]T, *string, error)
	index  int
	item   T
	items  []T
}

func newPager[T any](ctx context.Context, cursor string, fetch func(context.Context, string) ([]T, *string, error)) *Pager[T] {
	return &Pager[T]{ctx: ctx, cursor: cursor, fetch: fetch}
}

// Next advances to the next item, fetching another page when necessary.
func (p *Pager[T]) Next() bool {
	for {
		if p.err != nil || (p.done && p.index >= len(p.items)) {
			return false
		}
		if p.index < len(p.items) {
			p.item = p.items[p.index]
			p.index++
			return true
		}
		items, nextCursor, err := p.fetch(p.ctx, p.cursor)
		if err != nil {
			p.err = err
			return false
		}
		p.items = items
		p.index = 0
		if nextCursor == nil {
			p.done = true
		} else {
			p.cursor = *nextCursor
		}
	}
}

// Item returns the current item after Next returns true.
func (p *Pager[T]) Item() T { return p.item }

// Err returns the first page-fetch error, if any.
func (p *Pager[T]) Err() error { return p.err }
