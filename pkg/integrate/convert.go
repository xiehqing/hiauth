package integrate

import (
	"github.com/xiehqing/hiauth/internal/authentication"
	"github.com/xiehqing/hiauth/internal/db/entity"
	"github.com/xiehqing/infra/pkg/ormx"
)

func convertLoginResponse(response *authentication.LoginResponse) *LoginResponse {
	if response == nil {
		return nil
	}
	return &LoginResponse{
		AccessToken: response.AccessToken,
		TokenType:   response.TokenType,
		User:        convertUserPtr(response.User),
		Roles:       append([]string(nil), response.Roles...),
		Permissions: append([]string(nil), response.Permissions...),
		Menus:       convertMenuTrees(response.Menus),
		Department:  convertDepartment(response.Department),
	}
}

func convertUserPtr(user *entity.User) *User {
	if user == nil {
		return nil
	}
	result := convertUser(*user)
	return &result
}

func convertUsers(users []entity.User) []User {
	result := make([]User, 0, len(users))
	for _, user := range users {
		result = append(result, convertUser(user))
	}
	return result
}

func convertUser(user entity.User) User {
	return User{
		BaseModel:    convertBaseModel(user.BaseModel),
		Status:       user.Status,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Phone:        user.Phone,
		Email:        user.Email,
		Roles:        convertRoles(user.Roles),
		DepartmentID: user.DepartmentID,
		Department:   convertDepartment(user.Department),
	}
}

func convertRoles(roles []entity.Role) []Role {
	result := make([]Role, 0, len(roles))
	for _, role := range roles {
		result = append(result, Role{
			BaseModel:   convertBaseModel(role.BaseModel),
			DisplayName: role.DisplayName,
			Name:        role.Name,
			Description: role.Description,
			BuiltIn:     role.BuiltIn,
			Menus:       convertMenus(role.Menus),
		})
	}
	return result
}

func convertDepartment(department entity.Department) Department {
	return Department{
		BaseModel: convertBaseModel(department.BaseModel),
		ParentID:  department.ParentID,
		Name:      department.Name,
		Sort:      department.Sort,
	}
}

func convertMenus(menus []entity.Menu) []Menu {
	result := make([]Menu, 0, len(menus))
	for _, menu := range menus {
		result = append(result, convertMenu(menu))
	}
	return result
}

func convertMenu(menu entity.Menu) Menu {
	return Menu{
		BaseModel: convertBaseModel(menu.BaseModel),
		Type:      menu.Type,
		ParentID:  menu.ParentID,
		Name:      menu.Name,
		Route:     menu.Route,
		Sort:      menu.Sort,
		Icon:      menu.Icon,
		Show:      menu.Show,
	}
}

func convertMenuTrees(trees []authentication.MenuTree) []MenuTree {
	result := make([]MenuTree, 0, len(trees))
	for _, tree := range trees {
		result = append(result, MenuTree{
			Menu:     convertMenu(tree.Menu),
			Children: convertMenuTrees(tree.Children),
		})
	}
	return result
}

func convertBaseModel(model ormx.BaseModel) BaseModel {
	return BaseModel{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedAt: model.UpdatedAt,
		UpdatedBy: model.UpdatedBy,
	}
}
