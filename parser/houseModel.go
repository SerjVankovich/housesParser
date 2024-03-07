package parser

type HouseModel struct {
	Street      string  `json:"street"`
	Levels      int     `json:"levels"`
	Area        float64 `json:"area"`
	HouseNumber string  `json:"houseNumber"`
}
