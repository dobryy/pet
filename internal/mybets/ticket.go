package mybets

type Ticket struct {
	ClientId     int64
	TicketNumber int64
	Amount       float64
	Odds         *Odds
}

type Odds struct {
	Id    int64
	Value float64
}
