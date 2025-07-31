package validation

type AccountInput struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=user admin"`
}

type ProductInput struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Description string  `json:"description" validate:"required,min=5"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
}

type OrderedProductInput struct {
	ID       string `json:"id" validate:"required,alphanum,min=10,max=40"`
	Quantity int    `json:"quantity" validate:"required,gt=0"`
}

type OrderInput struct {
	AccountID string                `json:"accountId" validate:"required,alphanum,min=10,max=40"`
	Products  []*OrderedProductInput `json:"products" validate:"required,min=1,dive"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RefreshTokenInput struct {
	UserID string `json:"userId" validate:"required,alphanum,min=10,max=40"`
}

type ResetPasswordInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogoutInput struct {
	UserID string `json:"userId" validate:"required,alphanum,min=10,max=40"`
}

type Pagination struct {
	Skip int `json:"skip" validate:"omitempty,gte=0"`
	Take int `json:"take" validate:"omitempty,gte=1,lte=100"`
}

type ProductIDInput struct {
	ProductID string `json:"productId" validate:"required,alphanum,min=10,max=40"`
}

type RestockProductInput struct {
	ProductID string `json:"productId" validate:"required,alphanum,min=10,max=40"`
	NewStock  int    `json:"newStock" validate:"required,gt=0"`
}

type UserIDInput struct {
	UserID string `json:"userId" validate:"required,alphanum,min=10,max=40"`
}
