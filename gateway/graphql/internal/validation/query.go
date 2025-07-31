package validation

type AccountsQueryInput struct {
	ID         string      `json:"id" validate:"omitempty,alphanum,min=10,max=40"`
	Name       string      `json:"name" validate:"omitempty,min=2"`
	Pagination *Pagination `json:"pagination" validate:"omitempty,dive"`
}

type ProductsQueryInput struct {
	Query      string      `json:"query" validate:"omitempty,min=2"`
	ID         string      `json:"id" validate:"omitempty,alphanum,min=10,max=40"`
	Pagination *Pagination `json:"pagination" validate:"omitempty,dive"`
}


type CurrentUsersQueryInput struct {
	Role       string      `json:"role" validate:"omitempty,oneof=user admin"`
	Pagination *Pagination `json:"pagination" validate:"omitempty,dive"`
}

type SuggestProductsQueryInput struct {
	Query  string `json:"query" validate:"required,min=2"`
	Size   int    `json:"size" validate:"omitempty,gte=1,lte=20"`
	// UseAI  bool   `json:"useAI"` // no validation needed for boolean
}



