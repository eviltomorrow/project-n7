package helper

var App = appHelper{
	Name: "unknown",
}

type appHelper struct {
	Name string `json:"app-name"`
}
