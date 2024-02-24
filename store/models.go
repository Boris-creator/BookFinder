package store

type BookModel struct {
	Title       string `db:"title"`
	Description string `db:"description"`
	Isbn        string `db:"isbn"`
	Source      int    `db:"source"`
}
