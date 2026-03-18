package service

const (
	RoleAdmin   = "ADMIN"
	RoleManager = "MANAGER"
	RoleMember  = "MEMBER"
	RoleViewer  = "VIEWER"

	PermAll               = "*"
	PermGateOpen          = "gate:open"
	PermGateClose         = "gate:close"
	PermGateViewStatus    = "gate:view_status"
	PermGateConfigure     = "gate:configure"
	PermGateManageMembers = "gate:manage_members"

	ActionTypeOpen   = "OPEN"
	ActionTypeClose  = "CLOSE"
	ActionTypeStatus = "STATUS"

	TransportMQTT = "MQTT"
	TransportHTTP = "HTTP"
	TransportNone = "NONE"
)
