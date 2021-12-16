package entity

type TypeOne struct {
	Name                 string `json:"name"`
	SpecificTypeOneValue int    `json:"typeOneValue"`
}

type TypeTwo struct {
	Name                 string `json:"name"`
	SpecificTypeTwoValue int    `json:"typeTwoValue"`
}
