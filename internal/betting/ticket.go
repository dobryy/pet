package betting

import "time"

type Ticket struct {
	ClientId      int
	TicketNumber  int
	Amount        float64
	Odds          *Odds
	Time          time.Time
	PerfStartTime time.Time
}

type Odds struct {
	Id    int64
	Value float64
}
