package domain

type Article struct {
	Id      int64
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  Author `json:"author"`
}

type Author struct {
	Id   int64
	Name string
}
