package rbac

// Role map role to the constant
type Role string

// user role constants
const (
	RoleAdmin            Role = "ADMIN"
	RoleProductManager   Role = "PRODUCT_MANAGER"
	RoleCashiers         Role = "CASHIER"
	RoleFinancialAuditor Role = "FINANCIAL_AUDITOR"
	RoleInternalService  Role = "INTERNAL_SERVICE"
)

// Resource is a resource
type Resource string

// Resource constants
const (
	ResourceUser        Resource = "user"
	ResourceProduct     Resource = "product"
	ResourceTransaction Resource = "transaction"
)

// Action is an action
type Action string

// Action constants
const (
	ActionCreateAny  Action = "create_any"
	ActionViewAny    Action = "view_any"
	ActionEditAny    Action = "edit_any"
	ActionDeleteAny  Action = "delete_any"
	ActionChangeRole Action = "change_role"
)

// TraversePermission traverse the built in permission
func TraversePermission(cb func(role Role, rsc Resource, act Action)) {
	for rra, roles := range _permissions {
		for _, role := range roles {
			cb(role, rra.Resource, rra.Action)
		}
	}
}

type resourceAction struct {
	Resource Resource
	Action   Action
}

// describe the permission for the Role here
var _permissions = map[resourceAction][]Role{
	{ResourceUser, ActionCreateAny}: {RoleAdmin},
	{ResourceUser, ActionViewAny}:   {RoleAdmin},
	{ResourceUser, ActionEditAny}:   {RoleAdmin},
	{ResourceUser, ActionDeleteAny}: {RoleAdmin},

	{ResourceProduct, ActionViewAny}:    {RoleAdmin, RoleProductManager, RoleFinancialAuditor},
	{ResourceProduct, ActionEditAny}:    {RoleAdmin, RoleProductManager},
	{ResourceProduct, ActionChangeRole}: {RoleAdmin, RoleProductManager},
	{ResourceProduct, ActionDeleteAny}:  {RoleAdmin, RoleProductManager},

	{ResourceTransaction, ActionViewAny}:    {RoleAdmin, RoleCashiers, RoleFinancialAuditor},
	{ResourceTransaction, ActionEditAny}:    {RoleAdmin, RoleCashiers},
	{ResourceTransaction, ActionChangeRole}: {RoleAdmin, RoleCashiers},
	{ResourceTransaction, ActionDeleteAny}:  {RoleAdmin, RoleCashiers},
}
