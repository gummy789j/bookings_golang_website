package forms

type errors map[string][]string

func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e errors) Get(field string) string {

	if len(e[field]) == 0 {
		return ""
	}

	return e[field][0]
}
