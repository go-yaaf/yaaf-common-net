package model

type Role = int

type role struct {
	UNDEFINED  Role `Undefined [0]`
	SALES      Role `Sales [1]`
	OPERATIONS Role `Operations [2]`
	SUPPORT    Role `Support [4]`
	FINANCE    Role `Finance [8]`
	MANAGEMENT Role `Management [16]`
	IsValid    func(int) bool
	String     func(int) string
}

var Roles = &role{
	UNDEFINED:  0,  // Undefined [0]
	SALES:      1,  // Sales [1]
	OPERATIONS: 2,  // Operations [2]
	SUPPORT:    4,  // Support [4]
	FINANCE:    8,  // Finance [8]
	MANAGEMENT: 16, // Management [16]
	IsValid:    isValidRole,
	String:     stringRole,
}

func isValidRole(code int) bool {
	return code >= 0 && code < 32
}

func stringRole(code int) string {
	if isValidRole(code) {
		return "ROLE"
	} else {
		return "UNKNOWN"
	}
}
