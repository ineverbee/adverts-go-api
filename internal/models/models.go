package models

//Advert defines a structure for an item in advert catalog
type Advert struct {
	ID             int      `json:"-"`
	Title          string   `json:"title,omitempty"`
	Ad_description string   `json:"ad_description,omitempty"`
	Price          int      `json:"price,omitempty"`
	Photos         []string `json:"photos,omitempty"`
}
