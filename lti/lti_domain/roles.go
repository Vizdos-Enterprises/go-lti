package lti_domain

// Role represents a normalized internal role identifier (e.g. MEMBERSHIP_MEMBER).
type Role string

// https://www.imsglobal.org/spec/lti/v1p3#lis-vocabulary-for-system-roles
const (
	// --- System Roles ---
	SYSTEM_ADMINISTRATOR Role = "SYSTEM_ADMINISTRATOR"
	SYSTEM_NONE          Role = "SYSTEM_NONE"
	SYSTEM_ACCOUNT_ADMIN Role = "SYSTEM_ACCOUNT_ADMIN"
	SYSTEM_CREATOR       Role = "SYSTEM_CREATOR"
	SYSTEM_SYSADMIN      Role = "SYSTEM_SYSADMIN"
	SYSTEM_SYSSUPPORT    Role = "SYSTEM_SYSSUPPORT"
	SYSTEM_USER          Role = "SYSTEM_USER"
	SYSTEM_TESTUSER      Role = "SYSTEM_TESTUSER"

	// --- Institution Roles ---
	INSTITUTION_ADMINISTRATOR       Role = "INSTITUTION_ADMINISTRATOR"
	INSTITUTION_FACULTY             Role = "INSTITUTION_FACULTY"
	INSTITUTION_GUEST               Role = "INSTITUTION_GUEST"
	INSTITUTION_NONE                Role = "INSTITUTION_NONE"
	INSTITUTION_OTHER               Role = "INSTITUTION_OTHER"
	INSTITUTION_STAFF               Role = "INSTITUTION_STAFF"
	INSTITUTION_STUDENT             Role = "INSTITUTION_STUDENT"
	INSTITUTION_ALUMNI              Role = "INSTITUTION_ALUMNI"
	INSTITUTION_INSTRUCTOR          Role = "INSTITUTION_INSTRUCTOR"
	INSTITUTION_LEARNER             Role = "INSTITUTION_LEARNER"
	INSTITUTION_MEMBER              Role = "INSTITUTION_MEMBER"
	INSTITUTION_MENTOR              Role = "INSTITUTION_MENTOR"
	INSTITUTION_OBSERVER            Role = "INSTITUTION_OBSERVER"
	INSTITUTION_PROSPECTIVE_STUDENT Role = "INSTITUTION_PROSPECTIVE_STUDENT"

	// --- Membership / Context Roles ---
	MEMBERSHIP_ADMINISTRATOR Role = "MEMBERSHIP_ADMINISTRATOR"
	MEMBERSHIP_CONTENT_DEV   Role = "MEMBERSHIP_CONTENT_DEV"
	MEMBERSHIP_INSTRUCTOR    Role = "MEMBERSHIP_INSTRUCTOR"
	MEMBERSHIP_LEARNER       Role = "MEMBERSHIP_LEARNER"
	MEMBERSHIP_MENTOR        Role = "MEMBERSHIP_MENTOR"
	MEMBERSHIP_MANAGER       Role = "MEMBERSHIP_MANAGER"
	MEMBERSHIP_MEMBER        Role = "MEMBERSHIP_MEMBER"
	MEMBERSHIP_OFFICER       Role = "MEMBERSHIP_OFFICER"

	UNKNOWN Role = "UNKNOWN"
)

// uriToRole provides complete URI → internal constant mapping.
var uriToRole = map[string]Role{
	// --- System roles ---
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#Administrator": SYSTEM_ADMINISTRATOR,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#None":          SYSTEM_NONE,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#AccountAdmin":  SYSTEM_ACCOUNT_ADMIN,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#Creator":       SYSTEM_CREATOR,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#SysAdmin":      SYSTEM_SYSADMIN,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#SysSupport":    SYSTEM_SYSSUPPORT,
	"http://purl.imsglobal.org/vocab/lis/v2/system/person#User":          SYSTEM_USER,
	"http://purl.imsglobal.org/vocab/lti/system/person#TestUser":         SYSTEM_TESTUSER,

	// --- Core Institution roles ---
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Administrator": INSTITUTION_ADMINISTRATOR,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Faculty":       INSTITUTION_FACULTY,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Guest":         INSTITUTION_GUEST,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#None":          INSTITUTION_NONE,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Other":         INSTITUTION_OTHER,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Staff":         INSTITUTION_STAFF,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Student":       INSTITUTION_STUDENT,

	// --- Non-Core Institution roles ---
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Alumni":             INSTITUTION_ALUMNI,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Instructor":         INSTITUTION_INSTRUCTOR,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Learner":            INSTITUTION_LEARNER,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Member":             INSTITUTION_MEMBER,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Mentor":             INSTITUTION_MENTOR,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Observer":           INSTITUTION_OBSERVER,
	"http://purl.imsglobal.org/vocab/lis/v2/institution/person#ProspectiveStudent": INSTITUTION_PROSPECTIVE_STUDENT,

	// --- Core Membership / Context roles ---
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Administrator":    MEMBERSHIP_ADMINISTRATOR,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#ContentDeveloper": MEMBERSHIP_CONTENT_DEV,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor":       MEMBERSHIP_INSTRUCTOR,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner":          MEMBERSHIP_LEARNER,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Mentor":           MEMBERSHIP_MENTOR,

	// --- Non-Core Membership / Context roles ---
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Manager": MEMBERSHIP_MANAGER,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Member":  MEMBERSHIP_MEMBER,
	"http://purl.imsglobal.org/vocab/lis/v2/membership#Officer": MEMBERSHIP_OFFICER,

	"UNKNOWN": UNKNOWN,
}

// ParseRoleURI converts a URI string into an internal Role constant.
// Returns empty Role("") if unknown or unsupported.
func ParseRoleURI(uri string) Role {
	if r, ok := uriToRole[uri]; ok {
		return r
	}
	return UNKNOWN
}
