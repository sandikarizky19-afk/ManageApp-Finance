package model

import "time"

type AccountTypes struct {
	Id          int     `json:"id"`
	TypeCode    string  `json:"type_code"`
	TypeName    string  `json:"type_name"`
	Description *string `json:"description"`
}

type AccountTypesModel struct {
	ID          int       `json:"id"`
	TypeCode    string    `json:"type_code"`
	TypeName    string    `json:"type_name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AccountTypesRequest struct {
	TypeCode    string  `json:"type_code"`
	TypeName    string  `json:"type_name"`
	Description *string `json:"description"`
}

type AccountTypesResponse struct {
	ID          int       `json:"id"`
	TypeCode    string    `json:"type_code"`
	TypeName    string    `json:"type_name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateAccountTypeRequest struct {
	ID          int     `json:"id"`
	TypeCode    string  `json:"type_code"`
	TypeName    string  `json:"type_name"`
	Description *string `json:"description"`
}

// type AccountTypeResponse struct {
// 	ID          int64      `json:"id"`
// 	TypeCode    string     `json:"type_code"`
// 	TypeName    string     `json:"type_name"`
// 	Description *string    `json:"description"`
// 	CreatedAt   time.Time  `json:"created_at"`
// 	UpdatedAt   time.Time  `json:"updated_at"`
// 	DeletedAt   *time.Time `json:"deleted_at"`
// }
