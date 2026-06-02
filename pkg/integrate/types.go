package integrate

import "time"

type BaseModel struct {
	ID        int64      `json:"id"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	CreatedBy string     `json:"createdBy,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy string     `json:"updatedBy,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Device   string `json:"device"`
}

type LoginResponse struct {
	AccessToken string     `json:"accessToken"`
	TokenType   string     `json:"tokenType"`
	User        *User      `json:"user"`
	Roles       []string   `json:"roles"`
	Permissions []string   `json:"permissions"`
	Menus       []MenuTree `json:"menus"`
	Department  Department `json:"department"`
}

type User struct {
	BaseModel
	Status       int        `json:"status"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	Phone        string     `json:"phone"`
	Email        string     `json:"email"`
	Roles        []Role     `json:"roles,omitempty"`
	DepartmentID int64      `json:"departmentId"`
	Department   Department `json:"department,omitempty"`
}

type Role struct {
	BaseModel
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BuiltIn     int    `json:"builtIn"`
	Menus       []Menu `json:"menus,omitempty"`
}

type Department struct {
	BaseModel
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Sort     int    `json:"sort"`
}

type Menu struct {
	BaseModel
	Type     int    `json:"type"`
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Route    string `json:"route"`
	Sort     int    `json:"sort"`
	Icon     string `json:"icon"`
	Show     int    `json:"show"`
}

type MenuTree struct {
	Menu
	Children []MenuTree `json:"children"`
}
