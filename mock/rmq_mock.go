package mock

type MockDelivery struct {
	Acks        int
	Rejected    int
	Pushed      int
	PayloadImpl func() string
}

func (r *MockDelivery) Payload() string {
	return r.PayloadImpl()
}

func (r *MockDelivery) Ack() error {
	r.Acks++
	return nil
}
func (r *MockDelivery) Reject() error {
	r.Rejected++
	return nil
}
func (r *MockDelivery) Push() error {
	r.Pushed++
	return nil
}
