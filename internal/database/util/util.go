package util

import "github.com/google/uuid"

func CleanStrings(s []string) []string {
	res := make([]string, 0, len(s))
	for _, str := range s {
		if str != "" {
			res = append(res, str)
		}
	}
	return res
}

func CleanUUIDs(uuids uuid.UUIDs) uuid.UUIDs {
	res := make(uuid.UUIDs, 0, len(uuids))
	for _, id := range uuids {
		if id != uuid.Nil {
			res = append(res, id)
		}
	}
	return res
}
