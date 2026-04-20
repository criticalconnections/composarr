package models

type Stack struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Slug        string `db:"slug" json:"slug"`
	Description string `db:"description" json:"description"`
	ComposePath string `db:"compose_path" json:"composePath"`
	Status      string `db:"status" json:"status"`
	AutoUpdate  bool   `db:"auto_update" json:"autoUpdate"`
	CreatedAt   Time   `db:"created_at" json:"createdAt"`
	UpdatedAt   Time   `db:"updated_at" json:"updatedAt"`
}

type StackDependency struct {
	ID             string `db:"id" json:"id"`
	StackID        string `db:"stack_id" json:"stackId"`
	DependsOnID    string `db:"depends_on_id" json:"dependsOnId"`
	DependencyType string `db:"dependency_type" json:"dependencyType"`
	CreatedAt      Time   `db:"created_at" json:"createdAt"`
}
