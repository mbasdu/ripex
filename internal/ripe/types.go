package ripe

type Organisation struct {
	ID      string
	Name    string
	Country string
}

type Inetnum struct {
	Key     string
	OrgID   string
	Country string
	NetName string
	Status  string
}

type AutNum struct {
	ASN    string
	AsName string
	OrgID  string
}

type Route struct {
	Prefix string
	Origin string
}

type SnapshotData struct {
	Organisations []Organisation
	Inetnums      []Inetnum
	AutNums       []AutNum
	Routes        []Route
}

type ParseCounts struct {
	Organisations int `json:"organisations"`
	Inetnums      int `json:"inetnums"`
	AutNums       int `json:"aut_nums"`
	Routes        int `json:"routes"`
}
