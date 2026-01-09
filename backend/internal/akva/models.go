package akva

import "time"

// DetalleAlimentacion represents a row from TB_DetalleAlimentacion
type DetalleAlimentacion struct {
	ID            string    `db:"ID"`
	Name          string    `db:"Name"`
	UnitName      string    `db:"UnitName"`
	FechaHora     time.Time `db:"FechaHora"`
	Dia           time.Time `db:"Dia"`
	Inicio        string    `db:"inicio"`
	Fin           string    `db:"Fin"`
	Dif           int       `db:"dif"`
	AmountGrams   float64   `db:"AmountGrams"`
	PelletFishMin float64   `db:"pelletfishmin"`
	FishCount     int       `db:"FisCount"`
	PesoProm      float64   `db:"PesoProm"`
	Biomasa       float64   `db:"Biomasa"`
	PelletPK      float64   `db:"pelletpK"`
	FeedName      string    `db:"Feedname"`
	SiloName      string    `db:"SiloName"`
	DoserName     string    `db:"DoserName"`
	GramsPerSec   float64   `db:"gramspersec"`
	KgTonMin      float64   `db:"kgtonmin"`
	Marca         int       `db:"Marca"`
}
