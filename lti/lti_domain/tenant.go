package lti_domain

import "fmt"

type TenantID any

func TenantIDString(id TenantID) string {
	switch v := id.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
