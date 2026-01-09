package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type User struct {
	Email string `json:"email"`
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(map[string]int)
	scanner := bufio.NewScanner(r)
	var user User
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := json.Unmarshal(line, &user); err != nil {
			continue
		}
		if user.Email == "" {
			continue
		}
		atIdx := strings.IndexByte(user.Email, '@')
		if atIdx == -1 {
			continue
		}
		emailDomain := strings.ToLower(user.Email[atIdx+1:])
		if strings.HasSuffix(emailDomain, "."+domain) {
			result[emailDomain]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return DomainStat(result), nil
}
