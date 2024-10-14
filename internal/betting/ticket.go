package betting

type Ticket struct {
	ClientId     int
	TicketNumber int
	Amount       float64
	Odds         *Odds
}

type Odds struct {
	Id    int64
	Value float64
}
