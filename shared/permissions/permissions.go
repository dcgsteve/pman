package permissions

import (
	"fmt"
	"strings"
)

type GroupAccess struct {
	GroupName  string
	Permission string // "ro" or "rw"
}

func ParseGroups(groupsStr string) ([]GroupAccess, error) {
	if groupsStr == "" {
		return []GroupAccess{}, nil
	}

	var groups []GroupAccess
	parts := strings.Split(groupsStr, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		groupParts := strings.Split(part, ":")
		if len(groupParts) != 2 {
			return nil, fmt.Errorf("invalid group format: %s (expected format: groupname:permission)", part)
		}

		groupName := strings.TrimSpace(groupParts[0])
		permission := strings.TrimSpace(groupParts[1])

		if permission != "ro" && permission != "rw" {
			return nil, fmt.Errorf("invalid permission '%s' for group '%s' (must be 'ro' or 'rw')", permission, groupName)
		}

		groups = append(groups, GroupAccess{
			GroupName:  groupName,
			Permission: permission,
		})
	}

	return groups, nil
}

func FormatGroups(groups []GroupAccess) string {
	var parts []string
	for _, group := range groups {
		parts = append(parts, fmt.Sprintf("%s:%s", group.GroupName, group.Permission))
	}
	return strings.Join(parts, ",")
}

func HasGroupAccess(userGroups string, requiredGroup string, requireWrite bool) bool {
	groups, err := ParseGroups(userGroups)
	if err != nil {
		return false
	}

	for _, group := range groups {
		if group.GroupName == requiredGroup {
			if requireWrite {
				return group.Permission == "rw" 
			}
			return true // ro or rw both allow read
		}
	}

	return false
}

func GetUserGroups(userGroups string) []string {
	groups, err := ParseGroups(userGroups)
	if err != nil {
		return []string{}
	}

	var groupNames []string
	for _, group := range groups {
		groupNames = append(groupNames, group.GroupName)
	}

	return groupNames
}