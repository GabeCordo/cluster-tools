package cli

type Template struct {
	Title string
}

type Help struct {
	output map[string]Template
}
