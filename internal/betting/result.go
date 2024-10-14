package betting

type Result struct {
	Ticket *Ticket
	Result func() error
}
